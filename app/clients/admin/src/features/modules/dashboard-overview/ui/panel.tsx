import { Button } from "antd";
import type { ReactNode } from "react";
import { m } from "#/paraglide/messages";

interface PanelProps {
  title: string;
  sub?: string;
  legend?: ReactNode;
  actions?: ReactNode;
  children: ReactNode;
}

// Panel 是所有部件的统一外壳：标题 + 可选副标题 + 图例 + 右侧操作 + 主体。
export function Panel({ title, sub, legend, actions, children }: Readonly<PanelProps>) {
  return (
    <section className="rounded-lg border border-line bg-container">
      <header className="flex items-center gap-3 border-line border-b px-4 py-3">
        <h2 className="font-semibold text-ink text-sm">{title}</h2>
        {sub ? <span className="font-normal text-ink-tertiary text-xs">{sub}</span> : null}
        <span className="flex-1" />
        {legend}
        {actions}
      </header>
      {children}
    </section>
  );
}

interface LegendItem {
  label: string;
  color: string;
  shape?: "line" | "square";
}

export function Legend({ items }: Readonly<{ items: readonly LegendItem[] }>) {
  return (
    <span className="flex items-center gap-3.5 text-ink-secondary text-xs">
      {items.map((item) => (
        <span className="inline-flex items-center gap-1.5" key={item.label}>
          <i
            className="block rounded-[2px]"
            style={{
              background: item.color,
              height: item.shape === "square" ? 10 : 3,
              width: item.shape === "square" ? 10 : 14,
            }}
          />
          {item.label}
        </span>
      ))}
    </span>
  );
}

// TableToggle 是「数据表 / 图表」切换按钮，三个部件共用同一形态。
export function TableToggle({
  active,
  onClick,
}: Readonly<{ active: boolean; onClick: () => void }>) {
  return (
    <Button
      color={active ? "primary" : "default"}
      onClick={onClick}
      size="small"
      variant={active ? "outlined" : "text"}
    >
      {active ? m.dashboard_overview_toggle_chart() : m.dashboard_overview_toggle_table()}
    </Button>
  );
}

// Note 是部件底部的口径注（含 code 片段），三个部件共用。
export function Note({ children }: Readonly<{ children: ReactNode }>) {
  return (
    <p className="px-4 pb-3 text-ink-tertiary text-xs leading-relaxed [&_b]:text-ink-secondary [&_code]:rounded [&_code]:bg-fill-tertiary [&_code]:px-1 [&_code]:py-px [&_code]:font-mono [&_code]:text-[10.5px]">
      {children}
    </p>
  );
}
