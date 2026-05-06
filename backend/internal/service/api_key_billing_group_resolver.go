package service

import (
	"context"
	"errors"
	"sort"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrNoUsableBillingGroup = infraerrors.Forbidden("NO_USABLE_BILLING_GROUP", "no usable billing group is available")

type apiKeySubscriptionResolver interface {
	GetActiveSubscription(ctx context.Context, userID, groupID int64) (*UserSubscription, error)
	ValidateAndCheckLimits(sub *UserSubscription, group *Group) (bool, error)
	DoWindowMaintenance(sub *UserSubscription)
}

func (s *APIKeyService) ResolveBillingGroupForRequest(
	ctx context.Context,
	apiKey *APIKey,
	subscriptionResolver apiKeySubscriptionResolver,
	skipBilling bool,
	targetPlatform string,
) (*UserSubscription, error) {
	if apiKey == nil {
		return nil, ErrNoUsableBillingGroup
	}

	candidates := apiKey.allowedBillingGroups()
	if targetPlatform != "" {
		filtered := candidates[:0]
		for _, group := range candidates {
			if group.Platform == targetPlatform {
				filtered = append(filtered, group)
			}
		}
		candidates = filtered
	}
	if len(candidates) == 0 {
		return nil, ErrNoUsableBillingGroup
	}

	subscriptionGroups, standardGroups := splitBillingGroups(candidates)
	var lastSubscriptionErr error
	if subscriptionResolver != nil {
		for i := range subscriptionGroups {
			group := &subscriptionGroups[i]
			subscription, err := subscriptionResolver.GetActiveSubscription(ctx, apiKey.UserID, group.ID)
			if err != nil {
				lastSubscriptionErr = err
				continue
			}
			if !skipBilling {
				needsMaintenance, err := subscriptionResolver.ValidateAndCheckLimits(subscription, group)
				if err != nil {
					lastSubscriptionErr = err
					continue
				}
				if needsMaintenance {
					maintenanceCopy := *subscription
					subscriptionResolver.DoWindowMaintenance(&maintenanceCopy)
				}
			}
			applyBillingGroup(apiKey, group)
			return subscription, nil
		}
	}

	if len(standardGroups) > 0 && !skipBilling && (apiKey.User == nil || apiKey.User.Balance <= 0) {
		if lastSubscriptionErr != nil && !errors.Is(lastSubscriptionErr, ErrSubscriptionNotFound) {
			return nil, lastSubscriptionErr
		}
		return nil, ErrInsufficientBalance
	}

	for i := range standardGroups {
		group := &standardGroups[i]
		applyBillingGroup(apiKey, group)
		return nil, nil
	}

	if lastSubscriptionErr != nil && !errors.Is(lastSubscriptionErr, ErrSubscriptionNotFound) {
		return nil, lastSubscriptionErr
	}
	return nil, ErrNoUsableBillingGroup
}

func (k *APIKey) allowedBillingGroups() []Group {
	if k == nil {
		return nil
	}
	if len(k.AllowedGroups) > 0 {
		groups := make([]Group, 0, len(k.AllowedGroups))
		for _, group := range k.AllowedGroups {
			if group.IsActive() {
				groups = append(groups, group)
			}
		}
		return groups
	}
	if k.Group != nil && k.Group.IsActive() {
		return []Group{*k.Group}
	}
	return nil
}

func splitBillingGroups(groups []Group) ([]Group, []Group) {
	subscriptionGroups := make([]Group, 0, len(groups))
	standardGroups := make([]Group, 0, len(groups))
	for _, group := range groups {
		if group.IsSubscriptionType() {
			subscriptionGroups = append(subscriptionGroups, group)
			continue
		}
		standardGroups = append(standardGroups, group)
	}
	sort.SliceStable(subscriptionGroups, func(i, j int) bool {
		left, right := subscriptionGroups[i], subscriptionGroups[j]
		leftLimit, leftHasLimit := positiveDailyLimit(left)
		rightLimit, rightHasLimit := positiveDailyLimit(right)
		if leftHasLimit != rightHasLimit {
			return leftHasLimit
		}
		if leftHasLimit && leftLimit != rightLimit {
			return leftLimit < rightLimit
		}
		return groupOrderLess(left, right)
	})
	sort.SliceStable(standardGroups, func(i, j int) bool {
		return groupOrderLess(standardGroups[i], standardGroups[j])
	})
	return subscriptionGroups, standardGroups
}

func positiveDailyLimit(group Group) (float64, bool) {
	if group.DailyLimitUSD == nil || *group.DailyLimitUSD <= 0 {
		return 0, false
	}
	return *group.DailyLimitUSD, true
}

func groupOrderLess(left, right Group) bool {
	if left.SortOrder != right.SortOrder {
		return left.SortOrder < right.SortOrder
	}
	return left.ID < right.ID
}

func applyBillingGroup(apiKey *APIKey, group *Group) {
	if apiKey == nil || group == nil {
		return
	}
	previousGroupID := apiKey.GroupID
	groupCopy := *group
	apiKey.Group = &groupCopy
	apiKey.GroupID = &groupCopy.ID
	if apiKey.User != nil && (previousGroupID == nil || *previousGroupID != groupCopy.ID) {
		// UserGroupRPMOverride is scoped to a specific (user, group) pair.
		// Once billing switches to another allowed group, the cached override
		// can no longer be trusted and must fall back to a fresh lookup.
		apiKey.User.UserGroupRPMOverride = nil
	}
}
