import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'

import HomeView from '../HomeView.vue'

const { authState, appState, checkAuth, fetchPublicSettings } = vi.hoisted(() => ({
  authState: {
    isAuthenticated: false,
    isAdmin: false,
    user: null as { email: string } | null,
  },
  appState: {
    cachedPublicSettings: null as Record<string, string> | null,
    siteName: 'Sub2API',
    siteLogo: '',
    docUrl: '',
    publicSettingsLoaded: true,
  },
  checkAuth: vi.fn(),
  fetchPublicSettings: vi.fn(),
}))

vi.mock('@/stores', () => ({
  useAuthStore: () => ({ ...authState, checkAuth }),
  useAppStore: () => ({ ...appState, fetchPublicSettings }),
}))

vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key }),
}))

const RouterLinkStub = {
  props: ['to'],
  template: '<a :href="to"><slot /></a>',
}

function mountHome() {
  return mount(HomeView, {
    global: {
      stubs: {
        RouterLink: RouterLinkStub,
        LocaleSwitcher: true,
        Icon: true,
      },
    },
  })
}

describe('HomeView', () => {
  beforeEach(() => {
    authState.isAuthenticated = false
    authState.isAdmin = false
    authState.user = null
    appState.cachedPublicSettings = null
    appState.siteName = 'Sub2API'
    appState.siteLogo = ''
    appState.docUrl = ''
    appState.publicSettingsLoaded = true
    vi.clearAllMocks()
  })

  it('renders the package entry and tutorial images safely', () => {
    const wrapper = mountHome()

    const purchaseLink = wrapper.get('a[href="https://pay.ldxp.cn/shop/FED14QEA"]')
    expect(purchaseLink.attributes('target')).toBe('_blank')
    expect(purchaseLink.attributes('rel')).toContain('noreferrer')
    expect(wrapper.find('img[src="/dis1.png"]').exists()).toBe(true)
    expect(wrapper.find('img[src="/dis2.png"]').exists()).toBe(true)
  })

  it('routes authenticated administrators to the official admin dashboard', () => {
    authState.isAuthenticated = true
    authState.isAdmin = true
    authState.user = { email: 'admin@example.com' }

    const wrapper = mountHome()

    expect(wrapper.find('a[href="/admin/dashboard"]').exists()).toBe(true)
  })

  it('keeps unsafe documentation URLs out of the page', () => {
    appState.cachedPublicSettings = { doc_url: 'javascript:alert(1)' }

    const wrapper = mountHome()

    expect(wrapper.find('a[href="javascript:alert(1)"]').exists()).toBe(false)
  })
})
