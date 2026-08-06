package service

import (
	"context"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var (
	ErrAPIKeyPlatformForbidden     = infraerrors.Forbidden("API_KEY_PLATFORM_FORBIDDEN", "api key is not authorized for the resolved platform")
	ErrPlatformEndpointUnsupported = infraerrors.Forbidden("PLATFORM_ENDPOINT_UNSUPPORTED", "the resolved platform does not support this endpoint")
)

// GatewayPlatformAssetContext carries the explicit V2 route for one request.
// PricingGroupID is a read-only compatibility reference only; schedulers must
// use SchedulingScope rather than this legacy group ID.
type GatewayPlatformAssetContext struct {
	Platform        *ResolvedPlatformModel
	BillingAsset    *ResolvedBillingAsset
	SchedulingScope PlatformSchedulingScope
	PricingGroupID  *int64
}

// UsesPlatformAssetPermissions reports whether an API Key has at least one
// explicit platform grant. New model requests require this grant; the result
// is not a signal to fall back to legacy group routing.
func UsesPlatformAssetPermissions(apiKey *APIKey) bool {
	return apiKey != nil && len(apiKey.AllowedPlatformIDs) > 0
}

// ResolvePlatformAssetRequest resolves a V2 request in the required order:
// model -> authorized platform -> endpoint capability -> billing asset.
func (s *APIKeyService) ResolvePlatformAssetRequest(
	ctx context.Context,
	apiKey *APIKey,
	resolver PlatformModelResolver,
	subscriptions apiKeySubscriptionResolver,
	requestedModel, endpoint string,
	skipBilling bool,
) (*GatewayPlatformAssetContext, error) {
	if apiKey == nil || !UsesPlatformAssetPermissions(apiKey) {
		return nil, ErrAPIKeyPlatformForbidden
	}
	resolution, err := NewPlatformAssetProductCoreAdapter(s, subscriptions, resolver).
		Resolve(ctx, apiKey, requestedModel, endpoint, skipBilling)
	if err != nil {
		return nil, err
	}
	return gatewayPlatformAssetContextFromDecision(resolution.Decision, resolution.Subscription), nil
}

func cloneResolvedPlatformModel(value *ResolvedPlatformModel) *ResolvedPlatformModel {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.EndpointCapabilities = append([]string(nil), value.EndpointCapabilities...)
	cloned.LegacyGroupID = clonePlatformInt64Pointer(value.LegacyGroupID)
	return &cloned
}

func cloneResolvedBillingAsset(value *ResolvedBillingAsset) *ResolvedBillingAsset {
	if value == nil {
		return nil
	}
	cloned := *value
	if value.SubscriptionID != nil {
		subscriptionID := *value.SubscriptionID
		cloned.SubscriptionID = &subscriptionID
	}
	if value.PlanID != nil {
		planID := *value.PlanID
		cloned.PlanID = &planID
	}
	cloned.Subscription = cloneUserSubscription(value.Subscription)
	return &cloned
}

// PlatformAssetPricingGroupIDFromContext exposes the optional legacy pricing
// reference without making it part of account selection or billing eligibility.
func PlatformAssetPricingGroupIDFromContext(ctx context.Context) (*int64, bool) {
	route, ok := GatewayPlatformAssetContextFromContext(ctx)
	if !ok || route.PricingGroupID == nil {
		return nil, false
	}
	return clonePlatformInt64Pointer(route.PricingGroupID), true
}

func effectivePricingGroupID(ctx context.Context, apiKey *APIKey) *int64 {
	if _, ok := GatewayPlatformAssetContextFromContext(ctx); ok {
		// A V2 route is priced by its resolved adapter. Never fall back to the
		// API key's historical GroupID after the route has been resolved.
		if groupID, ok := PlatformAssetPricingGroupIDFromContext(ctx); ok {
			return groupID
		}
		return nil
	}
	if groupID, ok := PlatformAssetPricingGroupIDFromContext(ctx); ok {
		return groupID
	}
	if apiKey == nil {
		return nil
	}
	return clonePlatformInt64Pointer(apiKey.GroupID)
}

func effectivePricingAdapter(ctx context.Context, apiKey *APIKey) string {
	if route, ok := GatewayPlatformAssetContextFromContext(ctx); ok && route.Platform != nil {
		if adapter := strings.TrimSpace(route.Platform.AccountPlatform); adapter != "" {
			return adapter
		}
	}
	if platform, ok := ResolvedTargetPlatformFromContext(ctx); ok {
		return strings.TrimSpace(platform)
	}
	// Legacy API keys continue to resolve pricing through GroupID. Their group
	// platform is a routing hint only and must not silently become a new
	// independent pricing selector.
	return ""
}

func pricingInputForRequest(ctx context.Context, apiKey *APIKey, model string) PricingInput {
	input := PricingInput{
		Model:   model,
		Adapter: effectivePricingAdapter(ctx, apiKey),
		GroupID: effectivePricingGroupID(ctx, apiKey),
	}
	if input.Adapter != "" {
		input.GroupID = nil
	}
	return input
}

func overridePlatformAssetBillingMultipliers(ctx context.Context, token, image, video float64) (float64, float64, float64) {
	route, ok := GatewayPlatformAssetContextFromContext(ctx)
	if !ok || route.BillingAsset == nil {
		return token, image, video
	}
	multiplier := nonNegativeMultiplier(route.BillingAsset.RateMultiplier)
	return multiplier, multiplier, multiplier
}

func applyPlatformAssetUsageAttribution(ctx context.Context, usageLog *UsageLog) {
	if usageLog == nil {
		return
	}
	route, ok := GatewayPlatformAssetContextFromContext(ctx)
	if !ok || route.Platform == nil {
		return
	}
	platformID := route.Platform.PlatformID
	usageLog.PlatformID = &platformID
	if route.BillingAsset == nil || route.BillingAsset.Source == "" {
		return
	}
	source := route.BillingAsset.Source
	usageLog.BillingSourceType = &source
}
