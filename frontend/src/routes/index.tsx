import { createFileRoute } from "@tanstack/react-router"

import {
  Dashboard,
  DashboardLoading,
  DashboardUnavailable,
} from "@/components/dashboard/dashboard"
import { getCurrenciesQueryOptions } from "@/lib/api/client"

export const Route = createFileRoute("/")({
  loader: async ({ context }) => {
    await context.queryClient.ensureQueryData(getCurrenciesQueryOptions())
  },
  pendingComponent: DashboardLoading,
  errorComponent: DashboardUnavailable,
  component: Dashboard,
})
