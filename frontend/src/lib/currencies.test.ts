import { describe, expect, it } from "vitest"

import { joinCurrencies } from "./currencies"

describe("joinCurrencies", () => {
  it("keeps provider-supported ISO currencies with provider names", () => {
    expect(
      joinCurrencies([
        { code: "USD", name: "Provider dollar" },
        { code: "EUR", name: "" },
        { code: "ZZZ", name: "Unknown" },
      ])
    ).toEqual([
      { code: "EUR", name: "Euro" },
      { code: "USD", name: "Provider dollar" },
    ])
  })
})
