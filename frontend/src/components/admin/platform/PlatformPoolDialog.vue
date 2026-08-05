<template>
  <BaseDialog
    :show="show"
    :title="platform ? t('admin.platforms.edit') : t('admin.platforms.create')"
    width="wide"
    @close="emit('close')"
  >
    <form class="space-y-5" @submit.prevent="submit">
      <div class="grid gap-4 md:grid-cols-2">
        <label class="space-y-1.5">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('admin.platforms.code') }}</span>
          <input
            v-model="form.code"
            data-test="platform-code"
            class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100"
            autocomplete="off"
            maxlength="50"
            required
          />
        </label>
        <label class="space-y-1.5">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('admin.platforms.name') }}</span>
          <input
            v-model="form.name"
            data-test="platform-name"
            class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100"
            maxlength="100"
            required
          />
        </label>
        <label class="space-y-1.5">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('admin.platforms.accountPlatform') }}</span>
          <select
            v-model="form.account_platform"
            data-test="platform-account-platform"
            class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100"
            required
          >
            <option v-for="option in accountPlatformOptions" :key="option.value" :value="option.value">
              {{ option.label }}
            </option>
          </select>
        </label>
        <label class="space-y-1.5">
          <span class="text-sm font-medium text-gray-700 dark:text-gray-200">{{ t('admin.platforms.status') }}</span>
          <select
            v-model="form.status"
            data-test="platform-status"
            class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100"
          >
            <option value="active">{{ t('common.active') }}</option>
            <option value="disabled">{{ t('common.disabled') }}</option>
          </select>
        </label>
      </div>

      <section class="space-y-3 border-t border-gray-200 pt-5 dark:border-dark-700">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h3 class="text-sm font-semibold text-gray-900 dark:text-gray-100">{{ t('admin.platforms.modelRules') }}</h3>
            <p class="mt-1 text-xs text-gray-500 dark:text-gray-400">{{ t('admin.platforms.modelRulesHint') }}</p>
          </div>
          <button type="button" class="btn btn-secondary gap-1.5" data-test="add-model-rule" @click="addRule">
            <Icon name="plus" size="sm" />
            <span>{{ t('admin.platforms.addRule') }}</span>
          </button>
        </div>

        <div v-if="form.model_rules.length === 0" class="rounded-md border border-dashed border-gray-300 px-4 py-5 text-sm text-gray-500 dark:border-dark-600 dark:text-gray-400">
          {{ t('admin.platforms.noRules') }}
        </div>

        <div v-for="(rule, index) in form.model_rules" :key="rule.key" class="rounded-md border border-gray-200 p-4 dark:border-dark-700">
          <div class="flex items-start justify-between gap-3">
            <div class="grid min-w-0 flex-1 gap-3 md:grid-cols-2">
              <label class="space-y-1.5">
                <span class="text-xs font-medium text-gray-600 dark:text-gray-300">{{ t('admin.platforms.modelPattern') }}</span>
                <input
                  v-model="rule.model_pattern"
                  :data-test="`model-pattern-${index}`"
                  class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 font-mono text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100"
                  :placeholder="t('admin.platforms.modelPatternPlaceholder')"
                  required
                />
              </label>
              <label class="space-y-1.5">
                <span class="text-xs font-medium text-gray-600 dark:text-gray-300">{{ t('admin.platforms.upstreamModel') }}</span>
                <input
                  v-model="rule.upstream_model"
                  :data-test="`upstream-model-${index}`"
                  class="w-full rounded-md border border-gray-300 bg-white px-3 py-2 font-mono text-sm text-gray-900 outline-none transition focus:border-primary-500 focus:ring-2 focus:ring-primary-500/20 dark:border-dark-600 dark:bg-dark-800 dark:text-gray-100"
                  :placeholder="t('admin.platforms.upstreamModelPlaceholder')"
                />
              </label>
            </div>
            <button
              type="button"
              class="icon-button text-gray-500 hover:bg-red-50 hover:text-red-600 dark:hover:bg-red-900/20 dark:hover:text-red-300"
              :title="t('common.delete')"
              :aria-label="t('common.delete')"
              @click="removeRule(index)"
            >
              <Icon name="trash" size="sm" />
            </button>
          </div>

          <div class="mt-3 flex flex-wrap items-center gap-x-5 gap-y-2">
            <span class="text-xs font-medium text-gray-600 dark:text-gray-300">{{ t('admin.platforms.endpointCapabilities') }}</span>
            <label v-for="endpoint in endpointOptions" :key="endpoint.value" class="flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200">
              <input
                v-model="rule.endpoint_capabilities"
                type="checkbox"
                :value="endpoint.value"
                :data-test="`endpoint-${endpoint.value}-${index}`"
                class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-800"
              />
              <span>{{ endpoint.label }}</span>
            </label>
            <label class="ml-auto flex items-center gap-2 text-sm text-gray-700 dark:text-gray-200">
              <input v-model="rule.enabled" type="checkbox" class="h-4 w-4 rounded border-gray-300 text-primary-600 focus:ring-primary-500 dark:border-dark-600 dark:bg-dark-800" />
              <span>{{ t('common.enabled') }}</span>
            </label>
          </div>
        </div>
      </section>

      <p v-if="validationError" class="text-sm text-red-600 dark:text-red-400">{{ validationError }}</p>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" :disabled="submitting" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="submitting" data-test="save-platform" @click="submit">
          {{ submitting ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, reactive, watch } from 'vue'
import { useI18n } from 'vue-i18n'
import BaseDialog from '@/components/common/BaseDialog.vue'
import Icon from '@/components/icons/Icon.vue'
import type { AccountPlatform, CreatePlatformPoolRequest, PlatformPool } from '@/types'

type ModelRuleForm = {
  key: string
  model_pattern: string
  upstream_model: string
  endpoint_capabilities: string[]
  enabled: boolean
}

const props = defineProps<{
  show: boolean
  platform: PlatformPool | null
  submitting: boolean
}>()

const emit = defineEmits<{
  close: []
  save: [input: CreatePlatformPoolRequest]
}>()

const { t } = useI18n()
const accountPlatformOptions: Array<{ value: AccountPlatform; label: string }> = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'grok', label: 'Grok' },
  { value: 'antigravity', label: 'Antigravity' },
]
const endpointOptions = [
  { value: 'chat_completions', label: 'Chat Completions' },
  { value: 'responses', label: 'Responses' },
]

const form = reactive({
  code: '',
  name: '',
  account_platform: 'openai' as AccountPlatform,
  status: 'active' as 'active' | 'disabled',
  model_rules: [] as ModelRuleForm[],
})
const validationError = computed(() => {
  if (!form.code.trim() || !form.name.trim()) return t('admin.platforms.requiredFields')
  if (!/^[a-z0-9_-]+$/.test(form.code.trim().toLowerCase())) return t('admin.platforms.invalidCode')
  if (form.model_rules.some(rule => !rule.model_pattern.trim() || rule.endpoint_capabilities.length === 0)) {
    return t('admin.platforms.invalidRule')
  }
  return ''
})

function newRule(rule?: PlatformPool['model_rules'][number]): ModelRuleForm {
  return {
    key: `${Date.now()}-${Math.random()}`,
    model_pattern: rule?.model_pattern ?? '',
    upstream_model: rule?.upstream_model ?? '',
    endpoint_capabilities: [...(rule?.endpoint_capabilities ?? [])],
    enabled: rule?.enabled ?? true,
  }
}

function resetForm() {
  form.code = props.platform?.code ?? ''
  form.name = props.platform?.name ?? ''
  form.account_platform = props.platform?.account_platform ?? 'openai'
  form.status = props.platform?.status ?? 'active'
  form.model_rules = (props.platform?.model_rules ?? []).map(newRule)
}

function addRule() {
  form.model_rules.push(newRule())
}

function removeRule(index: number) {
  form.model_rules.splice(index, 1)
}

function submit() {
  if (validationError.value) return
  emit('save', {
    code: form.code.trim().toLowerCase(),
    name: form.name.trim(),
    account_platform: form.account_platform,
    status: form.status,
    model_rules: form.model_rules.map(rule => ({
      model_pattern: rule.model_pattern.trim(),
      upstream_model: rule.upstream_model.trim(),
      endpoint_capabilities: [...rule.endpoint_capabilities],
      enabled: rule.enabled,
    })),
  })
}

watch(() => [props.show, props.platform] as const, ([show]) => {
  if (show) resetForm()
}, { immediate: true })
</script>
