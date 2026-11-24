import type { Funnel } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { m } from "#/paraglide/messages";
import { fmt, pct } from "./format";
import { ORDINAL_FUNNEL } from "./palette";

// toJson 把 proto 枚举转成字符串名，Record 键用字符串。
const STAGE_LABEL: Record<string, () => string> = {
  FUNNEL_STAGE_KEY_UNSPECIFIED: () => m.dashboard_overview_funnel_stage_complete_test(),
  FUNNEL_STAGE_KEY_COMPLETE_TEST: () => m.dashboard_overview_funnel_stage_complete_test(),
  FUNNEL_STAGE_KEY_REPORT_VIEW: () => m.dashboard_overview_funnel_stage_report_view(),
  FUNNEL_STAGE_KEY_REPORT_CENTER_VIEW: () => m.dashboard_overview_funnel_stage_report_center(),
  FUNNEL_STAGE_KEY_SHARE: () => m.dashboard_overview_funnel_stage_share(),
  FUNNEL_STAGE_KEY_MEMBER_ACTIVATE: () => m.dashboard_overview_funnel_stage_member(),
};

const ROW = "grid grid-cols-[120px_1fr_58px_72px_66px] items-center gap-2.5 text-xs";

function enumKey(e: number): string {
  return e as unknown as string;
}

function stageLabel(key: number): string {
  const fn = STAGE_LABEL[enumKey(key)] ?? STAGE_LABEL.FUNNEL_STAGE_KEY_UNSPECIFIED;
  return fn ? fn() : "";
}

export function FunnelChart({ funnel }: Readonly<{ funnel: Funnel }>) {
  return (
    <div className="flex flex-col gap-2 px-4 pt-3 pb-3">
      <div className={`${ROW} text-[10.5px] text-ink-tertiary`}>
        <span>{m.dashboard_overview_funnel_col_stage()}</span>
        <span />
        <span className="text-right">{m.dashboard_overview_funnel_col_count()}</span>
        <span className="text-right">{m.dashboard_overview_funnel_col_step()}</span>
        <span className="text-right">{m.dashboard_overview_funnel_col_cumulative()}</span>
      </div>
      {funnel.stages.map((stage, i) => (
        <div className={ROW} key={stage.key}>
          <span className="whitespace-nowrap text-ink-secondary">{stageLabel(stage.key)}</span>
          <span className="h-3 overflow-hidden rounded-md bg-fill-tertiary">
            <span
              className="block h-full rounded-md"
              style={{ background: ORDINAL_FUNNEL[i], width: `${stage.ofFirst.toFixed(1)}%` }}
            />
          </span>
          <span className="text-right font-bold text-ink-secondary tabular-nums">
            {fmt(Number(stage.count))}
          </span>
          <span className="text-right text-ink-tertiary tabular-nums">
            {stage.stepConversion === undefined ? "—" : pct(stage.stepConversion)}
          </span>
          <span className="text-right text-ink-tertiary tabular-nums">{pct(stage.ofFirst)}</span>
        </div>
      ))}
    </div>
  );
}

export function FunnelTable({ funnel }: Readonly<{ funnel: Funnel }>) {
  return (
    <div className="px-4 pt-2 pb-3">
      <table className="w-full border-collapse text-xs">
        <thead>
          <tr className="text-ink-tertiary">
            {[
              m.dashboard_overview_funnel_col_stage(),
              m.dashboard_overview_funnel_col_count(),
              m.dashboard_overview_funnel_col_step(),
              m.dashboard_overview_funnel_col_cumulative(),
            ].map((title) => (
              <th className="border-line border-b px-3 py-2 text-left font-normal" key={title}>
                {title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {funnel.stages.map((stage) => (
            <tr className="text-ink-secondary" key={stage.key}>
              <td className="border-line-soft border-b px-3 py-2">{stageLabel(stage.key)}</td>
              <td className="border-line-soft border-b px-3 py-2 tabular-nums">
                {fmt(Number(stage.count))}
              </td>
              <td className="border-line-soft border-b px-3 py-2 tabular-nums">
                {stage.stepConversion === undefined ? "—" : pct(stage.stepConversion)}
              </td>
              <td className="border-line-soft border-b px-3 py-2 tabular-nums">
                {pct(stage.ofFirst)}
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
