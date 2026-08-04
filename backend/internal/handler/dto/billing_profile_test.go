//go:build unit

package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBillingProfileFromService(t *testing.T) {
	price := 0.12
	got := BillingProfileFromService(&service.BillingProfile{
		GroupID:               7,
		BalanceRateMultiplier: 1.25,
		PeakRateEnabled:       true,
		PeakStart:             "09:00",
		PeakEnd:               "18:00",
		PeakRateMultiplier:    1.5,
		ImagePrice1K:          &price,
	})

	require.Equal(t, int64(7), got.GroupID)
	require.Equal(t, 1.25, got.BalanceRateMultiplier)
	require.True(t, got.PeakRateEnabled)
	require.Equal(t, "09:00", got.PeakStart)
	require.Equal(t, &price, got.ImagePrice1K)
}
