//go:build unit

package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestResolveBillingMultipliersSubscriptionSnapshotOverridesBalanceProfile(t *testing.T) {
	group := &Group{
		RateMultiplier:       2,
		ImageRateIndependent: true,
		ImageRateMultiplier:  3,
		VideoRateIndependent: true,
		VideoRateMultiplier:  4,
		PeakRateEnabled:      true,
		PeakStart:            "00:00",
		PeakEnd:              "23:59",
		PeakRateMultiplier:   1.5,
	}
	apiKey := &APIKey{Group: group}
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)

	token, image, video := resolveBillingMultipliers(apiKey, nil, 9, now)
	require.Equal(t, 3.0, token)
	require.Equal(t, 3.0, image)
	require.Equal(t, 4.0, video)

	subscription := &UserSubscription{RateMultiplierSnapshot: 0.5}
	token, image, video = resolveBillingMultipliers(apiKey, subscription, 9, now)
	require.Equal(t, 0.5, token)
	require.Equal(t, 0.5, image)
	require.Equal(t, 0.5, video)
}

func TestResolveBillingMultipliersAllowsFreeSubscription(t *testing.T) {
	token, image, video := resolveBillingMultipliers(
		&APIKey{Group: &Group{RateMultiplier: 3}},
		&UserSubscription{RateMultiplierSnapshot: 0},
		1,
		time.Now(),
	)

	require.Equal(t, 0.0, token)
	require.Equal(t, 0.0, image)
	require.Equal(t, 0.0, video)
}
