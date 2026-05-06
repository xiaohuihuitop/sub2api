import { beforeEach, describe, expect, it, vi } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createMemoryHistory, createRouter } from 'vue-router'
import HomeView from '@/views/HomeView.vue'

const checkAuthMock = vi.fn()
const fetchPublicSettingsMock = vi.fn()

const authState = {
  isAuthenticated: false,
  isAdmin: false,
  user: { email: 'user@example.com' }
}

const appState = {
  publicSettingsLoaded: true,
  siteName: 'Sub2API',
  siteLogo: '',
  docUrl: '',
  cachedPublicSettings: {
    site_name: 'Sub2API',
    site_logo: '',
    site_subtitle: '简单、直接、可用',
    home_content: ''
  }
}

vi.mock('@/stores', () => ({
  useAuthStore: () => ({
    ...authState,
    checkAuth: checkAuthMock
  }),
  useAppStore: () => ({
    ...appState,
    fetchPublicSettings: fetchPublicSettingsMock
  })
}))

vi.mock('vue-i18n', () => ({
  useI18n: () => ({
    t: (key: string) => {
      const messages: Record<string, string> = {
        'home.switchToLight': '切换到浅色模式',
        'home.switchToDark': '切换到深色模式',
        'home.planLabel': 'Plan',
        'home.consoleLabel': 'Console',
        'home.planTitle': '套餐',
        'home.consoleTitle': '控制台',
        'home.planDescription': '购买新套餐，或使用兑换码激活套餐。',
        'home.consoleDescription': '登录后进入后台查看额度、密钥和使用情况。',
        'home.buyPlan': '购买套餐',
        'home.redeemPlan': '兑换套餐',
        'home.loginConsole': '登录控制台',
        'home.enterConsole': '进入控制台'
      }
      return messages[key] ?? key
    }
  })
}))

vi.mock('@/components/common/LocaleSwitcher.vue', () => ({
  default: { template: '<div class="locale-switcher-stub" />' }
}))

vi.mock('@/components/icons/Icon.vue', () => ({
  default: {
    props: ['name'],
    template: '<span class="icon-stub">{{ name }}</span>'
  }
}))

function createTestRouter() {
  return createRouter({
    history: createMemoryHistory(),
    routes: [
      { path: '/', component: { template: '<div>root</div>' } },
      { path: '/login', component: { template: '<div>login</div>' } },
      { path: '/dashboard', component: { template: '<div>dashboard</div>' } },
      { path: '/admin/dashboard', component: { template: '<div>admin dashboard</div>' } },
      { path: '/redeem', component: { template: '<div>redeem</div>' } }
    ]
  })
}

describe('HomeView', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    authState.isAuthenticated = false
    authState.isAdmin = false
    authState.user = { email: 'user@example.com' }
    appState.publicSettingsLoaded = true
    appState.siteName = 'Sub2API'
    appState.siteLogo = ''
    appState.cachedPublicSettings.site_name = 'Sub2API'
    appState.cachedPublicSettings.site_logo = ''
    appState.cachedPublicSettings.site_subtitle = '简单、直接、可用'
    appState.cachedPublicSettings.home_content = ''

    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: vi.fn().mockReturnValue({
        matches: false,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn()
      })
    })
  })

  it('未登录时展示固定首页、套餐和登录控制台入口', async () => {
    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, {
      global: { plugins: [router] }
    })

    expect(wrapper.text()).toContain('Code Token')
    expect(wrapper.text()).toContain('稳定&流畅的AI编程体验。。。')
    expect(wrapper.text()).toContain('套餐')
    expect(wrapper.text()).toContain('控制台')
    expect(wrapper.text()).toContain('购买套餐')
    expect(wrapper.text()).toContain('兑换入口')
    expect(wrapper.text()).toContain('进入控制台')
  })

  it('未登录时控制台入口指向 /login，兑换入口指向 /redeem', async () => {
    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, {
      global: { plugins: [router] }
    })

    const hrefs = wrapper
      .findAll('a')
      .map((link) => link.attributes('href'))
      .filter((href): href is string => Boolean(href))

    expect(hrefs).toContain('/login')
    expect(hrefs).toContain('/redeem')
    expect(hrefs).toContain('https://pay.ldxp.cn/shop/FED14QEA')
  })

  it('普通用户已登录时控制台入口指向 /dashboard', async () => {
    authState.isAuthenticated = true

    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, {
      global: { plugins: [router] }
    })

    expect(wrapper.text()).toContain('进入控制台')
    expect(wrapper.html()).toContain('/dashboard')
  })

  it('管理员已登录时控制台入口指向 /admin/dashboard', async () => {
    authState.isAuthenticated = true
    authState.isAdmin = true

    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, {
      global: { plugins: [router] }
    })

    expect(wrapper.text()).toContain('进入控制台')
    expect(wrapper.html()).toContain('/admin/dashboard')
  })

  it('home_content 非空时仍优先使用覆盖内容', async () => {
    appState.cachedPublicSettings.home_content = '<div class="custom-home">custom</div>'

    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, {
      global: { plugins: [router] }
    })

    expect(wrapper.find('.custom-home').exists()).toBe(true)
    expect(wrapper.text()).not.toContain('购买套餐')
  })

  it('home_content 是链接时使用 iframe 模式', async () => {
    appState.cachedPublicSettings.home_content = 'https://example.com/home'

    const router = createTestRouter()
    await router.push('/')
    await router.isReady()

    const wrapper = mount(HomeView, {
      global: { plugins: [router] }
    })

    const iframe = wrapper.find('iframe')
    expect(iframe.exists()).toBe(true)
    expect(iframe.attributes('src')).toBe('https://example.com/home')
    expect(wrapper.text()).not.toContain('购买套餐')
  })
})
