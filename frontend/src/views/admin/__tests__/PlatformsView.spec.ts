import { defineComponent } from 'vue'
import { flushPromises, mount } from '@vue/test-utils'
import { beforeEach, describe, expect, it, vi } from 'vitest'

import PlatformsView from '../PlatformsView.vue'
import { adminAPI } from '@/api/admin'

const { showError, showSuccess } = vi.hoisted(() => ({
  showError: vi.fn(),
  showSuccess: vi.fn(),
}))

vi.mock('vue-i18n', async () => {
  const actual = await vi.importActual<typeof import('vue-i18n')>('vue-i18n')
  return {
    ...actual,
    useI18n: () => ({ t: (key: string) => key }),
  }
})

vi.mock('@/api/admin', () => ({
  adminAPI: {
    platforms: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
    },
  },
}))

vi.mock('@/stores/app', () => ({
  useAppStore: () => ({ showError, showSuccess }),
}))

const DataTableStub = defineComponent({
  props: {
    columns: { type: Array, required: true },
    data: { type: Array, required: true },
  },
  template: `
    <div data-test="platform-table">
      <div v-for="row in data" :key="row.id" data-test="platform-row">
        <span v-for="column in columns" :key="column.key">
          <slot :name="\`cell-\${column.key}\`" :row="row" :value="row[column.key]" />
        </span>
      </div>
    </div>
  `,
})

const mountView = () => mount(PlatformsView, {
  global: {
    stubs: {
      AppLayout: { template: '<div><slot /></div>' },
      DataTable: DataTableStub,
      Icon: true,
      PlatformIcon: true,
      PlatformPoolDialog: { template: '<div />' },
    },
  },
})

describe('PlatformsView', () => {
  beforeEach(() => {
    vi.mocked(adminAPI.platforms.list).mockReset()
  })

  it('renders platforms from the platform-level endpoint contract', async () => {
    vi.mocked(adminAPI.platforms.list).mockResolvedValue([
      {
        id: 1,
        code: 'codex',
        name: 'Codex',
        account_platform: 'openai',
        status: 'active',
        endpoint_capabilities: ['chat_completions', 'responses'],
        model_rules: [{ id: 1, model_pattern: 'gpt-*', upstream_model: 'gpt-5.6', enabled: true }],
      },
      {
        id: 2,
        code: 'glm',
        name: 'GLM',
        account_platform: 'openai',
        status: 'active',
        endpoint_capabilities: ['chat_completions'],
        model_rules: [{ id: 2, model_pattern: 'glm*', upstream_model: '', enabled: true }],
      },
    ])

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.findAll('[data-test="platform-row"]')).toHaveLength(2)
    expect(wrapper.text()).toContain('codex')
    expect(wrapper.text()).toContain('gpt-*')
    expect(wrapper.text()).toContain('gpt-5.6')
    expect(wrapper.text()).toContain('Chat Completions')
    expect(wrapper.text()).toContain('Responses')
  })

  it('shows a retryable error state when the platform list fails', async () => {
    vi.mocked(adminAPI.platforms.list).mockRejectedValue(new Error('offline'))

    const wrapper = mountView()
    await flushPromises()

    expect(wrapper.find('[data-test="platform-load-error"]').exists()).toBe(true)
    expect(showError).toHaveBeenCalled()
  })
})
