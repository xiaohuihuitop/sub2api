//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type platformModelResolverStub struct {
	resolved *ResolvedPlatformModel
	err      error
}

func (s platformModelResolverStub) ResolveModel(context.Context, string) (*ResolvedPlatformModel, error) {
	return s.resolved, s.err
}

func (s platformModelResolverStub) ResolveModelCandidates(context.Context, string) ([]*ResolvedPlatformModel, error) {
	if s.resolved == nil {
		return nil, s.err
	}
	return []*ResolvedPlatformModel{s.resolved}, s.err
}

func TestResolvePlatformAssetRequestRejectsUnapprovedPlatform(t *testing.T) {
	apiKey := &APIKey{
		UserID:             7,
		AllowedPlatformIDs: []int64{2},
		AllowBalance:       true,
		User:               &User{ID: 7, Balance: 10},
	}
	resolver := platformModelResolverStub{resolved: &ResolvedPlatformModel{
		PlatformID:      1,
		PlatformCode:    "gpt",
		AccountPlatform: PlatformOpenAI,
	}}

	_, err := (&APIKeyService{}).ResolvePlatformAssetRequest(
		context.Background(), apiKey, resolver, nil, "gpt-4o", "/v1/chat/completions", false,
	)

	require.ErrorIs(t, err, ErrAPIKeyPlatformForbidden)
}

func TestResolvePlatformAssetRequestHonorsEndpointAndSelectsBalance(t *testing.T) {
	apiKey := &APIKey{
		UserID:             7,
		AllowedPlatformIDs: []int64{1},
		AllowBalance:       true,
		User:               &User{ID: 7, Balance: 10},
	}
	resolver := platformModelResolverStub{resolved: &ResolvedPlatformModel{
		PlatformID:           1,
		PlatformCode:         "gpt",
		AccountPlatform:      PlatformOpenAI,
		EndpointCapabilities: []string{string(OpenAIEndpointCapabilityChatCompletions)},
	}}

	route, err := (&APIKeyService{}).ResolvePlatformAssetRequest(
		context.Background(), apiKey, resolver, nil, "gpt-4o", "/v1/chat/completions", false,
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), route.Platform.PlatformID)
	require.Equal(t, BillingSourceBalance, route.BillingAsset.Source)
	require.Equal(t, PlatformOpenAI, route.SchedulingScope.AccountPlatform)

	_, err = (&APIKeyService{}).ResolvePlatformAssetRequest(
		context.Background(), apiKey, resolver, nil, "gpt-4o", "/v1/responses", false,
	)
	require.ErrorIs(t, err, ErrPlatformEndpointUnsupported)
}

func TestGatewayPlatformAssetContextSetsSchedulingAndModelRouting(t *testing.T) {
	legacyGroupID := int64(9)
	route := &GatewayPlatformAssetContext{
		Platform: &ResolvedPlatformModel{
			PlatformID:      1,
			PlatformCode:    "gpt",
			AccountPlatform: PlatformOpenAI,
			RequestedModel:  "public-gpt",
			UpstreamModel:   "gpt-4o-2024-08-06",
			LegacyGroupID:   &legacyGroupID,
		},
		SchedulingScope: PlatformSchedulingScope{
			PlatformID:      1,
			PlatformCode:    "gpt",
			AccountPlatform: PlatformOpenAI,
		},
		PricingGroupID: &legacyGroupID,
	}

	ctx := WithGatewayPlatformAssetContext(context.Background(), route)

	got, ok := GatewayPlatformAssetContextFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(1), got.Platform.PlatformID)
	require.Equal(t, int64(9), *got.PricingGroupID)
	scope, ok := PlatformSchedulingScopeFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, int64(1), scope.PlatformID)
	platform, ok := ResolvedTargetPlatformFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, PlatformOpenAI, platform)
	upstreamModel, ok := ResolvedUpstreamModelFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "gpt-4o-2024-08-06", upstreamModel)
	publicModel, ok := RequestedPublicModelFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, "public-gpt", publicModel)
}

func TestPlatformAssetBillingFactsOverrideLegacyGroupValues(t *testing.T) {
	platformID := int64(3)
	pricingGroupID := int64(9)
	subscriptionID := int64(22)
	ctx := WithGatewayPlatformAssetContext(context.Background(), &GatewayPlatformAssetContext{
		Platform: &ResolvedPlatformModel{PlatformID: platformID, AccountPlatform: PlatformOpenAI},
		BillingAsset: &ResolvedBillingAsset{
			Source:         BillingSourceSubscription,
			SubscriptionID: &subscriptionID,
			RateMultiplier: 0.5,
		},
		SchedulingScope: PlatformSchedulingScope{PlatformID: platformID, AccountPlatform: PlatformOpenAI},
		PricingGroupID:  &pricingGroupID,
	})

	token, image, video := overridePlatformAssetBillingMultipliers(ctx, 3, 4, 5)
	require.Equal(t, 0.5, token)
	require.Equal(t, 0.5, image)
	require.Equal(t, 0.5, video)
	pricingGroup, ok := PlatformAssetPricingGroupIDFromContext(ctx)
	require.True(t, ok)
	require.Equal(t, pricingGroupID, *pricingGroup)

	usageLog := &UsageLog{}
	applyPlatformAssetUsageAttribution(ctx, usageLog)
	require.Equal(t, platformID, *usageLog.PlatformID)
	require.Equal(t, BillingSourceSubscription, *usageLog.BillingSourceType)
}

func TestEffectivePricingGroupIDDoesNotFallBackInsidePlatformRoute(t *testing.T) {
	legacyGroupID := int64(99)
	ctx := WithGatewayPlatformAssetContext(context.Background(), &GatewayPlatformAssetContext{
		Platform:        &ResolvedPlatformModel{PlatformID: 1, AccountPlatform: PlatformOpenAI},
		SchedulingScope: PlatformSchedulingScope{PlatformID: 1, AccountPlatform: PlatformOpenAI},
		PricingGroupID:  nil,
	})
	apiKey := &APIKey{GroupID: &legacyGroupID}

	require.Nil(t, effectivePricingGroupID(ctx, apiKey))
	require.Equal(t, PlatformOpenAI, effectivePricingAdapter(ctx, apiKey))
}
