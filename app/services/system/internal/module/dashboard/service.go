package dashboard

import (
	"context"
	"log/slog"

	"github.com/go-kratos/kratos/v3/transport/grpc"
	"github.com/go-kratos/kratos/v3/transport/http"

	connecttransport "cyber-ecosystem/shared-go/kratos/transport/connect"
	"cyber-ecosystem/shared-go/utils"

	systempb "cyber-ecosystem/gen/go/cyber/system/v1"
)

// Struct --------------------------------------------------------------------------------------------------------------

type DashboardService struct {
	systempb.UnimplementedDashboardServiceServer

	log         *slog.Logger
	dashboardUC *DashboardUC
}

func NewDashboardService(logger *slog.Logger, dashboardUC *DashboardUC) *DashboardService {
	return &DashboardService{
		log:         logger.With("module", "module/dashboard_service"),
		dashboardUC: dashboardUC,
	}
}

func (s *DashboardService) RegisterGRPC(srv *grpc.Server) {
	systempb.RegisterDashboardServiceServer(srv, s)
}

func (s *DashboardService) RegisterHTTP(srv *http.Server) {
	systempb.RegisterDashboardServiceHTTPServer(srv, s)
}

func (s *DashboardService) RegisterConnect(srv *connecttransport.Server) {
	systempb.RegisterDashboardServiceConnectServer(srv, s)
}

// Handler -------------------------------------------------------------------------------------------------------------

func (s *DashboardService) GetOverview(ctx context.Context, in *systempb.GetOverviewRequest) (*systempb.GetOverviewResponse, error) {
	out, err := s.dashboardUC.Overview(ctx, toWindow(in.GetWindow()))
	if err != nil {
		return nil, err
	}
	return &systempb.GetOverviewResponse{
		MetricCards:    utils.SliceMap(out.Cards, toMetricCard),
		Funnel:         toFunnel(out.Funnel),
		Viral:          toViral(out.Viral),
		TopInviters:    utils.SliceMap(out.TopInviters, toInviter),
		Quality:        utils.SliceMap(out.Quality, toQualityRow),
		RecentActivity: utils.SliceMap(out.Recent, toActivity),
	}, nil
}

func (s *DashboardService) GetTrend(ctx context.Context, in *systempb.GetTrendRequest) (*systempb.GetTrendResponse, error) {
	out, err := s.dashboardUC.Trend(ctx, toGranularity(in.GetGranularity()))
	if err != nil {
		return nil, err
	}
	return &systempb.GetTrendResponse{
		Labels:         out.Labels,
		Registrations:  out.Registrations,
		CompletedTests: out.CompletedTests,
	}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func toWindow(w systempb.Window) Window {
	switch w {
	case systempb.Window_WINDOW_7D:
		return Window7D
	case systempb.Window_WINDOW_30D:
		return Window30D
	default:
		return WindowDay
	}
}

func toGranularity(g systempb.Granularity) Granularity {
	switch g {
	case systempb.Granularity_GRANULARITY_MONTH:
		return GranularityMonth
	case systempb.Granularity_GRANULARITY_YEAR:
		return GranularityYear
	default:
		return GranularityWeek
	}
}

func toMetricCard(d *MetricCard) *systempb.MetricCard {
	return &systempb.MetricCard{
		Key:      toMetricKey(d.Key),
		Value:    d.Value,
		SubKind:  toSubMetricKind(d.SubKind),
		SubValue: utils.Wrap(d.SubValue, utils.DoubleW),
		SubNote:  utils.StringWEmptyNil(d.SubNote),
		Drill:    toCardDrill(d.Drill),
	}
}

func toFunnel(d *Funnel) *systempb.Funnel {
	if d == nil {
		return nil
	}
	return &systempb.Funnel{
		EnteredTests: d.EnteredTests,
		Stages: utils.SliceMap(d.Stages, func(s *FunnelStage) *systempb.FunnelStage {
			return &systempb.FunnelStage{
				Key:            toFunnelStageKey(s.Key),
				Count:          s.Count,
				StepConversion: utils.Wrap(s.StepConversion, utils.DoubleW),
				OfFirst:        s.OfFirst,
			}
		}),
	}
}

func toViral(d *Viral) *systempb.Viral {
	if d == nil {
		return nil
	}
	return &systempb.Viral{
		NewReferrals:   d.NewReferrals,
		Inviters:       d.Inviters,
		PerInviter:     d.PerInviter,
		ConversionRate: d.ConversionRate,
		ActiveUsers:    d.ActiveUsers,
		KFactor:        d.KFactor,
		Levels: utils.SliceMap(d.Levels, func(l *FissionLevel) *systempb.FissionLevel {
			return &systempb.FissionLevel{Level: l.Level, Count: l.Count, Share: l.Share}
		}),
	}
}

func toInviter(d *Inviter) *systempb.Inviter {
	return &systempb.Inviter{
		UserId:         utils.StringWEmptyNil(d.UserID),
		Name:           d.Name,
		Invites:        d.Invites,
		CompletionRate: d.CompletionRate,
	}
}

func toQualityRow(d *QualityRow) *systempb.QualityRow {
	return &systempb.QualityRow{
		Key:     toQualityKey(d.Key),
		Invited: d.Invited,
		Organic: d.Organic,
		DeltaPp: d.DeltaPP,
	}
}

func toActivity(d *Activity) *systempb.Activity {
	return &systempb.Activity{
		OccurredAt: utils.ToTimestamp(&d.OccurredAt),
		Type:       d.Type,
		Content:    d.Content,
	}
}

func toMetricKey(key string) systempb.MetricKey {
	switch key {
	case MetricTest:
		return systempb.MetricKey_METRIC_KEY_TEST
	case MetricActive:
		return systempb.MetricKey_METRIC_KEY_ACTIVE
	case MetricPay:
		return systempb.MetricKey_METRIC_KEY_PAY
	default:
		return systempb.MetricKey_METRIC_KEY_REGISTER
	}
}

func toSubMetricKind(kind string) systempb.SubMetricKind {
	switch kind {
	case SubAvgPerUser:
		return systempb.SubMetricKind_SUB_METRIC_KIND_AVG_PER_USER
	case SubPeak:
		return systempb.SubMetricKind_SUB_METRIC_KIND_PEAK
	case SubManual:
		return systempb.SubMetricKind_SUB_METRIC_KIND_MANUAL
	default:
		return systempb.SubMetricKind_SUB_METRIC_KIND_DELTA_PREV
	}
}

func toCardDrill(drill string) systempb.CardDrill {
	switch drill {
	case DrillTestRecord:
		return systempb.CardDrill_CARD_DRILL_TEST_RECORD
	case DrillOrder:
		return systempb.CardDrill_CARD_DRILL_ORDER
	default:
		return systempb.CardDrill_CARD_DRILL_USER_LIST
	}
}

func toFunnelStageKey(key string) systempb.FunnelStageKey {
	switch key {
	case EventReportView:
		return systempb.FunnelStageKey_FUNNEL_STAGE_KEY_REPORT_VIEW
	case EventReportCenterView:
		return systempb.FunnelStageKey_FUNNEL_STAGE_KEY_REPORT_CENTER_VIEW
	case EventShare:
		return systempb.FunnelStageKey_FUNNEL_STAGE_KEY_SHARE
	case EventMemberActivate:
		return systempb.FunnelStageKey_FUNNEL_STAGE_KEY_MEMBER_ACTIVATE
	default:
		return systempb.FunnelStageKey_FUNNEL_STAGE_KEY_COMPLETE_TEST
	}
}

func toQualityKey(key string) systempb.QualityKey {
	switch key {
	case QualityPayRate:
		return systempb.QualityKey_QUALITY_KEY_PAY_RATE
	case QualityRetention7:
		return systempb.QualityKey_QUALITY_KEY_RETENTION_7D
	default:
		return systempb.QualityKey_QUALITY_KEY_TEST_RATE
	}
}
