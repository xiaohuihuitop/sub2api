//go:build unit

package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type fakeAPIKeySubscriptionResolver struct {
	subs          map[int64]*UserSubscription
	validateErrs  map[int64]error
	activeErrs    map[int64]error
	checkedGroups []int64
}

func (f *fakeAPIKeySubscriptionResolver) GetActiveSubscription(ctx context.Context, userID, groupID int64) (*UserSubscription, error) {
	if f.activeErrs != nil {
		if err := f.activeErrs[groupID]; err != nil {
			return nil, err
		}
	}
	if f.subs == nil {
		return nil, ErrSubscriptionNotFound
	}
	sub := f.subs[groupID]
	if sub == nil {
		return nil, ErrSubscriptionNotFound
	}
	cp := *sub
	return &cp, nil
}

func (f *fakeAPIKeySubscriptionResolver) ValidateAndCheckLimits(sub *UserSubscription, group *Group) (bool, error) {
	if group != nil {
		f.checkedGroups = append(f.checkedGroups, group.ID)
		if f.validateErrs != nil {
			if err := f.validateErrs[group.ID]; err != nil {
				return false, err
			}
		}
	}
	return false, nil
}

func (f *fakeAPIKeySubscriptionResolver) DoWindowMaintenance(sub *UserSubscription) {}

func TestAPIKeyService_ResolveBillingGroupForRequest_UsesSmallestDailyLimitSubscriptionFirst(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	low := testSubscriptionGroup(10, "low", PlatformAnthropic, 5, 2)
	high := testSubscriptionGroup(20, "high", PlatformAnthropic, 20, 1)
	apiKey := testAPIKeyWithAllowedGroups(7, high, []Group{*high, *low}, 0)
	resolver := &fakeAPIKeySubscriptionResolver{
		subs: map[int64]*UserSubscription{
			low.ID:  testActiveSubscription(7, low.ID),
			high.ID: testActiveSubscription(7, high.ID),
		},
	}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, "")

	require.NoError(t, err)
	require.NotNil(t, subscription)
	require.Equal(t, low.ID, subscription.GroupID)
	require.NotNil(t, apiKey.Group)
	require.Equal(t, low.ID, apiKey.Group.ID)
	require.NotNil(t, apiKey.GroupID)
	require.Equal(t, low.ID, *apiKey.GroupID)
	require.Equal(t, []int64{low.ID}, resolver.checkedGroups)
}

func TestAPIKeyService_ResolveBillingGroupForRequest_SkipsExceededSubscription(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	low := testSubscriptionGroup(10, "low", PlatformAnthropic, 5, 1)
	high := testSubscriptionGroup(20, "high", PlatformAnthropic, 20, 2)
	apiKey := testAPIKeyWithAllowedGroups(7, low, []Group{*low, *high}, 0)
	resolver := &fakeAPIKeySubscriptionResolver{
		subs: map[int64]*UserSubscription{
			low.ID:  testActiveSubscription(7, low.ID),
			high.ID: testActiveSubscription(7, high.ID),
		},
		validateErrs: map[int64]error{
			low.ID: ErrDailyLimitExceeded,
		},
	}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, "")

	require.NoError(t, err)
	require.NotNil(t, subscription)
	require.Equal(t, high.ID, subscription.GroupID)
	require.Equal(t, high.ID, apiKey.Group.ID)
	require.Equal(t, []int64{low.ID, high.ID}, resolver.checkedGroups)
}

func TestAPIKeyService_ResolveBillingGroupForRequest_ClearsScopedRPMOverrideWhenBillingGroupChanges(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	primary := testSubscriptionGroup(10, "primary", PlatformAnthropic, 20, 2)
	selected := testSubscriptionGroup(20, "selected", PlatformAnthropic, 5, 1)
	override := 7
	apiKey := testAPIKeyWithAllowedGroups(7, primary, []Group{*primary, *selected}, 0)
	apiKey.User.UserGroupRPMOverride = &override
	resolver := &fakeAPIKeySubscriptionResolver{
		subs: map[int64]*UserSubscription{
			primary.ID:  testActiveSubscription(7, primary.ID),
			selected.ID: testActiveSubscription(7, selected.ID),
		},
	}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, "")

	require.NoError(t, err)
	require.NotNil(t, subscription)
	require.Equal(t, selected.ID, subscription.GroupID)
	require.NotNil(t, apiKey.User)
	require.Nil(t, apiKey.User.UserGroupRPMOverride, "override must be cleared after switching to another billing group")
}

func TestAPIKeyService_ResolveBillingGroupForRequest_FallsBackToBalanceGroup(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	subGroup := testSubscriptionGroup(10, "sub", PlatformAnthropic, 5, 1)
	standardGroup := testStandardGroup(30, "balance", PlatformAnthropic, 3)
	apiKey := testAPIKeyWithAllowedGroups(7, subGroup, []Group{*subGroup, *standardGroup}, 2)
	resolver := &fakeAPIKeySubscriptionResolver{
		subs: map[int64]*UserSubscription{
			subGroup.ID: testActiveSubscription(7, subGroup.ID),
		},
		validateErrs: map[int64]error{
			subGroup.ID: ErrDailyLimitExceeded,
		},
	}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, "")

	require.NoError(t, err)
	require.Nil(t, subscription)
	require.NotNil(t, apiKey.Group)
	require.Equal(t, standardGroup.ID, apiKey.Group.ID)
	require.NotNil(t, apiKey.GroupID)
	require.Equal(t, standardGroup.ID, *apiKey.GroupID)
}

func TestAPIKeyService_ResolveBillingGroupForRequest_FiltersByTargetPlatform(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	anthropic := testSubscriptionGroup(10, "anthropic", PlatformAnthropic, 5, 1)
	gemini := testSubscriptionGroup(20, "gemini", PlatformGemini, 20, 2)
	apiKey := testAPIKeyWithAllowedGroups(7, anthropic, []Group{*anthropic, *gemini}, 0)
	resolver := &fakeAPIKeySubscriptionResolver{
		subs: map[int64]*UserSubscription{
			anthropic.ID: testActiveSubscription(7, anthropic.ID),
			gemini.ID:   testActiveSubscription(7, gemini.ID),
		},
	}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, PlatformGemini)

	require.NoError(t, err)
	require.NotNil(t, subscription)
	require.Equal(t, gemini.ID, subscription.GroupID)
	require.Equal(t, gemini.ID, apiKey.Group.ID)
	require.Equal(t, []int64{gemini.ID}, resolver.checkedGroups)
}

func TestAPIKeyService_ResolveBillingGroupForRequest_ReturnsLastSubscriptionLimitErrorWhenNoFallback(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	subGroup := testSubscriptionGroup(10, "sub", PlatformAnthropic, 5, 1)
	apiKey := testAPIKeyWithAllowedGroups(7, subGroup, []Group{*subGroup}, 0)
	resolver := &fakeAPIKeySubscriptionResolver{
		subs: map[int64]*UserSubscription{
			subGroup.ID: testActiveSubscription(7, subGroup.ID),
		},
		validateErrs: map[int64]error{
			subGroup.ID: ErrDailyLimitExceeded,
		},
	}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, "")

	require.Nil(t, subscription)
	require.ErrorIs(t, err, ErrDailyLimitExceeded)
}

func testAPIKeyWithAllowedGroups(userID int64, primary *Group, allowed []Group, balance float64) *APIKey {
	var groupID *int64
	if primary != nil {
		id := primary.ID
		groupID = &id
	}
	allowedIDs := make([]int64, 0, len(allowed))
	for _, group := range allowed {
		allowedIDs = append(allowedIDs, group.ID)
	}
	return &APIKey{
		ID:              1,
		UserID:          userID,
		GroupID:         groupID,
		Group:           primary,
		AllowedGroupIDs: allowedIDs,
		AllowedGroups:   allowed,
		User: &User{
			ID:      userID,
			Status:  StatusActive,
			Role:    RoleUser,
			Balance: balance,
		},
		Status: StatusActive,
	}
}

func testSubscriptionGroup(id int64, name, platform string, dailyLimit float64, sortOrder int) *Group {
	return &Group{
		ID:               id,
		Name:             name,
		Platform:         platform,
		Status:           StatusActive,
		Hydrated:         true,
		SubscriptionType: SubscriptionTypeSubscription,
		DailyLimitUSD:    &dailyLimit,
		RateMultiplier:   1,
		SortOrder:        sortOrder,
	}
}

func testStandardGroup(id int64, name, platform string, sortOrder int) *Group {
	return &Group{
		ID:               id,
		Name:             name,
		Platform:         platform,
		Status:           StatusActive,
		Hydrated:         true,
		SubscriptionType: SubscriptionTypeStandard,
		RateMultiplier:   1,
		SortOrder:        sortOrder,
	}
}

func testActiveSubscription(userID, groupID int64) *UserSubscription {
	return &UserSubscription{
		ID:        groupID * 100,
		UserID:    userID,
		GroupID:   groupID,
		Status:    SubscriptionStatusActive,
		StartsAt:  time.Now().Add(-time.Hour),
		ExpiresAt: time.Now().Add(time.Hour),
	}
}

var _ apiKeySubscriptionResolver = (*fakeAPIKeySubscriptionResolver)(nil)

func TestAPIKeyService_ResolveBillingGroupForRequest_SkipsMissingSubscriptionAndUsesNext(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	missing := testSubscriptionGroup(10, "missing", PlatformAnthropic, 5, 1)
	next := testSubscriptionGroup(20, "next", PlatformAnthropic, 10, 2)
	apiKey := testAPIKeyWithAllowedGroups(7, missing, []Group{*missing, *next}, 0)
	resolver := &fakeAPIKeySubscriptionResolver{
		subs: map[int64]*UserSubscription{
			next.ID: testActiveSubscription(7, next.ID),
		},
		activeErrs: map[int64]error{
			missing.ID: ErrSubscriptionNotFound,
		},
	}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, "")

	require.NoError(t, err)
	require.NotNil(t, subscription)
	require.Equal(t, next.ID, subscription.GroupID)
}

func TestAPIKeyService_ResolveBillingGroupForRequest_ReturnsNoUsableGroupForPlatformMismatch(t *testing.T) {
	svc := NewAPIKeyService(nil, nil, nil, nil, nil, nil, nil)
	anthropic := testStandardGroup(10, "anthropic", PlatformAnthropic, 1)
	apiKey := testAPIKeyWithAllowedGroups(7, anthropic, []Group{*anthropic}, 10)
	resolver := &fakeAPIKeySubscriptionResolver{}

	subscription, err := svc.ResolveBillingGroupForRequest(context.Background(), apiKey, resolver, false, PlatformGemini)

	require.Nil(t, subscription)
	require.True(t, errors.Is(err, ErrNoUsableBillingGroup))
}
