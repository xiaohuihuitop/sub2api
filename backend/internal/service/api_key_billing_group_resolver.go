package service

import (
	"context"
	"errors"
	"strings"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

var ErrNoUsableBillingGroup = infraerrors.Forbidden("NO_USABLE_BILLING_GROUP", "no usable billing group is available")

type apiKeySubscriptionResolver interface {
	GetActiveSubscription(ctx context.Context, userID, groupID int64) (*UserSubscription, error)
	ValidateAndCheckLimits(sub *UserSubscription, group *Group) (bool, error)
	EnsureWindowMaintenance(ctx context.Context, sub *UserSubscription) (*UserSubscription, error)
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

	plans, balances := splitAPIKeyBillingGroups(candidates)
	var lastUnavailable error
	for i := range plans {
		group := &plans[i]
		if skipBilling {
			applyAPIKeyBillingGroup(apiKey, group)
			return nil, nil
		}
		if subscriptions == nil {
			continue
		}
		subscription, err := subscriptions.GetActiveSubscription(ctx, apiKey.UserID, group.ID)
		if err != nil {
			if errors.Is(err, ErrSubscriptionNotFound) {
				lastUnavailable = err
				continue
			}
			return nil, err
		}
		needsMaintenance, err := subscriptions.ValidateAndCheckLimits(subscription, group)
		if needsMaintenance {
			subscription, err = subscriptions.EnsureWindowMaintenance(ctx, subscription)
			if err == nil {
				_, err = subscriptions.ValidateAndCheckLimits(subscription, group)
			}
		}
		if err != nil {
			if isSubscriptionUsageLimitError(err) {
				lastUnavailable = err
				continue
			}
			return nil, err
		}
		applyAPIKeyBillingGroup(apiKey, group)
		return subscription, nil
	}

	if len(balances) > 0 {
		if !skipBilling && (apiKey.User == nil || apiKey.User.Balance <= 0) {
			return nil, ErrInsufficientBalance
		}
		applyAPIKeyBillingGroup(apiKey, &balances[0])
		return nil, nil
	}
	if lastUnavailable != nil {
		return nil, lastUnavailable
	}
	return nil, ErrNoUsableBillingGroup
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

func splitAPIKeyBillingGroups(groups []Group) ([]Group, []Group) {
	plans := make([]Group, 0, len(groups))
	balances := make([]Group, 0, len(groups))
	for _, group := range groups {
		if group.IsSubscriptionType() {
			plans = append(plans, group)
		} else {
			balances = append(balances, group)
		}
	}
	return plans, balances
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
