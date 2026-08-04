//go:build unit

package admin

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAssignSubscriptionRequestAllowsPlanWithoutGroup(t *testing.T) {
	req := AssignSubscriptionRequest{UserID: 7, PlanID: 8}

	require.NoError(t, req.ValidateAssignmentSource())
	require.True(t, req.UsesPlan())
}

func TestAssignSubscriptionRequestRequiresPlanOrGroup(t *testing.T) {
	req := AssignSubscriptionRequest{UserID: 7}

	require.Error(t, req.ValidateAssignmentSource())
	require.False(t, req.UsesPlan())
}

func TestBulkAssignSubscriptionRequestAllowsPlanWithoutGroup(t *testing.T) {
	req := BulkAssignSubscriptionRequest{UserIDs: []int64{7, 8}, PlanID: 9}

	require.NoError(t, req.ValidateAssignmentSource())
	require.True(t, req.UsesPlan())
}
