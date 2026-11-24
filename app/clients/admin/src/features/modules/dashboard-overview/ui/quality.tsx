import type { QualityRow } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { m } from "#/paraglide/messages";
import { pct, signedPP } from "./format";
import { S1, S2 } from "./palette";

// 三指标同为比率 → 统一 0–80% 单轴（不做双轴）；小值柱虽短，但柱顶已直标数值，
// 另给百分点差值确保结论可读。
const MAX = 80;

const ROW = "grid grid-cols-[96px_1fr_74px] items-end gap-3";

// toJson 把 proto 枚举转成字符串名，Record 键用字符串。
const QUALITY_LABEL: Record<string, () => string> = {
  QUALITY_KEY_UNSPECIFIED: () => m.dashboard_overview_quality_test_rate(),
  QUALITY_KEY_TEST_RATE: () => m.dashboard_overview_quality_test_rate(),
  QUALITY_KEY_PAY_RATE: () => m.dashboard_overview_quality_pay_rate(),
  QUALITY_KEY_RETENTION_7D: () => m.dashboard_overview_quality_retention(),
};

function enumKey(e: number): string {
  return e as unknown as string;
}

function qualityLabel(key: number): string {
  const fn = QUALITY_LABEL[enumKey(key)] ?? QUALITY_LABEL.QUALITY_KEY_UNSPECIFIED;
  return fn ? fn() : "";
}

export function QualityChart({ rows }: Readonly<{ rows: readonly QualityRow[] }>) {
  return (
    <div className="flex flex-col gap-3 px-4 pt-3 pb-2">
      <div className={`${ROW} border-line border-b pb-0.5 text-[10.5px] text-ink-tertiary`}>
        <span>{m.dashboard_overview_quality_col_metric()}</span>
        <span>{m.dashboard_overview_quality_col_axis()}</span>
        <span className="text-right">{m.dashboard_overview_quality_col_delta()}</span>
      </div>
      {rows.map((row) => {
        const bars = [
          { key: "invited", value: row.invited, color: S1 },
          { key: "organic", value: row.organic, color: S2 },
        ];
        return (
          <div className={ROW} key={row.key}>
            <span className="pb-0.5 text-ink-secondary text-xs">{qualityLabel(row.key)}</span>
            <div className="flex h-[74px] items-end gap-2.5 border-line border-b">
              {bars.map((bar) => (
                <div className="flex h-full flex-col justify-end" key={bar.key}>
                  <span className="mb-1 text-[10.5px] text-ink-secondary tabular-nums">
                    {pct(bar.value)}
                  </span>
                  <span
                    className="block w-[34px] rounded-t"
                    style={{
                      background: bar.color,
                      height: `${((bar.value / MAX) * 100).toFixed(1)}%`,
                      minHeight: 2,
                    }}
                  />
                </div>
              ))}
            </div>
            <span
              className={`pb-0.5 text-right font-extrabold text-xs tabular-nums ${
                row.deltaPp >= 0 ? "text-success" : "text-error"
              }`}
            >
              {signedPP(row.deltaPp)}
            </span>
          </div>
        );
      })}
    </div>
  );
}
