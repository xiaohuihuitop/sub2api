import type { UserSubscription } from '@/types'

export type SubscriptionLimitWindow = 'daily' | 'weekly' | 'monthly'

export function getSubscriptionPlanName(subscription: UserSubscription): string {
  return subscription.plan_name_snapshot || subscription.group?.name || `Group #${subscription.group_id}`
}

export function getSubscriptionLimit(
  subscription: UserSubscription,
  window: SubscriptionLimitWindow,
): number | null {
  const snapshot = getSnapshotLimit(subscription, window)
  return snapshot === undefined ? getLegacyGroupLimit(subscription, window) : snapshot
}

export function getSubscriptionRateMultiplier(subscription: UserSubscription): number {
  return subscription.rate_multiplier_snapshot ?? subscription.group?.rate_multiplier ?? 1
}

function getSnapshotLimit(
  subscription: UserSubscription,
  window: SubscriptionLimitWindow,
): number | null | undefined {
  if (window === 'daily') return subscription.daily_limit_usd_snapshot
  if (window === 'weekly') return subscription.weekly_limit_usd_snapshot
  return subscription.monthly_limit_usd_snapshot
}

function getLegacyGroupLimit(
  subscription: UserSubscription,
  window: SubscriptionLimitWindow,
): number | null {
  if (window === 'daily') return subscription.group?.daily_limit_usd ?? null
  if (window === 'weekly') return subscription.group?.weekly_limit_usd ?? null
  return subscription.group?.monthly_limit_usd ?? null
}
