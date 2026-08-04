package service

import (
	"context"
	"errors"
	"reflect"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrNoUsableBillingGroup = infraerrors.Forbidden("NO_USABLE_BILLING_GROUP", "no usable billing group is available")

type apiKeySubscriptionResolver interface {
	GetActiveSubscription(ctx context.Context, userID, groupID int64) (*UserSubscription, error)
	ValidateAndCheckLimits(sub *UserSubscription, group *Group) (bool, error)
	EnsureWindowMaintenance(ctx context.Context, sub *UserSubscription) (*UserSubscription, error)
}

type apiKeySubscriptionCandidateLister interface {
	ListActiveSubscriptions(ctx context.Context, userID, groupID int64) ([]UserSubscription, error)
}

func (s *APIKeyService) ResolveBillingGroupForRequest(
	ctx context.Context,
	apiKey *APIKey,
	subscriptions apiKeySubscriptionResolver,
	skipBilling bool,
	targetPlatform string,
	requestEndpoint string,
) (*UserSubscription, error) {
	candidates := billingGroupCandidates(apiKey, targetPlatform, requestEndpoint)
	if len(candidates) == 0 {
		if apiKey != nil && apiKey.GroupID == nil && len(apiKey.AllowedGroupIDs) == 0 {
			if skipBilling || (apiKey.User != nil && apiKey.User.Balance > 0) {
				return nil, nil
			}
			return nil, ErrInsufficientBalance
		}
		return nil, ErrNoUsableBillingGroup
	}

	if skipBilling {
		applyAPIKeyBillingGroup(apiKey, &candidates[0])
		return nil, nil
	}

	if hasSubscriptionResolver(subscriptions) {
		for i := range candidates {
			group := &candidates[i]
			subscriptionsForGroup, err := listBillingGroupSubscriptions(ctx, subscriptions, apiKey.UserID, group.ID)
			if err != nil {
				if isSubscriptionCandidateUnavailableError(err) {
					continue
				}
				return nil, err
			}

			for i := range subscriptionsForGroup {
				subscription := &subscriptionsForGroup[i]
				needsMaintenance, err := subscriptions.ValidateAndCheckLimits(subscription, group)
				if needsMaintenance {
					subscription, err = subscriptions.EnsureWindowMaintenance(ctx, subscription)
					if err == nil {
						_, err = subscriptions.ValidateAndCheckLimits(subscription, group)
					}
				}
				if err != nil {
					if isSubscriptionCandidateUnavailableError(err) {
						continue
					}
					return nil, err
				}
				applyAPIKeyBillingGroup(apiKey, group)
				return subscription, nil
			}
		}
	}

	if apiKey.User == nil || apiKey.User.Balance <= 0 {
		return nil, ErrInsufficientBalance
	}
	applyAPIKeyBillingGroup(apiKey, &candidates[0])
	return nil, nil
}

func hasSubscriptionResolver(subscriptions apiKeySubscriptionResolver) bool {
	if subscriptions == nil {
		return false
	}
	value := reflect.ValueOf(subscriptions)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}

func billingGroupCandidates(apiKey *APIKey, platform, endpoint string) []Group {
	if apiKey == nil {
		return nil
	}
	groups := apiKey.AllowedGroups
	if len(groups) == 0 && apiKey.Group != nil {
		groups = []Group{*apiKey.Group}
	}
	capability := billingEndpointCapability(endpoint)
	candidates := make([]Group, 0, len(groups))
	for _, group := range groups {
		if !group.IsActive() || (platform != "" && group.Platform != platform) {
			continue
		}
		if capability != "" && group.Platform == PlatformOpenAI && !groupSupportsBillingEndpoint(group, capability) {
			continue
		}
		candidates = append(candidates, group)
	}
	return sortAPIKeyAllowedGroups(candidates)
}

func listBillingGroupSubscriptions(
	ctx context.Context,
	subscriptions apiKeySubscriptionResolver,
	userID, groupID int64,
) ([]UserSubscription, error) {
	if lister, ok := subscriptions.(apiKeySubscriptionCandidateLister); ok {
		return lister.ListActiveSubscriptions(ctx, userID, groupID)
	}

	subscription, err := subscriptions.GetActiveSubscription(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	if subscription == nil {
		return nil, ErrSubscriptionNotFound
	}
	return []UserSubscription{*subscription}, nil
}

func billingEndpointCapability(endpoint string) OpenAIEndpointCapability {
	switch {
	case strings.Contains(strings.ToLower(endpoint), "/chat/completions"):
		return OpenAIEndpointCapabilityChatCompletions
	case strings.Contains(strings.ToLower(endpoint), "/responses"):
		return OpenAIEndpointCapabilityResponses
	default:
		return ""
	}
}

func groupSupportsBillingEndpoint(group Group, capability OpenAIEndpointCapability) bool {
	if group.OpenAIEndpointCapabilities == nil {
		return true
	}
	return group.OpenAIEndpointCapabilities[string(capability)]
}

func isSubscriptionUsageLimitError(err error) bool {
	return errors.Is(err, ErrDailyLimitExceeded) ||
		errors.Is(err, ErrWeeklyLimitExceeded) ||
		errors.Is(err, ErrMonthlyLimitExceeded)
}

func isSubscriptionCandidateUnavailableError(err error) bool {
	return isSubscriptionUsageLimitError(err) ||
		errors.Is(err, ErrSubscriptionNotFound) ||
		errors.Is(err, ErrSubscriptionInvalid) ||
		errors.Is(err, ErrSubscriptionExpired) ||
		errors.Is(err, ErrSubscriptionSuspended)
}

func applyAPIKeyBillingGroup(apiKey *APIKey, group *Group) {
	if apiKey == nil || group == nil {
		return
	}
	changed := apiKey.GroupID == nil || *apiKey.GroupID != group.ID
	groupCopy := *group
	apiKey.GroupID = &groupCopy.ID
	apiKey.Group = &groupCopy
	if changed && apiKey.User != nil {
		apiKey.User.UserGroupRPMOverride = nil
	}
}
