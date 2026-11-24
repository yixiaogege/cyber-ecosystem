import type { GetTrendResponse } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { Granularity } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { fmt, niceStepMax } from "./format";
import { AXIS_TEXT, GRID, S1, S2 } from "./palette";

// 画布几何与 manage 原型一致：左轴留 58、右侧留 70 给末端数值标注、底部 36 给横轴标签。
const W = 1160;
const H = 262;
const PL = 58;
const PR = 70;
const PT = 18;
const PB = 36;
const X0 = PL;
const X1 = W - PR;
const Y0 = PT;
const Y1 = H - PB;

interface Series {
  key: string;
  label: string;
  color: string;
  values: number[];
}

function seriesOf(data: GetTrendResponse): Series[] {
  return [
    {
      key: "reg",
      label: m.dashboard_overview_trend_registrations(),
      color: S1,
      values: data.registrations.map(Number),
    },
    {
      key: "test",
      label: m.dashboard_overview_trend_tests(),
      color: S2,
      values: data.completedTests.map(Number),
    },
  ];
}

export function TrendChart({ data }: Readonly<{ data: GetTrendResponse }>) {
  const [hover, setHover] = useState<number | null>(null);
  const labels = data.labels;
  const series = seriesOf(data);
  const count = labels.length;
  const max = niceStepMax(Math.max(0, ...series.flatMap((s) => s.values)));

  const x = (i: number) => (count === 1 ? (X0 + X1) / 2 : X0 + ((X1 - X0) * i) / (count - 1));
  const y = (v: number) => Y1 - (Y1 - Y0) * (v / max);

  return (
    <div className="relative px-4 pt-3 pb-1">
      <svg
        aria-label={m.dashboard_overview_trend_aria()}
        className="block h-auto w-full"
        role="img"
        viewBox={`0 0 ${W} ${H}`}
      >
        {/* 刻度：4 等分 → 上限取整步长，故标签恒为整数 */}
        {[4, 3, 2, 1, 0].map((i) => {
          const value = (max * i) / 4;
          const gy = y(value);
          return (
            <g key={i}>
              <line stroke={GRID} strokeWidth={1} x1={X0} x2={X1} y1={gy} y2={gy} />
              <text
                fill={AXIS_TEXT}
                fontSize={10.5}
                style={{ fontVariantNumeric: "tabular-nums" }}
                textAnchor="end"
                x={X0 - 10}
                y={gy + 4}
              >
                {fmt(value)}
              </text>
            </g>
          );
        })}
        {labels.map((label, i) => (
          <text
            fill={AXIS_TEXT}
            fontSize={10.5}
            key={label}
            textAnchor="middle"
            x={x(i)}
            y={Y1 + 19}
          >
            {label}
          </text>
        ))}
        {hover !== null ? (
          <line stroke="#94a3b8" strokeWidth={1} x1={x(hover)} x2={x(hover)} y1={Y0} y2={Y1} />
        ) : null}
        {series.map((s) => (
          <g key={s.key}>
            <polyline
              fill="none"
              points={s.values.map((v, i) => `${x(i)},${y(v)}`).join(" ")}
              stroke={s.color}
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
            />
            {count > 0 ? (
              <>
                <circle
                  cx={x(count - 1)}
                  cy={y(s.values[count - 1] ?? 0)}
                  fill={s.color}
                  r={4.5}
                  stroke="var(--ant-color-bg-container, #fff)"
                  strokeWidth={2}
                />
                <text
                  fill={AXIS_TEXT}
                  fontSize={11}
                  fontWeight={700}
                  style={{ fontVariantNumeric: "tabular-nums" }}
                  x={x(count - 1) + 11}
                  y={y(s.values[count - 1] ?? 0) + 4}
                >
                  {fmt(s.values[count - 1] ?? 0)}
                </text>
              </>
            ) : null}
          </g>
        ))}
        {/* biome-ignore lint/a11y/noStaticElementInteractions: 十字准星是指针增强；同一份数据由 svg 的 aria-label 与「数据表」开关无障碍提供 */}
        <rect
          fill="transparent"
          height={Y1 - Y0}
          onMouseLeave={() => setHover(null)}
          onMouseMove={(event) => {
            const rect = event.currentTarget.getBoundingClientRect();
            const vx = ((event.clientX - rect.left) / rect.width) * W;
            const step = (X1 - X0) / Math.max(1, count - 1);
            const index = Math.max(0, Math.min(count - 1, Math.round((vx - X0) / step)));
            setHover(index);
          }}
          style={{ cursor: "crosshair" }}
          width={X1 - X0}
          x={X0}
          y={Y0}
        />
      </svg>
      {hover !== null && labels[hover] !== undefined ? (
        <div
          className="pointer-events-none absolute top-6 left-6 z-10 rounded-lg border border-line bg-elevated px-3 py-2 text-ink text-xs leading-relaxed shadow-layer-md"
          style={{ fontVariantNumeric: "tabular-nums" }}
        >
          <div className="text-ink-tertiary">{labels[hover]}</div>
          {series.map((s) => (
            <div className="flex items-center gap-1.5" key={s.key}>
              <i className="inline-block size-2.5 rounded-[2px]" style={{ background: s.color }} />
              {s.label} <b>{fmt(s.values[hover] ?? 0)}</b>
            </div>
          ))}
        </div>
      ) : null}
    </div>
  );
}

export function TrendTable({ data }: Readonly<{ data: GetTrendResponse }>) {
  const reg = data.registrations.map(Number);
  const test = data.completedTests.map(Number);
  const sum = (values: number[]) => values.reduce((a, b) => a + b, 0);
  return (
    <div className="px-4 pt-2 pb-3">
      <table className="w-full border-collapse text-xs">
        <thead>
          <tr className="text-ink-tertiary">
            <th className="border-line border-b px-3 py-2 text-left font-normal">
              {m.dashboard_overview_trend_col_bucket()}
            </th>
            <th className="border-line border-b px-3 py-2 text-left font-normal">
              {m.dashboard_overview_trend_registrations()}
            </th>
            <th className="border-line border-b px-3 py-2 text-left font-normal">
              {m.dashboard_overview_trend_tests()}
            </th>
          </tr>
        </thead>
        <tbody>
          {data.labels.map((label, i) => (
            <tr className="text-ink-secondary" key={label}>
              <td className="border-line-soft border-b px-3 py-2">{label}</td>
              <td className="border-line-soft border-b px-3 py-2 tabular-nums">
                {fmt(reg[i] ?? 0)}
              </td>
              <td className="border-line-soft border-b px-3 py-2 tabular-nums">
                {fmt(test[i] ?? 0)}
              </td>
            </tr>
          ))}
          <tr className="text-ink">
            <td className="px-3 py-2 font-semibold">{m.dashboard_overview_trend_total()}</td>
            <td className="px-3 py-2 font-semibold tabular-nums">{fmt(sum(reg))}</td>
            <td className="px-3 py-2 font-semibold tabular-nums">{fmt(sum(test))}</td>
          </tr>
        </tbody>
      </table>
    </div>
  );
}

export function trendSubtitle(granularity: Granularity): string {
  if (granularity === Granularity.MONTH) return m.dashboard_overview_trend_sub_month();
  if (granularity === Granularity.YEAR) return m.dashboard_overview_trend_sub_year();
  return m.dashboard_overview_trend_sub_week();
}
