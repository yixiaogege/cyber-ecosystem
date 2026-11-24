const NUMBER = new Intl.NumberFormat("en-US");

// fmt 与 manage 原型的 fmt 一致：千分位，整数展示（proto 侧 int64 已保证整数值）。
export const fmt = (value: number | bigint): string => NUMBER.format(value);

export const pct = (value: number, digits = 1): string => `${value.toFixed(digits)}%`;

export const fixed = (value: number, digits = 1): string => value.toFixed(digits);

// signedPP 用于百分点差值：正数显式带 +，负数自带 -。
export const signedPP = (value: number): string => `${value > 0 ? "+" : ""}${value.toFixed(1)}pp`;

// 步长阶梯与 manage 原型一致；取第一个不小于 n 的档位，都超出则退到 10。
const STEPS = [1, 1.5, 2, 2.5, 3, 4, 5, 6, 7.5] as const;

// niceStepMax 先取步长再定上限，使 4 等分刻度恒为整数（避免 6,250 这类刻度）。
export function niceStepMax(value: number): number {
  if (value <= 0) return 4;
  const raw = value / 4;
  const power = 10 ** Math.floor(Math.log10(raw));
  const step = STEPS.find((candidate) => raw / power <= candidate) ?? 10;
  return step * power * 4;
}
