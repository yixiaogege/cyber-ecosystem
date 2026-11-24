import type { Viral } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { m } from "#/paraglide/messages";
import { fixed, fmt, pct } from "./format";
import { ORDINAL_LEVEL } from "./palette";

const LEVEL_LABEL = [
  m.dashboard_overview_level_1,
  m.dashboard_overview_level_2,
  m.dashboard_overview_level_3,
  m.dashboard_overview_level_4,
];

const RANK_ROW = "grid grid-cols-[74px_1fr_54px_60px] items-center gap-2.5 text-xs";

// toJson 把 int64 转成字符串，Number() 还原。
function num(value: bigint | string | number): number {
  return Number(value);
}

export function ViralKpi({ viral }: Readonly<{ viral: Viral }>) {
  const cells = [
    {
      key: "referrals",
      label: m.dashboard_overview_viral_new_referrals(),
      value: fmt(num(viral.newReferrals)),
      note: m.dashboard_overview_viral_new_referrals_note(),
    },
    {
      key: "inviters",
      label: m.dashboard_overview_viral_inviters(),
      value: fmt(num(viral.inviters)),
      small: m.dashboard_overview_viral_per_inviter({ value: fixed(viral.perInviter, 2) }),
      note: m.dashboard_overview_viral_inviters_note(),
    },
    {
      key: "conversion",
      label: m.dashboard_overview_viral_conversion(),
      value: pct(viral.conversionRate),
      note: m.dashboard_overview_viral_conversion_note(),
    },
    {
      key: "k",
      label: m.dashboard_overview_viral_k(),
      value: fixed(viral.kFactor, 3),
      note: m.dashboard_overview_viral_k_note({
        referrals: fmt(num(viral.newReferrals)),
        active: fmt(num(viral.activeUsers)),
      }),
    },
  ];
  return (
    <div className="divide-line-soft border-line grid grid-cols-4 divide-x border-b">
      {cells.map((cell) => (
        <div className="px-4 py-3" key={cell.key}>
          <div className="text-ink-tertiary text-[11.5px]">{cell.label}</div>
          <div className="text-ink mt-0.5 text-[23px] font-black tabular-nums">
            {cell.value}
            {cell.small ? (
              <small className="text-ink-secondary ml-1 text-xs font-semibold">{cell.small}</small>
            ) : null}
          </div>
          <div className="text-ink-tertiary mt-0.5 text-[10.5px] leading-normal">{cell.note}</div>
        </div>
      ))}
    </div>
  );
}

export function ViralLevels({ viral }: Readonly<{ viral: readonly FissionLevelView[] }>) {
  const levels = viral;
  const max = Math.max(1, ...levels.map((l) => l.count));
  return (
    <div className="flex flex-col gap-2 px-4 pt-2 pb-3">
      <div className={`${RANK_ROW} border-line text-ink-tertiary border-b pb-1 text-[10.5px]`}>
        <span>{m.dashboard_overview_level_col()}</span>
        <span />
        <span className="text-right">{m.dashboard_overview_level_col_count()}</span>
        <span className="text-right">{m.dashboard_overview_level_col_share()}</span>
      </div>
      {levels.map((level, i) => (
        <div className={RANK_ROW} key={level.label}>
          <span className="text-ink-secondary whitespace-nowrap">{level.label}</span>
          <span className="bg-fill-tertiary h-2.5 overflow-hidden rounded-[5px]">
            <span
              className="block h-full rounded-[5px]"
              style={{
                background: ORDINAL_LEVEL[i],
                width: `${((level.count / max) * 100).toFixed(1)}%`,
              }}
            />
          </span>
          <span className="text-ink-secondary text-right font-bold tabular-nums">
            {fmt(level.count)}
          </span>
          <span className="text-ink-tertiary text-right tabular-nums">{pct(level.share)}</span>
        </div>
      ))}
    </div>
  );
}

export interface FissionLevelView {
  label: string;
  count: number;
  share: number;
}

// levelViews 把服务端层级（1..4，其中 4 表示「4 级+」）映射为可渲染的标签序列。
export function levelViews(
  levels: readonly { level: number; count: bigint; share: number }[],
): FissionLevelView[] {
  const byLevel = new Map(levels.map((l) => [l.level, l]));
  return LEVEL_LABEL.map((label, i) => {
    const found = byLevel.get(i + 1);
    return { label: label(), count: num(found?.count ?? 0), share: found?.share ?? 0 };
  });
}

export function ViralTable({
  viral,
  levels,
}: Readonly<{ viral: Viral; levels: readonly FissionLevelView[] }>) {
  const rows: [string, string][] = [
    [m.dashboard_overview_viral_new_referrals(), fmt(num(viral.newReferrals))],
    [m.dashboard_overview_viral_inviters(), fmt(num(viral.inviters))],
    [m.dashboard_overview_viral_per_inviter_row(), fixed(viral.perInviter, 2)],
    [m.dashboard_overview_viral_conversion(), pct(viral.conversionRate)],
    [m.dashboard_overview_viral_k(), fixed(viral.kFactor, 3)],
    [m.dashboard_overview_viral_active(), fmt(num(viral.activeUsers))],
    ...levels.map((level): [string, string] => [
      m.dashboard_overview_viral_level_row({ level: level.label }),
      fmt(level.count),
    ]),
  ];
  return (
    <div className="px-4 pt-2 pb-3">
      <table className="w-full border-collapse text-xs">
        <thead>
          <tr className="text-ink-tertiary">
            {[
              m.dashboard_overview_viral_table_metric(),
              m.dashboard_overview_viral_table_value(),
            ].map((title) => (
              <th className="border-line border-b px-3 py-2 text-left font-normal" key={title}>
                {title}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {rows.map(([label, value]) => (
            <tr className="text-ink-secondary" key={label}>
              <td className="border-line-soft border-b px-3 py-2">{label}</td>
              <td className="border-line-soft border-b px-3 py-2 tabular-nums">{value}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
