<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import { opsAPI, type OpsAccountSwitchRecord, type OpsAccountSwitchSummary } from '@/api/admin/ops'

interface Props {
  platformFilter?: string
  groupIdFilter?: number | null
  timeRange?: string
  customStartTime?: string | null
  customEndTime?: string | null
  refreshToken: number
}

const props = withDefaults(defineProps<Props>(), {
  platformFilter: '',
  groupIdFilter: null,
  timeRange: '1h',
  customStartTime: null,
  customEndTime: null
})

const { t, locale } = useI18n()

const loading = ref(false)
const errorMessage = ref('')
const summary = ref<OpsAccountSwitchSummary | null>(null)

const currentRecord = computed(() => summary.value?.current ?? null)
const recentSwitches = computed(() => summary.value?.recent_switches ?? [])

function formatWhen(value?: string | null): string {
  if (!value) return '--'
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return '--'
  return new Intl.DateTimeFormat(locale.value === 'zh' ? 'zh-CN' : 'en-US', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    hour12: false
  }).format(date)
}

function formatFromLabel(record: OpsAccountSwitchRecord): string {
  if (record.from_account_name) return record.from_account_name
  if (record.from_account_id) return `#${record.from_account_id}`
  return '--'
}

function formatToLabel(record: OpsAccountSwitchRecord): string {
  if (record.to_account_name) return record.to_account_name
  return `#${record.to_account_id}`
}

async function loadData() {
  loading.value = true
  errorMessage.value = ''
  try {
    summary.value = await opsAPI.getAccountSwitchSummary(
      props.platformFilter,
      props.groupIdFilter,
      props.timeRange,
      props.customStartTime,
      props.customEndTime
    )
  } catch (err: any) {
    console.error('[OpsAccountSwitchCard] Failed to load data', err)
    errorMessage.value = err?.response?.data?.detail || t('admin.ops.accountSwitch.loadFailed')
  } finally {
    loading.value = false
  }
}

watch(
  () => [props.refreshToken, props.platformFilter, props.groupIdFilter, props.timeRange, props.customStartTime, props.customEndTime],
  () => {
    loadData()
  },
  { immediate: true }
)
</script>

<template>
  <div class="rounded-3xl bg-white p-6 shadow-sm ring-1 ring-gray-900/5 dark:bg-dark-800 dark:ring-dark-700">
    <div class="mb-4 flex items-center justify-between gap-3">
      <h3 class="flex items-center gap-2 text-sm font-bold text-gray-900 dark:text-white">
        <svg class="h-4 w-4 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 16l4-4m0 0l-4-4m4 4H7" />
        </svg>
        {{ t('admin.ops.accountSwitch.title') }}
      </h3>
      <button
        class="flex items-center gap-1 rounded-lg bg-gray-100 px-2 py-1 text-[11px] font-semibold text-gray-700 transition-colors hover:bg-gray-200 disabled:cursor-not-allowed disabled:opacity-50 dark:bg-dark-700 dark:text-gray-300 dark:hover:bg-dark-600"
        :disabled="loading"
        :title="t('common.refresh')"
        @click="loadData"
      >
        <svg class="h-3 w-3" :class="{ 'animate-spin': loading }" fill="none" viewBox="0 0 24 24" stroke="currentColor">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
        </svg>
      </button>
    </div>

    <div v-if="errorMessage" class="mb-3 rounded-xl bg-red-50 p-2.5 text-xs text-red-600 dark:bg-red-900/20 dark:text-red-400">
      {{ errorMessage }}
    </div>

    <div class="overflow-hidden rounded-xl border border-gray-200 dark:border-dark-700">
      <div class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-gray-50 dark:bg-dark-900">
            <tr>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.ops.accountSwitch.currentAccount') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.ops.alertEvents.table.platform') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.ops.accountSwitch.lastSelectedAt') }}
              </th>
            </tr>
          </thead>
          <tbody class="bg-white dark:bg-dark-800">
            <tr>
              <td class="px-4 py-3 text-sm font-semibold text-gray-900 dark:text-white">
                {{ currentRecord ? formatToLabel(currentRecord) : t('admin.ops.accountSwitch.noCurrentAccount') }}
              </td>
              <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">
                {{ currentRecord?.platform || '--' }}
              </td>
              <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-200">
                {{ currentRecord ? formatWhen(currentRecord.switched_at) : '--' }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div class="mt-5 overflow-hidden rounded-xl border border-gray-200 dark:border-dark-700">
      <div class="flex items-center justify-between border-b border-gray-200 bg-gray-50 px-4 py-3 dark:border-dark-700 dark:bg-dark-900">
        <div class="text-xs font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
          {{ t('admin.ops.accountSwitch.recentSwitches') }}
        </div>
        <div class="text-[11px] text-gray-500 dark:text-gray-400">
          {{ t('admin.ops.accountSwitch.totalRows', { count: recentSwitches.length }) }}
        </div>
      </div>

      <div v-if="recentSwitches.length === 0" class="px-4 py-8 text-center text-sm text-gray-500 dark:text-gray-400">
        {{ t('admin.ops.accountSwitch.noSwitchHistory') }}
      </div>

      <div v-else class="overflow-x-auto">
        <table class="min-w-full divide-y divide-gray-200 dark:divide-dark-700">
          <thead class="bg-white dark:bg-dark-800">
            <tr>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('common.from') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('common.to') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.ops.alertEvents.table.platform') }}
              </th>
              <th class="px-4 py-3 text-left text-[11px] font-bold uppercase tracking-wider text-gray-500 dark:text-gray-400">
                {{ t('admin.ops.accountSwitch.lastSelectedAt') }}
              </th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200 bg-white dark:divide-dark-700 dark:bg-dark-800">
            <tr
              v-for="(item, index) in recentSwitches"
              :key="`${item.switched_at}-${item.to_account_id}-${item.from_account_id ?? 0}-${index}`"
              class="hover:bg-gray-50 dark:hover:bg-dark-700/50"
            >
              <td class="px-4 py-3 text-sm font-semibold text-gray-900 dark:text-white">
                {{ formatFromLabel(item) }}
              </td>
              <td class="px-4 py-3 text-sm font-semibold text-gray-900 dark:text-white">
                <div class="flex items-center gap-2">
                  <svg class="h-3.5 w-3.5 shrink-0 text-emerald-500" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                    <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 7l5 5m0 0l-5 5m5-5H6" />
                  </svg>
                  <span>{{ formatToLabel(item) }}</span>
                </div>
              </td>
              <td class="px-4 py-3 text-xs text-gray-500 dark:text-gray-400">
                {{ item.platform || '--' }}
              </td>
              <td class="px-4 py-3 text-sm text-gray-700 dark:text-gray-200">
                {{ formatWhen(item.switched_at) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
