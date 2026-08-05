package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type subscriptionCandidatesRepoStub struct {
	userSubRepoNoop
	candidates []UserSubscription
	calls      int
}

type subscriptionPlanCandidatesRepoStub struct {
	userSubRepoNoop
	candidates       []UserSubscription
	requestedPlanIDs []int64
	calls            int
}

func (r *subscriptionPlanCandidatesRepoStub) ListActiveByUserIDAndPlanIDs(
	_ context.Context,
	_ int64,
	planIDs []int64,
) ([]UserSubscription, error) {
	r.calls++
	r.requestedPlanIDs = append([]int64(nil), planIDs...)
	return append([]UserSubscription(nil), r.candidates...), nil
}

type subscriptionInvalidationPubSubStub struct {
	billingCacheWorkerStub
	publishedKeys     []string
	invalidatedSubIDs []int64
}

func (s *subscriptionInvalidationPubSubStub) InvalidateSubscriptionCache(
	_ context.Context,
	_ int64,
	subscriptionID int64,
) error {
	s.invalidatedSubIDs = append(s.invalidatedSubIDs, subscriptionID)
	return nil
}

func (s *subscriptionInvalidationPubSubStub) PublishSubscriptionCacheInvalidation(
	_ context.Context,
	cacheKey string,
) error {
	s.publishedKeys = append(s.publishedKeys, cacheKey)
	return nil
}

func (s *subscriptionInvalidationPubSubStub) SubscribeSubscriptionCacheInvalidation(
	context.Context,
	func(string),
) error {
	return nil
}

func (r *subscriptionCandidatesRepoStub) ListActiveByUserIDAndGroupID(
	_ context.Context,
	_, _ int64,
) ([]UserSubscription, error) {
	r.calls++
	return append([]UserSubscription(nil), r.candidates...), nil
}

func TestGetActiveSubscriptionSelectsEarliestCandidate(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	repo := &subscriptionCandidatesRepoStub{
		candidates: []UserSubscription{
			{ID: 1, UserID: 10, GroupID: 20, ExpiresAt: now.Add(24 * time.Hour)},
			{ID: 2, UserID: 10, GroupID: 20, ExpiresAt: now.Add(48 * time.Hour)},
		},
	}
	svc := &SubscriptionService{userSubRepo: repo}

	subscription, err := svc.GetActiveSubscription(context.Background(), 10, 20)

	require.NoError(t, err)
	require.Equal(t, int64(1), subscription.ID)
	require.Equal(t, 1, repo.calls)
}

func TestListActiveSubscriptionsReturnsAllCandidates(t *testing.T) {
	repo := &subscriptionCandidatesRepoStub{
		candidates: []UserSubscription{{ID: 1}, {ID: 2}},
	}
	svc := &SubscriptionService{userSubRepo: repo}

	subscriptions, err := svc.ListActiveSubscriptions(context.Background(), 10, 20)

	require.NoError(t, err)
	require.Equal(t, []int64{1, 2}, []int64{subscriptions[0].ID, subscriptions[1].ID})
}

func TestListActiveSubscriptionsByPlanIDsUsesDistinctSortedPlanIDs(t *testing.T) {
	repo := &subscriptionPlanCandidatesRepoStub{
		candidates: []UserSubscription{{ID: 11}, {ID: 12}},
	}
	svc := &SubscriptionService{userSubRepo: repo}

	subscriptions, err := svc.ListActiveSubscriptionsByPlanIDs(
		context.Background(),
		10,
		[]int64{30, 10, 30, 20},
	)

	require.NoError(t, err)
	require.Equal(t, []int64{10, 20, 30}, repo.requestedPlanIDs)
	require.Equal(t, []int64{11, 12}, []int64{subscriptions[0].ID, subscriptions[1].ID})
	require.Equal(t, 1, repo.calls)
}

func TestListActiveSubscriptionsByPlanIDsSkipsRepositoryForNoPlans(t *testing.T) {
	repo := &subscriptionPlanCandidatesRepoStub{}
	svc := &SubscriptionService{userSubRepo: repo}

	subscriptions, err := svc.ListActiveSubscriptionsByPlanIDs(context.Background(), 10, nil)

	require.NoError(t, err)
	require.Empty(t, subscriptions)
	require.Zero(t, repo.calls)
}

func TestInvalidateSubscriptionCachesTargetsConcreteInstanceAndPublishesSelectionKeys(t *testing.T) {
	cache := &subscriptionInvalidationPubSubStub{}
	svc := &SubscriptionService{
		billingCacheService: &BillingCacheService{cache: cache},
	}

	err := svc.invalidateSubscriptionCaches(10, 20, 101)

	require.NoError(t, err)
	require.Equal(t, []int64{101}, cache.invalidatedSubIDs)
	require.Equal(t, []string{
		subCacheKey(10, 20),
		subCandidateCacheKey(10, 20),
	}, cache.publishedKeys)
}
