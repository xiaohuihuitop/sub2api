//go:build unit

package repository

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/pkg/openai_compat"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

func TestOpenAIAccountBillingEndpointCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		account  *service.Account
		expected []service.OpenAIEndpointCapability
	}{
		{
			name:    "oauth accepts both inbound text endpoints",
			account: &service.Account{Platform: service.PlatformOpenAI, Type: service.AccountTypeOAuth},
			expected: []service.OpenAIEndpointCapability{
				service.OpenAIEndpointCapabilityChatCompletions,
				service.OpenAIEndpointCapabilityResponses,
			},
		},
		{
			name: "auto api key accepts both inbound endpoints when probe supports responses",
			account: &service.Account{
				Platform: service.PlatformOpenAI,
				Type:     service.AccountTypeAPIKey,
				Extra: map[string]any{
					openai_compat.ExtraKeyResponsesSupported: true,
				},
			},
			expected: []service.OpenAIEndpointCapability{
				service.OpenAIEndpointCapabilityChatCompletions,
				service.OpenAIEndpointCapabilityResponses,
			},
		},
		{
			name: "auto api key accepts both inbound endpoints when probe requires chat fallback",
			account: &service.Account{
				Platform: service.PlatformOpenAI,
				Type:     service.AccountTypeAPIKey,
				Extra: map[string]any{
					openai_compat.ExtraKeyResponsesSupported: false,
				},
			},
			expected: []service.OpenAIEndpointCapability{
				service.OpenAIEndpointCapabilityChatCompletions,
				service.OpenAIEndpointCapabilityResponses,
			},
		},
		{
			name: "force responses accepts only inbound responses",
			account: &service.Account{
				Platform: service.PlatformOpenAI,
				Type:     service.AccountTypeAPIKey,
				Extra: map[string]any{
					openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceResponses),
				},
			},
			expected: []service.OpenAIEndpointCapability{
				service.OpenAIEndpointCapabilityResponses,
			},
		},
		{
			name: "force chat completions accepts only inbound chat completions",
			account: &service.Account{
				Platform: service.PlatformOpenAI,
				Type:     service.AccountTypeAPIKey,
				Extra: map[string]any{
					openai_compat.ExtraKeyResponsesMode: string(openai_compat.ResponsesSupportModeForceChatCompletions),
				},
			},
			expected: []service.OpenAIEndpointCapability{
				service.OpenAIEndpointCapabilityChatCompletions,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, openAIAccountBillingEndpointCapabilities(tt.account))
		})
	}
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
