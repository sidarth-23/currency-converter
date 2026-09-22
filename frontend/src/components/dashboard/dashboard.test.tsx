// @vitest-environment jsdom

import { cleanup, render, screen, waitFor } from "@testing-library/react"
import { RouterProvider } from "@tanstack/react-router"
import { afterEach, describe, expect, it, vi } from "vitest"
window.scrollTo = vi.fn()

const mocks = vi.hoisted(() => ({
  getCurrenciesQueryOptions: vi.fn(),
}))

vi.mock("@/lib/api/client", () => ({
  getCurrenciesQueryOptions: mocks.getCurrenciesQueryOptions,
}))

vi.mock("@tanstack/react-devtools", () => ({
  TanStackDevtools: () => null,
}))

vi.mock("@tanstack/react-router-devtools", () => ({
  TanStackRouterDevtoolsPanel: () => null,
}))

import { DashboardLoading } from "@/components/dashboard/dashboard"
import { getRouter } from "@/router"
import { Route } from "@/routes/index"

afterEach(() => {
  cleanup()
  mocks.getCurrenciesQueryOptions.mockReset()
})

describe("currency catalog route", () => {
  it("renders the dashboard heading without a watch count while loading", () => {
    render(<DashboardLoading />)

    expect(screen.getByText("Currency watcher")).toBeTruthy()
    expect(
      screen.getByRole("heading", { name: "Keep an eye on your rates." })
    ).toBeTruthy()
    expect(screen.queryByText(/watched$/)).toBeNull()
    expect(
      screen.getByRole("status", { name: "Loading your watchlist…" })
    ).toBeTruthy()
  })
  it("loads the same currency query options used by the dashboard", async () => {
    const queryOptions = { queryKey: ["currencies"] }
    const ensureQueryData = vi.fn().mockResolvedValue(undefined)

    mocks.getCurrenciesQueryOptions.mockReturnValue(queryOptions)

    const loader = Route.options.loader
    if (typeof loader !== "function") {
      throw new Error("The index route must define a loader.")
    }

    await loader({ context: { queryClient: { ensureQueryData } } } as never)

    expect(ensureQueryData).toHaveBeenCalledWith(queryOptions)
  })

  it("renders only the unavailable state when the catalog query rejects", async () => {
    mocks.getCurrenciesQueryOptions.mockReturnValue({
      queryKey: ["currencies"],
      queryFn: () => Promise.reject(new Error("catalog unavailable")),
      staleTime: 60 * 60 * 1000,
    })
    const router = getRouter()

    await router.load()
    render(<RouterProvider router={router} />)

    await waitFor(() => {
      expect(screen.getByText("Currency catalog unavailable")).toBeTruthy()
    })
    expect(screen.getByText("Try again shortly.")).toBeTruthy()
    expect(screen.queryByText("Add a pair")).toBeNull()
    expect(screen.queryByText("No pairs yet")).toBeNull()
  })
})
