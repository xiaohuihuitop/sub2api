package productcore

import (
	"context"
	"strings"
)

type Authorizer struct {
	platforms PlatformCatalog
	assets    AssetSelector
}

func NewAuthorizer(platforms PlatformCatalog, assets AssetSelector) *Authorizer {
	return &Authorizer{platforms: platforms, assets: assets}
}

func (a *Authorizer) Resolve(ctx context.Context, grant AccessGrant, request Request) (*Decision, error) {
	if a == nil || a.platforms == nil {
		return nil, ErrModelUnavailable
	}
	platform, err := a.platforms.ResolveModel(ctx, request.Model)
	if err != nil {
		return nil, err
	}
	if platform == nil || !allowsPlatform(grant.PlatformIDs, platform.ID) {
		return nil, ErrPlatformForbidden
	}
	if !supportsEndpoint(platform.EndpointCapabilities, request.EndpointCapability) {
		return nil, ErrEndpointUnsupported
	}
	if a.assets == nil {
		return nil, ErrNoBillingAsset
	}
	asset, err := a.assets.Select(ctx, grant, request.SkipBilling)
	if err != nil {
		return nil, err
	}
	if asset == nil && !request.SkipBilling {
		return nil, ErrNoBillingAsset
	}
	return &Decision{Platform: clonePlatform(*platform), BillingAsset: cloneBillingAsset(asset)}, nil
}

func allowsPlatform(allowed []int64, platformID int64) bool {
	for _, candidate := range allowed {
		if candidate == platformID {
			return true
		}
	}
	return false
}

func supportsEndpoint(configured []string, requested string) bool {
	requested = strings.TrimSpace(requested)
	if requested == "" || len(configured) == 0 {
		return true
	}
	for _, capability := range configured {
		if strings.EqualFold(strings.TrimSpace(capability), requested) {
			return true
		}
	}
	return false
}

func clonePlatform(value Platform) Platform {
	value.EndpointCapabilities = append([]string(nil), value.EndpointCapabilities...)
	value.LegacyPricingGroupID = cloneInt64(value.LegacyPricingGroupID)
	return value
}

func cloneBillingAsset(value *BillingAsset) *BillingAsset {
	if value == nil {
		return nil
	}
	cloned := *value
	cloned.SubscriptionID = cloneInt64(value.SubscriptionID)
	cloned.PlanID = cloneInt64(value.PlanID)
	return &cloned
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
