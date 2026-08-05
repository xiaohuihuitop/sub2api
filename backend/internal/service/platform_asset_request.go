package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/ctxkey"
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

// UsesPlatformAssetPermissions distinguishes explicitly migrated API keys from
// legacy keys. A key without platform IDs must remain on the legacy path until
// an administrator grants its new permissions explicitly.
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
	if resolver == nil {
		return nil, fmt.Errorf("%w: platform model resolver is required", ErrPlatformInvalid)
	}

	platform, err := resolver.ResolveModel(ctx, requestedModel)
	if err != nil {
		return nil, err
	}
	if platform == nil || !apiKeyAllowsPlatform(apiKey, platform.PlatformID) {
		return nil, ErrAPIKeyPlatformForbidden
	}
	if !platformSupportsRequestEndpoint(platform, endpoint) {
		return nil, ErrPlatformEndpointUnsupported
	}

	asset, err := s.ResolveBillingAssetForRequest(ctx, apiKey, subscriptions, skipBilling)
	if err != nil {
		return nil, err
	}
	scope := PlatformSchedulingScope{
		PlatformID:      platform.PlatformID,
		PlatformCode:    platform.PlatformCode,
		AccountPlatform: platform.AccountPlatform,
	}
	if _, ok := normalizePlatformSchedulingScope(scope); !ok {
		return nil, fmt.Errorf("%w: resolved platform has no account adapter", ErrPlatformInvalid)
	}

	return &GatewayPlatformAssetContext{
		Platform:        cloneResolvedPlatformModel(platform),
		BillingAsset:    cloneResolvedBillingAsset(asset),
		SchedulingScope: scope,
		PricingGroupID:  clonePlatformInt64Pointer(platform.LegacyGroupID),
	}, nil
}

func apiKeyAllowsPlatform(apiKey *APIKey, platformID int64) bool {
	for _, allowedID := range apiKey.AllowedPlatformIDs {
		if allowedID == platformID {
			return true
		}
	}
	return false
}

func platformSupportsRequestEndpoint(platform *ResolvedPlatformModel, endpoint string) bool {
	capability := billingEndpointCapability(endpoint)
	if capability == "" || platform == nil || len(platform.EndpointCapabilities) == 0 {
		return true
	}
	for _, configured := range platform.EndpointCapabilities {
		if strings.EqualFold(strings.TrimSpace(configured), string(capability)) {
			return true
		}
	}
	return false
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

// WithGatewayPlatformAssetContext installs the V2 request route and the
// related compatibility values. The platform scheduling scope remains the
// authority for account selection; a pricing group, when present, is only
// available to legacy model-pricing readers.
func WithGatewayPlatformAssetContext(ctx context.Context, route *GatewayPlatformAssetContext) context.Context {
	if ctx == nil || route == nil || route.Platform == nil {
		return ctx
	}
	scope, ok := normalizePlatformSchedulingScope(route.SchedulingScope)
	if !ok {
		return ctx
	}
	cloned := &GatewayPlatformAssetContext{
		Platform:        cloneResolvedPlatformModel(route.Platform),
		BillingAsset:    cloneResolvedBillingAsset(route.BillingAsset),
		SchedulingScope: scope,
		PricingGroupID:  clonePlatformInt64Pointer(route.PricingGroupID),
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

// GatewayPlatformAssetContextFromContext returns an isolated copy so callers
// cannot mutate values shared across a request's downstream stages.
func GatewayPlatformAssetContextFromContext(ctx context.Context) (*GatewayPlatformAssetContext, bool) {
	if ctx == nil {
		return nil, false
	}
	route, ok := ctx.Value(ctxkey.GatewayPlatformAsset).(*GatewayPlatformAssetContext)
	if !ok || route == nil || route.Platform == nil {
		return nil, false
	}
	cloned := &GatewayPlatformAssetContext{
		Platform:        cloneResolvedPlatformModel(route.Platform),
		BillingAsset:    cloneResolvedBillingAsset(route.BillingAsset),
		SchedulingScope: route.SchedulingScope,
		PricingGroupID:  clonePlatformInt64Pointer(route.PricingGroupID),
	}
	return cloned, true
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
	if groupID, ok := PlatformAssetPricingGroupIDFromContext(ctx); ok {
		return groupID
	}
	if apiKey == nil {
		return nil
	}
	return clonePlatformInt64Pointer(apiKey.GroupID)
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
