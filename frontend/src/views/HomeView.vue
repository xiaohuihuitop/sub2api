<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="homeContentUrl"
      :src="homeContentUrl"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <!-- homeContent is an administrator-controlled setting. -->
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="min-h-screen bg-stone-100 text-stone-900 dark:bg-dark-950 dark:text-stone-100"
  >
    <header
      class="border-b border-stone-200/80 bg-white/85 px-4 backdrop-blur-xl dark:border-dark-800 dark:bg-dark-950/85 sm:px-6"
    >
      <nav class="mx-auto flex h-16 max-w-6xl items-center justify-between gap-4">
        <router-link :to="dashboardEntryPath" class="flex min-w-0 items-center gap-3">
          <img
            :src="siteLogo || '/logo.png'"
            :alt="siteName"
            class="h-9 w-9 shrink-0 rounded-lg object-contain shadow-sm"
          />
          <span class="truncate text-sm font-semibold text-stone-900 dark:text-white sm:text-base">
            {{ siteName }}
          </span>
        </router-link>

        <div class="flex shrink-0 items-center gap-2">
          <LocaleSwitcher />
          <a
            v-if="docUrl"
            :href="docUrl"
            target="_blank"
            rel="noopener noreferrer"
            class="inline-flex h-9 w-9 items-center justify-center rounded-lg text-stone-500 transition-colors hover:bg-stone-100 hover:text-primary-700 dark:text-dark-300 dark:hover:bg-dark-800 dark:hover:text-primary-300"
            :title="t('home.viewDocs')"
          >
            <Icon name="book" size="md" />
          </a>
          <router-link
            :to="dashboardEntryPath"
            class="inline-flex h-9 items-center gap-2 rounded-lg bg-stone-900 px-3 text-xs font-semibold text-white transition-colors hover:bg-stone-700 dark:bg-primary-500 dark:hover:bg-primary-400 sm:text-sm"
          >
            <span
              v-if="isAuthenticated"
              class="flex h-5 w-5 items-center justify-center rounded-full bg-primary-500 text-[10px] text-white dark:bg-stone-900"
            >
              {{ userInitial }}
            </span>
            {{ isAuthenticated ? t('home.dashboard') : t('home.login') }}
          </router-link>
        </div>
      </nav>
    </header>

    <main>
      <section class="relative overflow-hidden border-b border-stone-200/80 dark:border-dark-800">
        <div class="pointer-events-none absolute inset-0 bg-mesh-gradient opacity-80 dark:opacity-50"></div>
        <div class="relative mx-auto max-w-6xl px-4 pb-12 pt-14 sm:px-6 md:pb-16 md:pt-20">
          <div class="mx-auto max-w-3xl text-center">
            <p class="mb-4 text-xs font-semibold uppercase text-primary-700 dark:text-primary-300">
              {{ t('home.custom.eyebrow') }}
            </p>
            <h1 class="text-4xl font-bold text-stone-950 dark:text-white sm:text-5xl md:text-6xl">
              {{ siteName }}
            </h1>
            <p class="mx-auto mt-5 max-w-2xl text-base leading-7 text-stone-600 dark:text-dark-300 sm:text-lg">
              {{ siteSubtitle }}
            </p>
            <p class="mx-auto mt-2 max-w-2xl text-sm leading-6 text-stone-500 dark:text-dark-400">
              {{ t('home.custom.heroDescription') }}
            </p>
            <div class="mt-8 flex flex-wrap items-center justify-center gap-3">
              <router-link :to="dashboardEntryPath" class="btn btn-primary px-6 py-3">
                {{ isAuthenticated ? t('home.goToDashboard') : t('home.getStarted') }}
                <Icon name="arrowRight" size="sm" class="ml-1" :stroke-width="2" />
              </router-link>
              <a href="#tutorial" class="btn btn-secondary px-6 py-3">
                {{ t('home.custom.tutorialLink') }}
              </a>
            </div>
          </div>
        </div>
      </section>

      <section class="border-b border-stone-200/80 bg-stone-100/70 dark:border-dark-800 dark:bg-dark-950">
        <div class="mx-auto grid max-w-6xl gap-4 px-4 py-10 sm:px-6 md:grid-cols-2">
            <article
              class="rounded-lg border border-stone-200 bg-white/90 p-5 shadow-card backdrop-blur-sm dark:border-dark-700 dark:bg-dark-900/80 sm:p-6"
            >
              <div class="flex items-start gap-4">
                <div class="stat-icon stat-icon-primary shrink-0">
                  <Icon name="creditCard" size="md" :stroke-width="2" />
                </div>
                <div class="min-w-0">
                  <p class="text-xs font-semibold text-primary-700 dark:text-primary-300">
                    {{ t('home.custom.package') }}
                  </p>
                  <h2 class="mt-1 text-xl font-semibold text-stone-900 dark:text-white">
                    {{ t('home.custom.packageTitle') }}
                  </h2>
                  <p class="mt-2 text-sm leading-6 text-stone-600 dark:text-dark-300">
                    {{ t('home.custom.packageDescription') }}
                  </p>
                </div>
              </div>
              <div class="mt-5 grid grid-cols-2 gap-3">
                <a
                  href="https://pay.ldxp.cn/shop/FED14QEA"
                  target="_blank"
                  rel="noopener noreferrer"
                  class="btn btn-primary w-full"
                >
                  {{ t('home.custom.buyPackage') }}
                </a>
                <router-link to="/redeem" class="btn btn-secondary w-full">
                  {{ t('home.custom.redeem') }}
                </router-link>
              </div>
            </article>

            <article
              class="rounded-lg border border-stone-200 bg-white/90 p-5 shadow-card backdrop-blur-sm dark:border-dark-700 dark:bg-dark-900/80 sm:p-6"
            >
              <div class="flex items-start gap-4">
                <div class="stat-icon shrink-0 bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-300">
                  <Icon name="key" size="md" :stroke-width="2" />
                </div>
                <div class="min-w-0">
                  <p class="text-xs font-semibold text-blue-700 dark:text-blue-300">
                    {{ t('home.custom.console') }}
                  </p>
                  <h2 class="mt-1 text-xl font-semibold text-stone-900 dark:text-white">
                    {{ t('home.custom.consoleTitle') }}
                  </h2>
                  <p class="mt-2 text-sm leading-6 text-stone-600 dark:text-dark-300">
                    {{ t('home.custom.consoleDescription') }}
                  </p>
                </div>
              </div>
              <router-link :to="dashboardEntryPath" class="btn btn-secondary mt-5 w-full">
                {{ t('home.custom.openConsole') }}
                <Icon name="arrowRight" size="sm" class="ml-1" :stroke-width="2" />
              </router-link>
            </article>
          </div>
      </section>

      <section id="tutorial" class="mx-auto max-w-6xl scroll-mt-6 px-4 py-16 sm:px-6 md:py-20">
        <div class="mx-auto max-w-2xl text-center">
          <p class="text-xs font-semibold uppercase text-primary-700 dark:text-primary-300">
            {{ t('home.custom.tutorialEyebrow') }}
          </p>
          <h2 class="mt-3 text-3xl font-bold text-stone-950 dark:text-white">
            {{ t('home.custom.tutorialTitle') }}
          </h2>
          <p class="mt-4 text-sm leading-7 text-stone-600 dark:text-dark-300">
            {{ t('home.custom.tutorialDescription') }}
          </p>
          <div class="mt-6 flex flex-wrap justify-center gap-3">
            <a
              href="https://www.codex-docs.com/using-codex/app/"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary"
            >
              {{ t('home.custom.downloadCodex') }}
            </a>
            <a
              href="https://github.com/farion1231/cc-switch/releases/download/v3.10.3/CC-Switch-v3.10.3-Windows-Portable.zip"
              target="_blank"
              rel="noopener noreferrer"
              class="btn btn-secondary"
            >
              {{ t('home.custom.downloadSwitch') }}
            </a>
          </div>
        </div>

        <div class="mt-10 space-y-5">
          <article
            class="grid overflow-hidden rounded-lg border border-stone-200 bg-white shadow-card dark:border-dark-700 dark:bg-dark-900 md:grid-cols-[0.9fr_1.1fr]"
          >
            <div class="p-6 md:p-8">
              <span class="step-badge">1</span>
              <h3 class="mt-4 text-xl font-semibold text-stone-900 dark:text-white">
                {{ t('home.custom.stepOneTitle') }}
              </h3>
              <p class="mt-3 text-sm leading-7 text-stone-600 dark:text-dark-300">
                {{ t('home.custom.stepOneDescription') }}
              </p>
            </div>
            <div class="border-t border-stone-200 bg-stone-50 p-3 dark:border-dark-700 dark:bg-dark-950 md:border-l md:border-t-0">
              <img
                src="/dis2.png"
                :alt="t('home.custom.stepOneTitle')"
                class="h-full min-h-52 w-full rounded-lg object-contain"
              />
            </div>
          </article>

          <article
            class="grid overflow-hidden rounded-lg border border-stone-200 bg-white shadow-card dark:border-dark-700 dark:bg-dark-900 md:grid-cols-[1.1fr_0.9fr]"
          >
            <div class="border-b border-stone-200 bg-stone-50 p-3 dark:border-dark-700 dark:bg-dark-950 md:border-b-0 md:border-r">
              <img
                src="/dis1.png"
                :alt="t('home.custom.stepTwoTitle')"
                class="h-full min-h-52 w-full rounded-lg object-contain"
              />
            </div>
            <div class="p-6 md:p-8">
              <span class="step-badge">2</span>
              <h3 class="mt-4 text-xl font-semibold text-stone-900 dark:text-white">
                {{ t('home.custom.stepTwoTitle') }}
              </h3>
              <p class="mt-3 text-sm leading-7 text-stone-600 dark:text-dark-300">
                {{ t('home.custom.stepTwoDescription') }}
              </p>
              <dl class="mt-5 space-y-3 text-sm">
                <div class="flex flex-col gap-1 sm:flex-row sm:justify-between sm:gap-4">
                  <dt class="text-stone-500 dark:text-dark-400">{{ t('home.custom.providerName') }}</dt>
                  <dd class="font-mono text-stone-900 dark:text-white">x</dd>
                </div>
                <div class="flex flex-col gap-1 sm:flex-row sm:justify-between sm:gap-4">
                  <dt class="text-stone-500 dark:text-dark-400">{{ t('home.custom.apiAddress') }}</dt>
                  <dd class="break-all font-mono text-stone-900 dark:text-white">https://api.xhhtop.top/v1</dd>
                </div>
              </dl>
            </div>
          </article>
        </div>

        <div class="mt-5 overflow-hidden rounded-lg border border-stone-800 bg-stone-950 shadow-card">
          <div class="border-b border-stone-800 px-5 py-3 text-sm font-medium text-stone-200">
            {{ t('home.custom.configTitle') }}
          </div>
          <pre class="overflow-x-auto p-5 text-xs leading-6 text-stone-300 sm:text-sm"><code>model_provider = "custom"
model = "gpt-5.5"
model_reasoning_effort = "high"
disable_response_storage = true
service_tier = "fast"

[model_providers.custom]
name = "custom"
wire_api = "responses"
requires_openai_auth = true
base_url = "https://api.xhhtop.top/v1"</code></pre>
        </div>
      </section>
    </main>

    <footer class="border-t border-stone-200 bg-white/70 px-4 py-7 dark:border-dark-800 dark:bg-dark-950 sm:px-6">
      <div class="mx-auto flex max-w-6xl flex-col gap-3 text-sm text-stone-500 dark:text-dark-400 sm:flex-row sm:items-center sm:justify-between">
        <p>&copy; {{ currentYear }} {{ siteName }}. {{ t('home.footer.allRightsReserved') }}</p>
        <div class="flex items-center gap-4">
          <a v-if="docUrl" :href="docUrl" target="_blank" rel="noopener noreferrer">{{ t('home.docs') }}</a>
          <a :href="githubUrl" target="_blank" rel="noopener noreferrer">GitHub</a>
        </div>
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'
import { sanitizeUrl } from '@/utils/url'

const { t } = useI18n()
const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '', {
  allowRelative: true,
  allowDataUrl: true,
}))
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const docUrl = computed(() => sanitizeUrl(appStore.cachedPublicSettings?.doc_url || appStore.docUrl || ''))
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')
const homeContentUrl = computed(() => {
  const value = homeContent.value.trim()
  if (!value.startsWith('http://') && !value.startsWith('https://')) return ''
  return sanitizeUrl(value)
})

const githubUrl = 'https://github.com/Wei-Shaw/sub2api'
const isAuthenticated = computed(() => authStore.isAuthenticated)
const dashboardPath = computed(() => authStore.isAdmin ? '/admin/dashboard' : '/dashboard')
const dashboardEntryPath = computed(() => isAuthenticated.value ? dashboardPath.value : '/login')
const userInitial = computed(() => authStore.user?.email?.charAt(0).toUpperCase() || '')
const currentYear = computed(() => new Date().getFullYear())

onMounted(() => {
  authStore.checkAuth()
  if (!appStore.publicSettingsLoaded) appStore.fetchPublicSettings()
})
</script>

<style scoped>
.step-badge {
  display: inline-flex;
  height: 2rem;
  width: 2rem;
  align-items: center;
  justify-content: center;
  border-radius: 9999px;
  background: linear-gradient(135deg, #f59e0b, #b45309);
  color: white;
  font-size: 0.875rem;
  font-weight: 700;
}
</style>
