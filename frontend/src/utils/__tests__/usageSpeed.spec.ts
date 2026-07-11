import { describe, expect, it } from 'vitest'

import { formatOutputSpeed } from '../usageSpeed'

describe('formatOutputSpeed', () => {
  it('formats output tokens per second', () => {
    expect(formatOutputSpeed(120, 2000)).toBe('60.0 t/s')
  })

  it.each([
    [0, 2000],
    [120, 0],
    [undefined, undefined],
    [120, -1],
  ])('returns a placeholder for invalid values', (tokens, duration) => {
    expect(formatOutputSpeed(tokens, duration)).toBe('-')
  })
})
