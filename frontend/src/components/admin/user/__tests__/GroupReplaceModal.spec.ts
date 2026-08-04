import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupReplaceModal from '../GroupReplaceModal.vue'
import type { AdminGroup, AdminUser } from '@/types'

vi.mock('@/api/admin', () => ({
  adminAPI: { users: { replaceGroup: vi.fn() } },
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

function group(partial: Partial<AdminGroup>): AdminGroup {
  return {
    id: 0,
    name: '',
    status: 'active',
    is_exclusive: true,
    subscription_type: 'standard',
    platform: 'openai',
    ...partial,
  } as AdminGroup
}

describe('GroupReplaceModal', () => {
  it('offers an active exclusive routing group even when legacy data marks it as subscription', () => {
    const wrapper = mount(GroupReplaceModal, {
      props: {
        show: true,
        user: { id: 1 } as AdminUser,
        oldGroup: { id: 1, name: 'Old routing group' },
        allGroups: [
          group({
            id: 2,
            name: 'Replacement routing group',
            subscription_type: 'subscription',
          }),
        ],
      },
      global: {
        stubs: {
          BaseDialog: { props: ['show'], template: '<div v-if="show"><slot /><slot name="footer" /></div>' },
          Icon: true,
        },
      },
    })

    expect(wrapper.text()).toContain('Replacement routing group')
  })
})
