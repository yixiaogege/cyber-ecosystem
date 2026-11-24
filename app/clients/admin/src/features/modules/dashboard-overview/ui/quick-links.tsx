import { Link } from "@tanstack/react-router";
import { m } from "#/paraglide/messages";

interface QuickLink {
  key: string;
  label: () => string;
  to?: string;
}

// 入口清单对齐 manage 原型的 13 项。本仓只落地了用户画像与用户列表；
// 其余属于上游业务域（会员 / 订单 / SKU / 题库 …），本轮不建，故渲染为置灰占位，
// 与原型里「报告配置（非本期）」的灰标同一处理，不做死链。
const LINKS: readonly QuickLink[] = [
  {
    key: "user_portrait",
    label: m.dashboard_overview_link_user_portrait,
    to: "/dashboard/home/user-portrait",
  },
  { key: "user_list", label: m.dashboard_overview_link_user_list, to: "/dashboard/system/users" },
  { key: "member", label: m.dashboard_overview_link_member },
  { key: "order", label: m.dashboard_overview_link_order },
  { key: "sku", label: m.dashboard_overview_link_sku },
  { key: "tenant", label: m.dashboard_overview_link_tenant },
  { key: "report_config", label: m.dashboard_overview_link_report_config },
  { key: "question_bank", label: m.dashboard_overview_link_question_bank },
  { key: "rating", label: m.dashboard_overview_link_rating },
  { key: "activation", label: m.dashboard_overview_link_activation },
  { key: "announcement", label: m.dashboard_overview_link_announcement },
  { key: "app_version", label: m.dashboard_overview_link_app_version },
  { key: "feedback", label: m.dashboard_overview_link_feedback },
];

const SHELL =
  "inline-flex items-center gap-2 rounded-[10px] border border-line bg-container px-3 py-2 text-xs";

export function QuickLinks() {
  return (
    <div className="flex flex-wrap gap-2 px-4 py-3">
      {LINKS.map((link) =>
        link.to ? (
          <Link
            className={`${SHELL} text-primary hover:border-line-hover`}
            key={link.key}
            to={link.to}
          >
            {link.label()}
          </Link>
        ) : (
          <span aria-disabled className={`${SHELL} text-ink-disabled`} key={link.key}>
            {link.label()}
            <span className="rounded border border-line px-1 text-[9.5px]">
              {m.dashboard_overview_link_out_of_scope()}
            </span>
          </span>
        ),
      )}
    </div>
  );
}
