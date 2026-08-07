import { describe, expect, it } from 'vitest'
import { buildPlatformModelRules, resolvePlatformModelPreset, splitPlatformModelRules } from '../platformModelRules'

describe('platformModelRules', () => {
  it('splits self mappings from explicit mappings', () => {
    expect(splitPlatformModelRules([
      { model_pattern: 'gpt-5.6', upstream_model: 'gpt-5.6', enabled: true },
      { model_pattern: 'gpt-latest', upstream_model: 'gpt-5.6', enabled: true },
    ])).toEqual({
      allowedModels: ['gpt-5.6'],
      mappings: [{ from: 'gpt-latest', to: 'gpt-5.6' }],
    })
  })

  it('lets an explicit mapping replace a same-name self mapping', () => {
    expect(buildPlatformModelRules(
      ['gpt-latest'],
      [{ from: 'gpt-latest', to: 'gpt-5.6' }],
    )).toEqual([
      { model_pattern: 'gpt-latest', upstream_model: 'gpt-5.6', enabled: true },
    ])
  })

  it('uses GLM presets for an OpenAI-compatible GLM platform', () => {
    expect(resolvePlatformModelPreset('glm-primary', 'openai')).toBe('glm')
  })
})
