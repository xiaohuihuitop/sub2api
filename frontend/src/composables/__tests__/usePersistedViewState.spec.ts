import { beforeEach, describe, expect, it, vi } from 'vitest'

import {
  isISODateString,
  isValidDateRange,
  readPersistedViewState,
  writePersistedViewState,
} from '../usePersistedViewState'

interface TestState {
  startDate: string
  granularity: 'day' | 'hour'
}

const fallback: TestState = {
  startDate: '2026-07-01',
  granularity: 'day',
}

const isTestState = (value: unknown): value is TestState => {
  if (!value || typeof value !== 'object') return false
  const state = value as Partial<TestState>
  return typeof state.startDate === 'string'
    && (state.granularity === 'day' || state.granularity === 'hour')
}

describe('persisted view state', () => {
  beforeEach(() => {
    localStorage.clear()
    vi.restoreAllMocks()
  })

  it('returns the fallback when persisted JSON is invalid', () => {
    localStorage.setItem('view', '{broken')

    expect(readPersistedViewState('view', fallback, isTestState)).toEqual(fallback)
  })

  it('returns the fallback when persisted data fails validation', () => {
    localStorage.setItem('view', JSON.stringify({ startDate: 12, granularity: 'week' }))

    expect(readPersistedViewState('view', fallback, isTestState)).toEqual(fallback)
  })

  it('round-trips a validated state', () => {
    const state: TestState = { startDate: '2026-07-10', granularity: 'hour' }

    writePersistedViewState('view', state)

    expect(readPersistedViewState('view', fallback, isTestState)).toEqual(state)
  })

  it('keeps the page usable when localStorage throws', () => {
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('storage disabled')
    })

    expect(() => writePersistedViewState('view', fallback)).not.toThrow()
  })

  it('rejects impossible ISO dates and reversed ranges', () => {
    expect(isISODateString('2026-02-29')).toBe(false)
    expect(isISODateString('2026-12-11')).toBe(true)
    expect(isValidDateRange('2026-07-11', '2026-07-10')).toBe(false)
    expect(isValidDateRange('2026-07-10', '2026-07-11')).toBe(true)
  })
})
