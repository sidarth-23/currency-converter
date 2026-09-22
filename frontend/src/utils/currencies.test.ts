import { describe, expect, it } from "vitest"

import { joinCurrencies } from "./currencies"

describe("joinCurrencies", () => {
  it("retains provider records and names in code order", () => {
    expect(
      joinCurrencies([
        { code: "USD", name: "Provider dollar" },
        { code: "EUR", name: "Provider euro" },
        { code: "ZZZ", name: "Unknown" },
      ])
    ).toEqual([
      { code: "EUR", name: "Provider euro" },
      { code: "USD", name: "Provider dollar" },
      { code: "ZZZ", name: "Unknown" },
    ])
  })
})
