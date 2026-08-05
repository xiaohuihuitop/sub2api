package gatewayruntime

import (
	"context"

	"github.com/Wei-Shaw/sub2api/internal/productcore"
)

type DispatchIntent struct {
	Platform     productcore.Platform
	BillingAsset *productcore.BillingAsset
}

type dispatchIntentContextKey struct{}

func WithDispatchIntent(ctx context.Context, intent *DispatchIntent) context.Context {
	if ctx == nil || intent == nil {
		return ctx
	}
	return context.WithValue(ctx, dispatchIntentContextKey{}, cloneIntent(intent))
}

func DispatchIntentFromContext(ctx context.Context) (*DispatchIntent, bool) {
	if ctx == nil {
		return nil, false
	}
	intent, ok := ctx.Value(dispatchIntentContextKey{}).(*DispatchIntent)
	if !ok || intent == nil {
		return nil, false
	}
	return cloneIntent(intent), true
}

func cloneIntent(intent *DispatchIntent) *DispatchIntent {
	if intent == nil {
		return nil
	}
	return &DispatchIntent{
		Platform: productcore.Platform{
			ID:                   intent.Platform.ID,
			Code:                 intent.Platform.Code,
			AccountPlatform:      intent.Platform.AccountPlatform,
			RequestedModel:       intent.Platform.RequestedModel,
			UpstreamModel:        intent.Platform.UpstreamModel,
			EndpointCapabilities: append([]string(nil), intent.Platform.EndpointCapabilities...),
			LegacyPricingGroupID: cloneInt64(intent.Platform.LegacyPricingGroupID),
		},
		BillingAsset: cloneBillingAsset(intent.BillingAsset),
	}
}

func cloneBillingAsset(asset *productcore.BillingAsset) *productcore.BillingAsset {
	if asset == nil {
		return nil
	}
	cloned := *asset
	cloned.SubscriptionID = cloneInt64(asset.SubscriptionID)
	cloned.PlanID = cloneInt64(asset.PlanID)
	return &cloned
}

func cloneInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
