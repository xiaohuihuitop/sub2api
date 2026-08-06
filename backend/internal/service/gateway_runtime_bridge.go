package service

import (
	"context"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/gatewayruntime"
	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
	"github.com/Wei-Shaw/sub2api/internal/productcore"
)

func gatewayPlatformAssetContextFromDecision(
	decision *productcore.Decision,
	subscription *UserSubscription,
) *GatewayPlatformAssetContext {
	if decision == nil {
		return nil
	}
	platform := decision.Platform
	return &GatewayPlatformAssetContext{
		Platform: &ResolvedPlatformModel{
			PlatformID:           platform.ID,
			PlatformCode:         platform.Code,
			AccountPlatform:      platform.AccountPlatform,
			RequestedModel:       platform.RequestedModel,
			UpstreamModel:        platform.UpstreamModel,
			EndpointCapabilities: append([]string(nil), platform.EndpointCapabilities...),
		},
		BillingAsset: resolvedBillingAssetFromProduct(decision.BillingAsset, subscription),
		SchedulingScope: PlatformSchedulingScope{
			PlatformID:      platform.ID,
			PlatformCode:    platform.Code,
			AccountPlatform: platform.AccountPlatform,
		},
	}
}

func resolvedBillingAssetFromProduct(
	asset *productcore.BillingAsset,
	subscription *UserSubscription,
) *ResolvedBillingAsset {
	if asset == nil {
		return nil
	}
	return &ResolvedBillingAsset{
		Source:         asset.Source,
		SubscriptionID: clonePlatformInt64Pointer(asset.SubscriptionID),
		PlanID:         clonePlatformInt64Pointer(asset.PlanID),
		RateMultiplier: asset.RateMultiplier,
		Subscription:   cloneUserSubscription(subscription),
	}
}

func productBillingAssetFromResolved(asset *ResolvedBillingAsset) *productcore.BillingAsset {
	if asset == nil {
		return nil
	}
	return &productcore.BillingAsset{
		Source:         asset.Source,
		SubscriptionID: clonePlatformInt64Pointer(asset.SubscriptionID),
		PlanID:         clonePlatformInt64Pointer(asset.PlanID),
		RateMultiplier: asset.RateMultiplier,
	}
}

func dispatchIntentFromGatewayRoute(route *GatewayPlatformAssetContext) *gatewayruntime.DispatchIntent {
	if route == nil || route.Platform == nil {
		return nil
	}
	platform := productPlatformFromResolved(route.Platform)
	if platform == nil {
		return nil
	}
	if route.PricingGroupID != nil {
		platform.LegacyPricingGroupID = clonePlatformInt64Pointer(route.PricingGroupID)
	}
	return &gatewayruntime.DispatchIntent{
		Platform:     *platform,
		BillingAsset: productBillingAssetFromResolved(route.BillingAsset),
	}
}

func cloneGatewayPlatformAssetContext(route *GatewayPlatformAssetContext) *GatewayPlatformAssetContext {
	if route == nil {
		return nil
	}
	return &GatewayPlatformAssetContext{
		Platform:        cloneResolvedPlatformModel(route.Platform),
		BillingAsset:    cloneResolvedBillingAsset(route.BillingAsset),
		SchedulingScope: route.SchedulingScope,
		PricingGroupID:  clonePlatformInt64Pointer(route.PricingGroupID),
	}
}

func attachGatewayPlatformAssetRoute(ctx context.Context, route *GatewayPlatformAssetContext) context.Context {
	if ctx == nil || route == nil || route.Platform == nil {
		return ctx
	}
	scope, ok := normalizePlatformSchedulingScope(route.SchedulingScope)
	if !ok {
		return ctx
	}
	cloned := cloneGatewayPlatformAssetContext(route)
	if intent := dispatchIntentFromGatewayRoute(cloned); intent != nil {
		ctx = gatewayruntime.WithDispatchIntent(ctx, intent)
	}
	ctx = context.WithValue(ctx, ctxkey.GatewayPlatformAsset, cloned)
	ctx = WithPlatformSchedulingScope(ctx, scope)
	ctx = WithResolvedTargetPlatform(ctx, scope.AccountPlatform)
	if model := strings.TrimSpace(cloned.Platform.UpstreamModel); model != "" {
		ctx = context.WithValue(ctx, ctxkey.ResolvedUpstreamModel, model)
	}
	if model := strings.TrimSpace(cloned.Platform.RequestedModel); model != "" {
		ctx = context.WithValue(ctx, ctxkey.RequestedPublicModel, model)
	}
	return ctx
}

func attachProductDecision(
	ctx context.Context,
	decision *productcore.Decision,
	subscription *UserSubscription,
) context.Context {
	return attachGatewayPlatformAssetRoute(ctx, gatewayPlatformAssetContextFromDecision(decision, subscription))
}

func AttachPlatformAssetResolution(ctx context.Context, resolution *PlatformAssetResolution) context.Context {
	if resolution == nil {
		return ctx
	}
	return attachProductDecision(ctx, resolution.Decision, resolution.Subscription)
}

// WithGatewayPlatformAssetContext installs the V2 request route and both the
// new runtime intent and the compatibility context values.
func WithGatewayPlatformAssetContext(ctx context.Context, route *GatewayPlatformAssetContext) context.Context {
	return attachGatewayPlatformAssetRoute(ctx, route)
}

// GatewayPlatformAssetContextFromContext returns an isolated copy. The legacy
// context remains preferred so the subscription object is preserved for current
// usage and quota accounting; the new intent is a fallback for future readers.
func GatewayPlatformAssetContextFromContext(ctx context.Context) (*GatewayPlatformAssetContext, bool) {
	if ctx == nil {
		return nil, false
	}
	if route, ok := ctx.Value(ctxkey.GatewayPlatformAsset).(*GatewayPlatformAssetContext); ok && route != nil && route.Platform != nil {
		return cloneGatewayPlatformAssetContext(route), true
	}
	intent, ok := gatewayruntime.DispatchIntentFromContext(ctx)
	if !ok {
		return nil, false
	}
	route := gatewayPlatformAssetContextFromDecision(&productcore.Decision{
		Platform:     intent.Platform,
		BillingAsset: intent.BillingAsset,
	}, nil)
	if route == nil {
		return nil, false
	}
	return route, true
}
