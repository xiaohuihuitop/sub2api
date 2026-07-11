export type PersistedGranularity = 'day' | 'hour'

export interface PersistedDateRange {
  startDate: string
  endDate: string
  granularity?: PersistedGranularity
}

const DATE_PATTERN = /^\d{4}-\d{2}-\d{2}$/

const isValidDate = (value: unknown): value is string => {
  if (typeof value !== 'string' || !DATE_PATTERN.test(value)) return false

  const [year, month, day] = value.split('-').map(Number)
  const date = new Date(year, month - 1, day)
  return date.getFullYear() === year && date.getMonth() === month - 1 && date.getDate() === day
}

const isValidGranularity = (value: unknown): value is PersistedGranularity =>
  value === 'day' || value === 'hour'

export function getPersistedDateRange(
  storageKey: string,
  fallback: PersistedDateRange,
): PersistedDateRange {
  if (typeof window === 'undefined') return fallback

  try {
    const raw = window.localStorage.getItem(storageKey)
    if (!raw) return fallback

    const saved: unknown = JSON.parse(raw)
    if (!saved || typeof saved !== 'object') return fallback

    const { startDate, endDate, granularity } = saved as Partial<PersistedDateRange>
    if (!isValidDate(startDate) || !isValidDate(endDate) || startDate > endDate) return fallback

    const restored: PersistedDateRange = { startDate, endDate }
    if (fallback.granularity) {
      restored.granularity = isValidGranularity(granularity) ? granularity : fallback.granularity
    }
    return restored
  } catch (error) {
    console.warn('Failed to read persisted date range:', error)
    return fallback
  }
}

export function setPersistedDateRange(storageKey: string, value: PersistedDateRange): void {
  if (typeof window === 'undefined') return
  if (!isValidDate(value.startDate) || !isValidDate(value.endDate) || value.startDate > value.endDate) {
    return
  }

  try {
    window.localStorage.setItem(storageKey, JSON.stringify(value))
  } catch (error) {
    console.warn('Failed to persist date range:', error)
  }
}
