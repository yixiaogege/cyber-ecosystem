import { Granularity, Window } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { Segmented } from "antd";
import { m } from "#/paraglide/messages";

// 唯一筛选行：时间窗驱动指标卡 / 漏斗 / 裂变 / 动态，趋势粒度只驱动趋势图——两个筛选器互不影响。
export function FilterBar({
  window,
  granularity,
  onWindow,
  onGranularity,
}: Readonly<{
  window: Window;
  granularity: Granularity;
  onWindow: (value: Window) => void;
  onGranularity: (value: Granularity) => void;
}>) {
  return (
    <div className="flex flex-wrap items-center gap-5">
      <div className="flex items-center gap-2">
        <span className="text-ink-tertiary text-xs">{m.dashboard_overview_filter_window()}</span>
        <Segmented
          onChange={(value) => onWindow(value as Window)}
          options={[
            { label: m.dashboard_overview_window_day(), value: Window.WINDOW_DAY },
            { label: m.dashboard_overview_window_7d(), value: Window.WINDOW_7D },
            { label: m.dashboard_overview_window_30d(), value: Window.WINDOW_30D },
          ]}
          value={window}
        />
      </div>
      <div className="flex items-center gap-2">
        <span className="text-ink-tertiary text-xs">
          {m.dashboard_overview_filter_granularity()}
        </span>
        <Segmented
          onChange={(value) => onGranularity(value as Granularity)}
          options={[
            { label: m.dashboard_overview_granularity_week(), value: Granularity.WEEK },
            { label: m.dashboard_overview_granularity_month(), value: Granularity.MONTH },
            { label: m.dashboard_overview_granularity_year(), value: Granularity.YEAR },
          ]}
          value={granularity}
        />
      </div>
    </div>
  );
}

// windowLabel 是各部件副标题用的窗口名，与筛选行同源。
export function windowLabel(window: Window): string {
  if (window === Window.WINDOW_7D) return m.dashboard_overview_window_7d();
  if (window === Window.WINDOW_30D) return m.dashboard_overview_window_30d();
  return m.dashboard_overview_window_day();
}
