import { useMemo } from "react"

import { useForm } from "@tanstack/react-form"
import { z } from "zod"

import { Button } from "@/components/ui/button"
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
import type { CurrencyOption } from "@/lib/currencies"
import type { WatchPair } from "@/lib/watchlist/database"

type WatchPairFormProps = {
  options: readonly CurrencyOption[]
  pairs: readonly WatchPair[]
  onAdd: (pair: WatchPair) => Promise<void>
}

function defaultWatchPair(options: readonly CurrencyOption[]): WatchPair {
  const codes = new Set(options.map((option) => option.code))
  if (codes.has("USD") && codes.has("EUR")) {
    return { base: "USD", target: "EUR" }
  }
  return { base: options[0]!.code, target: options[1]!.code }
}

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

export function WatchPairForm({ options, pairs, onAdd }: WatchPairFormProps) {
  const schema = useMemo(() => createWatchPairSchema(pairs), [pairs])
  const form = useForm({
    defaultValues: defaultWatchPair(options),
    validators: {
      onChange: schema,
      onSubmit: schema,
    },
    onSubmit: async ({ value }) => {
      await onAdd(value)
    },
  })

  return (
    <Card>
      <CardHeader>
        <CardTitle>Add a pair</CardTitle>
        <CardDescription>
          Track one quote currency against a base.
        </CardDescription>
      </CardHeader>
      <CardContent>
        <form
          className="grid gap-4 sm:grid-cols-[1fr_auto_1fr_auto] sm:items-end"
          onSubmit={(event) => {
            event.preventDefault()
            void form.handleSubmit()
          }}
        >
          <form.Field name="base">
            {(field) => (
              <Field>
                <FieldLabel htmlFor="base-currency">Base currency</FieldLabel>
                <Select
                  value={field.state.value}
                  onValueChange={(value) => {
                    if (value !== null) {
                      field.handleChange(value)
                    }
                  }}
                >
                  <SelectTrigger id="base-currency" className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {options.map((option) => (
                      <SelectItem key={option.code} value={option.code}>
                        {option.code} — {option.name}
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
              <Field
                data-invalid={field.state.meta.errors.length > 0 || undefined}
              >
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
                  <SelectTrigger id="target-currency" className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {options.map((option) => (
                      <SelectItem key={option.code} value={option.code}>
                        {option.code} — {option.name}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FieldError errors={field.state.meta.errors} />
              </Field>
            )}
          </form.Field>
          <form.Subscribe
            selector={(state) => [state.canSubmit, state.isSubmitting]}
          >
            {([canSubmit, isSubmitting]) => (
              <Button type="submit" disabled={!canSubmit || isSubmitting}>
                Add pair
              </Button>
            )}
          </form.Subscribe>
        </form>
      </CardContent>
    </Card>
  )
}
