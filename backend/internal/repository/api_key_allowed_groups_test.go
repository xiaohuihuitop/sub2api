//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/ent/apikey"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAccountBillingEndpointCapabilities(t *testing.T) {
	oauth := &service.Account{Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth}
	autoChatFallback := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
		Extra: map[string]any{
			openai_compat.ExtraKeyResponsesMode:      string(openai_compat.ResponsesSupportModeAuto),
			openai_compat.ExtraKeyResponsesSupported: false,
		},
	}
	forceResponses := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
		Extra: map[string]any{
			openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceResponses),
		},
	}
	forceChatCompletions := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
		Extra: map[string]any{
			openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceChatCompletions),
		},
	}
	chatOnly := &service.Account{
		Platform: service.PlatformOpenAI,
		Type:     service.AccountTypeAPIKey,
		Credentials: map[string]any{
			"openai_capabilities": []any{"chat_completions"},
		},
		Extra: map[string]any{openai_compat.ExtraKeyResponsesSupported: false},
	}

	require.Equal(t, []service.OpenAIEndpointCapability{service.OpenAIEndpointCapabilityResponses}, openAIAccountBillingEndpointCapabilities(oauth))
	require.Equal(t, []service.OpenAIEndpointCapability{
		service.OpenAIEndpointCapabilityChatCompletions,
		service.OpenAIEndpointCapabilityResponses,
	}, openAIAccountBillingEndpointCapabilities(autoChatFallback))
	require.Equal(t, []service.OpenAIEndpointCapability{service.OpenAIEndpointCapabilityResponses}, openAIAccountBillingEndpointCapabilities(forceResponses))
	require.Equal(t, []service.OpenAIEndpointCapability{service.OpenAIEndpointCapabilityChatCompletions}, openAIAccountBillingEndpointCapabilities(forceChatCompletions))
	require.Equal(t, []service.OpenAIEndpointCapability{
		service.OpenAIEndpointCapabilityChatCompletions,
		service.OpenAIEndpointCapabilityResponses,
	}, openAIAccountBillingEndpointCapabilities(chatOnly))
}

func TestAPIKeyRepositoryReplaceAllowedGroupsUsesAtomicStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newAPIKeyRepositoryWithSQL(nil, db)
	mock.ExpectExec(`(?s)WITH deleted AS.*DELETE FROM api_key_allowed_groups.*INSERT INTO api_key_allowed_groups`).
		WithArgs(int64(7), pq.Array([]int64{10, 20})).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = repo.ReplaceAllowedGroups(context.Background(), 7, []int64{10, 20})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyRepositoryReplaceAllowedGroupsClearsWithEmptyArray(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newAPIKeyRepositoryWithSQL(nil, db)
	mock.ExpectExec(`(?s)WITH deleted AS.*DELETE FROM api_key_allowed_groups.*INSERT INTO api_key_allowed_groups`).
		WithArgs(int64(8), pq.Array([]int64{})).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = repo.ReplaceAllowedGroups(context.Background(), 8, []int64{})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyRepositoryReplaceAssetPermissionsUsesAtomicStatement(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newAPIKeyRepositoryWithSQL(nil, db)
	permissions := service.APIKeyAssetPermissions{
		PlatformIDs:         []int64{30, 10, 30},
		SubscriptionPlanIDs: []int64{80, 20, 80},
		AllowBalance:        false,
	}
	mock.ExpectExec(`(?s)WITH target AS.*normalized_platforms AS.*normalized_plans AS.*DELETE FROM api_key_platforms.*NOT EXISTS.*DELETE FROM api_key_subscription_plans.*NOT EXISTS.*INSERT INTO api_key_subscription_plans`).
		WithArgs(int64(7), false, pq.Array([]int64{10, 30}), pq.Array([]int64{20, 80})).
		WillReturnResult(sqlmock.NewResult(0, 2))

	err = repo.ReplaceAssetPermissions(context.Background(), 7, permissions)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyRepositoryLoadAssetPermissions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newAPIKeyRepositoryWithSQL(nil, db)
	keys := []*service.APIKey{{ID: 7}, {ID: 8}}

	mock.ExpectQuery(`(?s)SELECT api_key_id, platform_id.*FROM api_key_platforms`).
		WithArgs(pq.Array([]int64{7, 8})).
		WillReturnRows(sqlmock.NewRows([]string{"api_key_id", "platform_id"}).
			AddRow(int64(7), int64(30)).
			AddRow(int64(7), int64(10)).
			AddRow(int64(8), int64(20)))
	mock.ExpectQuery(`(?s)SELECT api_key_id, subscription_plan_id.*FROM api_key_subscription_plans`).
		WithArgs(pq.Array([]int64{7, 8})).
		WillReturnRows(sqlmock.NewRows([]string{"api_key_id", "subscription_plan_id"}).
			AddRow(int64(7), int64(80)).
			AddRow(int64(8), int64(20)).
			AddRow(int64(8), int64(60)))

	err = repo.loadAssetPermissions(context.Background(), keys)

	require.NoError(t, err)
	require.Equal(t, []int64{10, 30}, keys[0].AllowedPlatformIDs)
	require.Equal(t, []int64{20}, keys[1].AllowedPlatformIDs)
	require.Equal(t, []int64{80}, keys[0].AllowedSubscriptionPlanIDs)
	require.Equal(t, []int64{20, 60}, keys[1].AllowedSubscriptionPlanIDs)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyAuthFieldSelectionIncludesAssetPermissions(t *testing.T) {
	fields := apiKeyAuthFieldSelection()

	require.Contains(t, fields, apikey.FieldAllowBalance)
}

func TestAPIKeyRepositoryClearGroupRemovesAllowedLinkAndReassignsPrimary(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newAPIKeyRepositoryWithSQL(nil, db)
	mock.ExpectExec(`(?s)WITH affected AS.*DELETE FROM api_key_allowed_groups.*UPDATE api_keys`).
		WithArgs(int64(20)).
		WillReturnResult(sqlmock.NewResult(0, 2))

	affected, err := repo.ClearGroupIDByGroupID(context.Background(), 20)

	require.NoError(t, err)
	require.Equal(t, int64(2), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestAPIKeyRepositoryMigrateGroupUpdatesAllowedLinksAndPrimary(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	repo := newAPIKeyRepositoryWithSQL(nil, db)
	mock.ExpectExec(`(?s)WITH affected AS.*INSERT INTO api_key_allowed_groups.*DELETE FROM api_key_allowed_groups.*UPDATE api_keys`).
		WithArgs(int64(7), int64(20), int64(30)).
		WillReturnResult(sqlmock.NewResult(0, 3))

	affected, err := repo.UpdateGroupIDByUserAndGroup(context.Background(), 7, 20, 30)

	require.NoError(t, err)
	require.Equal(t, int64(3), affected)
	require.NoError(t, mock.ExpectationsWereMet())
}
