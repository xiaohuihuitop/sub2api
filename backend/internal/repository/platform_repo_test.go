package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestPlatformRepositoryCreateWritesPlatformAndRulesAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	createdAt := time.Date(2026, 8, 4, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	mock.ExpectBegin()
	mock.ExpectQuery("INSERT INTO platforms").
		WithArgs("gpt", "GPT", service.PlatformOpenAI, service.StatusActive, []byte(`["chat_completions","responses"]`), nil).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(7), createdAt, updatedAt))
	mock.ExpectQuery("INSERT INTO platform_model_rules").
		WithArgs(int64(7), "gpt-4o", "gpt-4o-2024-08-06", service.StatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(11), createdAt, updatedAt))
	mock.ExpectCommit()

	repo := newPlatformRepository(db)
	platform := &service.Platform{
		Code:                 "gpt",
		Name:                 "GPT",
		AccountPlatform:      service.PlatformOpenAI,
		Status:               service.StatusActive,
		EndpointCapabilities: []string{"chat_completions", "responses"},
		ModelRules: []service.PlatformModelRule{{
			ModelPattern:         "gpt-4o",
			UpstreamModel:        "gpt-4o-2024-08-06",
			EndpointCapabilities: []string{"chat_completions", "responses"},
			Enabled:              true,
		}},
	}

	err = repo.Create(context.Background(), platform)

	require.NoError(t, err)
	require.Equal(t, int64(7), platform.ID)
	require.Equal(t, createdAt, platform.CreatedAt)
	require.Equal(t, int64(11), platform.ModelRules[0].ID)
	require.Equal(t, int64(7), platform.ModelRules[0].PlatformID)
	require.Equal(t, "gpt", platform.ModelRules[0].PlatformCode)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlatformRepositoryListModelRulesQueriesOnlyActivePlatformsAndRules(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	mock.ExpectQuery("SELECT r.id, r.platform_id, p.code.*p.endpoint_capabilities").
		WithArgs(service.StatusActive, service.StatusActive).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "platform_id", "code", "account_platform", "legacy_group_id", "model_pattern", "upstream_model", "endpoint_capabilities", "created_at", "updated_at",
		}).AddRow(
			int64(11), int64(7), "gpt", service.PlatformOpenAI, int64(9), "gpt-4o", "gpt-4o-2024-08-06", []byte(`["chat_completions"]`), time.Now(), time.Now(),
		))

	rules, err := newPlatformRepository(db).ListModelRules(context.Background())

	require.NoError(t, err)
	require.Len(t, rules, 1)
	require.True(t, rules[0].Enabled)
	require.Equal(t, service.PlatformOpenAI, rules[0].AccountPlatform)
	require.Equal(t, int64(9), *rules[0].LegacyGroupID)
	require.Equal(t, []string{"chat_completions"}, rules[0].EndpointCapabilities)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlatformRepositoryUpdateReplacesRulesAtomically(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	updatedAt := time.Date(2026, 8, 4, 11, 0, 0, 0, time.UTC)
	mock.ExpectBegin()
	mock.ExpectExec("UPDATE platforms").
		WithArgs("gpt", "GPT", service.PlatformOpenAI, service.StatusActive, []byte(`["chat_completions"]`), nil, int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM platform_model_rules").
		WithArgs(int64(7)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("INSERT INTO platform_model_rules").
		WithArgs(int64(7), "gpt-5", "", service.StatusActive).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at", "updated_at"}).
			AddRow(int64(12), updatedAt, updatedAt))
	mock.ExpectCommit()

	platform := &service.Platform{
		ID:                   7,
		Code:                 "gpt",
		Name:                 "GPT",
		AccountPlatform:      service.PlatformOpenAI,
		Status:               service.StatusActive,
		EndpointCapabilities: []string{"chat_completions"},
		ModelRules: []service.PlatformModelRule{{
			ModelPattern: "gpt-5",
			Enabled:      true,
		}},
	}

	err = newPlatformRepository(db).Update(context.Background(), platform)

	require.NoError(t, err)
	require.Equal(t, int64(12), platform.ModelRules[0].ID)
	require.Equal(t, int64(7), platform.ModelRules[0].PlatformID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlatformRepositoryGetByIDLoadsDisabledRulesForEditing(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	createdAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT id, code, name, account_platform, status, endpoint_capabilities, legacy_group_id, created_at, updated_at FROM platforms").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "account_platform", "status", "endpoint_capabilities", "legacy_group_id", "created_at", "updated_at",
		}).AddRow(int64(7), "gpt", "GPT", service.PlatformOpenAI, service.StatusActive, []byte(`["chat_completions"]`), nil, createdAt, createdAt))
	mock.ExpectQuery("SELECT id, platform_id, model_pattern, upstream_model, status, created_at, updated_at").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "platform_id", "model_pattern", "upstream_model", "status", "created_at", "updated_at",
		}).AddRow(int64(11), int64(7), "gpt-4o", "", service.StatusDisabled, createdAt, createdAt))

	platform, err := newPlatformRepository(db).GetByID(context.Background(), 7)

	require.NoError(t, err)
	require.Equal(t, "gpt", platform.Code)
	require.Equal(t, service.PlatformOpenAI, platform.AccountPlatform)
	require.Len(t, platform.ModelRules, 1)
	require.False(t, platform.ModelRules[0].Enabled)
	require.Equal(t, []string{"chat_completions"}, platform.EndpointCapabilities)
	require.Equal(t, platform.EndpointCapabilities, platform.ModelRules[0].EndpointCapabilities)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestPlatformRepositoryListLoadsAllPlatformsAndRules(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	createdAt := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	mock.ExpectQuery("SELECT id, code, name, account_platform, status, endpoint_capabilities, legacy_group_id, created_at, updated_at FROM platforms").
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "code", "name", "account_platform", "status", "endpoint_capabilities", "legacy_group_id", "created_at", "updated_at",
		}).AddRow(int64(7), "gpt", "GPT", service.PlatformOpenAI, service.StatusActive, []byte(`["chat_completions"]`), nil, createdAt, createdAt))
	mock.ExpectQuery("SELECT id, platform_id, model_pattern, upstream_model, status, created_at, updated_at").
		WithArgs(int64(7)).
		WillReturnRows(sqlmock.NewRows([]string{
			"id", "platform_id", "model_pattern", "upstream_model", "status", "created_at", "updated_at",
		}).AddRow(int64(11), int64(7), "gpt-4o", "gpt-4o-2024-08-06", service.StatusActive, createdAt, createdAt))

	platforms, err := newPlatformRepository(db).List(context.Background())

	require.NoError(t, err)
	require.Len(t, platforms, 1)
	require.Equal(t, "gpt", platforms[0].Code)
	require.Len(t, platforms[0].ModelRules, 1)
	require.Equal(t, "gpt-4o", platforms[0].ModelRules[0].ModelPattern)
	require.Equal(t, []string{"chat_completions"}, platforms[0].EndpointCapabilities)
	require.Equal(t, platforms[0].EndpointCapabilities, platforms[0].ModelRules[0].EndpointCapabilities)
	require.NoError(t, mock.ExpectationsWereMet())
}
