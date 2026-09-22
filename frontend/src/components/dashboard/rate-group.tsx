import { useQuery } from "@tanstack/react-query"
import { Trash2Icon } from "lucide-react"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { getRatesQueryOptions } from "@/lib/api/client"
import type { WatchPair } from "@/stores/watchlist"
import type { RatesOutputBody } from "@/lib/api/generated"

type Rates = RatesOutputBody["rates"]

/** earliestExpiry returns the earliest valid rate expiry, or undefined when no usable rate is present. */
function earliestExpiry(rates: Rates | undefined) {
  let earliest: number | undefined
  for (const { expiresAt } of Object.values(rates ?? {})) {
    const expiry = Date.parse(expiresAt)
    if (
      Number.isFinite(expiry) &&
      (earliest === undefined || expiry < earliest)
    ) {
      earliest = expiry
    }
  }
  return earliest
}

/** hasExpiredRate reports whether the earliest valid rate has expired; missing rates are not expired. */
function hasExpiredRate(rates: Rates | undefined) {
  const expiry = earliestExpiry(rates)
  return expiry !== undefined && expiry <= Date.now()
}

/** nextRefetchDelay schedules the earliest expiry with a one-millisecond minimum, or disables scheduling when rates are absent. */
function nextRefetchDelay(rates: Rates | undefined) {
  const expiry = earliestExpiry(rates)
  return expiry === undefined ? false : Math.max(expiry - Date.now(), 1)
}

const failedRefreshRetryDelay = 30_000

type RateGroupProps = {
  base: string
  pairs: readonly WatchPair[]
  onRemove: (pair: WatchPair) => Promise<void>
}

/** RateGroup renders one base currency's rates and refetches when cached results expire or refresh failures retry. */
export function RateGroup({ base, pairs, onRemove }: RateGroupProps) {
  const targets = pairs.map((pair) => pair.target)
  const ratesQuery = useQuery({
    ...getRatesQueryOptions(base, targets),
    staleTime: Infinity,
    gcTime: 60 * 60 * 1000,
    retry: false,
    refetchIntervalInBackground: false,
    refetchInterval: (query) => {
      if (query.state.errorUpdatedAt > query.state.dataUpdatedAt) {
        return failedRefreshRetryDelay
      }
      return nextRefetchDelay(query.state.data?.rates)
    },
    refetchOnWindowFocus: (query) => hasExpiredRate(query.state.data?.rates),
  })

  return (
    <Card className="overflow-hidden py-0">
      <CardHeader className="flex-row items-center justify-between border-b border-border px-5 py-4 sm:px-6">
        <div>
          <CardDescription className="text-xs font-semibold tracking-[0.2em] uppercase">
            Base
          </CardDescription>
          <CardTitle className="mt-1 text-2xl">{base}</CardTitle>
        </div>
        {ratesQuery.isFetching ? (
          <span className="text-sm text-muted-foreground">Updating…</span>
        ) : null}
      </CardHeader>
      <CardContent className="p-0">
        {ratesQuery.isPending ? (
          <p className="px-5 py-6 text-sm text-muted-foreground sm:px-6">
            Loading latest rates…
          </p>
        ) : ratesQuery.isError && ratesQuery.data === undefined ? (
          <p className="px-5 py-6 text-sm text-destructive sm:px-6">
            Rates are unavailable right now. Try again shortly.
          </p>
        ) : (
          <div className="divide-y divide-border">
            {pairs.map((pair) => {
              const rate = ratesQuery.data?.rates[pair.target]?.rate
              return (
                <div
                  className="flex items-center justify-between gap-4 px-5 py-4 sm:px-6"
                  key={`${pair.base}-${pair.target}`}
                >
                  <div>
                    <p className="font-medium">{pair.target}</p>
                    <p className="text-sm text-muted-foreground">
                      1 {pair.base} equals
                    </p>
                  </div>
                  <div className="flex items-center gap-4">
                    <p className="text-xl font-semibold tabular-nums">
                      {rate === undefined
                        ? "—"
                        : rate.toLocaleString(undefined, {
                            maximumFractionDigits: 6,
                          })}
                    </p>
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      aria-label={`Remove ${pair.base} to ${pair.target}`}
                      onClick={() => {
                        void onRemove(pair)
                      }}
                    >
                      <Trash2Icon />
                    </Button>
                  </div>
                </div>
              )
            })}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
