import { useEffect, useMemo, useState } from "react"
import type { FormEvent } from "react"
import { useQuery } from "@tanstack/react-query"
import { createFileRoute } from "@tanstack/react-router"

import { Button } from "@/components/ui/button"
import { getRatesQueryOptions } from "@/lib/api/client"
const STORAGE_KEY = "currency-watcher:watch-pairs"

const CURRENCIES = [
  "AUD",
  "CAD",
  "CHF",
  "CNY",
  "DKK",
  "EUR",
  "GBP",
  "HKD",
  "INR",
  "JPY",
  "NOK",
  "NZD",
  "SEK",
  "SGD",
  "USD",
] as const

type CurrencyCode = (typeof CURRENCIES)[number]

type WatchPair = {
  base: string
  target: string
}

type RateGroupProps = {
  base: string
  pairs: readonly WatchPair[]
  onRemove: (pair: WatchPair) => void
}

export const Route = createFileRoute("/")({ component: Dashboard })

function isCurrencyCode(value: string): value is CurrencyCode {
  return CURRENCIES.some((currency) => currency === value)
}

function isWatchPair(value: unknown): value is WatchPair {
  if (typeof value !== "object" || value === null) {
    return false
  }

  const candidate = value as Record<string, unknown>
  return typeof candidate.base === "string" && typeof candidate.target === "string"
}

function deduplicatePairs(pairs: readonly WatchPair[]): WatchPair[] {
  const seen = new Set<string>()
  return pairs.filter((pair) => {
    const key = `${pair.base}-${pair.target}`
    if (seen.has(key)) {
      return false
    }
    seen.add(key)
    return true
  })
}

function Dashboard() {
  const [pairs, setPairs] = useState<WatchPair[]>([])
  const [hydrated, setHydrated] = useState(false)
  const [base, setBase] = useState<CurrencyCode>("USD")
  const [target, setTarget] = useState<CurrencyCode>("EUR")

  useEffect(() => {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    if (stored !== null) {
      try {
        const parsed: unknown = JSON.parse(stored)
        if (Array.isArray(parsed)) {
          setPairs(deduplicatePairs(parsed.filter(isWatchPair)))
        }
      } catch {
        window.localStorage.removeItem(STORAGE_KEY)
      }
    }
    setHydrated(true)
  }, [])

  const groupedPairs = useMemo(() => {
    const groups: Record<string, WatchPair[]> = {}
    for (const pair of pairs) {
      const group = groups[pair.base]
      if (group === undefined) {
        groups[pair.base] = [pair]
      } else {
        group.push(pair)
      }
    }
    return groups
  }, [pairs])

  const addPair = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    if (base === target) {
      return
    }

    setPairs((current) => {
      if (current.some((pair) => pair.base === base && pair.target === target)) {
        return current
      }
      return [...current, { base, target }]
    })
  }

  const removePair = (pairToRemove: WatchPair) => {
    setPairs((current) =>
      current.filter(
        (pair) => pair.base !== pairToRemove.base || pair.target !== pairToRemove.target,
      ),
    )
  }

  const pairAlreadyAdded = pairs.some(
    (pair) => pair.base === base && pair.target === target,
  )
  const groupedEntries = Object.entries(groupedPairs)

  return (
    <main className="min-h-svh bg-muted/30 px-4 py-10 sm:px-8">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-8">
        <header className="flex flex-col gap-3">
          <p className="text-sm font-semibold tracking-[0.2em] text-primary uppercase">
            Currency watcher
          </p>
          <div className="flex flex-col justify-between gap-4 sm:flex-row sm:items-end">
            <div>
              <h1 className="text-3xl font-semibold tracking-tight sm:text-4xl">
                Keep an eye on your rates.
              </h1>
              <p className="mt-2 max-w-xl text-muted-foreground">
                Choose currency pairs to follow. Rates refresh automatically while your watchlist
                stays in this browser.
              </p>
            </div>
            <div className="rounded-full border border-border bg-card px-4 py-2 text-sm text-muted-foreground shadow-sm">
              {hydrated ? `${pairs.length} ${pairs.length === 1 ? "pair" : "pairs"} watched` : "Loading watchlist"}
            </div>
          </div>
        </header>

        <section className="rounded-2xl border border-border bg-card p-5 shadow-sm sm:p-6">
          <div className="mb-5">
            <h2 className="text-lg font-semibold">Add a pair</h2>
            <p className="text-sm text-muted-foreground">Track one quote currency against a base.</p>
          </div>
          <form className="grid gap-4 sm:grid-cols-[1fr_auto_1fr_auto] sm:items-end" onSubmit={addPair}>
            <label className="grid gap-2 text-sm font-medium" htmlFor="base-currency">
              Base currency
              <select
                id="base-currency"
                className="h-10 rounded-lg border border-input bg-background px-3 font-normal outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                value={base}
                onChange={(event) => {
                  const selected = event.currentTarget.value
                  if (isCurrencyCode(selected)) {
                    setBase(selected)
                  }
                }}
              >
                {CURRENCIES.map((currency) => (
                  <option key={currency} value={currency}>
                    {currency}
                  </option>
                ))}
              </select>
            </label>
            <span className="hidden pb-2 text-center text-muted-foreground sm:block" aria-hidden="true">
              →
            </span>
            <label className="grid gap-2 text-sm font-medium" htmlFor="target-currency">
              Target currency
              <select
                id="target-currency"
                className="h-10 rounded-lg border border-input bg-background px-3 font-normal outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
                value={target}
                onChange={(event) => {
                  const selected = event.currentTarget.value
                  if (isCurrencyCode(selected)) {
                    setTarget(selected)
                  }
                }}
              >
                {CURRENCIES.map((currency) => (
                  <option key={currency} value={currency}>
                    {currency}
                  </option>
                ))}
              </select>
            </label>
            <Button type="submit" disabled={base === target || pairAlreadyAdded}>
              Add pair
            </Button>
          </form>
          {base === target ? (
            <p className="mt-3 text-sm text-destructive">Choose two different currencies.</p>
          ) : pairAlreadyAdded ? (
            <p className="mt-3 text-sm text-muted-foreground">That pair is already on your watchlist.</p>
          ) : null}
        </section>

        {!hydrated ? (
          <section className="rounded-2xl border border-dashed border-border bg-card p-8 text-center text-muted-foreground">
            Loading your watchlist…
          </section>
        ) : groupedEntries.length === 0 ? (
          <section className="rounded-2xl border border-dashed border-border bg-card p-10 text-center">
            <h2 className="text-lg font-semibold">No pairs yet</h2>
            <p className="mt-2 text-sm text-muted-foreground">
              Add a currency pair above to start watching exchange rates.
            </p>
          </section>
        ) : (
          <div className="grid gap-6">
            {groupedEntries.map(([groupBase, groupPairs]) => (
              <RateGroup
                key={groupBase}
                base={groupBase}
                pairs={groupPairs}
                onRemove={removePair}
              />
            ))}
          </div>
        )}
      </div>
    </main>
  )
}

function RateGroup({ base, pairs, onRemove }: RateGroupProps) {
  const targets = pairs.map((pair) => pair.target)
  const ratesQuery = useQuery({
    ...getRatesQueryOptions(base, targets),
    select: (data) => data.rates,
    staleTime: 5 * 60 * 1000,
    gcTime: 60 * 60 * 1000,
  })

  return (
    <section className="overflow-hidden rounded-2xl border border-border bg-card shadow-sm">
      <div className="flex items-center justify-between border-b border-border px-5 py-4 sm:px-6">
        <div>
          <p className="text-xs font-semibold tracking-[0.2em] text-muted-foreground uppercase">Base</p>
          <h2 className="mt-1 text-2xl font-semibold">{base}</h2>
        </div>
        {ratesQuery.isFetching ? (
          <span className="text-sm text-muted-foreground">Updating…</span>
        ) : null}
      </div>
      {ratesQuery.isPending ? (
        <p className="px-5 py-6 text-sm text-muted-foreground sm:px-6">Loading latest rates…</p>
      ) : ratesQuery.isError ? (
        <p className="px-5 py-6 text-sm text-destructive sm:px-6">
          Rates are unavailable right now. Try again shortly.
        </p>
      ) : (
        <div className="divide-y divide-border">
          {pairs.map((pair) => {
            const rate = ratesQuery.data?.[pair.target]
            return (
              <div
                className="flex items-center justify-between gap-4 px-5 py-4 sm:px-6"
                key={`${pair.base}-${pair.target}`}
              >
                <div>
                  <p className="font-medium">{pair.target}</p>
                  <p className="text-sm text-muted-foreground">1 {pair.base} equals</p>
                </div>
                <div className="flex items-center gap-4">
                  <p className="text-xl font-semibold tabular-nums">
                    {rate === undefined
                      ? "—"
                      : rate.toLocaleString(undefined, { maximumFractionDigits: 6 })}
                  </p>
                  <Button
                    type="button"
                    variant="ghost"
                    size="sm"
                    aria-label={`Remove ${pair.base} to ${pair.target}`}
                    onClick={() => onRemove(pair)}
                  >
                    Remove
                  </Button>
                </div>
              </div>
            )
          })}
        </div>
      )}
    </section>
  )
}
