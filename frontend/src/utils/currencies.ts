import type { CurrencyRecord } from "@/lib/api/client"

export type CurrencyOption = {
  code: string
  name: string
}

/** joinCurrencies maps API records to code-and-name options sorted by currency code. */
export function joinCurrencies(
  currencies: readonly CurrencyRecord[]
): CurrencyOption[] {
  const options: CurrencyOption[] = []

  for (const currency of currencies) {
    options.push({
      code: currency.code,
      name: currency.name,
    })
  }

  return options.sort((left, right) => left.code.localeCompare(right.code))
}
