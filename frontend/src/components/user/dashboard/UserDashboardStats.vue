<template>
  <div class="space-y-4">
    <div class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
      <div v-if="!isSimple" class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-emerald-100 p-2 dark:bg-emerald-900/30">
            <svg
              class="h-5 w-5 text-emerald-600 dark:text-emerald-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2"
                d="M2.25 18.75a60.07 60.07 0 0115.797 2.101c.727.198 1.453-.342 1.453-1.096V18.75M3.75 4.5v.75A.75.75 0 013 6h-.75m0 0v-.375c0-.621.504-1.125 1.125-1.125H20.25M2.25 6v9m18-10.5v.75c0 .414.336.75.75.75h.75m-1.5-1.5h.375c.621 0 1.125.504 1.125 1.125v9.75c0 .621-.504 1.125-1.125 1.125h-.375m1.5-1.5H21a.75.75 0 00-.75.75v.75m0 0H3.75m0 0h-.375a1.125 1.125 0 01-1.125-1.125V15m1.5 1.5v-.75A.75.75 0 003 15h-.75M15 10.5a3 3 0 11-6 0 3 3 0 016 0zm3 0h.008v.008H18V10.5zm-12 0h.008v.008H6V10.5z"
              />
            </svg>
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('dashboard.balance') }}
            </p>
            <p class="text-xl font-bold text-emerald-600 dark:text-emerald-400">
              ${{ formatBalance(balance) }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('common.available') }}</p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-purple-100 p-2 dark:bg-purple-900/30">
            <Icon
              name="dollar"
              size="md"
              class="text-purple-600 dark:text-purple-400"
              :stroke-width="2"
            />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('dashboard.todayCost') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              <span class="text-purple-600 dark:text-purple-400" :title="t('dashboard.actual')">
                ${{ formatCost(stats?.today_actual_cost || 0) }}
              </span>
              <span
                class="text-sm font-normal text-gray-400 dark:text-gray-500"
                :title="t('dashboard.standard')"
              >
                / ${{ formatCost(stats?.today_cost || 0) }}
              </span>
            </p>
            <p class="text-xs">
              <span class="text-gray-500 dark:text-gray-400">{{ t('common.total') }}: </span>
              <span class="text-purple-600 dark:text-purple-400" :title="t('dashboard.actual')">
                ${{ formatCost(stats?.total_actual_cost || 0) }}
              </span>
              <span class="text-gray-400 dark:text-gray-500" :title="t('dashboard.standard')">
                / ${{ formatCost(stats?.total_cost || 0) }}
              </span>
            </p>
          </div>
        </div>
      </div>

      <div class="card p-4">
        <div class="flex items-center gap-3">
          <div class="rounded-lg bg-amber-100 p-2 dark:bg-amber-900/30">
            <Icon
              name="cube"
              size="md"
              class="text-amber-600 dark:text-amber-400"
              :stroke-width="2"
            />
          </div>
          <div>
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('dashboard.todayTokens') }}
            </p>
            <p class="text-xl font-bold text-gray-900 dark:text-white">
              {{ formatTokens(stats?.today_tokens || 0) }}
            </p>
            <p class="text-xs text-gray-500 dark:text-gray-400">
              {{ t('dashboard.input') }}: {{ formatTokens(stats?.today_input_tokens || 0) }} /
              {{ t('dashboard.output') }}: {{ formatTokens(stats?.today_output_tokens || 0) }}
            </p>
          </div>
        </div>
      </div>
    </div>

    <div
      v-if="displayedSubscriptions.length > 0"
      class="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4"
    >
      <div
        v-for="subscription in displayedSubscriptions"
        :key="subscription.id"
        class="card p-4"
      >
        <div class="flex items-start gap-3">
          <div class="rounded-lg bg-primary-100 p-2 dark:bg-primary-900/30">
            <Icon
              name="creditCard"
              size="md"
              class="text-primary-600 dark:text-primary-400"
              :stroke-width="2"
            />
          </div>
          <div class="min-w-0 flex-1">
            <p class="text-xs font-medium text-gray-500 dark:text-gray-400">
              {{ t('nav.mySubscriptions') }}
            </p>
            <div class="mt-0.5 flex items-start justify-between gap-3">
              <p class="min-w-0 truncate text-base font-bold text-gray-900 dark:text-white">
                {{ subscription.group?.name || `Group #${subscription.group_id}` }}
              </p>
              <span
                v-if="primaryUsageMetric(subscription)"
                class="shrink-0 text-right text-[11px] text-gray-500 dark:text-gray-400"
              >
                {{ t('userSubscriptions.resetIn', { time: formatResetTime(subscription) }) }}
              </span>
            </div>
            <p
              v-if="subscription.expires_at"
              :class="['mt-1 text-xs', expirationClass(subscription.expires_at)]"
            >
              {{ formatExpiration(subscription.expires_at) }}
            </p>
            <p v-else class="mt-1 text-xs text-gray-500 dark:text-gray-400">
              {{ t('userSubscriptions.noExpiration') }}
            </p>

            <div v-if="primaryUsageMetric(subscription)" class="mt-3 space-y-1.5">
              <div class="flex items-center justify-between gap-3">
                <span class="text-xs font-medium text-gray-600 dark:text-gray-300">
                  {{ primaryUsageMetric(subscription)?.label }}
                </span>
                <span class="text-xs text-gray-500 dark:text-gray-400">
                  {{ primaryUsageMetric(subscription)?.usageText }}
                </span>
              </div>
              <div
                class="relative h-2 overflow-hidden rounded-full bg-emerald-500/25 dark:bg-emerald-400/25"
              >
                <div
                  class="absolute inset-y-0 left-0 rounded-full bg-red-200 dark:bg-red-400/35"
                  :style="{ width: getUsedProgressWidth(subscription) }"
                ></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import Icon from '@/components/icons/Icon.vue'
import type { UserDashboardStats as UserStatsType } from '@/api/usage'
import type { UserSubscription } from '@/types'

interface UsageMetric {
  label: string
  usageText: string
  used: number
  limit: number
  remaining: number
}

const props = defineProps<{
  stats: UserStatsType
  balance: number
  isSimple: boolean
  subscriptions: UserSubscription[]
}>()

const { t } = useI18n()
const displayedSubscriptions = computed(() => props.subscriptions.slice(0, 4))

const formatBalance = (b: number) =>
  new Intl.NumberFormat('en-US', {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2
  }).format(b)

const formatCost = (c: number) => c.toFixed(4)

const formatTokens = (value: number) => {
  if (value >= 1_000_000) return `${(value / 1_000_000).toFixed(1)}M`
  if (value >= 1000) return `${(value / 1000).toFixed(1)}K`
  return value.toString()
}

function primaryUsageMetric(subscription: UserSubscription): UsageMetric | null {
  const group = subscription.group

  if (group?.daily_limit_usd) {
    const limit = group.daily_limit_usd
    const used = subscription.daily_usage_usd || 0
    const remaining = Math.max(limit - used, 0)
    return {
      label: t('userSubscriptions.daily'),
      usageText: `$${limit.toFixed(2)} / $${remaining.toFixed(2)}`,
      used,
      limit,
      remaining
    }
  }

  if (group?.weekly_limit_usd) {
    const limit = group.weekly_limit_usd
    const used = subscription.weekly_usage_usd || 0
    const remaining = Math.max(limit - used, 0)
    return {
      label: t('userSubscriptions.weekly'),
      usageText: `$${limit.toFixed(2)} / $${remaining.toFixed(2)}`,
      used,
      limit,
      remaining
    }
  }

  if (group?.monthly_limit_usd) {
    const limit = group.monthly_limit_usd
    const used = subscription.monthly_usage_usd || 0
    const remaining = Math.max(limit - used, 0)
    return {
      label: t('userSubscriptions.monthly'),
      usageText: `$${limit.toFixed(2)} / $${remaining.toFixed(2)}`,
      used,
      limit,
      remaining
    }
  }

  return null
}

function getUsedProgressWidth(subscription: UserSubscription): string {
  const metric = primaryUsageMetric(subscription)
  if (!metric || metric.limit <= 0) return '0%'
  return `${Math.min((metric.used / metric.limit) * 100, 100)}%`
}

function getWindowStart(subscription: UserSubscription): string | null {
  const group = subscription.group
  if (group?.daily_limit_usd) return subscription.daily_window_start || null
  if (group?.weekly_limit_usd) return subscription.weekly_window_start || null
  if (group?.monthly_limit_usd) return subscription.monthly_window_start || null
  return null
}

function getWindowHours(subscription: UserSubscription): number {
  const group = subscription.group
  if (group?.daily_limit_usd) return 24
  if (group?.weekly_limit_usd) return 168
  if (group?.monthly_limit_usd) return 720
  return 0
}

function formatResetTime(subscription: UserSubscription): string {
  const windowStart = getWindowStart(subscription)
  const windowHours = getWindowHours(subscription)

  if (!windowStart || windowHours <= 0) {
    return t('userSubscriptions.windowNotActive')
  }

  const start = new Date(windowStart)
  const end = new Date(start.getTime() + windowHours * 60 * 60 * 1000)
  const now = new Date()
  const diff = end.getTime() - now.getTime()

  if (diff <= 0) {
    return t('userSubscriptions.windowNotActive')
  }

  const hours = Math.floor(diff / (1000 * 60 * 60))
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60))

  if (hours > 24) {
    const days = Math.floor(hours / 24)
    const remainingHours = hours % 24
    return `${days}d ${remainingHours}h`
  }

  if (hours > 0) {
    return `${hours}h ${minutes}m`
  }

  return `${minutes}m`
}

function formatExpiration(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days < 0) return t('userSubscriptions.status.expired')
  if (days === 0) return t('common.today')
  if (days === 1) return t('common.tomorrow')
  return t('userSubscriptions.daysRemaining', { days })
}

function expirationClass(expiresAt: string): string {
  const now = new Date()
  const expires = new Date(expiresAt)
  const diff = expires.getTime() - now.getTime()
  const days = Math.ceil(diff / (1000 * 60 * 60 * 24))

  if (days <= 0) return 'text-red-600 dark:text-red-400 font-medium'
  if (days <= 3) return 'text-red-600 dark:text-red-400'
  if (days <= 7) return 'text-orange-600 dark:text-orange-400'
  return 'text-gray-500 dark:text-gray-400'
}
</script>
