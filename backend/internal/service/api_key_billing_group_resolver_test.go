//go:build unit

package service

import (
	"context"
	"testing"
	"time"

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
