//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildPlatformMigrationPreflightReportsAmbiguities(t *testing.T) {
	report := buildPlatformMigrationPreflight(legacyPlatformMigrationSnapshot{
		ModelRules: []legacyPlatformModelRule{
			{GroupID: 10, Platform: PlatformOpenAI, Pattern: "gpt-*"},
			{GroupID: 11, Platform: PlatformGrok, Pattern: "gpt-4o"},
		},
		Accounts: []legacyAccountPlatformBinding{
			{AccountID: 7, Platforms: []string{PlatformOpenAI, PlatformGemini}},
		},
		GroupPlanIDs: map[int64][]int64{
			10: {100, 101},
			11: {200},
		},
		GroupSubscriptionTypes: map[int64]string{
			10: SubscriptionTypeSubscription,
			11: SubscriptionTypeSubscription,
		},
		APIKeys: []legacyAPIKeyGroupBinding{
			{APIKeyID: 3, GroupIDs: []int64{10}},
		},
	})

	require.False(t, report.Ready)
	require.ElementsMatch(t, []string{
		platformMigrationConflictModelPatternOverlap,
		platformMigrationConflictAccountPlatformAmbiguity,
		platformMigrationConflictAPIKeyPlanAmbiguity,
	}, platformMigrationConflictCodes(report.Conflicts))
}

func TestBuildPlatformMigrationPreflightAllowsExplicitMultiplePlatforms(t *testing.T) {
	report := buildPlatformMigrationPreflight(legacyPlatformMigrationSnapshot{
		GroupPlatforms: map[int64]string{
			10: PlatformOpenAI,
			11: PlatformGrok,
		},
		GroupPlanIDs: map[int64][]int64{
			10: {100},
			11: {200},
		},
		GroupSubscriptionTypes: map[int64]string{
			10: SubscriptionTypeSubscription,
			11: SubscriptionTypeSubscription,
		},
		APIKeys: []legacyAPIKeyGroupBinding{
			{APIKeyID: 3, GroupIDs: []int64{10, 11}},
		},
	})

	require.True(t, report.Ready)
	require.Empty(t, report.Conflicts)
}

func TestBuildPlatformMigrationPreflightRejectsCatchAllModelRuleOverlap(t *testing.T) {
	report := buildPlatformMigrationPreflight(legacyPlatformMigrationSnapshot{
		ModelRules: []legacyPlatformModelRule{
			{GroupID: 10, Platform: PlatformOpenAI, Pattern: "*"},
			{GroupID: 11, Platform: PlatformGrok, Pattern: "gpt-4o"},
		},
	})

	require.False(t, report.Ready)
	require.Equal(t, []string{platformMigrationConflictModelPatternOverlap}, platformMigrationConflictCodes(report.Conflicts))
}

func TestBuildPlatformMigrationPreflightAllowsSubscriptionAndBalanceGroups(t *testing.T) {
	report := buildPlatformMigrationPreflight(legacyPlatformMigrationSnapshot{
		GroupPlatforms: map[int64]string{
			10: PlatformOpenAI,
			20: PlatformOpenAI,
		},
		GroupSubscriptionTypes: map[int64]string{
			10: SubscriptionTypeSubscription,
			20: SubscriptionTypeStandard,
		},
		GroupPlanIDs: map[int64][]int64{
			10: {100},
		},
		APIKeys: []legacyAPIKeyGroupBinding{
			{APIKeyID: 3, GroupIDs: []int64{10, 20}},
		},
	})

	require.True(t, report.Ready)
	require.Empty(t, report.Conflicts)
}

func TestBuildPlatformMigrationPreflightRejectsUnknownGroupAssetType(t *testing.T) {
	report := buildPlatformMigrationPreflight(legacyPlatformMigrationSnapshot{
		GroupPlatforms: map[int64]string{10: PlatformOpenAI},
		GroupPlanIDs:   map[int64][]int64{10: {100}},
		APIKeys: []legacyAPIKeyGroupBinding{
			{APIKeyID: 3, GroupIDs: []int64{10}},
		},
	})

	require.False(t, report.Ready)
	require.Equal(t, []string{platformMigrationConflictAPIKeyAssetTypeUnmapped}, platformMigrationConflictCodes(report.Conflicts))
}
