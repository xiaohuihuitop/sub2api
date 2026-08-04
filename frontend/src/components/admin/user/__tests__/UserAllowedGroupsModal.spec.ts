import { flushPromises, mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import UserAllowedGroupsModal from '../UserAllowedGroupsModal.vue'
import type { AdminGroup, AdminUser } from '@/types'

const { listGroups } = vi.hoisted(() => ({
  listGroups: vi.fn(),
}))

vi.mock('@/api/admin', () => ({
  adminAPI: {
    groups: { list: listGroups },
    users: { update: vi.fn() },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showSuccess: vi.fn() }),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

const user = {
  id: 1,
  email: 'user@example.com',
  allowed_groups: [],
} as AdminUser

function group(partial: Partial<AdminGroup>): AdminGroup {
  return {
    id: 0,
    name: '',
    status: 'active',
    is_exclusive: false,
    subscription_type: 'standard',
    platform: 'openai',
    ...partial,
  } as AdminGroup
}

describe('UserAllowedGroupsModal', () => {
  it('allows active routing groups regardless of their legacy subscription type', async () => {
    listGroups.mockResolvedValue({
      items: [
        group({
          id: 7,
          name: 'Legacy subscription routing group',
          is_exclusive: true,
          subscription_type: 'subscription',
        }),
      ],
    })

    const wrapper = mount(UserAllowedGroupsModal, {
      props: { show: false, user },
      global: {
        stubs: {
          BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /></div>' },
          PlatformIcon: true,
        },
      },
    })

    await wrapper.setProps({ show: true })
    await flushPromises()

    expect(wrapper.text()).toContain('Legacy subscription routing group')
  })
})
