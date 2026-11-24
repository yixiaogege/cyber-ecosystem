import { useQuery } from "@connectrpc/connect-query";
import { Granularity, Window } from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import {
  getOverview,
  getTrend,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard-DashboardService_connectquery";
import { useState } from "react";
import {
  ActivitySection,
  FunnelSection,
  InviterPanel,
  LinksSection,
  MetricSection,
  QualityPanel,
  TrendSection,
  ViralSection,
} from "./sections";
import { FilterBar } from "./ui/filter-bar";

// 运营总览：一个时间窗 + 一个趋势粒度驱动全部部件，页面本身只组装取数与编排。
export function OverviewPage() {
  const [window, setWindow] = useState<Window>(Window.WINDOW_DAY);
  const [granularity, setGranularity] = useState<Granularity>(Granularity.WEEK);

  const overview = useQuery(getOverview, { window });
  const trend = useQuery(getTrend, { granularity });
  const isDay = window === Window.WINDOW_DAY;

  return (
    <div className="flex flex-col gap-3 p-4">
      <FilterBar
        granularity={granularity}
        onGranularity={setGranularity}
        onWindow={setWindow}
        window={window}
      />

      <MetricSection data={overview.data} isDay={isDay} isPending={overview.isPending} />

      <TrendSection data={trend.data} granularity={granularity} isPending={trend.isPending} />

      <FunnelSection data={overview.data} window={window} />
      <ViralSection data={overview.data} window={window} />

      <div className="grid grid-cols-[1fr_1.15fr] gap-3">
        <InviterPanel data={overview.data} />
        <QualityPanel data={overview.data} />
      </div>

      <ActivitySection data={overview.data} isDay={isDay} />
      <LinksSection />
    </div>
  );
}
