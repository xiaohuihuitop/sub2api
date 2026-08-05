import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'

import PlatformPoolDialog from '../PlatformPoolDialog.vue'

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const mountDialog = (platform = null) => mount(PlatformPoolDialog, {
  props: {
    show: true,
    platform,
    submitting: false,
  },
  global: {
    stubs: {
      BaseDialog: {
        props: ['show'],
        template: '<div v-if="show"><slot /><slot name="footer" /></div>',
      },
      Icon: true,
    },
  },
})

describe('PlatformPoolDialog', () => {
  it('emits an asset-free platform pool payload', async () => {
    const wrapper = mountDialog()

    await wrapper.get('[data-test="platform-code"]').setValue('openai-primary')
    await wrapper.get('[data-test="platform-name"]').setValue('OpenAI Primary')
    await wrapper.get('[data-test="platform-account-platform"]').setValue('openai')
    await wrapper.get('[data-test="add-model-rule"]').trigger('click')
    await wrapper.get('[data-test="model-pattern-0"]').setValue('gpt-*')
    await wrapper.get('[data-test="endpoint-responses-0"]').setValue(true)
    await wrapper.get('form').trigger('submit')

    expect(wrapper.emitted('save')).toEqual([[
      {
        code: 'openai-primary',
        name: 'OpenAI Primary',
        account_platform: 'openai',
        status: 'active',
        model_rules: [
          {
            model_pattern: 'gpt-*',
            upstream_model: '',
            endpoint_capabilities: ['responses'],
            enabled: true,
          },
        ],
      },
    ]])
  })

  it('loads an existing platform pool for editing', () => {
    const wrapper = mountDialog({
      id: 12,
      code: 'grok-accounts',
      name: 'Grok Accounts',
      account_platform: 'grok',
      status: 'disabled',
      legacy_group_id: 7,
      model_rules: [
        {
          id: 5,
          model_pattern: 'grok-*',
          upstream_model: 'grok-4',
          endpoint_capabilities: ['chat_completions'],
          enabled: true,
        },
      ],
    })

    expect((wrapper.get('[data-test="platform-code"]').element as HTMLInputElement).value).toBe('grok-accounts')
    expect((wrapper.get('[data-test="platform-account-platform"]').element as HTMLSelectElement).value).toBe('grok')
    expect((wrapper.get('[data-test="endpoint-chat_completions-0"]').element as HTMLInputElement).checked).toBe(true)
    expect(wrapper.text()).not.toContain('legacy_group_id')
  })
})
