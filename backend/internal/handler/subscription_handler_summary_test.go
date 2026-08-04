//go:build unit

package handler

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSubscriptionSummaryItemUsesInstanceSnapshotLimits(t *testing.T) {
	legacyLimit := 100.0
	dailyLimit := 10.0
	weeklyLimit := 20.0
	monthlyLimit := 30.0
	subscription := service.UserSubscription{
		ID:                      1,
		GroupID:                 2,
		DailyLimitUSDSnapshot:   &dailyLimit,
		WeeklyLimitUSDSnapshot:  &weeklyLimit,
		MonthlyLimitUSDSnapshot: &monthlyLimit,
		Group: &service.Group{
			Name:            "Routing group",
			DailyLimitUSD:   &legacyLimit,
			WeeklyLimitUSD:  &legacyLimit,
			MonthlyLimitUSD: &legacyLimit,
		},
	}

	item := subscriptionSummaryItemFromService(subscription)

	require.Equal(t, "Routing group", item.GroupName)
	require.Equal(t, dailyLimit, item.DailyLimitUSD)
	require.Equal(t, weeklyLimit, item.WeeklyLimitUSD)
	require.Equal(t, monthlyLimit, item.MonthlyLimitUSD)
}
