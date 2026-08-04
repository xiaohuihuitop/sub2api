import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import GroupBillingProfileDialog from '../GroupBillingProfileDialog.vue'
import type { AdminGroup } from '@/types'

const { getBillingProfile, updateBillingProfile } = vi.hoisted(() => ({
  getBillingProfile: vi.fn(),
  updateBillingProfile: vi.fn(),
}))

const { showError, showSuccess } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('@/api/admin/groups', () => ({
  getBillingProfile,
  updateBillingProfile,
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({ t: (key: string) => key }),
}))

const BaseDialogStub = defineComponent({
  props: { show: Boolean, title: String },
  template: '<div v-if="show"><slot /><slot name="footer" /></div>',
})

const profile = {
  group_id: 1,
  balance_rate_multiplier: 1,
  peak_rate_enabled: false,
  peak_start: '',
  peak_end: '',
  peak_rate_multiplier: 1,
  image_rate_independent: false,
  image_rate_multiplier: 1,
  image_price_1k: null,
  image_price_2k: null,
  image_price_4k: null,
  batch_image_discount_multiplier: 1,
  batch_image_hold_multiplier: 1,
  video_rate_independent: false,
  video_rate_multiplier: 1,
  video_price_480p: null,
  video_price_720p: null,
  video_price_1080p: null,
  web_search_price_per_call: null,
}

const group = {
  id: 1,
  name: 'OpenAI routing',
  platform: 'openai',
} as AdminGroup

describe('GroupBillingProfileDialog', () => {
  beforeEach(() => {
    getBillingProfile.mockReset().mockResolvedValue(profile)
    updateBillingProfile.mockReset().mockResolvedValue(profile)
  })

  it('loads and saves balance pricing without subscription terms', async () => {
    const wrapper = mount(GroupBillingProfileDialog, {
      props: { show: true, group },
      global: {
        stubs: {
          BaseDialog: BaseDialogStub,
          GroupBadge: true,
          Icon: true,
        },
      },
    })

    await flushPromises()
    await wrapper.get('[data-testid="balance-rate-multiplier"]').setValue('0.8')
    await wrapper.get('[data-testid="web-search-price"]').setValue('0.02')
    await wrapper.get('form').trigger('submit')

    expect(updateBillingProfile).toHaveBeenCalledWith(1, expect.objectContaining({
      balance_rate_multiplier: 0.8,
      web_search_price_per_call: 0.02,
      image_rate_multiplier: 1,
      video_rate_multiplier: 1,
    }))
    expect(Object.keys(updateBillingProfile.mock.calls[0][1])).not.toContain('daily_limit_usd')
    expect(Object.keys(updateBillingProfile.mock.calls[0][1])).not.toContain('rate_multiplier')
    expect(Object.keys(updateBillingProfile.mock.calls[0][1])).not.toContain('group_id')
  })
})
