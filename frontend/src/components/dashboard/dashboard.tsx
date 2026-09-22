import { useEffect, useMemo, useState } from "react"
import { useQuery } from "@tanstack/react-query"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getCurrenciesQueryOptions } from "@/lib/api/client"
import { joinCurrencies } from "@/lib/currencies"
import {
  deduplicateWatchPairs,
  loadWatchPairs,
  saveWatchPairs,
  type WatchPair,
} from "@/lib/watch-pairs"

import { RateGroup } from "./rate-group"
import { WatchPairForm } from "./watch-pair-form"

export function Dashboard() {
  const currenciesQuery = useQuery(getCurrenciesQueryOptions())
  const options = useMemo(
    () => joinCurrencies(currenciesQuery.data ?? []),
    [currenciesQuery.data]
  )
  const availableCodes = useMemo(
    () => new Set(options.map((option) => option.code)),
    [options]
  )
  const [pairs, setPairs] = useState<WatchPair[]>([])
  const [hydrated, setHydrated] = useState(false)

  useEffect(() => {
    if (!currenciesQuery.isSuccess || hydrated) {
      return
    }
    const storedPairs = loadWatchPairs(availableCodes)
    saveWatchPairs(storedPairs)
    setPairs(storedPairs)
    setHydrated(true)
  }, [availableCodes, currenciesQuery.isSuccess, hydrated])

  const groups = useMemo(() => {
    const groupedPairs = new Map<string, WatchPair[]>()
    for (const pair of pairs) {
      const group = groupedPairs.get(pair.base)
      if (group === undefined) {
        groupedPairs.set(pair.base, [pair])
      } else {
        group.push(pair)
      }
    }
    return [...groupedPairs]
  }, [pairs])

  const addPair = (pair: WatchPair) => {
    setPairs((current) => {
      const next = deduplicateWatchPairs([...current, pair])
      if (next.length > current.length) {
        saveWatchPairs(next)
      }
      return next
    })
  }

  const removePair = (pairToRemove: WatchPair) => {
    setPairs((current) => {
      const next = current.filter(
        (pair) =>
          pair.base !== pairToRemove.base || pair.target !== pairToRemove.target
      )
      saveWatchPairs(next)
      return next
    })
  }

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
                Choose currency pairs to follow. Rates refresh automatically
                while your watchlist stays in this browser.
              </p>
            </div>
            <div className="rounded-full border border-border bg-card px-4 py-2 text-sm text-muted-foreground shadow-sm">
              {hydrated
                ? `${pairs.length} ${pairs.length === 1 ? "pair" : "pairs"} watched`
                : "Loading watchlist"}
            </div>
          </div>
        </header>

        {currenciesQuery.isPending ? (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              Loading available currencies…
            </CardContent>
          </Card>
        ) : currenciesQuery.isError ? (
          <Alert variant="destructive">
            <AlertTitle>Currency catalog unavailable</AlertTitle>
            <AlertDescription>Try again shortly.</AlertDescription>
          </Alert>
        ) : options.length < 2 ? (
          <Alert variant="destructive">
            <AlertTitle>Currency catalog unavailable</AlertTitle>
            <AlertDescription>
              The provider did not return enough supported currencies.
            </AlertDescription>
          </Alert>
        ) : (
          <WatchPairForm options={options} pairs={pairs} onAdd={addPair} />
        )}

        {!hydrated ? (
          <Card>
            <CardContent className="py-8 text-center text-muted-foreground">
              Loading your watchlist…
            </CardContent>
          </Card>
        ) : groups.length === 0 ? (
          <Card className="border-dashed">
            <CardHeader className="items-center py-10 text-center">
              <CardTitle>No pairs yet</CardTitle>
              <p className="text-sm text-muted-foreground">
                Add a currency pair above to start watching exchange rates.
              </p>
            </CardHeader>
          </Card>
        ) : (
          <div className="grid gap-6">
            {groups.map(([base, groupPairs]) => (
              <RateGroup
                key={base}
                base={base}
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
