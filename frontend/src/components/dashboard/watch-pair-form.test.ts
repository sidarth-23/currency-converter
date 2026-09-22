import { describe, expect, it } from "vitest"

import { createWatchPairSchema } from "./watch-pair-form"

describe("WatchPairForm validation", () => {
  it("rejects equal and already watched pairs", () => {
    const equalPair = createWatchPairSchema([]).safeParse({
      base: "USD",
      target: "USD",
    })
    expect(equalPair.success).toBe(false)
    if (!equalPair.success) {
      expect(equalPair.error.issues).toContainEqual(
        expect.objectContaining({
          message: "Choose two different currencies.",
          path: ["target"],
        })
      )
    }

    const duplicatePair = createWatchPairSchema([
      { base: "USD", target: "EUR" },
    ]).safeParse({ base: "USD", target: "EUR" })
    expect(duplicatePair.success).toBe(false)
    if (!duplicatePair.success) {
      expect(duplicatePair.error.issues).toContainEqual(
        expect.objectContaining({
          message: "That pair is already on your watchlist.",
          path: ["target"],
        })
      )
    }
  })

  it("accepts a new pair", () => {
    expect(
      createWatchPairSchema([{ base: "USD", target: "GBP" }]).safeParse({
        base: "USD",
        target: "EUR",
      }).success
    ).toBe(true)
  })
})
