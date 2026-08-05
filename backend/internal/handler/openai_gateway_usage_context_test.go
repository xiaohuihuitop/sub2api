package handler

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestSubmitUsageRecordTaskCopiesRequestContext(t *testing.T) {
	parent := context.WithValue(context.Background(), ctxkey.ClientRequestID, "client-request-123")
	parent = context.WithValue(parent, ctxkey.RequestID, "request-456")

	var gotClientRequestID string
	var gotRequestID string
	h := &GatewayHandler{}
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		gotClientRequestID, _ = ctx.Value(ctxkey.ClientRequestID).(string)
		gotRequestID, _ = ctx.Value(ctxkey.RequestID).(string)
	})

	require.Equal(t, "client-request-123", gotClientRequestID)
	require.Equal(t, "request-456", gotRequestID)
}

func TestOpenAISubmitUsageRecordTaskCopiesRequestContext(t *testing.T) {
	parent := context.WithValue(context.Background(), ctxkey.ClientRequestID, "openai-client-request-123")
	parent = context.WithValue(parent, ctxkey.RequestID, "openai-request-456")

	var gotClientRequestID string
	var gotRequestID string
	h := &OpenAIGatewayHandler{}
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		gotClientRequestID, _ = ctx.Value(ctxkey.ClientRequestID).(string)
		gotRequestID, _ = ctx.Value(ctxkey.RequestID).(string)
	})

	require.Equal(t, "openai-client-request-123", gotClientRequestID)
	require.Equal(t, "openai-request-456", gotRequestID)
}

func TestOpenAISubmitUsageRecordTaskPreservesPlatformAssetContext(t *testing.T) {
	platformID := int64(42)
	parent := service.WithGatewayPlatformAssetContext(context.Background(), &service.GatewayPlatformAssetContext{
		Platform: &service.ResolvedPlatformModel{
			PlatformID:      platformID,
			PlatformCode:    "openai",
			AccountPlatform: service.PlatformOpenAI,
		},
		BillingAsset: &service.ResolvedBillingAsset{
			Source:         service.BillingSourceBalance,
			RateMultiplier: 1.25,
		},
		SchedulingScope: service.PlatformSchedulingScope{
			PlatformID:      platformID,
			PlatformCode:    "openai",
			AccountPlatform: service.PlatformOpenAI,
		},
	})

	var got *service.GatewayPlatformAssetContext
	h := &OpenAIGatewayHandler{}
	h.submitUsageRecordTask(parent, func(ctx context.Context) {
		got, _ = service.GatewayPlatformAssetContextFromContext(ctx)
	})

	require.NotNil(t, got)
	require.Equal(t, platformID, got.Platform.PlatformID)
	require.Equal(t, service.BillingSourceBalance, got.BillingAsset.Source)
	require.Equal(t, 1.25, got.BillingAsset.RateMultiplier)
}

func TestUsageRecordContextPreservesRouteAfterParentCancellation(t *testing.T) {
	parent, cancel := context.WithCancel(context.Background())
	parent = service.WithGatewayPlatformAssetContext(parent, &service.GatewayPlatformAssetContext{
		Platform: &service.ResolvedPlatformModel{
			PlatformID:      42,
			PlatformCode:    "openai",
			AccountPlatform: service.PlatformOpenAI,
		},
		BillingAsset: &service.ResolvedBillingAsset{
			Source:         service.BillingSourceSubscription,
			SubscriptionID: int64Pointer(7),
			RateMultiplier: 1.5,
		},
		SchedulingScope: service.PlatformSchedulingScope{
			PlatformID:      42,
			PlatformCode:    "openai",
			AccountPlatform: service.PlatformOpenAI,
		},
	})
	cancel()

	workerContext := usageRecordContext(parent, context.Background())
	route, ok := service.GatewayPlatformAssetContextFromContext(workerContext)

	require.NoError(t, workerContext.Err())
	require.True(t, ok)
	require.Equal(t, service.BillingSourceSubscription, route.BillingAsset.Source)
	require.Equal(t, int64(7), *route.BillingAsset.SubscriptionID)
}

func int64Pointer(value int64) *int64 {
	return &value
}
