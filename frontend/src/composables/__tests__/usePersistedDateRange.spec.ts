import { beforeEach, describe, expect, it } from 'vitest'

import { getPersistedDateRange, setPersistedDateRange } from '../usePersistedDateRange'

const storageKey = 'test-date-range'
const fallback = {
  startDate: '2026-07-01',
  endDate: '2026-07-11',
  granularity: 'day' as const,
}

describe('usePersistedDateRange', () => {
  beforeEach(() => {
    localStorage.clear()
  })

  it('restores a saved date range and granularity after a refresh', () => {
    setPersistedDateRange(storageKey, {
      startDate: '2026-07-05',
      endDate: '2026-07-10',
      granularity: 'hour',
    })

    expect(getPersistedDateRange(storageKey, fallback)).toEqual({
      startDate: '2026-07-05',
      endDate: '2026-07-10',
      granularity: 'hour',
    })
  })

  it('falls back when saved dates are invalid or reversed', () => {
    localStorage.setItem(storageKey, JSON.stringify({
      startDate: '2026-07-12',
      endDate: '2026-07-10',
      granularity: 'hour',
    }))

    expect(getPersistedDateRange(storageKey, fallback)).toEqual(fallback)
  })
})
