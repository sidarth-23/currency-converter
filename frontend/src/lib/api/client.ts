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

export function getCurrenciesQueryOptions() {
  return getCurrenciesOptions()
}
