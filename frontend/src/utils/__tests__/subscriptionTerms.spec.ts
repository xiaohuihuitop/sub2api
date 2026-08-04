import { describe, expect, it } from 'vitest'

import {
  getSubscriptionLimit,
  getSubscriptionPlanName,
  getSubscriptionRateMultiplier,
} from '../subscriptionTerms'
import type { UserSubscription } from '@/types'

const subscription = {
  id: 1,
  user_id: 2,
  group_id: 3,
  status: 'active',
  starts_at: '2026-08-01T00:00:00Z',
  expires_at: '2026-09-01T00:00:00Z',
  daily_usage_usd: 0,
  weekly_usage_usd: 0,
  monthly_usage_usd: 0,
  daily_window_start: null,
  weekly_window_start: null,
  monthly_window_start: null,
  created_at: '2026-08-01T00:00:00Z',
  updated_at: '2026-08-01T00:00:00Z',
  plan_name_snapshot: 'Snapshot plan',
  daily_limit_usd_snapshot: 12,
  weekly_limit_usd_snapshot: null,
  monthly_limit_usd_snapshot: 80,
  rate_multiplier_snapshot: 0.5,
  group: {
    name: 'Legacy routing group',
    daily_limit_usd: 99,
    weekly_limit_usd: 50,
    monthly_limit_usd: 999,
    rate_multiplier: 1.5,
  },
} as UserSubscription

describe('subscriptionTerms', () => {
  it('uses the subscription snapshot instead of mutable group terms', () => {
    expect(getSubscriptionPlanName(subscription)).toBe('Snapshot plan')
    expect(getSubscriptionLimit(subscription, 'daily')).toBe(12)
    expect(getSubscriptionLimit(subscription, 'weekly')).toBeNull()
    expect(getSubscriptionLimit(subscription, 'monthly')).toBe(80)
    expect(getSubscriptionRateMultiplier(subscription)).toBe(0.5)
  })

  it('falls back to legacy group terms only when a snapshot field is absent', () => {
    const legacy = {
      ...subscription,
      plan_name_snapshot: '',
      daily_limit_usd_snapshot: undefined,
      rate_multiplier_snapshot: undefined,
    }

    expect(getSubscriptionPlanName(legacy)).toBe('Legacy routing group')
    expect(getSubscriptionLimit(legacy, 'daily')).toBe(99)
    expect(getSubscriptionRateMultiplier(legacy)).toBe(1.5)
  })
})
