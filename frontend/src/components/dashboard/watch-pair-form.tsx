import { useMemo, useState } from "react"

import { Dialog } from "@base-ui/react/dialog"
import { useForm } from "@tanstack/react-form"
import { z } from "zod"

import { Button } from "@/components/ui/button"
import { PlusIcon, XIcon } from "lucide-react"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Field, FieldError, FieldLabel } from "@/components/ui/field"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import type { WatchPair } from "@/stores/watchlist"
import type { CurrencyOption } from "@/utils/currencies"

type WatchPairFormProps = {
  options: readonly CurrencyOption[]
  pairs: readonly WatchPair[]
  onAdd: (pair: WatchPair) => Promise<void>
}

/** defaultWatchPair prefers USD-to-EUR when available, otherwise selects the first two currency options. */
function defaultWatchPair(options: readonly CurrencyOption[]): WatchPair {
  const codes = new Set(options.map((option) => option.code))
  if (codes.has("USD") && codes.has("EUR")) {
    return { base: "USD", target: "EUR" }
  }
  return { base: options[0]!.code, target: options[1]!.code }
}

/** createWatchPairSchema validates that a pair uses distinct currencies and is not already persisted. */
export function createWatchPairSchema(pairs: readonly WatchPair[]) {
  return z
    .object({
      base: z.string(),
      target: z.string(),
    })
    .superRefine((value, context) => {
      if (value.base === value.target) {
        context.addIssue({
          code: "custom",
          message: "Choose two different currencies.",
          path: ["target"],
        })
      }
      if (
        pairs.some(
          (pair) => pair.base === value.base && pair.target === value.target
        )
      ) {
        context.addIssue({
          code: "custom",
          message: "That pair is already on your watchlist.",
          path: ["target"],
        })
      }
    })
}

/** WatchPairForm collects a new pair, validates it as selections change and submit, then persists it through onAdd. */
export function WatchPairForm({ options, pairs, onAdd }: WatchPairFormProps) {
  const [open, setOpen] = useState(false)
  const schema = useMemo(() => createWatchPairSchema(pairs), [pairs])
  const form = useForm({
    defaultValues: defaultWatchPair(options),
    validators: {
      onChange: schema,
      onSubmit: schema,
    },
    onSubmit: async ({ value }) => {
      await onAdd(value)
      setOpen(false)
    },
  })

  return (
    <Dialog.Root open={open} onOpenChange={setOpen}>
      <Dialog.Trigger
        render={
          <Button
            type="button"
            size="icon-lg"
            className="fixed right-6 bottom-6 z-40 rounded-full shadow-lg sm:right-8"
            aria-label="Add currency pair"
          />
        }
      >
        <PlusIcon />
      </Dialog.Trigger>
      <Dialog.Portal>
        <Dialog.Backdrop className="fixed inset-0 z-50 bg-black/50 backdrop-blur-[1px]" />
        <Dialog.Viewport className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <Dialog.Popup className="w-full max-w-lg outline-none">
            <Card>
              <CardHeader className="relative">
                <Dialog.Title render={<CardTitle />}>Add a pair</Dialog.Title>
                <Dialog.Description render={<CardDescription />}>
                  Track one quote currency against a base.
                </Dialog.Description>
                <Dialog.Close
                  render={
                    <Button
                      type="button"
                      variant="ghost"
                      size="icon-sm"
                      className="absolute top-4 right-4"
                      aria-label="Close add pair dialog"
                    />
                  }
                >
                  <XIcon />
                </Dialog.Close>
              </CardHeader>
              <CardContent>
                <form
                  className="grid gap-4"
                  onSubmit={(event) => {
                    event.preventDefault()
                    void form.handleSubmit()
                  }}
                >
                  <div className="grid gap-4 sm:grid-cols-[1fr_auto_1fr_auto] sm:items-end">
                    <form.Field name="base">
                      {(field) => (
                        <Field>
                          <FieldLabel htmlFor="base-currency">
                            Base currency
                          </FieldLabel>
                          <Select
                            value={field.state.value}
                            onValueChange={(value) => {
                              if (value !== null) {
                                field.handleChange(value)
                              }
                            }}
                          >
                            <SelectTrigger
                              id="base-currency"
                              className="w-full"
                            >
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {options.map((option) => (
                                <SelectItem
                                  key={option.code}
                                  value={option.code}
                                >
                                  <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-xs font-medium text-muted-foreground">
                                    {option.code}
                                  </kbd>
                                  <span>{option.name}</span>
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </Field>
                      )}
                    </form.Field>
                    <span
                      className="hidden pb-2 text-center text-muted-foreground sm:block"
                      aria-hidden="true"
                    >
                      →
                    </span>
                    <form.Field name="target">
                      {(field) => (
                        <Field>
                          <FieldLabel htmlFor="target-currency">
                            Target currency
                          </FieldLabel>
                          <Select
                            value={field.state.value}
                            onValueChange={(value) => {
                              if (value !== null) {
                                field.handleChange(value)
                              }
                            }}
                          >
                            <SelectTrigger
                              id="target-currency"
                              className="w-full"
                            >
                              <SelectValue />
                            </SelectTrigger>
                            <SelectContent>
                              {options.map((option) => (
                                <SelectItem
                                  key={option.code}
                                  value={option.code}
                                >
                                  <kbd className="rounded border border-border bg-muted px-1.5 py-0.5 font-mono text-xs font-medium text-muted-foreground">
                                    {option.code}
                                  </kbd>
                                  <span>{option.name}</span>
                                </SelectItem>
                              ))}
                            </SelectContent>
                          </Select>
                        </Field>
                      )}
                    </form.Field>
                    <form.Subscribe
                      selector={(state) => [
                        state.canSubmit,
                        state.isSubmitting,
                      ]}
                    >
                      {([canSubmit, isSubmitting]) => (
                        <Button
                          type="submit"
                          disabled={!canSubmit || isSubmitting}
                        >
                          Add pair
                        </Button>
                      )}
                    </form.Subscribe>
                  </div>
                  <form.Field name="target">
                    {(field) => (
                      <FieldError
                        className="border-t border-border pt-4"
                        errors={field.state.meta.errors}
                      />
                    )}
                  </form.Field>
                </form>
              </CardContent>
            </Card>
          </Dialog.Popup>
        </Dialog.Viewport>
      </Dialog.Portal>
    </Dialog.Root>
  )
}
