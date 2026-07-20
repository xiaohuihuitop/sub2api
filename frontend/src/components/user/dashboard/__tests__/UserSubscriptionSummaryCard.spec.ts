import { describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import UserSubscriptionSummaryCard from '../UserSubscriptionSummaryCard.vue'
import type { UserSubscription } from '@/types'

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({
    t: (key: string, params?: Record<string, unknown>) =>
      params ? `${key}:${JSON.stringify(params)}` : key,
  }),
}))

const subscription = {
  id: 1,
  group_id: 2,
  status: 'active',
  starts_at: '2026-07-01T00:00:00Z',
  expires_at: '2026-08-01T00:00:00Z',
  daily_usage_usd: 3,
  weekly_usage_usd: 0,
  monthly_usage_usd: 0,
  daily_window_start: '2026-07-11T00:00:00Z',
  weekly_window_start: null,
  monthly_window_start: null,
  group: {
    name: 'Premium',
    daily_limit_usd: 10,
  },
} as UserSubscription

describe('UserSubscriptionSummaryCard', () => {
  it('shows the active quota, remaining amount, and progress', () => {
    const wrapper = mount(UserSubscriptionSummaryCard, {
      props: { subscription },
      global: { stubs: { Icon: true } },
    })

    expect(wrapper.text()).toContain('Premium')
    expect(wrapper.text()).toContain('$3.00 / $10.00')
    expect(wrapper.text()).toContain('$7.00')
    expect(wrapper.get('[data-testid="subscription-progress"]').attributes('style')).toContain('30%')
  })

  it('shows unlimited when the subscription has no quota', () => {
    const wrapper = mount(UserSubscriptionSummaryCard, {
      props: {
        subscription: {
          ...subscription,
          group: { name: 'Unlimited' },
        },
      },
      global: { stubs: { Icon: true } },
    })

    expect(wrapper.text()).toContain('Unlimited')
    expect(wrapper.text()).toContain('dashboard.subscriptionSummary.unlimited')
    expect(wrapper.find('[data-testid="subscription-progress"]').exists()).toBe(false)
  })
})
