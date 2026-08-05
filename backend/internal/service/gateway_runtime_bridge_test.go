//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/Wei-Shaw/sub2api/internal/productcore"
	"github.com/stretchr/testify/require"
)

func TestGatewayRuntimeBridgePreservesSchedulerAndPricingFacts(t *testing.T) {
	legacyGroupID := int64(9)
	subscriptionID := int64(22)
	decision := &productcore.Decision{
		Platform: productcore.Platform{
			ID: 3, Code: "gpt", AccountPlatform: PlatformOpenAI,
			RequestedModel: "public-gpt", UpstreamModel: "gpt-4o-2024-08-06",
			LegacyPricingGroupID: &legacyGroupID,
		},
		BillingAsset: &productcore.BillingAsset{
			Source: "subscription", SubscriptionID: &subscriptionID, RateMultiplier: 0.5,
		},
	}

	ctx := attachProductDecision(context.Background(), decision, nil)
	intent, ok := gatewayruntime.DispatchIntentFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(3), intent.Platform.ID)

	legacy, ok := GatewayPlatformAssetContextFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, PlatformOpenAI, legacy.SchedulingScope.AccountPlatform)
	require.Equal(t, legacyGroupID, *legacy.PricingGroupID)
	require.Equal(t, 0.5, legacy.BillingAsset.RateMultiplier)
}

func TestGatewayRuntimeBridgeUsesExplicitPricingGroupReference(t *testing.T) {
	platformGroupID := int64(3)
	explicitPricingGroupID := int64(9)
	route := &GatewayPlatformAssetContext{
		Platform: &ResolvedPlatformModel{
			PlatformID:      3,
			AccountPlatform: PlatformOpenAI,
			LegacyGroupID:   &platformGroupID,
		},
		SchedulingScope: PlatformSchedulingScope{PlatformID: 3, AccountPlatform: PlatformOpenAI},
		PricingGroupID:  &explicitPricingGroupID,
	}

	ctx := WithGatewayPlatformAssetContext(context.Background(), route)
	intent, ok := gatewayruntime.DispatchIntentFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, explicitPricingGroupID, *intent.Platform.LegacyPricingGroupID)
}
