import { createFileRoute } from "@tanstack/react-router";
import { MessageSquare } from "lucide-react";
import { pageTitle } from "#/config";
import { AgentsChatPage, agentsChatSearchSchema } from "#/features/modules/agents-chat";
import { m } from "#/paraglide/messages";

export const Route = createFileRoute("/dashboard/")({
  staticData: {
    title: "layout_dashboard_workbench",
    menu: { icon: MessageSquare, order: 10 },
    fullHeight: true,
  },
  validateSearch: agentsChatSearchSchema.parse,
  head: () => ({ meta: [{ title: pageTitle(m.layout_dashboard_workbench()) }] }),
  component: AgentsChatPage,
});
