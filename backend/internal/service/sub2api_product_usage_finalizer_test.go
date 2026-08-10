//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/applicationgateway"
	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/stretchr/testify/require"
)

func TestSub2APIProductUsageFinalizerSkipsFailedTerminalEvents(t *testing.T) {
	finalizer := NewSub2APIProductUsageFinalizer(nil, nil, nil)
	err := finalizer.Finalize(context.Background(), ProductUsageRecord{
		Snapshot: applicationgateway.DecisionSnapshot{},
		Event: gatewayruntime.UsageEvent{
			Success: false,
			Error:   gatewayruntime.NewRuntimeError(gatewayruntime.ErrorUpstream5xx, true, "upstream failed"),
		},
	})
	require.NoError(t, err)
}

func TestSub2APIProductUsageFinalizerRequiresCompleteDependenciesForSuccessfulEvent(t *testing.T) {
	finalizer := NewSub2APIProductUsageFinalizer(nil, nil, nil)
	err := finalizer.Finalize(context.Background(), ProductUsageRecord{
		Snapshot: applicationgateway.DecisionSnapshot{},
		Event: gatewayruntime.UsageEvent{
			Success: true,
			Facts:   gatewayruntime.UsageFacts{AccountID: 1},
		},
	})
	require.ErrorIs(t, err, ErrProductUsageFinalizerUnavailable)
}
