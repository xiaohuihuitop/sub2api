//go:build unit

package service

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

// TestBuildUsageBillingCommand_SubscriptionAppliesRateMultiplier locks in the fix
// that subscription-mode billing honours the group (and any user-specific) rate
// multiplier — i.e. cmd.SubscriptionCost tracks ActualCost (= TotalCost *
// RateMultiplier), not raw TotalCost.
func TestBuildUsageBillingCommand_SubscriptionAppliesRateMultiplier(t *testing.T) {
	t.Parallel()

	groupID := int64(7)
	subID := int64(42)

	tests := []struct {
		name           string
		totalCost      float64
		actualCost     float64
		isSubscription bool
		wantSub        float64
		wantBalance    float64
	}{
		{
			name:           "subscription with 2x multiplier consumes 2x quota",
			totalCost:      1.0,
			actualCost:     2.0,
			isSubscription: true,
			wantSub:        2.0,
			wantBalance:    0,
		},
		{
			name:           "subscription with 0.5x multiplier consumes 0.5x quota",
			totalCost:      1.0,
			actualCost:     0.5,
			isSubscription: true,
			wantSub:        0.5,
			wantBalance:    0,
		},
		{
			name:           "free subscription (multiplier 0) consumes no quota",
			totalCost:      1.0,
			actualCost:     0,
			isSubscription: true,
			wantSub:        0,
			wantBalance:    0,
		},
		{
			name:           "balance billing keeps using ActualCost (regression)",
			totalCost:      1.0,
			actualCost:     2.0,
			isSubscription: false,
			wantSub:        0,
			wantBalance:    2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			p := &postUsageBillingParams{
				Cost:               &CostBreakdown{TotalCost: tt.totalCost, ActualCost: tt.actualCost},
				User:               &User{ID: 1},
				APIKey:             &APIKey{ID: 2, GroupID: &groupID},
				Account:            &Account{ID: 3},
				Subscription:       &UserSubscription{ID: subID},
				IsSubscriptionBill: tt.isSubscription,
			}

			cmd := buildUsageBillingCommand("req-1", nil, p)
			if cmd == nil {
				t.Fatal("buildUsageBillingCommand returned nil")
			}
			if cmd.SubscriptionCost != tt.wantSub {
				t.Errorf("SubscriptionCost = %v, want %v", cmd.SubscriptionCost, tt.wantSub)
			}
			if cmd.BalanceCost != tt.wantBalance {
				t.Errorf("BalanceCost = %v, want %v", cmd.BalanceCost, tt.wantBalance)
			}
		})
	}
}

func TestFinalizePostUsageBilling_UsesSubscriptionGroupForSubscriptionCache(t *testing.T) {
	t.Parallel()

	keyGroupID := int64(10)
	subscriptionGroupID := int64(20)

	cache := &billingCacheWorkerStub{}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{})
	t.Cleanup(svc.Stop)
	invalidator := &subscriptionCacheInvalidationStub{}

	finalizePostUsageBilling(&postUsageBillingParams{
		Cost:               &CostBreakdown{ActualCost: 1.25},
		User:               &User{ID: 1},
		APIKey:             &APIKey{ID: 2, GroupID: &keyGroupID},
		Subscription:       &UserSubscription{ID: 3, GroupID: subscriptionGroupID},
		IsSubscriptionBill: true,
	}, &billingDeps{
		billingCacheService: svc,
		subscriptionCache:   invalidator,
	}, &UsageBillingApplyResult{Applied: true})

	require.Equal(t, int64(1), invalidator.calls.Load())
	require.Equal(t, subscriptionGroupID, invalidator.groupID.Load())
	require.Eventually(t, func() bool {
		return cache.lastSubscriptionUpdateGroupID() == subscriptionGroupID
	}, 2*time.Second, 10*time.Millisecond)
}

func TestFinalizePostUsageBilling_SkipsSubscriptionCacheWhenBillingNotApplied(t *testing.T) {
	t.Parallel()

	keyGroupID := int64(10)
	subscriptionGroupID := int64(20)

	cache := &billingCacheWorkerStub{}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{})
	t.Cleanup(svc.Stop)
	invalidator := &subscriptionCacheInvalidationStub{}

	finalizePostUsageBilling(&postUsageBillingParams{
		Cost:               &CostBreakdown{ActualCost: 1.25},
		User:               &User{ID: 1},
		APIKey:             &APIKey{ID: 2, GroupID: &keyGroupID},
		Subscription:       &UserSubscription{ID: 3, GroupID: subscriptionGroupID},
		IsSubscriptionBill: true,
	}, &billingDeps{
		billingCacheService: svc,
		subscriptionCache:   invalidator,
	}, &UsageBillingApplyResult{Applied: false})

	require.Equal(t, int64(0), invalidator.calls.Load())
	require.Equal(t, int64(0), atomic.LoadInt64(&cache.subscriptionUpdates))
	require.Equal(t, int64(0), cache.lastSubscriptionUpdateGroupID())
}

func TestApplyUsageBillingFallback_RefreshesSubscriptionCachesAfterDBWrite(t *testing.T) {
	t.Parallel()

	keyGroupID := int64(10)
	subscriptionGroupID := int64(20)

	cache := &billingCacheWorkerStub{}
	svc := NewBillingCacheService(cache, nil, nil, nil, nil, nil, &config.Config{})
	t.Cleanup(svc.Stop)
	invalidator := &subscriptionCacheInvalidationStub{}
	subRepo := &fallbackUserSubRepoStub{}

	applied, err := applyUsageBilling(
		context.Background(),
		"req-1",
		nil,
		&postUsageBillingParams{
			Cost:               &CostBreakdown{ActualCost: 1.25},
			User:               &User{ID: 1},
			APIKey:             &APIKey{ID: 2, GroupID: &keyGroupID},
			Account:            &Account{ID: 3},
			Subscription:       &UserSubscription{ID: 4, GroupID: subscriptionGroupID},
			IsSubscriptionBill: true,
		},
		&billingDeps{
			userSubRepo:         subRepo,
			billingCacheService: svc,
			subscriptionCache:   invalidator,
		},
		nil,
	)

	require.NoError(t, err)
	require.True(t, applied)
	require.Equal(t, int64(1), subRepo.calls.Load())
	require.Equal(t, int64(1), invalidator.calls.Load())
	require.Equal(t, subscriptionGroupID, invalidator.groupID.Load())
	require.Eventually(t, func() bool {
		return cache.lastSubscriptionUpdateGroupID() == subscriptionGroupID
	}, 2*time.Second, 10*time.Millisecond)
}

type subscriptionCacheInvalidationStub struct {
	calls   atomic.Int64
	groupID atomic.Int64
}

func (s *subscriptionCacheInvalidationStub) InvalidateSubCache(userID, groupID int64) {
	s.calls.Add(1)
	s.groupID.Store(groupID)
}

func (s *subscriptionCacheInvalidationStub) InvalidateSubCacheSync(userID, groupID int64) {
	s.InvalidateSubCache(userID, groupID)
}

type fallbackUserSubRepoStub struct {
	UserSubscriptionRepository
	calls atomic.Int64
}

func (s *fallbackUserSubRepoStub) IncrementUsage(ctx context.Context, id int64, costUSD float64) error {
	s.calls.Add(1)
	return nil
}
