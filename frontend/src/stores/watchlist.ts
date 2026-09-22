import {
  createRxDatabase,
  toTypedRxJsonSchema,
  type ExtractDocumentTypeFromTypedRxJsonSchema,
  type RxCollection,
  type RxDatabase,
  type RxJsonSchema,
} from "rxdb"
import { getRxStorageDexie } from "rxdb/plugins/storage-dexie"

const watchPairSchemaLiteral = {
  title: "watch pair schema",
  version: 0,
  primaryKey: {
    key: "id",
    fields: ["base", "target"],
    separator: "|",
  },
  type: "object",
  properties: {
    id: {
      type: "string",
      maxLength: 7,
    },
    base: {
      type: "string",
      maxLength: 3,
      pattern: "^[A-Z]{3}$",
    },
    target: {
      type: "string",
      maxLength: 3,
      pattern: "^[A-Z]{3}$",
    },
  },
  required: ["id", "base", "target"],
  additionalProperties: false,
  indexes: [["base", "target"]],
} as const

const typedWatchPairSchema = toTypedRxJsonSchema(watchPairSchemaLiteral)
export type WatchPairDocument = ExtractDocumentTypeFromTypedRxJsonSchema<
  typeof typedWatchPairSchema
>
const watchPairSchema: RxJsonSchema<WatchPairDocument> = watchPairSchemaLiteral

export type WatchPair = Pick<WatchPairDocument, "base" | "target">

type WatchlistCollections = {
  watchpairs: RxCollection<WatchPairDocument>
}

export type WatchlistDatabase = RxDatabase<WatchlistCollections>

let databasePromise: Promise<WatchlistDatabase> | undefined

/** getWatchlistDatabase creates the singleton Dexie-backed watch-pair collection and reuses it thereafter. */
export function getWatchlistDatabase(): Promise<WatchlistDatabase> {
  databasePromise ??= createRxDatabase<WatchlistCollections>({
    name: "currencywatcher",
    storage: getRxStorageDexie(),
  }).then(async (database) => {
    await database.addCollections({
      watchpairs: { schema: watchPairSchema },
    })
    return database
  })

  return databasePromise
}

/** addWatchPair persists a pair idempotently, leaving an existing exact pair unchanged. */
export async function addWatchPair(
  database: WatchlistDatabase,
  pair: WatchPair
): Promise<void> {
  await database.watchpairs.insertIfNotExists(pair as WatchPairDocument)
}

/** removeWatchPair removes only the persisted document matching the exact base and target pair. */
export async function removeWatchPair(
  database: WatchlistDatabase,
  pair: WatchPair
): Promise<void> {
  const primaryKey = database.watchpairs.schema.getPrimaryOfDocumentData(
    pair as WatchPairDocument
  )
  const document = await database.watchpairs.findOne(primaryKey).exec()

  if (document !== null) {
    await document.remove()
  }
}
