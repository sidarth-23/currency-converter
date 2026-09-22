import { useQuery } from "@tanstack/react-query"

import { Button } from "@/components/ui/button"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { getRatesQueryOptions } from "@/lib/api/client"
import type { WatchPair } from "@/lib/watch-pairs"

type RateGroupProps = {
  base: string
  pairs: readonly WatchPair[]
  onRemove: (pair: WatchPair) => void
}

export function RateGroup({ base, pairs, onRemove }: RateGroupProps) {
  const targets = pairs.map((pair) => pair.target)
  const ratesQuery = useQuery({
    ...getRatesQueryOptions(base, targets),
    select: (data) => data.rates,
    staleTime: 5 * 60 * 1000,
    gcTime: 60 * 60 * 1000,
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
      </CardContent>
    </Card>
  )
}
