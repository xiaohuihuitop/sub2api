//go:build unit

package dto

import (
	"encoding/json"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyFromServiceDoesNotExposeLegacyGroups(t *testing.T) {
	src := &service.APIKey{
		ID:              1,
		AllowedGroupIDs: []int64{10, 20},
		AllowedGroups: []service.Group{
			{ID: 10, Name: "Balance", SortOrder: 3},
			{ID: 20, Name: "Plan"},
		},
	}

	out := APIKeyFromService(src)

	body, err := json.Marshal(out)
	require.NoError(t, err)
	require.NotContains(t, string(body), `"group_id"`)
	require.NotContains(t, string(body), `"group_ids"`)
	require.NotContains(t, string(body), `"group"`)
	require.NotContains(t, string(body), `"groups"`)
}

func TestAPIKeyFromServiceMapsAssetPermissions(t *testing.T) {
	src := &service.APIKey{
		ID:                         1,
		AllowedPlatformIDs:         []int64{10, 30},
		AllowedSubscriptionPlanIDs: []int64{20, 80},
		AllowBalance:               false,
	}

	out := APIKeyFromService(src)

	require.Equal(t, []int64{10, 30}, out.PlatformIDs)
	require.Equal(t, []int64{20, 80}, out.SubscriptionPlanIDs)
	require.False(t, out.AllowBalance)
}
