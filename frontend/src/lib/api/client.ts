import { env } from "@/env"

import { client } from "./generated/client.gen"
import {
  getCurrenciesOptions,
  getRatesOptions,
} from "./generated/@tanstack/react-query.gen"
import type {
  ClientOptions,
  Currency as CurrencyRecord,
  GetRatesData,
} from "./generated/types.gen"

export type { CurrencyRecord }

const configuredBaseUrl = env.VITE_API_BASE_URL
const apiConfig: ClientOptions = {
  baseUrl: `${configuredBaseUrl.replace(/\/$/, "")}/api`,
}

client.setConfig(apiConfig)

/** getRatesQueryOptions builds generated-client query options for a base currency and copied target list. */
export function getRatesQueryOptions(base: string, targets: readonly string[]) {
  const request: GetRatesData = {
    query: {
      base,
      targets: [...targets],
    },
    url: "/rates",
  }

  return getRatesOptions(request)
}

/** getCurrenciesQueryOptions builds generated-client currency query options with a one-hour stale period. */
export function getCurrenciesQueryOptions() {
  return {
    ...getCurrenciesOptions(),
    staleTime: 60 * 60 * 1000,
  }
}
