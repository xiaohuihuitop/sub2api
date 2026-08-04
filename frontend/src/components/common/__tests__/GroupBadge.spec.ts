import { mount } from '@vue/test-utils'
import { describe, expect, it, vi } from 'vitest'
import GroupBadge from '../GroupBadge.vue'
import GroupOptionItem from '../GroupOptionItem.vue'

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ cachedPublicSettings: {} }),
}))

const stubs = {
  PlatformIcon: { template: '<i data-platform-icon />' },
}

describe('routing group labels', () => {
  it('hides legacy group pricing and subscription labels by default', () => {
    const wrapper = mount(GroupBadge, {
      props: {
        name: 'OpenAI routing',
        platform: 'openai',
        subscriptionType: 'subscription',
        rateMultiplier: 1.5,
        peakRateEnabled: true,
        peakStart: '08:00',
        peakEnd: '10:00',
        peakRateMultiplier: 2,
      },
      global: { stubs },
    })

    expect(wrapper.text()).toContain('OpenAI routing')
    expect(wrapper.text()).not.toContain('1.5x')
    expect(wrapper.text()).not.toContain('groups.subscription')
    expect(wrapper.text()).not.toContain('08:00')
  })

  it('does not show legacy pricing inside a routing-group option', () => {
    const wrapper = mount(GroupOptionItem, {
      props: {
        name: 'OpenAI routing',
        platform: 'openai',
        subscriptionType: 'subscription',
        rateMultiplier: 1.5,
        peakRateEnabled: true,
        peakStart: '08:00',
        peakEnd: '10:00',
        peakRateMultiplier: 2,
      },
      global: { stubs },
    })

    expect(wrapper.text()).toContain('OpenAI routing')
    expect(wrapper.text()).not.toContain('1.5x')
    expect(wrapper.text()).not.toContain('08:00')
  })
})
