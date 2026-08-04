package repository

import (
	"context"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupEntityToService_PreservesMessagesDispatchModelConfig(t *testing.T) {
	group := &dbent.Group{
		ID:                    1,
		Name:                  "openai-dispatch",
		Platform:              service.PlatformOpenAI,
		Status:                service.StatusActive,
		SubscriptionType:      service.SubscriptionTypeStandard,
		RateMultiplier:        1,
		AllowMessagesDispatch: true,
		DefaultMappedModel:    "gpt-5.4",
		MessagesDispatchModelConfig: service.OpenAIMessagesDispatchModelConfig{
			OpusMappedModel:   "gpt-5.4-nano",
			SonnetMappedModel: "gpt-5.3-codex",
			HaikuMappedModel:  "gpt-5.4-mini",
			ExactModelMappings: map[string]string{
				"claude-sonnet-4.5": "gpt-5.4-nano",
			},
		},
	}

	got := groupEntityToService(group)
	require.NotNil(t, got)
	require.Equal(t, group.MessagesDispatchModelConfig, got.MessagesDispatchModelConfig)
}

func TestGroupEntityToService_UsesBillingProfileForBalancePricing(t *testing.T) {
	legacyImagePrice := 11.0
	profileImagePrice := 0.11
	group := &dbent.Group{
		ID:                           1,
		Name:                         "routing-only-group",
		Platform:                     service.PlatformOpenAI,
		Status:                       service.StatusActive,
		RateMultiplier:               9,
		PeakRateEnabled:              false,
		ImageRateIndependent:         false,
		ImageRateMultiplier:          9,
		ImagePrice1k:                 &legacyImagePrice,
		BatchImageDiscountMultiplier: 9,
		BatchImageHoldMultiplier:     9,
		VideoRateIndependent:         false,
		VideoRateMultiplier:          9,
		Edges: dbent.GroupEdges{
			BillingProfile: &dbent.BillingProfile{
				BalanceRateMultiplier:        1.25,
				PeakRateEnabled:              true,
				PeakStart:                    "08:00",
				PeakEnd:                      "20:00",
				PeakRateMultiplier:           1.5,
				ImageRateIndependent:         true,
				ImageRateMultiplier:          2.5,
				ImagePrice1k:                 &profileImagePrice,
				BatchImageDiscountMultiplier: 0.5,
				BatchImageHoldMultiplier:     0.6,
				VideoRateIndependent:         true,
				VideoRateMultiplier:          3.5,
			},
		},
	}

	got := groupEntityToService(group)

	require.Equal(t, 1.25, got.RateMultiplier)
	require.True(t, got.PeakRateEnabled)
	require.Equal(t, "08:00", got.PeakStart)
	require.Equal(t, "20:00", got.PeakEnd)
	require.Equal(t, 1.5, got.PeakRateMultiplier)
	require.True(t, got.ImageRateIndependent)
	require.Equal(t, 2.5, got.ImageRateMultiplier)
	require.Equal(t, &profileImagePrice, got.ImagePrice1K)
	require.Equal(t, 0.5, got.BatchImageDiscountMultiplier)
	require.Equal(t, 0.6, got.BatchImageHoldMultiplier)
	require.True(t, got.VideoRateIndependent)
	require.Equal(t, 3.5, got.VideoRateMultiplier)
}

func TestAPIKeyRepository_GetByKeyForAuth_PreservesMessagesDispatchModelConfig_SQLite(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "getbykey-auth-dispatch-unit@test.com")

	group, err := client.Group.Create().
		SetName("g-auth-dispatch-unit").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetSubscriptionType(service.SubscriptionTypeStandard).
		SetRateMultiplier(1).
		SetAllowMessagesDispatch(true).
		SetDefaultMappedModel("gpt-5.4").
		SetMessagesDispatchModelConfig(service.OpenAIMessagesDispatchModelConfig{
			OpusMappedModel:   "gpt-5.4-nano",
			SonnetMappedModel: "gpt-5.3-codex",
			HaikuMappedModel:  "gpt-5.4-mini",
			ExactModelMappings: map[string]string{
				"claude-sonnet-4.5": "gpt-5.4-nano",
			},
		}).
		Save(ctx)
	require.NoError(t, err)

	key := &service.APIKey{
		UserID:  user.ID,
		Key:     "sk-getbykey-auth-dispatch-unit",
		Name:    "Dispatch Key Unit",
		GroupID: &group.ID,
		Status:  service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	got, err := repo.GetByKeyForAuth(ctx, key.Key)
	require.NoError(t, err)
	require.Equal(t, key.Name, got.Name)
	require.NotNil(t, got.Group)
	require.Equal(t, group.MessagesDispatchModelConfig, got.Group.MessagesDispatchModelConfig)
}

func TestAPIKeyRepository_GetByKeyForAuth_UsesBillingProfileForBalancePricing_SQLite(t *testing.T) {
	repo, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	user := mustCreateAPIKeyRepoUser(t, ctx, client, "getbykey-auth-profile-unit@test.com")

	group, err := client.Group.Create().
		SetName("g-auth-profile-unit").
		SetPlatform(service.PlatformOpenAI).
		SetStatus(service.StatusActive).
		SetRateMultiplier(9).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.BillingProfile.Create().
		SetGroupID(group.ID).
		SetBalanceRateMultiplier(1.25).
		SetPeakRateEnabled(true).
		SetPeakStart("08:00").
		SetPeakEnd("20:00").
		SetPeakRateMultiplier(1.5).
		Save(ctx)
	require.NoError(t, err)

	key := &service.APIKey{
		UserID:  user.ID,
		Key:     "sk-getbykey-auth-profile-unit",
		Name:    "Profile Key Unit",
		GroupID: &group.ID,
		Status:  service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, key))

	got, err := repo.GetByKeyForAuth(ctx, key.Key)
	require.NoError(t, err)
	require.NotNil(t, got.Group)
	require.Equal(t, 1.25, got.Group.RateMultiplier)
	require.True(t, got.Group.PeakRateEnabled)
	require.Equal(t, 1.5, got.Group.PeakRateMultiplier)
}
