package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/ent/billingprofile"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestGroupRepositoryGetByIDLiteUsesBillingProfileForBalancePricing(t *testing.T) {
	_, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	group, err := client.Group.Create().
		SetName("billing-profile-lite-unit").
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

	repo := newGroupRepositoryWithSQL(client, nil)
	got, err := repo.GetByIDLite(ctx, group.ID)

	require.NoError(t, err)
	require.Equal(t, 1.25, got.RateMultiplier)
	require.True(t, got.PeakRateEnabled)
	require.Equal(t, 1.5, got.PeakRateMultiplier)
}

func TestGroupRepositoryCreateSeedsBillingProfile(t *testing.T) {
	_, client := newAPIKeyRepoSQLite(t)
	ctx := context.Background()
	repo := newGroupRepositoryWithSQL(client, nil)
	imagePrice := 0.12
	group := &service.Group{
		Name:                         "seeded-billing-profile-unit",
		Platform:                     service.PlatformOpenAI,
		Status:                       service.StatusActive,
		RateMultiplier:               1.25,
		PeakRateEnabled:              true,
		PeakStart:                    "09:00",
		PeakEnd:                      "18:00",
		PeakRateMultiplier:           1.5,
		ImageRateIndependent:         true,
		ImageRateMultiplier:          2,
		BatchImageDiscountMultiplier: 0.5,
		BatchImageHoldMultiplier:     0.6,
		ImagePrice1K:                 &imagePrice,
	}

	require.NoError(t, repo.Create(ctx, group))
	profile, err := client.BillingProfile.Query().Where(billingprofile.GroupIDEQ(group.ID)).Only(ctx)

	require.NoError(t, err)
	require.Equal(t, 1.25, profile.BalanceRateMultiplier)
	require.True(t, profile.PeakRateEnabled)
	require.Equal(t, "09:00", profile.PeakStart)
	require.Equal(t, 1.5, profile.PeakRateMultiplier)
	require.Equal(t, &imagePrice, profile.ImagePrice1k)
}
