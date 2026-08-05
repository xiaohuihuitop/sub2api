import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import KeyAssetPermissionsForm from '../KeyAssetPermissionsForm.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

describe('KeyAssetPermissionsForm', () => {
  it('allows independent multi-selection and keeps balance enabled by default', async () => {
    const wrapper = mount(KeyAssetPermissionsForm, {
      props: {
        platforms: [
          { id: 11, code: 'openai-primary', name: 'OpenAI Primary', account_platform: 'openai' },
          { id: 12, code: 'grok-primary', name: 'Grok Primary', account_platform: 'grok' },
        ],
        subscriptionPlans: [
          { id: 21, name: 'Pro' },
          { id: 22, name: 'Team' },
        ],
        platformIds: [],
        subscriptionPlanIds: [],
        allowBalance: true,
      },
    })

    expect((wrapper.get('[data-test="key-balance"]').element as HTMLInputElement).checked).toBe(true)

    await wrapper.get('[data-test="key-platform-11"]').setValue(true)
    await wrapper.get('[data-test="key-plan-21"]').setValue(true)
    await wrapper.get('[data-test="key-balance"]').setValue(false)

    expect(wrapper.emitted('update:platformIds')).toEqual([[[11]]])
    expect(wrapper.emitted('update:subscriptionPlanIds')).toEqual([[[21]]])
    expect(wrapper.emitted('update:allowBalance')).toEqual([[false]])
  })
})
