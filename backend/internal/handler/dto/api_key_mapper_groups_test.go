//go:build unit

package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyFromServiceMapsAllowedGroups(t *testing.T) {
	src := &service.APIKey{
		ID:              1,
		AllowedGroupIDs: []int64{10, 20},
		AllowedGroups: []service.Group{
			{ID: 10, Name: "Balance", SortOrder: 3},
			{ID: 20, Name: "Plan"},
		},
	}

	out := APIKeyFromService(src)

	require.Equal(t, []int64{10, 20}, out.GroupIDs)
	require.Len(t, out.Groups, 2)
	require.Equal(t, int64(10), out.Groups[0].ID)
	require.Equal(t, 3, out.Groups[0].SortOrder)
	require.Equal(t, int64(20), out.Groups[1].ID)
}
