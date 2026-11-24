package dashboard

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

// fixedNow anchors every window/trend derivation: 2026-11-18 is a Wednesday,
// so the ISO week containing it starts Monday 2026-11-16.
var fixedNow = time.Date(2026, 11, 18, 15, 4, 5, 0, time.Local)

type fakeDashboardRP struct {
	DashboardRP
	counts    map[string]int64
	rollups   []*DailyRollup
	referrals *ReferralStats
	levels    []*LevelCount
}

func (f *fakeDashboardRP) EventCounts(context.Context, time.Time, time.Time) (map[string]int64, error) {
	return f.counts, nil
}

func (f *fakeDashboardRP) DistinctUsers(context.Context, []string, time.Time, time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeDashboardRP) DailyActivePeak(context.Context, time.Time, time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeDashboardRP) ManualGrantCount(context.Context, time.Time, time.Time) (int64, error) {
	return 0, nil
}

func (f *fakeDashboardRP) ReferralStats(context.Context, time.Time, time.Time) (*ReferralStats, error) {
	return f.referrals, nil
}

func (f *fakeDashboardRP) ReferralLevels(context.Context, time.Time, time.Time) ([]*LevelCount, error) {
	return f.levels, nil
}

func (f *fakeDashboardRP) TopInviters(context.Context, time.Time, time.Time, int) ([]*Inviter, error) {
	return nil, nil
}

func (f *fakeDashboardRP) QualityCounts(context.Context, time.Time, time.Time) (*QualityCounts, error) {
	return nil, nil
}

func (f *fakeDashboardRP) RecentActivity(context.Context, time.Time, time.Time, int) ([]*Activity, error) {
	return nil, nil
}

func (f *fakeDashboardRP) DailyRollups(context.Context, time.Time, time.Time) ([]*DailyRollup, error) {
	return f.rollups, nil
}

func newDashboardUC(rp DashboardRP) *DashboardUC {
	uc := NewDashboardUC(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil, rp,
	)
	uc.clock = func() time.Time { return fixedNow }
	return uc
}

// window is a half-open [from, to): the to side is exclusive (tomorrow's
// midnight), so consecutive windows tile without overlap or gap.
func TestWindowIsHalfOpenPerDayCount(t *testing.T) {
	for _, tc := range []struct {
		w    Window
		span time.Duration
	}{
		{WindowDay, 24 * time.Hour},
		{Window7D, 7 * 24 * time.Hour},
		{Window30D, 30 * 24 * time.Hour},
	} {
		from, to := window(fixedNow, tc.w)
		if d := to.Sub(from); d != tc.span {
			t.Errorf("window(%v): span = %v, want %v", tc.w, d, tc.span)
		}
		if !from.Equal(startOfDay(fixedNow).AddDate(0, 0, -(tc.w.Days() - 1))) {
			t.Errorf("window(%v): from = %v, want today-(N-1) midnight", tc.w, from)
		}
	}
}

// startOfWeek must agree with postgres date_trunc('week'): Monday midnight.
func TestStartOfWeekIsMondayMidnight(t *testing.T) {
	got := startOfWeek(fixedNow)
	want := time.Date(2026, 11, 16, 0, 0, 0, 0, time.Local)
	if !got.Equal(want) {
		t.Fatalf("startOfWeek(wed) = %v, want %v", got, want)
	}
	if got.Weekday() != time.Monday {
		t.Fatalf("startOfWeek = %s, want Monday", got.Weekday())
	}
}

// Trend spans follow granularity only: 12 weeks / 12 months / 3 years,
// regardless of the current window.
func TestTrendRangeFollowsGranularityOnly(t *testing.T) {
	for _, tc := range []struct {
		g      Granularity
		points int
		first  time.Time
	}{
		{GranularityWeek, 12, time.Date(2026, 8, 31, 0, 0, 0, 0, time.Local)},
		{GranularityMonth, 12, time.Date(2025, 12, 1, 0, 0, 0, 0, time.Local)},
		{GranularityYear, 3, time.Date(2024, 1, 1, 0, 0, 0, 0, time.Local)},
	} {
		from, _, points := trendRange(fixedNow, tc.g)
		if points != tc.points {
			t.Errorf("trendRange(%v): points = %d, want %d", tc.g, points, tc.points)
		}
		if !from.Equal(tc.first) {
			t.Errorf("trendRange(%v): from = %v, want %v", tc.g, from, tc.first)
		}
	}
}

// The funnel is ordered and shares the event vocabulary; the first stage
// carries no step conversion (nothing precedes it).
func TestBuildFunnelOrdersAndFirstStageHasNoStepConversion(t *testing.T) {
	counts := map[string]int64{
		EventCompleteTest: 100, EventReportView: 50, EventReportCenterView: 25,
		EventShare: 10, EventMemberActivate: 5,
	}
	f := buildFunnel(counts)
	for i, s := range f.Stages {
		if s.Key != funnelStages[i] {
			t.Fatalf("stage %d key = %s, want %s", i, s.Key, funnelStages[i])
		}
		if i == 0 && s.StepConversion != nil {
			t.Errorf("first stage step conversion = %v, want nil", *s.StepConversion)
		}
	}
	if got := f.Stages[1].StepConversion; got == nil || *got != 0.5 {
		t.Errorf("stage 1 step conversion = %v, want 0.5", got)
	}
	if f.Stages[1].OfFirst != 50 {
		t.Errorf("stage 1 of-first = %v, want 50", f.Stages[1].OfFirst)
	}
}

// Null-vs-zero contract: an absent base renders nothing (nil), a real zero
// base renders 0 — the two must never collapse into one value.
func TestRatiosDistinguishAbsentFromZero(t *testing.T) {
	if got := deltaPercent(5, 0); got != nil {
		t.Errorf("deltaPercent(5, 0) = %v, want nil (no baseline renders no sub-label)", got)
	}
	if got := ratio(0, 0); got != nil {
		t.Errorf("ratio(0, 0) = %v, want nil", got)
	}
	if got := ratio(0, 10); got == nil || *got != 0 {
		t.Errorf("ratio(0, 10) = %v, want 0 (real zero, not absent)", got)
	}
	if got := percent(0, 0); got != 0 {
		t.Errorf("percent(0, 0) = %v, want 0", got)
	}
}

// The pay card folds pay_success and unlock_success into one metric.
func TestBuildCardsFoldsPayEvents(t *testing.T) {
	counts := map[string]int64{
		EventRegister: 10, EventCompleteTest: 8, EventPaySuccess: 3, EventUnlockSuccess: 2,
	}
	cards := buildCards(counts, map[string]int64{}, 6, 4, 2, 1, Window7D)
	byKey := map[string]*MetricCard{}
	for _, c := range cards {
		byKey[c.Key] = c
	}
	if byKey[MetricPay].Value != 5 {
		t.Errorf("pay card = %d, want 3+2", byKey[MetricPay].Value)
	}
	if byKey[MetricRegister].SubValue != nil {
		t.Errorf("register sub = %v, want nil (empty previous window)", *byKey[MetricRegister].SubValue)
	}
	if byKey[MetricActive].SubNote != "7" {
		t.Errorf("active note = %q, want 7", byKey[MetricActive].SubNote)
	}
}

// Trend aggregates rollups into granularity-aligned buckets and zero-fills
// the empty ones.
func TestTrendBucketsAndZeroFill(t *testing.T) {
	rp := &fakeDashboardRP{
		rollups: []*DailyRollup{
			{Date: time.Date(2026, 11, 16, 0, 0, 0, 0, time.Local), NewRegistrations: 4, CompletedTests: 2},
		},
	}
	got, err := newDashboardUC(rp).Trend(context.Background(), GranularityWeek)
	if err != nil {
		t.Fatalf("trend: %v", err)
	}
	if len(got.Labels) != 12 || len(got.Registrations) != 12 {
		t.Fatalf("trend points = %d, want 12", len(got.Labels))
	}
	// 2026-11-16 (Mon) opens the last bucket; earlier buckets zero-fill.
	if got.Registrations[11] != 4 || got.CompletedTests[11] != 2 {
		t.Errorf("last bucket = (%d, %d), want (4, 2)", got.Registrations[11], got.CompletedTests[11])
	}
	for i, v := range got.Registrations[:11] {
		if v != 0 {
			t.Errorf("bucket %d = %d, want 0 (zero-filled)", i, v)
		}
	}
}

// Viral caps the level pyramid at maxFissionLevel with "4级+" semantics and
// reports shares over the referral total.
func TestViralCapsFissionLevels(t *testing.T) {
	rp := &fakeDashboardRP{
		referrals: &ReferralStats{NewReferrals: 100, Inviters: 10, Converted: 30},
		levels: []*LevelCount{
			{Level: 1, Count: 60}, {Level: 2, Count: 30},
			{Level: 3, Count: 8}, {Level: 4, Count: 2}, {Level: 9, Count: 50},
		},
	}
	v, err := newDashboardUC(rp).viral(context.Background(), fixedNow, fixedNow.Add(24*time.Hour), 200)
	if err != nil {
		t.Fatalf("viral: %v", err)
	}
	if len(v.Levels) != maxFissionLevel {
		t.Fatalf("levels = %d, want cap %d", len(v.Levels), maxFissionLevel)
	}
	if v.Levels[0].Share != 40 { // 60 / (60+30+8+2) — level 9 rows are not counted
		t.Errorf("level1 share = %v, want 40", v.Levels[0].Share)
	}
	if v.KFactor != 0.5 { // 100 referrals / 200 active
		t.Errorf("k-factor = %v, want 0.5", v.KFactor)
	}
}
