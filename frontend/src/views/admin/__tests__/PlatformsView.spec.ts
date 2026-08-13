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
    useI18n: () => ({ t: (key: string) => key === 'admin.platforms.errors.PLATFORM_IN_USE' ? 'localized platform in use' : key }),
  }
})

vi.mock('@/api/admin', () => ({
  adminAPI: {
    platforms: {
      list: vi.fn(),
      create: vi.fn(),
      update: vi.fn(),
      previewDelete: vi.fn(),
      remove: vi.fn(),
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
      ConfirmDialog: defineComponent({
        props: ['show', 'title', 'message'],
        emits: ['confirm', 'cancel'],
        template: '<button v-if="show" data-test="confirm-platform-delete" @click="$emit(\'confirm\')">confirm</button>',
      }),
    },
  },
})

describe('PlatformsView', () => {
  beforeEach(() => {
    vi.mocked(adminAPI.platforms.list).mockReset()
    vi.mocked(adminAPI.platforms.previewDelete).mockReset()
    vi.mocked(adminAPI.platforms.remove).mockReset()
    showError.mockReset()
    showSuccess.mockReset()
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

  it('deletes a platform only after explicit confirmation', async () => {
    vi.mocked(adminAPI.platforms.list).mockResolvedValue([{
      id: 7, code: 'unused', name: 'Unused', account_platform: 'openai', status: 'disabled',
      endpoint_capabilities: [], model_rules: [],
    }])
    vi.mocked(adminAPI.platforms.remove).mockResolvedValue(undefined)
    vi.mocked(adminAPI.platforms.previewDelete).mockResolvedValue({
      accounts: 0, api_keys: 0, usage_logs: 3, audits: 0, ops: 2, configs: 0, can_delete: true,
    })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="delete-platform-7"]').trigger('click')
    await flushPromises()
    expect(adminAPI.platforms.previewDelete).toHaveBeenCalledWith(7)
    expect(adminAPI.platforms.remove).not.toHaveBeenCalled()
    await wrapper.get('[data-test="confirm-platform-delete"]').trigger('click')
    await flushPromises()

    expect(adminAPI.platforms.remove).toHaveBeenCalledWith(7)
    expect(showSuccess).toHaveBeenCalledWith('admin.platforms.deleted')
  })

  it('shows the localized safe-delete conflict instead of removing references', async () => {
    vi.mocked(adminAPI.platforms.list).mockResolvedValue([{
      id: 7, code: 'used', name: 'Used', account_platform: 'openai', status: 'active',
      endpoint_capabilities: ['responses'], model_rules: [],
    }])
    vi.mocked(adminAPI.platforms.remove).mockRejectedValue({
      reason: 'PLATFORM_IN_USE', metadata: { accounts: '1', api_keys: '2', usage_logs: '3', audits: '4', ops: '5', configs: '6' },
    })
    vi.mocked(adminAPI.platforms.previewDelete).mockResolvedValue({
      accounts: 0, api_keys: 0, usage_logs: 3, audits: 4, ops: 5, configs: 6, can_delete: true,
    })
    const wrapper = mountView()
    await flushPromises()

    await wrapper.get('[data-test="delete-platform-7"]').trigger('click')
    await flushPromises()
    await wrapper.get('[data-test="confirm-platform-delete"]').trigger('click')
    await flushPromises()

    expect(showError).toHaveBeenCalledWith('localized platform in use')
  })
})
