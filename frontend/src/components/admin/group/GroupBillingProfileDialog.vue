<template>
  <BaseDialog
    :show="show"
    :title="t('admin.groups.balanceBilling')"
    width="wide"
    @close="emit('close')"
  >
    <form class="space-y-5" @submit.prevent="save">
      <GroupBadge
        v-if="group"
        :name="group.name"
        :platform="group.platform"
        :show-rate="false"
      />

      <section class="space-y-3 border-t border-gray-200 pt-4 dark:border-dark-700">
        <label class="input-label">{{ t('admin.groups.form.rateMultiplier') }}</label>
        <input
          v-model.number="form.balance_rate_multiplier"
          data-testid="balance-rate-multiplier"
          type="number"
          min="0"
          step="0.001"
          class="input"
          required
        />
      </section>

      <section class="space-y-3 border-t border-gray-200 pt-4 dark:border-dark-700">
        <label class="flex items-center gap-2 text-sm font-medium text-gray-700 dark:text-gray-300">
          <input v-model="form.peak_rate_enabled" type="checkbox" class="checkbox" />
          {{ t('admin.groups.peakRate.enable') }}
        </label>
        <div v-if="form.peak_rate_enabled" class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <div>
            <label class="input-label">{{ t('admin.groups.peakRate.peakStart') }}</label>
            <input v-model="form.peak_start" type="time" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t('admin.groups.peakRate.peakEnd') }}</label>
            <input v-model="form.peak_end" type="time" class="input" required />
          </div>
          <div>
            <label class="input-label">{{ t('admin.groups.peakRate.peakMultiplier') }}</label>
            <input v-model.number="form.peak_rate_multiplier" type="number" min="0" step="0.001" class="input" required />
          </div>
        </div>
      </section>

      <section class="space-y-3 border-t border-gray-200 pt-4 dark:border-dark-700">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.groups.imagePricing.title') }}</label>
          <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <input v-model="form.image_rate_independent" type="checkbox" class="checkbox" />
            {{ t('admin.groups.imagePricing.independentMultiplier') }}
          </label>
        </div>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <NumberField v-model="form.image_rate_multiplier" :label="t('admin.groups.imagePricing.imageMultiplier')" required />
          <NumberField v-model="form.batch_image_discount_multiplier" :label="t('admin.groups.imagePricing.batchDiscountMultiplier')" required />
          <NumberField v-model="form.batch_image_hold_multiplier" :label="t('admin.groups.imagePricing.batchHoldMultiplier')" required />
          <NumberField v-model="form.image_price_1k" label="1K" optional />
          <NumberField v-model="form.image_price_2k" label="2K" optional />
          <NumberField v-model="form.image_price_4k" label="4K" optional />
        </div>
      </section>

      <section class="space-y-3 border-t border-gray-200 pt-4 dark:border-dark-700">
        <div class="flex flex-wrap items-center justify-between gap-3">
          <label class="text-sm font-medium text-gray-700 dark:text-gray-300">{{ t('admin.groups.videoPricing.title') }}</label>
          <label class="flex items-center gap-2 text-sm text-gray-600 dark:text-gray-300">
            <input v-model="form.video_rate_independent" type="checkbox" class="checkbox" />
            {{ t('admin.groups.videoPricing.independentMultiplier') }}
          </label>
        </div>
        <div class="grid grid-cols-1 gap-3 sm:grid-cols-3">
          <NumberField v-model="form.video_rate_multiplier" :label="t('admin.groups.videoPricing.videoMultiplier')" required />
          <NumberField v-model="form.video_price_480p" label="480p" optional />
          <NumberField v-model="form.video_price_720p" label="720p" optional />
          <NumberField v-model="form.video_price_1080p" label="1080p" optional />
        </div>
      </section>

      <section class="border-t border-gray-200 pt-4 dark:border-dark-700">
        <NumberField
          v-model="form.web_search_price_per_call"
          data-testid="web-search-price"
          :label="t('admin.groups.webSearchPricing.pricePerCall')"
          optional
        />
      </section>
    </form>

    <template #footer>
      <div class="flex justify-end gap-3">
        <button type="button" class="btn btn-secondary" @click="emit('close')">{{ t('common.cancel') }}</button>
        <button type="button" class="btn btn-primary" :disabled="saving || loading" @click="save">
          {{ saving ? t('common.saving') : t('common.save') }}
        </button>
      </div>
    </template>
  </BaseDialog>
</template>

<script setup lang="ts">
import { computed, defineComponent, h, reactive, ref, watch, type PropType } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import {
  getBillingProfile,
  updateBillingProfile,
  type BillingProfile,
  type UpdateBillingProfileRequest,
} from '@/api/admin/groups'
import type { AdminGroup } from '@/types'
import { extractApiErrorMessage } from '@/utils/apiError'
import BaseDialog from '@/components/common/BaseDialog.vue'
import GroupBadge from '@/components/common/GroupBadge.vue'

const NumberField = defineComponent({
  inheritAttrs: false,
  props: {
    modelValue: { type: [Number, String] as PropType<number | string | null>, default: null },
    label: { type: String, required: true },
    optional: Boolean,
    required: Boolean,
  },
  emits: ['update:modelValue'],
  setup(props, { emit, attrs }) {
    return () => h('div', [
      h('label', { class: 'input-label' }, props.label),
      h('input', {
        ...attrs,
        value: props.modelValue ?? '',
        type: 'number',
        min: 0,
        step: '0.001',
        class: 'input',
        required: props.required,
        onInput: (event: Event) => {
          const value = (event.target as HTMLInputElement).value
          emit('update:modelValue', value === '' ? null : Number(value))
        },
      }),
    ])
  },
})

const props = defineProps<{ show: boolean; group: AdminGroup | null }>()
const emit = defineEmits<{ close: []; saved: [profile: BillingProfile] }>()
const { t } = useI18n()
const appStore = useAppStore()
const loading = ref(false)
const saving = ref(false)

type FormState = UpdateBillingProfileRequest

const form = reactive<FormState>(defaultProfile())
const groupID = computed(() => props.group?.id ?? null)

watch([() => props.show, groupID], ([visible, id]) => {
  if (!visible || !id) return
  void load(id)
}, { immediate: true })

async function load(id: number) {
  loading.value = true
  try {
    applyProfile(await getBillingProfile(id))
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    loading.value = false
  }
}

async function save() {
  if (!groupID.value || saving.value) return
  saving.value = true
  try {
    const profile = await updateBillingProfile(groupID.value, toRequest(form))
    applyProfile(profile)
    appStore.showSuccess(t('common.saved'))
    emit('saved', profile)
    emit('close')
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('common.error')))
  } finally {
    saving.value = false
  }
}

function defaultProfile(): FormState {
  return {
    balance_rate_multiplier: 1,
    peak_rate_enabled: false,
    peak_start: '',
    peak_end: '',
    peak_rate_multiplier: 1,
    image_rate_independent: false,
    image_rate_multiplier: 1,
    image_price_1k: null,
    image_price_2k: null,
    image_price_4k: null,
    batch_image_discount_multiplier: 1,
    batch_image_hold_multiplier: 1,
    video_rate_independent: false,
    video_rate_multiplier: 1,
    video_price_480p: null,
    video_price_720p: null,
    video_price_1080p: null,
    web_search_price_per_call: null,
  }
}

function applyProfile(profile: BillingProfile): void {
  const { group_id: _groupID, ...editableProfile } = profile
  Object.assign(form, editableProfile)
}

function toRequest(profile: FormState): UpdateBillingProfileRequest {
  return {
    ...profile,
    image_price_1k: optionalNumber(profile.image_price_1k),
    image_price_2k: optionalNumber(profile.image_price_2k),
    image_price_4k: optionalNumber(profile.image_price_4k),
    video_price_480p: optionalNumber(profile.video_price_480p),
    video_price_720p: optionalNumber(profile.video_price_720p),
    video_price_1080p: optionalNumber(profile.video_price_1080p),
    web_search_price_per_call: optionalNumber(profile.web_search_price_per_call),
  }
}

function optionalNumber(value: number | null): number | null {
  return typeof value === 'number' && Number.isFinite(value) ? value : null
}
</script>
