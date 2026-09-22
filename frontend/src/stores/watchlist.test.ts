import { afterAll, describe, expect, it, vi } from "vitest"
import { addWatchPair, getWatchlistDatabase, removeWatchPair } from "./watchlist"

vi.mock("rxdb/plugins/storage-dexie", async () => {
  // The hoisted mock must load the test-only adapter inside its factory.
  const { getRxStorageMemory } = await import("rxdb/plugins/storage-memory")
  return { getRxStorageDexie: getRxStorageMemory }
})

const database = await getWatchlistDatabase()

afterAll(async () => {
  await database.remove()
})

describe("watchlist database", () => {
  it("keeps each composite pair once and removes only its exact key", async () => {
    await addWatchPair(database, { base: "USD", target: "EUR" })
    await addWatchPair(database, { base: "USD", target: "EUR" })

    expect(await database.watchpairs.find().exec()).toHaveLength(1)

    await addWatchPair(database, { base: "EUR", target: "USD" })

    expect(await database.watchpairs.find().exec()).toHaveLength(2)

    expect(
      (await database.watchpairs.find().exec()).map(({ base, target }) => ({
        base,
        target,
      }))
    ).toEqual(
      expect.arrayContaining([
        { base: "USD", target: "EUR" },
        { base: "EUR", target: "USD" },
      ])
    )
    await removeWatchPair(database, { base: "USD", target: "EUR" })

    expect(
      (await database.watchpairs.find().exec()).map(({ base, target }) => ({
        base,
        target,
      }))
    ).toEqual([{ base: "EUR", target: "USD" }])
  })
})
