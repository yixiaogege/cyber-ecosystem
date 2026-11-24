import { createFileRoute } from "@tanstack/react-router";
import { firstLeafRedirect } from "#/features/layout-dashboard/protocol/first-leaf-redirect";

export const Route = createFileRoute("/dashboard/home/")({
  component: firstLeafRedirect("/dashboard/home"),
});
