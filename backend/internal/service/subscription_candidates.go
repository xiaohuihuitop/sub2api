package service

import (
	"context"
	"time"
)

// GetActiveSubscription returns the first candidate selected by the stable
// repository ordering. It remains compatible with callers that need one
// subscription while allowing the resolver to inspect every candidate.
func (s *SubscriptionService) GetActiveSubscription(
	ctx context.Context,
	userID, groupID int64,
) (*UserSubscription, error) {
	subscriptions, err := s.ListActiveSubscriptions(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if len(subscriptions) == 0 {
		return nil, ErrSubscriptionNotFound
	}
	return cloneUserSubscription(&subscriptions[0]), nil
}

// ListActiveSubscriptions returns ordered active candidates for one group.
// Legacy repositories keep their one-subscription behavior until migrated.
func (s *SubscriptionService) ListActiveSubscriptions(
	ctx context.Context,
	userID, groupID int64,
) ([]UserSubscription, error) {
	lister, ok := s.userSubRepo.(ActiveUserSubscriptionLister)
	if !ok {
		subscription, err := s.getActiveSubscriptionLegacy(ctx, userID, groupID)
		if err != nil {
			return nil, err
		}
		return []UserSubscription{*cloneUserSubscription(subscription)}, nil
	}

	key := subCandidateCacheKey(userID, groupID)
	if s.subCacheL1 != nil {
		if value, found := s.subCacheL1.Get(key); found {
			if subscriptions, valid := value.([]UserSubscription); valid {
				return cloneUserSubscriptions(subscriptions), nil
			}
		}
	}

	value, err, _ := s.subCacheGroup.Do(key, func() (any, error) {
		subscriptions, queryErr := lister.ListActiveByUserIDAndGroupID(ctx, userID, groupID)
		if queryErr != nil {
			return nil, queryErr
		}
		if len(subscriptions) == 0 {
			return nil, ErrSubscriptionNotFound
		}

		cached := cloneUserSubscriptions(subscriptions)
		if s.subCacheL1 != nil {
			_ = s.subCacheL1.SetWithTTL(key, cached, 1, s.jitteredTTL(s.subCacheTTL))
		}
		return cached, nil
	})
	if err != nil {
		return nil, err
	}

	subscriptions, ok := value.([]UserSubscription)
	if !ok || len(subscriptions) == 0 {
		return nil, ErrSubscriptionNotFound
	}
	return cloneUserSubscriptions(subscriptions), nil
}

func cloneUserSubscriptions(subscriptions []UserSubscription) []UserSubscription {
	cloned := make([]UserSubscription, len(subscriptions))
	for index := range subscriptions {
		cloned[index] = *cloneUserSubscription(&subscriptions[index])
	}
	return cloned
}

func cloneUserSubscription(subscription *UserSubscription) *UserSubscription {
	if subscription == nil {
		return nil
	}

	cloned := *subscription
	cloned.SubscriptionPlanID = copyInt64(subscription.SubscriptionPlanID)
	cloned.DailyLimitUSDSnapshot = copyFloat64(subscription.DailyLimitUSDSnapshot)
	cloned.WeeklyLimitUSDSnapshot = copyFloat64(subscription.WeeklyLimitUSDSnapshot)
	cloned.MonthlyLimitUSDSnapshot = copyFloat64(subscription.MonthlyLimitUSDSnapshot)
	cloned.DailyWindowStart = copyTime(subscription.DailyWindowStart)
	cloned.WeeklyWindowStart = copyTime(subscription.WeeklyWindowStart)
	cloned.MonthlyWindowStart = copyTime(subscription.MonthlyWindowStart)
	cloned.AssignedBy = copyInt64(subscription.AssignedBy)
	cloned.DeletedAt = copyTime(subscription.DeletedAt)
	return &cloned
}

func copyInt64(value *int64) *int64 {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func copyTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}
