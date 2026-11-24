import { DashboardService } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { createFileRoute } from "@tanstack/react-router";
import { Gauge } from "lucide-react";
import { pageTitle } from "#/config";
import { op, requireOperations } from "#/features/layout-dashboard/auth/permissions";
import { OverviewPage } from "#/features/modules/dashboard-overview";
import { m } from "#/paraglide/messages";

const operations = [op(DashboardService, "getOverview"), op(DashboardService, "getTrend")];

export const Route = createFileRoute("/dashboard/home/overview")({
  staticData: {
    title: "dashboard_overview_title",
    menu: { icon: Gauge, order: 0 },
    operations,
  },
  head: () => ({ meta: [{ title: pageTitle(m.dashboard_overview_title()) }] }),
  beforeLoad: ({ context }) => requireOperations(operations, context.permissions),
  component: OverviewPage,
});
