import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import DashboardView from '../DashboardView.vue'

const { getDashboardStats, getDashboardTrend, getDashboardModels, refreshUser, fetchActiveSubscriptions } = vi.hoisted(() => ({
  getDashboardStats: vi.fn(),
  getDashboardTrend: vi.fn(),
  getDashboardModels: vi.fn(),
  refreshUser: vi.fn(),
  fetchActiveSubscriptions: vi.fn(),
}))

vi.mock('@/api/usage', () => ({
  usageAPI: { getDashboardStats, getDashboardTrend, getDashboardModels },
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { balance: 0 },
    isSimpleMode: false,
    refreshUser,
  }),
}))

vi.mock('@/stores', () => ({
  useSubscriptionStore: () => ({
    activeSubscriptions: [],
    fetchActiveSubscriptions,
  }),
}))

describe('user DashboardView', () => {
  beforeEach(() => {
    localStorage.clear()
    getDashboardStats.mockResolvedValue({})
    getDashboardTrend.mockResolvedValue({ trend: [] })
    getDashboardModels.mockResolvedValue({ models: [] })
    refreshUser.mockResolvedValue(undefined)
    fetchActiveSubscriptions.mockResolvedValue(undefined)
  })

  it('restores the saved dashboard range and granularity', async () => {
    localStorage.setItem('user-dashboard-date-range', JSON.stringify({
      startDate: '2026-07-03',
      endDate: '2026-07-09',
      granularity: 'hour',
    }))

    mount(DashboardView, {
      global: {
        stubs: {
          AppLayout: { template: '<div><slot /></div>' },
          LoadingSpinner: true,
          UserDashboardStats: true,
          UserDashboardCharts: true,
        },
      },
    })

    await flushPromises()

    expect(getDashboardTrend).toHaveBeenCalledWith({
      start_date: '2026-07-03',
      end_date: '2026-07-09',
      granularity: 'hour',
    })
  })
})
