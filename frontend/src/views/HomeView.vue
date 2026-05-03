<template>
  <div v-if="homeContent" class="min-h-screen">
    <iframe
      v-if="isHomeContentUrl"
      :src="homeContent.trim()"
      class="h-screen w-full border-0"
      allowfullscreen
    ></iframe>
    <div v-else v-html="homeContent"></div>
  </div>

  <div
    v-else
    class="min-h-screen bg-[radial-gradient(circle_at_top,rgba(245,158,11,0.08),transparent_35%),linear-gradient(180deg,#fafaf9_0%,#f5f5f4_100%)] text-stone-900 dark:bg-[radial-gradient(circle_at_top,rgba(245,158,11,0.12),transparent_28%),linear-gradient(180deg,#0c0a09_0%,#1c1917_100%)] dark:text-stone-100"
  >
    <header class="border-b border-stone-200/80 bg-white/80 backdrop-blur dark:border-stone-800 dark:bg-stone-950/75">
      <nav class="mx-auto flex max-w-5xl items-center justify-between px-6 py-4">
        <div class="flex items-center gap-3">
          <div class="flex h-10 w-10 items-center justify-center overflow-hidden rounded-xl bg-stone-100 dark:bg-stone-800">
            <img
              v-if="siteLogo"
              :src="siteLogo"
              alt="Logo"
              class="h-full w-full object-contain"
            />
            <span v-else class="text-sm font-semibold text-stone-500 dark:text-stone-300">
              {{ siteName.charAt(0) }}
            </span>
          </div>
          <div class="text-sm font-semibold tracking-tight text-stone-900 dark:text-stone-100">
            {{ siteName }}
          </div>
        </div>

        <div class="flex items-center gap-2">
          <LocaleSwitcher />
          <button
            @click="toggleTheme"
            class="rounded-xl border border-stone-200 bg-white p-2 text-stone-500 transition hover:text-stone-900 dark:border-stone-700 dark:bg-stone-900 dark:text-stone-300 dark:hover:text-white"
            :title="isDark ? t('home.switchToLight') : t('home.switchToDark')"
          >
            <Icon v-if="isDark" name="sun" size="md" />
            <Icon v-else name="moon" size="md" />
          </button>
        </div>
      </nav>
    </header>

    <main class="px-6 py-16 sm:py-24">
      <div class="mx-auto flex max-w-5xl flex-col gap-12">
        <section class="mx-auto max-w-2xl text-center">
          <p class="text-sm font-medium uppercase tracking-[0.24em] text-amber-700/80 dark:text-amber-300/80">
            {{ siteName }}
          </p>
          <h1 class="mt-4 text-4xl font-semibold tracking-tight text-stone-950 dark:text-white sm:text-5xl">
            {{ siteName }}
          </h1>
          <p class="mt-4 text-base leading-7 text-stone-600 dark:text-stone-300 sm:text-lg">
            {{ siteSubtitle }}
          </p>
        </section>

        <section class="grid gap-6 md:grid-cols-2">
          <article class="rounded-3xl border border-stone-200 bg-white/95 p-8 shadow-[0_24px_60px_-36px_rgba(28,25,23,0.35)] dark:border-stone-800 dark:bg-stone-950/85">
            <div class="mb-6">
              <p class="text-sm font-medium uppercase tracking-[0.2em] text-stone-500 dark:text-stone-400">
                {{ t('home.planLabel') }}
              </p>
              <h2 class="mt-3 text-2xl font-semibold text-stone-950 dark:text-white">
                {{ t('home.planTitle') }}
              </h2>
              <p class="mt-3 text-sm leading-6 text-stone-600 dark:text-stone-300">
                {{ t('home.planDescription') }}
              </p>
            </div>
            <div class="flex flex-col gap-3">
              <a
                href="https://pay.ldxp.cn/shop/FED14QEA"
                target="_blank"
                rel="noopener noreferrer"
                class="inline-flex items-center justify-center rounded-2xl bg-stone-950 px-5 py-3 text-sm font-medium text-white transition hover:bg-stone-800 dark:bg-amber-500 dark:text-stone-950 dark:hover:bg-amber-400"
              >
                {{ t('home.buyPlan') }}
              </a>
              <RouterLink
                to="/redeem"
                class="inline-flex items-center justify-center rounded-2xl border border-stone-200 bg-stone-50 px-5 py-3 text-sm font-medium text-stone-900 transition hover:bg-stone-100 dark:border-stone-700 dark:bg-stone-900 dark:text-stone-100 dark:hover:bg-stone-800"
              >
                {{ t('home.redeemPlan') }}
              </RouterLink>
            </div>
          </article>

          <article class="rounded-3xl border border-stone-200 bg-white/95 p-8 shadow-[0_24px_60px_-36px_rgba(28,25,23,0.35)] dark:border-stone-800 dark:bg-stone-950/85">
            <div class="mb-6">
              <p class="text-sm font-medium uppercase tracking-[0.2em] text-stone-500 dark:text-stone-400">
                {{ t('home.consoleLabel') }}
              </p>
              <h2 class="mt-3 text-2xl font-semibold text-stone-950 dark:text-white">
                {{ t('home.consoleTitle') }}
              </h2>
              <p class="mt-3 text-sm leading-6 text-stone-600 dark:text-stone-300">
                {{ t('home.consoleDescription') }}
              </p>
            </div>
            <RouterLink
              :to="dashboardEntryPath"
              class="inline-flex w-full items-center justify-center rounded-2xl bg-emerald-600 px-5 py-3 text-sm font-medium text-white transition hover:bg-emerald-500"
            >
              {{ isAuthenticated ? t('home.enterConsole') : t('home.loginConsole') }}
            </RouterLink>
          </article>
        </section>
      </div>
    </main>

    <footer class="border-t border-stone-200/80 px-6 py-6 dark:border-stone-800">
      <div class="mx-auto max-w-5xl text-center text-sm text-stone-500 dark:text-stone-400">
        &copy; {{ currentYear }} {{ siteName }}
      </div>
    </footer>
  </div>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useAuthStore, useAppStore } from '@/stores'
import LocaleSwitcher from '@/components/common/LocaleSwitcher.vue'
import Icon from '@/components/icons/Icon.vue'

const { t } = useI18n()

const authStore = useAuthStore()
const appStore = useAppStore()

const siteName = computed(() => appStore.cachedPublicSettings?.site_name || appStore.siteName || 'Sub2API')
const siteLogo = computed(() => appStore.cachedPublicSettings?.site_logo || appStore.siteLogo || '')
const siteSubtitle = computed(() => appStore.cachedPublicSettings?.site_subtitle || 'AI API Gateway Platform')
const homeContent = computed(() => appStore.cachedPublicSettings?.home_content || '')

const isHomeContentUrl = computed(() => {
  const content = homeContent.value.trim()
  return content.startsWith('http://') || content.startsWith('https://')
})

const isDark = ref(document.documentElement.classList.contains('dark'))
const isAuthenticated = computed(() => authStore.isAuthenticated)
const isAdmin = computed(() => authStore.isAdmin)
const dashboardEntryPath = computed(() => {
  if (!isAuthenticated.value) {
    return '/login'
  }

  return isAdmin.value ? '/admin/dashboard' : '/dashboard'
})
const currentYear = computed(() => new Date().getFullYear())

function toggleTheme() {
  isDark.value = !isDark.value
  document.documentElement.classList.toggle('dark', isDark.value)
  localStorage.setItem('theme', isDark.value ? 'dark' : 'light')
}

function initTheme() {
  const savedTheme = localStorage.getItem('theme')
  if (
    savedTheme === 'dark' ||
    (!savedTheme && window.matchMedia('(prefers-color-scheme: dark)').matches)
  ) {
    isDark.value = true
    document.documentElement.classList.add('dark')
  }
}

onMounted(() => {
  initTheme()
  authStore.checkAuth()

  if (!appStore.publicSettingsLoaded) {
    appStore.fetchPublicSettings()
  }
})
</script>
