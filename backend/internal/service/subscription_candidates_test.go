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
