package service

import "time"

// resolveBillingMultipliers keeps customer price selection independent from
// routing. A subscription snapshot always wins over the balance profile.
func resolveBillingMultipliers(
	apiKey *APIKey,
	subscription *UserSubscription,
	fallback float64,
	now time.Time,
) (token, image, video float64) {
	base := resolveBillingRateMultiplier(apiKey, subscription, fallback)
	image = resolveImageBillingMultiplier(apiKey, base, subscription != nil)
	video = resolveVideoBillingMultiplier(apiKey, base, subscription != nil)
	if subscription != nil {
		return base, image, video
	}
	peak := 1.0
	if apiKey != nil && apiKey.Group != nil {
		peak = apiKey.Group.PeakMultiplierAt(now)
	}
	return base * peak, image, video
}

func resolveBillingRateMultiplier(apiKey *APIKey, subscription *UserSubscription, fallback float64) float64 {
	if subscription != nil {
		return nonNegativeMultiplier(subscription.RateMultiplierSnapshot)
	}
	if apiKey != nil && apiKey.Group != nil {
		return nonNegativeMultiplier(apiKey.Group.RateMultiplier)
	}
	return nonNegativeMultiplier(fallback)
}

func resolveImageBillingMultiplier(apiKey *APIKey, base float64, subscriptionBilling bool) float64 {
	if subscriptionBilling {
		return base
	}
	return resolveImageRateMultiplier(apiKey, base)
}

func resolveVideoBillingMultiplier(apiKey *APIKey, base float64, subscriptionBilling bool) float64 {
	if subscriptionBilling {
		return base
	}
	return resolveVideoRateMultiplier(apiKey, base)
}

func nonNegativeMultiplier(multiplier float64) float64 {
	if multiplier < 0 {
		return 0
	}
	return multiplier
}
