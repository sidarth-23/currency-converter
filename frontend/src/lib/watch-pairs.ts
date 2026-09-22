import { z } from "zod"

export type WatchPair = {
  base: string
  target: string
}

const storageKey = "currency-watcher:watch-pairs"
const WatchPairRecord = z.object({
  base: z.string(),
  target: z.string(),
})

export function deduplicateWatchPairs(
  pairs: readonly WatchPair[]
): WatchPair[] {
  const seen = new Set<string>()

  return pairs.filter((pair) => {
    const key = `${pair.base}\u0000${pair.target}`
    if (seen.has(key)) {
      return false
    }
    seen.add(key)
    return true
  })
}

export function parseWatchPairs(
  raw: string | null,
  availableCodes: ReadonlySet<string>
): WatchPair[] {
  if (raw === null) {
    return []
  }

  try {
    return parseWatchPairValues(JSON.parse(raw), availableCodes)
  } catch {
    return []
  }
}

function parseWatchPairValues(
  values: unknown,
  availableCodes: ReadonlySet<string>
): WatchPair[] {
  if (!Array.isArray(values)) {
    return []
  }

  const pairs: WatchPair[] = []
  for (const value of values) {
    const candidate = WatchPairRecord.safeParse(value)
    if (
      !candidate.success ||
      candidate.data.base === candidate.data.target ||
      !availableCodes.has(candidate.data.base) ||
      !availableCodes.has(candidate.data.target)
    ) {
      continue
    }
    pairs.push(candidate.data)
  }
  return deduplicateWatchPairs(pairs)
}

export function serializeWatchPairs(pairs: readonly WatchPair[]): string {
  return JSON.stringify(deduplicateWatchPairs(pairs))
}

export function loadWatchPairs(
  availableCodes: ReadonlySet<string>
): WatchPair[] {
  const raw = window.localStorage.getItem(storageKey)
  if (raw === null) {
    return []
  }
  try {
    return parseWatchPairValues(JSON.parse(raw), availableCodes)
  } catch {
    window.localStorage.removeItem(storageKey)
    return []
  }
}

export function saveWatchPairs(pairs: readonly WatchPair[]): void {
  window.localStorage.setItem(storageKey, serializeWatchPairs(pairs))
}
