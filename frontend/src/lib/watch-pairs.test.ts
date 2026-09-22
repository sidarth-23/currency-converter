// @vitest-environment jsdom

import { describe, expect, it } from "vitest"

import {
  deduplicateWatchPairs,
  loadWatchPairs,
  parseWatchPairs,
  saveWatchPairs,
  serializeWatchPairs,
} from "./watch-pairs"

const availableCodes = new Set(["EUR", "GBP", "USD"])

describe("watch pair storage", () => {
  it("discards malformed, unsupported, duplicate, and self pairs", () => {
    expect(
      parseWatchPairs(
        JSON.stringify([
          { base: "USD", target: "EUR" },
          { base: "USD", target: "USD" },
          { base: "USD", target: "JPY" },
          { base: "USD", target: "EUR" },
          "invalid",
        ]),
        availableCodes
      )
    ).toEqual([{ base: "USD", target: "EUR" }])
  })

  it("removes malformed storage and persists validated pairs", () => {
    window.localStorage.setItem("currency-watcher:watch-pairs", "{")
    expect(loadWatchPairs(availableCodes)).toEqual([])
    expect(
      window.localStorage.getItem("currency-watcher:watch-pairs")
    ).toBeNull()

    const pairs = deduplicateWatchPairs([
      { base: "USD", target: "EUR" },
      { base: "USD", target: "EUR" },
      { base: "GBP", target: "USD" },
    ])
    saveWatchPairs(pairs)
    expect(loadWatchPairs(availableCodes)).toEqual(pairs)
    expect(serializeWatchPairs(pairs)).toBe(JSON.stringify(pairs))
  })
})
