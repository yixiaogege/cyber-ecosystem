import type { Inviter } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { m } from "#/paraglide/messages";
import { fmt, pct } from "./format";
import { S1 } from "./palette";

const ROW = "grid grid-cols-[74px_1fr_54px_60px] items-center gap-2.5 text-xs";

// toJson 把 int64 转成字符串，Number() 还原。
function num(value: bigint | string | number): number {
  return Number(value);
}

// 排行榜：条形长度按邀请数归一，单一色相（邀请数是同一种量，不是分类）。
export function InviterRank({ inviters }: Readonly<{ inviters: readonly Inviter[] }>) {
  const max = Math.max(1, ...inviters.map((i) => num(i.invites)));
  return (
    <div className="flex flex-col gap-2 px-4 pt-2 pb-3">
      <div className={`${ROW} border-line border-b pb-1 text-[10.5px] text-ink-tertiary`}>
        <span>{m.dashboard_overview_inviter_col()}</span>
        <span />
        <span className="text-right">{m.dashboard_overview_inviter_col_invites()}</span>
        <span className="text-right">{m.dashboard_overview_inviter_col_rate()}</span>
      </div>
      {inviters.map((inviter) => (
        <div className={ROW} key={inviter.userId ?? `${inviter.name}-${inviter.invites}`}>
          <span className="whitespace-nowrap text-ink-secondary">{inviter.name}</span>
          <span className="h-2.5 overflow-hidden rounded-[5px] bg-fill-tertiary">
            <span
              className="block h-full rounded-[5px]"
              style={{
                background: S1,
                width: `${((num(inviter.invites) / max) * 100).toFixed(1)}%`,
              }}
            />
          </span>
          <span className="text-right font-bold text-ink-secondary tabular-nums">
            {fmt(num(inviter.invites))}
          </span>
          <span className="text-right text-ink-tertiary tabular-nums">
            {pct(inviter.completionRate ?? 0)}
          </span>
        </div>
      ))}
    </div>
  );
}
