// @vitest-environment jsdom

import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { act, cleanup, render, screen } from "@testing-library/react"
import { afterEach, describe, expect, it, vi } from "vitest"

const mocks = vi.hoisted(() => ({
  getRatesQueryOptions: vi.fn(),
}))

vi.mock("@/lib/api/client", () => ({
  getRatesQueryOptions: mocks.getRatesQueryOptions,
}))

import { RateGroup } from "@/components/dashboard/rate-group"

const startedAt = new Date("2026-01-02T03:04:05Z").getTime()
const pairs = [
  { base: "USD", target: "EUR" },
  { base: "USD", target: "SGD" },
] as const

afterEach(() => {
  cleanup()
  mocks.getRatesQueryOptions.mockReset()
  vi.useRealTimers()
})

describe("RateGroup", () => {
  it("refreshes at the earliest server-supplied expiry", async () => {
    vi.useFakeTimers({ now: startedAt })
    const queryFn = vi
      .fn()
      .mockResolvedValueOnce({
        base: "USD",
        rates: {
          EUR: { rate: 0.85, expiresAt: "2026-01-02T03:04:15Z" },
          SGD: { rate: 1.35, expiresAt: "2026-01-02T03:04:25Z" },
        },
      })
      .mockResolvedValueOnce({
        base: "USD",
        rates: {
          EUR: { rate: 0.86, expiresAt: "2026-01-02T03:04:35Z" },
          SGD: { rate: 1.35, expiresAt: "2026-01-02T03:04:25Z" },
        },
      })
    mocks.getRatesQueryOptions.mockReturnValue({
      queryKey: ["rates", "USD", "EUR", "SGD"],
      queryFn,
    })
    const queryClient = new QueryClient({
      defaultOptions: { queries: { retry: false } },
    })

    render(
      <QueryClientProvider client={queryClient}>
        <RateGroup base="USD" pairs={pairs} onRemove={async () => {}} />
      </QueryClientProvider>
    )

    await act(async () => {
      await vi.advanceTimersByTimeAsync(0)
      await Promise.resolve()
    })
    expect(queryFn).toHaveBeenCalledTimes(1)
    expect(screen.getByText("0.85")).toBeTruthy()

    await act(async () => {
      await vi.advanceTimersByTimeAsync(9_999)
    })
    expect(queryFn).toHaveBeenCalledTimes(1)

    await act(async () => {
      await vi.advanceTimersByTimeAsync(1)
      await Promise.resolve()
      await vi.advanceTimersByTimeAsync(0)
    })
    expect(queryFn).toHaveBeenCalledTimes(2)
    expect(
      queryClient.getQueryData(["rates", "USD", "EUR", "SGD"])
    ).toMatchObject({
      rates: {
        EUR: { rate: 0.86, expiresAt: "2026-01-02T03:04:35Z" },
      },
    })
  })
})
