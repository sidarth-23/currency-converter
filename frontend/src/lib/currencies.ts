import { code as getISOCurrency } from "@orderlayer/iso4217-ts"

import type { CurrencyRecord } from "@/lib/api/client"

export type CurrencyOption = {
  code: string
  name: string
}

export function joinCurrencies(
  currencies: readonly CurrencyRecord[]
): CurrencyOption[] {
  const options: CurrencyOption[] = []

  for (const currency of currencies) {
    const metadata = getISOCurrency(currency.code)
    if (metadata === undefined) {
      continue
    }
    options.push({
      code: currency.code,
      name: currency.name || metadata.currency,
    })
  }

  return options.sort((left, right) => left.code.localeCompare(right.code))
}
