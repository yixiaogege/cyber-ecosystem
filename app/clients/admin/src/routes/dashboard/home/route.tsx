import { createFileRoute } from "@tanstack/react-router";
import { LayoutDashboard } from "lucide-react";

export const Route = createFileRoute("/dashboard/home")({
  staticData: {
    title: "layout_dashboard_dashboard",
    menu: { icon: LayoutDashboard, order: 210 },
  },
});
