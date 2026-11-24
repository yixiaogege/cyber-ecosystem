import type {
  GetOverviewResponse,
  GetTrendResponse,
  Granularity,
  Window,
} from "@cyber-ecosystem/gen-connect-ts/cyber/system/v1/dashboard_pb";
import { Skeleton } from "antd";
import { useState } from "react";
import { m } from "#/paraglide/messages";
import { windowLabel } from "./ui/filter-bar";
import { FunnelChart, FunnelTable } from "./ui/funnel";
import { InviterRank } from "./ui/inviters";
import { MetricCards } from "./ui/metric-cards";
import { S1, S2 } from "./ui/palette";
import { Legend, Note, Panel, TableToggle } from "./ui/panel";
import { QualityChart } from "./ui/quality";
import { QuickLinks } from "./ui/quick-links";
import { RecentActivity } from "./ui/recent-activity";
import { TrendChart, TrendTable, trendSubtitle } from "./ui/trend";
import { levelViews, ViralKpi, ViralLevels, ViralTable } from "./ui/viral";

// 各部件自持「图表 / 数据表」开关，页面只负责取数与编排。

const SKELETON_ROWS = { short: 2, medium: 3, tall: 4 };

export function MetricSection({
  data,
  isPending,
  isDay,
}: Readonly<{ data?: GetOverviewResponse; isPending: boolean; isDay: boolean }>) {
  if (isPending) return <Skeleton active paragraph={{ rows: SKELETON_ROWS.short }} />;
  return <MetricCards cards={data?.metricCards ?? []} isDay={isDay} />;
}

export function TrendSection({
  data,
  isPending,
  granularity,
}: Readonly<{ data?: GetTrendResponse; isPending: boolean; granularity: Granularity }>) {
  const [asTable, setAsTable] = useState(false);
  return (
    <Panel
      actions={<TableToggle active={asTable} onClick={() => setAsTable((v) => !v)} />}
      legend={
        <Legend
          items={[
            { color: S1, label: m.dashboard_overview_trend_registrations() },
            { color: S2, label: m.dashboard_overview_trend_tests() },
          ]}
        />
      }
      sub={trendSubtitle(granularity)}
      title={m.dashboard_overview_trend_title()}
    >
      {isPending || !data ? (
        <Skeleton active paragraph={{ rows: SKELETON_ROWS.tall }} />
      ) : (
        <TrendBody asTable={asTable} data={data} />
      )}
    </Panel>
  );
}

function TrendBody({ data, asTable }: Readonly<{ data: GetTrendResponse; asTable: boolean }>) {
  return asTable ? <TrendTable data={data} /> : <TrendChart data={data} />;
}

export function FunnelSection({
  data,
  window,
}: Readonly<{ data?: GetOverviewResponse; window: Window }>) {
  const [asTable, setAsTable] = useState(false);
  const funnel = data?.funnel;
  return (
    <Panel
      actions={<TableToggle active={asTable} onClick={() => setAsTable((v) => !v)} />}
      sub={windowLabel(window)}
      title={m.dashboard_overview_funnel_title()}
    >
      {funnel ? (
        asTable ? (
          <>
            <FunnelTable funnel={funnel} />
            <Note>{m.dashboard_overview_funnel_note_table()}</Note>
          </>
        ) : (
          <>
            <FunnelChart funnel={funnel} />
            <Note>
              {m.dashboard_overview_funnel_note({ entered: String(Number(funnel.enteredTests)) })}
            </Note>
          </>
        )
      ) : (
        <Skeleton active paragraph={{ rows: SKELETON_ROWS.medium }} />
      )}
    </Panel>
  );
}

export function ViralSection({
  data,
  window,
}: Readonly<{ data?: GetOverviewResponse; window: Window }>) {
  const [asTable, setAsTable] = useState(false);
  const viral = data?.viral;
  const levels = levelViews(viral?.levels ?? []);
  return (
    <Panel
      actions={<TableToggle active={asTable} onClick={() => setAsTable((v) => !v)} />}
      sub={windowLabel(window)}
      title={m.dashboard_overview_viral_title()}
    >
      {viral ? (
        <ViralBody asTable={asTable} levels={levels} viral={viral} />
      ) : (
        <Skeleton active paragraph={{ rows: SKELETON_ROWS.medium }} />
      )}
    </Panel>
  );
}

function ViralBody({
  viral,
  levels,
  asTable,
}: Readonly<{
  viral: NonNullable<GetOverviewResponse["viral"]>;
  levels: ReturnType<typeof levelViews>;
  asTable: boolean;
}>) {
  if (asTable) return <ViralTable levels={levels} viral={viral} />;
  return (
    <>
      <ViralKpi viral={viral} />
      <Note>{m.dashboard_overview_viral_note({ k: viral.kFactor.toFixed(3) })}</Note>
      <ViralLevels viral={levels} />
      <Note>{m.dashboard_overview_level_note()}</Note>
    </>
  );
}

export function InviterPanel({ data }: Readonly<{ data?: GetOverviewResponse }>) {
  return (
    <Panel sub={m.dashboard_overview_inviter_sub()} title={m.dashboard_overview_inviter_title()}>
      {data ? (
        <InviterRank inviters={data.topInviters} />
      ) : (
        <Skeleton active paragraph={{ rows: SKELETON_ROWS.medium }} />
      )}
    </Panel>
  );
}

export function QualityPanel({ data }: Readonly<{ data?: GetOverviewResponse }>) {
  return (
    <Panel
      legend={
        <Legend
          items={[
            { color: S1, label: m.dashboard_overview_quality_invited(), shape: "square" },
            { color: S2, label: m.dashboard_overview_quality_organic(), shape: "square" },
          ]}
        />
      }
      title={m.dashboard_overview_quality_title()}
    >
      {data ? (
        <>
          <QualityChart rows={data.quality} />
          <Note>{m.dashboard_overview_quality_note()}</Note>
        </>
      ) : (
        <Skeleton active paragraph={{ rows: SKELETON_ROWS.medium }} />
      )}
    </Panel>
  );
}

export function ActivitySection({
  data,
  isDay,
}: Readonly<{ data?: GetOverviewResponse; isDay: boolean }>) {
  return (
    <Panel title={m.dashboard_overview_activity_title()}>
      <RecentActivity isDay={isDay} items={data?.recentActivity ?? []} />
    </Panel>
  );
}

export function LinksSection() {
  return (
    <Panel title={m.dashboard_overview_links_title()}>
      <QuickLinks />
    </Panel>
  );
}
