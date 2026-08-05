//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type platformPoolSchedulingRepoStub struct {
	AccountRepository
	poolAccounts   []Account
	legacyAccounts map[string][]Account
	poolID         int64
	poolPlatform   string
	poolCalls      int
	legacyCalls    int
}

func (r *platformPoolSchedulingRepoStub) ListSchedulableByPlatformPool(
	_ context.Context,
	platformID int64,
	accountPlatform string,
) ([]Account, error) {
	r.poolCalls++
	r.poolID = platformID
	r.poolPlatform = accountPlatform
	return append([]Account(nil), r.poolAccounts...), nil
}

func (r *platformPoolSchedulingRepoStub) ListSchedulableUngroupedByPlatform(_ context.Context, platform string) ([]Account, error) {
	r.legacyCalls++
	return append([]Account(nil), r.legacyAccounts[platform]...), nil
}

func (r *platformPoolSchedulingRepoStub) ListSchedulableUngroupedByPlatforms(_ context.Context, platforms []string) ([]Account, error) {
	r.legacyCalls++
	accounts := make([]Account, 0, len(platforms))
	for _, platform := range platforms {
		accounts = append(accounts, r.legacyAccounts[platform]...)
	}
	return accounts, nil
}

func TestGatewayServiceSelectsOnlyExplicitPlatformPool(t *testing.T) {
	platformID := int64(42)
	repo := &platformPoolSchedulingRepoStub{
		poolAccounts: []Account{{
			ID:          420,
			PlatformID:  &platformID,
			Platform:    PlatformAnthropic,
			Status:      StatusActive,
			Schedulable: true,
		}},
		legacyAccounts: map[string][]Account{
			PlatformAnthropic: {{ID: 419, Platform: PlatformAnthropic, Status: StatusActive, Schedulable: true}},
		},
	}
	ctx := WithPlatformSchedulingScope(context.Background(), PlatformSchedulingScope{
		PlatformID:      platformID,
		PlatformCode:    "anthropic-primary",
		AccountPlatform: PlatformAnthropic,
	})

	account, err := (&GatewayService{accountRepo: repo}).SelectAccountForModelWithExclusions(ctx, nil, "", "", nil)

	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, int64(420), account.ID)
	require.Equal(t, 1, repo.poolCalls)
	require.Equal(t, platformID, repo.poolID)
	require.Equal(t, PlatformAnthropic, repo.poolPlatform)
	require.Zero(t, repo.legacyCalls)
}

func TestOpenAIGatewayServiceSelectsOnlyExplicitPlatformPool(t *testing.T) {
	platformID := int64(43)
	repo := &platformPoolSchedulingRepoStub{
		poolAccounts: []Account{{
			ID:          430,
			PlatformID:  &platformID,
			Platform:    PlatformOpenAI,
			Status:      StatusActive,
			Schedulable: true,
		}},
		legacyAccounts: map[string][]Account{
			PlatformOpenAI: {{ID: 429, Platform: PlatformOpenAI, Status: StatusActive, Schedulable: true}},
		},
	}
	ctx := WithPlatformSchedulingScope(context.Background(), PlatformSchedulingScope{
		PlatformID:      platformID,
		PlatformCode:    "gpt-primary",
		AccountPlatform: PlatformOpenAI,
	})

	account, err := (&OpenAIGatewayService{accountRepo: repo}).SelectAccountForModelWithExclusions(ctx, nil, "", "", nil)

	require.NoError(t, err)
	require.NotNil(t, account)
	require.Equal(t, int64(430), account.ID)
	require.Equal(t, 1, repo.poolCalls)
	require.Equal(t, platformID, repo.poolID)
	require.Equal(t, PlatformOpenAI, repo.poolPlatform)
	require.Zero(t, repo.legacyCalls)
}
