import type { AccountPlatform, PlatformModelRule } from '@/types'
import { getModelsByPlatform, isValidWildcardPattern, type ModelMappingEntry } from '@/composables/useModelWhitelist'

export type PlatformModelMapping = ModelMappingEntry

export function splitPlatformModelRules(rules: PlatformModelRule[]): {
  allowedModels: string[]
  mappings: PlatformModelMapping[]
} {
  const allowed = new Set<string>()
  const mappingBySource = new Map<string, string>()
  for (const rule of rules) {
    if (!rule.enabled) continue
    const from = rule.model_pattern.trim()
    const to = rule.upstream_model.trim()
    if (!from) continue
    if (!to || from === to) {
      allowed.add(from)
      continue
    }
    mappingBySource.set(from, to)
  }

  for (const source of mappingBySource.keys()) allowed.delete(source)
  return {
    allowedModels: [...allowed].sort(),
    mappings: [...mappingBySource.entries()]
      .map(([from, to]) => ({ from, to }))
      .sort((left, right) => left.from.localeCompare(right.from)),
  }
}

export function buildPlatformModelRules(
  allowedModels: string[],
  mappings: PlatformModelMapping[]
): PlatformModelRule[] {
  const rules = new Map<string, PlatformModelRule>()
  for (const rawModel of allowedModels) {
    const model = rawModel.trim()
    if (!model || !isValidWildcardPattern(model)) continue
    rules.set(model, { model_pattern: model, upstream_model: model, enabled: true })
  }

  for (const mapping of mappings) {
    const from = mapping.from.trim()
    const to = mapping.to.trim()
    if (!from || !to || !isValidWildcardPattern(from) || to.includes('*')) continue
    rules.set(from, { model_pattern: from, upstream_model: to, enabled: true })
  }

  return [...rules.values()].sort((left, right) => left.model_pattern.localeCompare(right.model_pattern))
}

export function resolvePlatformModelPreset(code: string, accountPlatform: AccountPlatform): string {
  const normalizedCode = code.trim().toLowerCase()
  if (normalizedCode.startsWith('glm') || normalizedCode.startsWith('zhipu')) return 'glm'
  if (normalizedCode.startsWith('grok') || normalizedCode.startsWith('xai')) return 'grok'
  return accountPlatform
}

export { getModelsByPlatform, isValidWildcardPattern }
