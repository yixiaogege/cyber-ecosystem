package dashboard

import (
	"context"
	"log/slog"
	"time"

	"cyber-ecosystem/app/services/system/internal/shared"
)

// 事件词表：聚合口径的唯一来源。C 端上报的 event_name 必须落在这里，否则不计入任何部件。
const (
	EventRegister         = "register"
	EventLogin            = "login"
	EventEnterTest        = "enter_test"
	EventCompleteTest     = "complete_test"
	EventReportView       = "report_view"
	EventReportCenterView = "report_center_view"
	EventShare            = "share"
	EventPaySuccess       = "pay_success"
	EventUnlockSuccess    = "unlock_success"
	EventMemberActivate   = "member_activate"
)

// 指标卡 / 漏斗 / 质量 的键：前端据此选 i18n 模板，服务端只给键与数值。
const (
	MetricRegister = "register"
	MetricTest     = "test"
	MetricActive   = "active"
	MetricPay      = "pay"

	SubDeltaPrev  = "delta_prev"  // 与上一等长窗相比的百分比增减
	SubAvgPerUser = "avg_per_user" // 人均次数
	SubPeak       = "peak"        // 窗口内单日最高（Note = 窗口天数）
	SubManual     = "manual"      // 其中手动开通数

	DrillUserList   = "user_list"
	DrillTestRecord = "test_record"
	DrillOrder      = "order"

	QualityTestRate   = "test_rate"
	QualityPayRate    = "pay_rate"
	QualityRetention7 = "retention_7d"
)

// funnelStages 是 C 端行为漏斗的有序阶段，键与事件名同源。
var funnelStages = []string{
	EventCompleteTest,
	EventReportView,
	EventReportCenterView,
	EventShare,
	EventMemberActivate,
}

// qualityKeys 与 qualityRates 的取值顺序一一对应。
var qualityKeys = []string{QualityTestRate, QualityPayRate, QualityRetention7}

const (
	topInviterLimit = 5
	activityLimit   = 8
	maxFissionLevel = 4 // 「4 级+」：链上溯到此为止

	// 同期质量窗：截止到 7 天前，好让队列的 7 日留存可观测。
	qualityWindowDays = 90
	qualityLagDays    = 7
)

// Window 统计时间窗，驱动指标卡、漏斗、裂变与最近动态。
type Window int

const (
	WindowDay Window = iota + 1
	Window7D
	Window30D
)

func (w Window) Days() int {
	switch w {
	case Window7D:
		return 7
	case Window30D:
		return 30
	default:
		return 1
	}
}

// Granularity 趋势聚合粒度。
type Granularity int

const (
	GranularityWeek Granularity = iota + 1
	GranularityMonth
	GranularityYear
)

// DO ------------------------------------------------------------------------------------------------------------------

type MetricCard struct {
	Key      string
	Value    int64
	SubKind  string
	SubValue *float64
	SubNote  string
	Drill    string
}

type FunnelStage struct {
	Key            string
	Count          int64
	StepConversion *float64 // 首阶段为空
	OfFirst        float64
}

type Funnel struct {
	EnteredTests int64
	Stages       []*FunnelStage
}

type FissionLevel struct {
	Level int32
	Count int64
	Share float64
}

type Viral struct {
	NewReferrals   int64
	Inviters       int64
	PerInviter     float64
	ConversionRate float64
	ActiveUsers    int64
	KFactor        float64
	Levels         []*FissionLevel
}

type Inviter struct {
	UserID         string
	Name           string
	Invites        int64
	CompletionRate float64
}

type QualityRow struct {
	Key     string
	Invited float64
	Organic float64
	DeltaPP float64
}

type Activity struct {
	OccurredAt time.Time
	Type       string
	Content    string
}

type Overview struct {
	Cards       []*MetricCard
	Funnel      *Funnel
	Viral       *Viral
	TopInviters []*Inviter
	Quality     []*QualityRow
	Recent      []*Activity
}

type Trend struct {
	Labels         []string
	Registrations  []int64
	CompletedTests []int64
}

// DailyRollup 是 platform_analytics 的一行日汇总。
type DailyRollup struct {
	Date             time.Time
	NewRegistrations int64
	CompletedTests   int64
}

// ReferralStats 是窗口内的邀请关系总量：新增关系数、去重邀请人数、被邀人完成过测评数。
type ReferralStats struct {
	NewReferrals int64
	Inviters     int64
	Converted    int64
}

// LevelCount 是某一裂变层级的被邀人数。
type LevelCount struct {
	Level int32
	Count int64
}

// GroupQuality 是一个人群（被邀 / 自然）在窗口内的规模与三项质量计数。
type GroupQuality struct {
	Total     int64
	Tested    int64
	Paid      int64
	Retained7 int64
}

// QualityCounts 按「是否被邀」分组，[0] = 被邀，[1] = 自然。
type QualityCounts struct {
	Invited *GroupQuality
	Organic *GroupQuality
}

// Port ----------------------------------------------------------------------------------------------------------------

type DashboardRP interface {
	EventCounts(ctx context.Context, from, to time.Time) (map[string]int64, error)
	DistinctUsers(ctx context.Context, eventNames []string, from, to time.Time) (int64, error)
	DailyActivePeak(ctx context.Context, from, to time.Time) (int64, error)
	ManualGrantCount(ctx context.Context, from, to time.Time) (int64, error)
	ReferralStats(ctx context.Context, from, to time.Time) (*ReferralStats, error)
	ReferralLevels(ctx context.Context, from, to time.Time) ([]*LevelCount, error)
	TopInviters(ctx context.Context, from, to time.Time, limit int) ([]*Inviter, error)
	QualityCounts(ctx context.Context, from, to time.Time) (*QualityCounts, error)
	RecentActivity(ctx context.Context, from, to time.Time, limit int) ([]*Activity, error)
	DailyRollups(ctx context.Context, from, to time.Time) ([]*DailyRollup, error)
}

// UC ------------------------------------------------------------------------------------------------------------------

type DashboardUC struct {
	shared.UC
	dashboardRP DashboardRP
}

func NewDashboardUC(logger *slog.Logger, tm shared.Transaction, dashboardRP DashboardRP) *DashboardUC {
	return &DashboardUC{
		UC:          shared.NewUC(logger.With("module", "module/dashboard"), tm),
		dashboardRP: dashboardRP,
	}
}

// Method --------------------------------------------------------------------------------------------------------------

func (uc *DashboardUC) Overview(ctx context.Context, w Window) (*Overview, error) {
	from, to := window(w)
	prevFrom := from.AddDate(0, 0, -w.Days())

	counts, err := uc.dashboardRP.EventCounts(ctx, from, to)
	if err != nil {
		return nil, err
	}
	prevCounts, err := uc.dashboardRP.EventCounts(ctx, prevFrom, from)
	if err != nil {
		return nil, err
	}
	active, err := uc.dashboardRP.DistinctUsers(ctx, []string{EventLogin}, from, to)
	if err != nil {
		return nil, err
	}
	testers, err := uc.dashboardRP.DistinctUsers(ctx, []string{EventCompleteTest}, from, to)
	if err != nil {
		return nil, err
	}
	peak, err := uc.dashboardRP.DailyActivePeak(ctx, from, to)
	if err != nil {
		return nil, err
	}
	manual, err := uc.dashboardRP.ManualGrantCount(ctx, from, to)
	if err != nil {
		return nil, err
	}
	cards := buildCards(counts, prevCounts, active, testers, peak, manual, w)

	funnel := buildFunnel(counts)

	viral, err := uc.viral(ctx, from, to, active)
	if err != nil {
		return nil, err
	}
	inviters, err := uc.dashboardRP.TopInviters(ctx, from, to, topInviterLimit)
	if err != nil {
		return nil, err
	}
	quality, err := uc.quality(ctx)
	if err != nil {
		return nil, err
	}
	recent, err := uc.dashboardRP.RecentActivity(ctx, from, to, activityLimit)
	if err != nil {
		return nil, err
	}
	return &Overview{
		Cards:       cards,
		Funnel:      funnel,
		Viral:       viral,
		TopInviters: inviters,
		Quality:     quality,
		Recent:      recent,
	}, nil
}

// Trend 只受粒度影响，与时间窗无关：周 = 近 12 周，月 = 近 12 个月，年 = 近 3 年。
func (uc *DashboardUC) Trend(ctx context.Context, g Granularity) (*Trend, error) {
	from, to, points := trendRange(time.Now(), g)
	rows, err := uc.dashboardRP.DailyRollups(ctx, from, to)
	if err != nil {
		return nil, err
	}
	reg := make([]int64, points)
	tests := make([]int64, points)
	labels := make([]string, points)
	index := make(map[time.Time]int, points)
	for i, b := range trendBuckets(from, g, points) {
		index[b] = i
		labels[i] = bucketLabel(b, g)
	}
	for _, r := range rows {
		i, ok := index[bucketOf(r.Date, g)]
		if !ok {
			continue
		}
		reg[i] += r.NewRegistrations
		tests[i] += r.CompletedTests
	}
	return &Trend{Labels: labels, Registrations: reg, CompletedTests: tests}, nil
}

// Private -------------------------------------------------------------------------------------------------------------

// window 把时间窗换算成 [from, to) 的绝对区间，日切按本机时区。
func window(w Window) (from, to time.Time) {
	today := startOfDay(time.Now())
	return today.AddDate(0, 0, -(w.Days() - 1)), today.AddDate(0, 0, 1)
}

func buildCards(counts, prevCounts map[string]int64, active, testers, peak, manual int64, w Window) []*MetricCard {
	register := counts[EventRegister]
	test := counts[EventCompleteTest]
	pay := counts[EventPaySuccess] + counts[EventUnlockSuccess]

	return []*MetricCard{
		{Key: MetricRegister, Value: register, SubKind: SubDeltaPrev, SubValue: deltaPercent(register, prevCounts[EventRegister]), Drill: DrillUserList},
		{Key: MetricTest, Value: test, SubKind: SubAvgPerUser, SubValue: ratio(test, testers), Drill: DrillTestRecord},
		{Key: MetricActive, Value: active, SubKind: SubPeak, SubValue: float64Ptr(float64(peak)), SubNote: dayCount(w), Drill: DrillUserList},
		{Key: MetricPay, Value: pay, SubKind: SubManual, SubValue: float64Ptr(float64(manual)), Drill: DrillOrder},
	}
}

func buildFunnel(counts map[string]int64) *Funnel {
	first := counts[funnelStages[0]]
	stages := make([]*FunnelStage, 0, len(funnelStages))
	for i, key := range funnelStages {
		count := counts[key]
		stage := &FunnelStage{Key: key, Count: count, OfFirst: percent(count, first)}
		if i > 0 {
			stage.StepConversion = ratio(count, counts[funnelStages[i-1]])
		}
		stages = append(stages, stage)
	}
	return &Funnel{EnteredTests: counts[EventEnterTest], Stages: stages}
}

func (uc *DashboardUC) viral(ctx context.Context, from, to time.Time, active int64) (*Viral, error) {
	stats, err := uc.dashboardRP.ReferralStats(ctx, from, to)
	if err != nil {
		return nil, err
	}
	levels, err := uc.dashboardRP.ReferralLevels(ctx, from, to)
	if err != nil {
		return nil, err
	}
	byLevel := make(map[int32]int64, len(levels))
	var total int64
	for _, l := range levels {
		byLevel[l.Level] = l.Count
		total += l.Count
	}
	out := make([]*FissionLevel, 0, maxFissionLevel)
	for level := int32(1); level <= maxFissionLevel; level++ {
		count := byLevel[level]
		out = append(out, &FissionLevel{Level: level, Count: count, Share: percent(count, total)})
	}
	return &Viral{
		NewReferrals:   stats.NewReferrals,
		Inviters:       stats.Inviters,
		PerInviter:     ratioFloat(stats.NewReferrals, stats.Inviters),
		ConversionRate: percent(stats.Converted, stats.NewReferrals),
		ActiveUsers:    active,
		KFactor:        ratioFloat(stats.NewReferrals, active),
		Levels:         out,
	}, nil
}

// quality 用固定同期窗而非请求时间窗：manage 的「被邀 vs 自然」是同期质量快照，
// 且 7 日留存要求队列注册满 7 天，所以窗口取 [今天-97 天, 今天-7 天)。
func (uc *DashboardUC) quality(ctx context.Context) ([]*QualityRow, error) {
	to := startOfDay(time.Now()).AddDate(0, 0, -qualityLagDays)
	from := to.AddDate(0, 0, -qualityWindowDays)
	counts, err := uc.dashboardRP.QualityCounts(ctx, from, to)
	if err != nil {
		return nil, err
	}
	invited := groupRates(counts.Invited)
	organic := groupRates(counts.Organic)
	rows := make([]*QualityRow, 0, len(qualityKeys))
	for i, key := range qualityKeys {
		rows = append(rows, &QualityRow{
			Key:     key,
			Invited: invited[i],
			Organic: organic[i],
			DeltaPP: invited[i] - organic[i],
		})
	}
	return rows, nil
}

// groupRates 返回三项比率（完成测评率 / 付费率 / 7 日留存），顺序与 qualityKeys 一致。
func groupRates(g *GroupQuality) [3]float64 {
	if g == nil || g.Total == 0 {
		return [3]float64{}
	}
	total := float64(g.Total)
	return [3]float64{
		float64(g.Tested) / total * 100,
		float64(g.Paid) / total * 100,
		float64(g.Retained7) / total * 100,
	}
}

func trendRange(now time.Time, g Granularity) (from, to time.Time, points int) {
	to = startOfDay(now).AddDate(0, 0, 1)
	switch g {
	case GranularityMonth:
		first := startOfMonth(now).AddDate(0, -11, 0)
		return first, to, 12
	case GranularityYear:
		first := time.Date(now.Year()-2, time.January, 1, 0, 0, 0, 0, now.Location())
		return first, to, 3
	default:
		first := startOfWeek(now).AddDate(0, 0, -11*7)
		return first, to, 12
	}
}

// trendBuckets 生成与粒度对齐的桶起点序列；无数据的桶由调用方补零。
func trendBuckets(from time.Time, g Granularity, points int) []time.Time {
	out := make([]time.Time, 0, points)
	for i := 0; i < points; i++ {
		switch g {
		case GranularityMonth:
			out = append(out, from.AddDate(0, i, 0))
		case GranularityYear:
			out = append(out, from.AddDate(i, 0, 0))
		default:
			out = append(out, from.AddDate(0, 0, i*7))
		}
	}
	return out
}

func bucketOf(t time.Time, g Granularity) time.Time {
	switch g {
	case GranularityMonth:
		return startOfMonth(t)
	case GranularityYear:
		return time.Date(t.Year(), time.January, 1, 0, 0, 0, 0, t.Location())
	default:
		return startOfWeek(t)
	}
}

func bucketLabel(t time.Time, g Granularity) string {
	switch g {
	case GranularityMonth:
		return t.Format("2006-01")
	case GranularityYear:
		return t.Format("2006")
	default:
		return t.Format("01-02")
	}
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

func startOfMonth(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
}

// startOfWeek 取周一零点：与 postgres date_trunc('week', ...) 的 ISO 周一致。
func startOfWeek(t time.Time) time.Time {
	day := startOfDay(t)
	offset := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -offset)
}

func dayCount(w Window) string {
	switch w {
	case Window7D:
		return "7"
	case Window30D:
		return "30"
	default:
		return "1"
	}
}

// deltaPercent 是与上一等长窗相比的百分比增减；无基期（上一窗为 0）时返回空，前端不渲染副标签。
func deltaPercent(cur, prev int64) *float64 {
	if prev == 0 {
		return nil
	}
	return float64Ptr(float64(cur-prev) / float64(prev) * 100)
}

// percent 以 first 为分母的占比；分母为 0 时返回 0。
func percent(v, first int64) float64 {
	if first == 0 {
		return 0
	}
	return float64(v) / float64(first) * 100
}

// ratio 比值；分母为 0 时返回空，与「无数据」和「比值 0」区分开。
func ratio(v, base int64) *float64 {
	if base == 0 {
		return nil
	}
	return float64Ptr(float64(v) / float64(base))
}

func ratioFloat(v, base int64) float64 {
	if base == 0 {
		return 0
	}
	return float64(v) / float64(base)
}

func float64Ptr(v float64) *float64 {
	return &v
}
