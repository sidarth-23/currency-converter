import { useMemo, type ReactNode } from "react"
import { useQuery, useSuspenseQuery } from "@tanstack/react-query"
import { ClientOnly } from "@tanstack/react-router"
import type { MangoQuery } from "rxdb"
import { RxDatabaseProvider, useLiveRxQuery } from "rxdb/plugins/react"

import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { getCurrenciesQueryOptions } from "@/lib/api/client"
import { joinCurrencies, type CurrencyOption } from "@/lib/currencies"
import {
  addWatchPair,
  getWatchlistDatabase,
  removeWatchPair,
  type WatchPair,
  type WatchPairDocument,
  type WatchlistDatabase,
} from "@/lib/watchlist/database"

import { RateGroup } from "./rate-group"
import { WatchPairForm } from "./watch-pair-form"

const watchPairsQuery: MangoQuery<WatchPairDocument> = {
  selector: {},
  sort: [{ base: "asc" }, { target: "asc" }],
}

const watchlistLoadingCard = (
  <Card>
    <CardContent className="py-8 text-center text-muted-foreground">
      Loading your watchlist…
    </CardContent>
  </Card>
)

export function DashboardLoading() {
  return <DashboardShell>{watchlistLoadingCard}</DashboardShell>
}

export function DashboardUnavailable() {
  return (
    <DashboardShell>
      <Alert variant="destructive">
        <AlertTitle>Currency catalog unavailable</AlertTitle>
        <AlertDescription>Try again shortly.</AlertDescription>
      </Alert>
    </DashboardShell>
  )
}

function DashboardShell({ children }: { children: ReactNode }) {
  return (
    <main className="min-h-svh bg-muted/30 px-4 py-10 sm:px-8">
      <div className="mx-auto flex w-full max-w-5xl flex-col gap-8">
        {children}
      </div>
    </main>
  )
}

export function Dashboard() {
  const { data: currencies } = useSuspenseQuery(getCurrenciesQueryOptions())
  const options = useMemo(() => joinCurrencies(currencies ?? []), [currencies])

  if (options.length < 2) {
    return <DashboardUnavailable />
  }

  return (
    <DashboardShell>
      <ClientOnly fallback={watchlistLoadingCard}>
        <Watchlist options={options} />
      </ClientOnly>
    </DashboardShell>
  )
}

type WatchlistProps = {
  options: readonly CurrencyOption[]
}

function Watchlist({ options }: WatchlistProps) {
  const databaseQuery = useQuery({
    queryKey: ["watchlist-database"],
    queryFn: getWatchlistDatabase,
    staleTime: Infinity,
    gcTime: Infinity,
    retry: false,
  })

  if (databaseQuery.isPending || databaseQuery.data === undefined) {
    return watchlistLoadingCard
  }

  if (databaseQuery.isError) {
    return (
      <Alert variant="destructive">
        <AlertTitle>Watchlist unavailable</AlertTitle>
        <AlertDescription>
          This browser could not open your saved watchlist.
        </AlertDescription>
      </Alert>
    )
  }

  return (
    <RxDatabaseProvider database={databaseQuery.data}>
      <WatchlistContent database={databaseQuery.data} options={options} />
    </RxDatabaseProvider>
  )
}

type WatchlistContentProps = WatchlistProps & {
  database: WatchlistDatabase
}

function WatchlistContent({ database, options }: WatchlistContentProps) {
  const { results } = useLiveRxQuery<WatchPairDocument>({
    collection: "watchpairs",
    query: watchPairsQuery,
  })
  const pairs = useMemo<WatchPair[]>(
    () => results.map(({ base, target }) => ({ base, target })),
    [results]
  )
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

  return (
    <>
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
              Choose currency pairs to follow. Rates refresh automatically while
              your watchlist stays in this browser.
            </p>
          </div>
          <div className="rounded-full border border-border bg-card px-4 py-2 text-sm text-muted-foreground shadow-sm">
            {pairs.length} {pairs.length === 1 ? "pair" : "pairs"} watched
          </div>
        </div>
      </header>

      <WatchPairForm
        options={options}
        pairs={pairs}
        onAdd={(pair) => addWatchPair(database, pair)}
      />

      {groups.length === 0 ? (
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
              onRemove={(pair) => removeWatchPair(database, pair)}
            />
          ))}
        </div>
      )}
    </>
  )
}
