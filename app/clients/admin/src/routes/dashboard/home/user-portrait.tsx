import { createFileRoute } from "@tanstack/react-router";
import { Users } from "lucide-react";
import { pageTitle } from "#/config";
import { UserPortraitPage } from "#/features/modules/dashboard-user-portrait";
import { m } from "#/paraglide/messages";

// 占位页无 operations：对所有人可见，不参与权限隐藏。
export const Route = createFileRoute("/dashboard/home/user-portrait")({
  staticData: {
    title: "dashboard_overview_user_portrait_title",
    menu: { icon: Users, order: 1 },
  },
  head: () => ({ meta: [{ title: pageTitle(m.dashboard_overview_user_portrait_title()) }] }),
  component: UserPortraitPage,
});
