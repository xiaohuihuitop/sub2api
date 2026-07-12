//go:build unit

package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNormalizeAPIKeyGroupIDsUsesExplicitGroups(t *testing.T) {
	legacy := int64(99)

	groupIDs := normalizeAPIKeyGroupIDs([]int64{30, 10, 30, 0, -1}, &legacy)

	require.Equal(t, []int64{30, 10}, groupIDs)
}

func TestNormalizeAPIKeyGroupIDsFallsBackToLegacyGroup(t *testing.T) {
	legacy := int64(20)

	groupIDs := normalizeAPIKeyGroupIDs(nil, &legacy)

	require.Equal(t, []int64{20}, groupIDs)
}

func TestSortAPIKeyAllowedGroupsUsesAdminOrderThenID(t *testing.T) {
	groups := []Group{
		{ID: 30, SortOrder: 2},
		{ID: 20, SortOrder: 1},
		{ID: 10, SortOrder: 1},
	}

	sorted := sortAPIKeyAllowedGroups(groups)

	require.Equal(t, []int64{10, 20, 30}, apiKeyGroupIDs(sorted))
	require.Equal(t, []int64{30, 20, 10}, apiKeyGroupIDs(groups), "sorting must not mutate repository data")
}
