import { client } from "./generated/client.gen"
import { getRatesOptions } from "./generated/@tanstack/react-query.gen"
import type { ClientOptions, GetRatesData } from "./generated/types.gen"

const configuredBaseUrl = import.meta.env.VITE_API_BASE_URL ?? "http://localhost:8080"
const apiConfig: ClientOptions = {
  baseUrl: `${configuredBaseUrl.replace(/\/$/, "")}/api`,
}

client.setConfig(apiConfig)

export function getRatesQueryOptions(
  base: string,
  targets: readonly string[],
) {
  const request: GetRatesData = {
    query: {
      base,
      targets: targets.join(","),
    },
  }

  return getRatesOptions(request)
}
