import type { Activity } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import dayjs from "dayjs";
import { m } from "#/paraglide/messages";

// 事件词表 → 展示标签。词表是服务端聚合口径的唯一来源（module/dashboard 的常量），
// 前端只负责翻译；未收录的 token 原样显示，便于上报侧新增事件时不被前端静默丢掉。
const TYPE_LABEL: Record<string, () => string> = {
  register: m.dashboard_overview_act_register,
  login: m.dashboard_overview_act_login,
  enter_test: m.dashboard_overview_act_enter_test,
  complete_test: m.dashboard_overview_act_complete_test,
  report_view: m.dashboard_overview_act_report_view,
  report_center_view: m.dashboard_overview_act_report_center,
  share: m.dashboard_overview_act_share,
  pay_success: m.dashboard_overview_act_pay,
  unlock_success: m.dashboard_overview_act_unlock,
  member_activate: m.dashboard_overview_act_member,
};

// toJson 把 Timestamp 转成 ISO 8601 字符串，dayjs 可直接解析。
function formatTime(occurredAt: Activity["occurredAt"], isDay: boolean): string {
  if (!occurredAt) return "—";
  const d = dayjs(occurredAt as unknown as string);
  return d.isValid() ? d.format(isDay ? "MM-DD HH:mm" : "MM-DD") : "—";
}

export function RecentActivity({
  items,
  isDay,
}: Readonly<{ items: readonly Activity[]; isDay: boolean }>) {
  if (items.length === 0) {
    return (
      <p className="px-4 py-6 text-center text-ink-tertiary text-xs">
        {m.dashboard_overview_act_empty()}
      </p>
    );
  }
  return (
    <table className="w-full border-collapse text-xs">
      <thead>
        <tr className="text-ink-tertiary">
          {[
            m.dashboard_overview_act_col_time(),
            m.dashboard_overview_act_col_type(),
            m.dashboard_overview_act_col_content(),
          ].map((title) => (
            <th className="border-line border-b px-4 py-2 text-left font-normal" key={title}>
              {title}
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {items.map((item) => (
          <tr
            className="text-ink-secondary"
            key={`${item.type}-${String(item.occurredAt)}-${item.content}`}
          >
            <td className="whitespace-nowrap border-line-soft border-b px-4 py-2 tabular-nums">
              {formatTime(item.occurredAt, isDay)}
            </td>
            <td className="whitespace-nowrap border-line-soft border-b px-4 py-2">
              {TYPE_LABEL[item.type]?.() ?? item.type}
            </td>
            <td className="border-line-soft border-b px-4 py-2">{item.content || "—"}</td>
          </tr>
        ))}
      </tbody>
    </table>
  );
}
