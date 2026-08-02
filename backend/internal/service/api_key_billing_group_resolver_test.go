//go:build unit

package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type billingGroupSubscriptionResolverStub struct {
	subs         map[int64]*UserSubscription
	validateErrs map[int64]error
	checked      []int64
}

func (s *billingGroupSubscriptionResolverStub) GetActiveSubscription(_ context.Context, userID, groupID int64) (*UserSubscription, error) {
	sub := s.subs[groupID]
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	copy := *sub
	return &copy, nil
}

func (s *billingGroupSubscriptionResolverStub) ValidateAndCheckLimits(_ *UserSubscription, group *Group) (bool, error) {
	s.checked = append(s.checked, group.ID)
	return false, s.validateErrs[group.ID]
}

func (s *billingGroupSubscriptionResolverStub) EnsureWindowMaintenance(_ context.Context, sub *UserSubscription) (*UserSubscription, error) {
	return sub, nil
}

func TestResolveBillingGroupUsesAdminOrder(t *testing.T) {
	first := billingTestGroup(10, SubscriptionTypeStandard, 1, PlatformAnthropic)
	first.SubscriptionType = SubscriptionTypeSubscription
	second := billingTestGroup(20, SubscriptionTypeSubscription, 2, PlatformAnthropic)
	apiKey := billingTestAPIKey(second, []Group{second, first}, 10)
	resolver := &billingGroupSubscriptionResolverStub{subs: map[int64]*UserSubscription{
		10: billingTestSubscription(10), 20: billingTestSubscription(20),
	}}

	sub, err := (&APIKeyService{}).ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, PlatformAnthropic, "/v1/messages")

	require.NoError(t, err)
	require.Equal(t, int64(10), sub.GroupID)
	require.Equal(t, int64(10), *apiKey.GroupID)
	require.Equal(t, []int64{10}, resolver.checked)
}

func TestResolveBillingGroupSkipsLimitedPlanThenUsesBalance(t *testing.T) {
	plan := billingTestGroup(10, SubscriptionTypeSubscription, 1, PlatformOpenAI)
	balance := billingTestGroup(20, SubscriptionTypeStandard, 2, PlatformOpenAI)
	apiKey := billingTestAPIKey(plan, []Group{plan, balance}, 5)
	resolver := &billingGroupSubscriptionResolverStub{
		subs:         map[int64]*UserSubscription{10: billingTestSubscription(10)},
		validateErrs: map[int64]error{10: ErrDailyLimitExceeded},
	}

	sub, err := (&APIKeyService{}).ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, PlatformOpenAI, "/v1/responses")

	require.NoError(t, err)
	require.Nil(t, sub)
	require.Equal(t, int64(20), *apiKey.GroupID)
}

func TestResolveBillingGroupFiltersPlatformAndEndpoint(t *testing.T) {
	chat := billingTestGroup(10, SubscriptionTypeSubscription, 1, PlatformOpenAI)
	chat.OpenAIEndpointCapabilities = map[string]bool{string(OpenAIEndpointCapabilityChatCompletions): true}
	responses := billingTestGroup(20, SubscriptionTypeSubscription, 2, PlatformOpenAI)
	responses.OpenAIEndpointCapabilities = map[string]bool{string(OpenAIEndpointCapabilityResponses): true}
	gemini := billingTestGroup(30, SubscriptionTypeSubscription, 0, PlatformGemini)
	apiKey := billingTestAPIKey(chat, []Group{gemini, chat, responses}, 0)
	resolver := &billingGroupSubscriptionResolverStub{subs: map[int64]*UserSubscription{
		10: billingTestSubscription(10), 20: billingTestSubscription(20), 30: billingTestSubscription(30),
	}}

	sub, err := (&APIKeyService{}).ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, PlatformOpenAI, "/v1/responses")

	require.NoError(t, err)
	require.Equal(t, int64(20), sub.GroupID)
	require.Equal(t, []int64{20}, resolver.checked)
}

func TestResolveBillingGroupRejectsIncompatibleEndpoint(t *testing.T) {
	chat := billingTestGroup(10, SubscriptionTypeSubscription, 1, PlatformOpenAI)
	chat.OpenAIEndpointCapabilities = map[string]bool{string(OpenAIEndpointCapabilityChatCompletions): true}
	apiKey := billingTestAPIKey(chat, []Group{chat}, 0)

	_, err := (&APIKeyService{}).ResolveBillingGroupForRequest(context.Background(), apiKey, &billingGroupSubscriptionResolverStub{}, false, PlatformOpenAI, "/v1/responses")

	require.ErrorIs(t, err, ErrNoUsableBillingGroup)
}

func TestResolveBillingGroupRejectsEndpointWhenCapabilityProbeFoundNoSupport(t *testing.T) {
	unsupported := billingTestGroup(10, SubscriptionTypeSubscription, 1, PlatformOpenAI)
	unsupported.OpenAIEndpointCapabilities = map[string]bool{}
	apiKey := billingTestAPIKey(unsupported, []Group{unsupported}, 0)

	_, err := (&APIKeyService{}).ResolveBillingGroupForRequest(
		context.Background(),
		apiKey,
		&billingGroupSubscriptionResolverStub{},
		false,
		PlatformOpenAI,
		"/v1/responses",
	)

	require.ErrorIs(t, err, ErrNoUsableBillingGroup)
}

func TestResolveBillingGroupAllowsUnboundKeyToUseBalance(t *testing.T) {
	apiKey := &APIKey{ID: 1, UserID: 7, User: &User{ID: 7, Balance: 5}}

	sub, err := (&APIKeyService{}).ResolveBillingGroupForRequest(context.Background(), apiKey, nil, false, "", "/v1/messages")

	require.NoError(t, err)
	require.Nil(t, sub)
	require.Nil(t, apiKey.GroupID)
}

func TestResolveBillingGroupUsesSecondPlanAfterSubscriptionAuthCacheInvalidation(t *testing.T) {
	const (
		userID        = int64(7)
		firstGroupID  = int64(10)
		secondGroupID = int64(20)
	)
	limit := 10.0
	first := billingTestGroup(firstGroupID, SubscriptionTypeSubscription, 1, PlatformOpenAI)
	first.DailyLimitUSD = &limit
	second := billingTestGroup(secondGroupID, SubscriptionTypeSubscription, 2, PlatformOpenAI)
	second.DailyLimitUSD = &limit
	apiKey := billingTestAPIKey(first, []Group{first, second}, 0)

	subRepo := &subscriptionAuthCacheUserSubRepoStub{newSubscriptionUserSubRepoStub()}
	now := time.Now()
	windowStart := now.Add(-time.Minute)
	subRepo.seed(&UserSubscription{
		ID: firstGroupID, UserID: userID, GroupID: firstGroupID, Status: SubscriptionStatusActive,
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), DailyWindowStart: &windowStart, DailyUsageUSD: 5,
	})
	subRepo.seed(&UserSubscription{
		ID: secondGroupID, UserID: userID, GroupID: secondGroupID, Status: SubscriptionStatusActive,
		StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour), DailyWindowStart: &windowStart, DailyUsageUSD: 0,
	})
	billingCacheService := &BillingCacheService{}
	subscriptions := NewSubscriptionService(nil, subRepo, billingCacheService, nil, &config.Config{
		SubscriptionCache: config.SubscriptionCacheConfig{L1Size: 1024, L1TTLSeconds: 60},
	})
	t.Cleanup(subscriptions.subCacheL1.Close)

	stale, err := subRepo.GetActiveByUserIDAndGroupID(context.Background(), userID, firstGroupID)
	require.NoError(t, err)
	require.True(t, subscriptions.subCacheL1.SetWithTTL(subCacheKey(userID, firstGroupID), stale, 1, time.Minute))
	subscriptions.subCacheL1.Wait()
	cachedRaw, cached := subscriptions.subCacheL1.Get(subCacheKey(userID, firstGroupID))
	require.True(t, cached)
	require.Equal(t, 5.0, cachedRaw.(*UserSubscription).DailyUsageUSD)
	subRepo.byID[firstGroupID].DailyUsageUSD = limit
	subRepo.byUserGroup[subRepo.key(userID, firstGroupID)].DailyUsageUSD = limit

	selected, err := (&APIKeyService{}).ResolveBillingGroupForRequest(
		context.Background(), apiKey, subscriptions, false, PlatformOpenAI, "/v1/chat/completions",
	)
	require.NoError(t, err)
	require.Equal(t, firstGroupID, selected.GroupID, "test setup must preserve the stale L1 snapshot")

	billingCacheService.QueueUpdateSubscriptionUsage(userID, firstGroupID, 5)
	_, cached = subscriptions.subCacheL1.Get(subCacheKey(userID, firstGroupID))
	require.False(t, cached)
	refreshed, err := subRepo.GetActiveByUserIDAndGroupID(context.Background(), userID, firstGroupID)
	require.NoError(t, err)
	require.Equal(t, limit, refreshed.DailyUsageUSD)

	selected, err = (&APIKeyService{}).ResolveBillingGroupForRequest(
		context.Background(), apiKey, subscriptions, false, PlatformOpenAI, "/v1/chat/completions",
	)
	require.NoError(t, err)
	require.Equal(t, secondGroupID, selected.GroupID)
	require.Equal(t, secondGroupID, *apiKey.GroupID)
}

type subscriptionAuthCacheUserSubRepoStub struct {
	*subscriptionUserSubRepoStub
}

func (s *subscriptionAuthCacheUserSubRepoStub) GetActiveByUserIDAndGroupID(
	ctx context.Context,
	userID, groupID int64,
) (*UserSubscription, error) {
	subscription, err := s.GetByUserIDAndGroupID(ctx, userID, groupID)
	if err != nil || !subscription.IsActive() {
		return nil, ErrSubscriptionNotFound
	}
	return subscription, nil
}

func billingTestGroup(id int64, subscriptionType string, order int, platform string) Group {
	return Group{ID: id, SubscriptionType: subscriptionType, SortOrder: order, Platform: platform, Status: StatusActive}
}

func billingTestAPIKey(primary Group, groups []Group, balance float64) *APIKey {
	primaryCopy := primary
	return &APIKey{
		ID: 1, UserID: 7, GroupID: &primaryCopy.ID, Group: &primaryCopy,
		AllowedGroupIDs: apiKeyGroupIDs(groups), AllowedGroups: groups,
		User: &User{ID: 7, Balance: balance},
	}
}

func billingTestSubscription(groupID int64) *UserSubscription {
	now := time.Now()
	return &UserSubscription{ID: groupID, UserID: 7, GroupID: groupID, Status: StatusActive, StartsAt: now.Add(-time.Hour), ExpiresAt: now.Add(time.Hour)}
}
