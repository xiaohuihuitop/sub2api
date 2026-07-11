import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'

import DashboardView from '../DashboardView.vue'

const {
  refreshUser,
  getDashboardStats,
  getDashboardTrend,
  getDashboardModels,
  getByDateRange,
  getMyPlatformQuotas,
  getActiveSubscriptions,
} = vi.hoisted(() => ({
  refreshUser: vi.fn(),
  getDashboardStats: vi.fn(),
  getDashboardTrend: vi.fn(),
  getDashboardModels: vi.fn(),
  getByDateRange: vi.fn(),
  getMyPlatformQuotas: vi.fn(),
  getActiveSubscriptions: vi.fn(),
}))

vi.mock('@/stores/auth', () => ({
  useAuthStore: () => ({
    user: { balance: 0 },
    isSimpleMode: false,
    refreshUser,
  }),
}))

vi.mock('@/api/usage', () => ({
  usageAPI: {
    getDashboardStats,
    getDashboardTrend,
    getDashboardModels,
    getByDateRange,
  },
}))

vi.mock('@/api/user', () => ({ getMyPlatformQuotas }))
vi.mock('@/api/subscriptions', () => ({ getActiveSubscriptions }))

describe('user DashboardView', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.clearAllMocks()
    refreshUser.mockResolvedValue(undefined)
    getDashboardStats.mockResolvedValue({ total_requests: 0 })
    getDashboardTrend.mockResolvedValue({ trend: [] })
    getDashboardModels.mockResolvedValue({ models: [] })
    getByDateRange.mockResolvedValue({ items: [] })
    getMyPlatformQuotas.mockResolvedValue({ platform_quotas: [] })
    getActiveSubscriptions.mockResolvedValue([])
  })

  it('restores the saved dashboard range and granularity', async () => {
    localStorage.setItem('sub2api:user-dashboard:view-state:v1', JSON.stringify({
      startDate: '2026-07-02',
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
          UserDashboardRecentUsage: true,
          UserDashboardQuickActions: true,
        },
      },
    })

    await flushPromises()

    expect(getDashboardTrend).toHaveBeenCalledWith({
      start_date: '2026-07-02',
      end_date: '2026-07-09',
      granularity: 'hour',
    })
    expect(getByDateRange).toHaveBeenCalledWith('2026-07-02', '2026-07-09')
    expect(getActiveSubscriptions).toHaveBeenCalledOnce()
  })
})
