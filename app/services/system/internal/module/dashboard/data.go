package dashboard

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	entsql "entgo.io/ent/dialect/sql"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/system/internal/platform"
	"cyber-ecosystem/app/services/system/internal/shared"
)

// 聚合口径全部落在这一层。这些查询是 GROUP BY / COUNT(DISTINCT) / FILTER (WHERE) /
// 递归 CTE，ent 的查询构造器表达不了（层级上溯的递归 CTE 尤其），所以统一走
// sql/execquery 的裸 SQL：只绑定占位符，不使用字符串拼接的用户输入。
// 三张表都用普通 tenant_id 列（没有 subject 派生的租户拦截器），因此每个查询都把
// 租户显式作为 $1 传入。
const (
	qEventCounts = `
SELECT event_name, COUNT(*) AS count
FROM event_log
WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at < $3
GROUP BY event_name`

	qDistinctUsers = `
SELECT COUNT(DISTINCT user_id) AS count
FROM event_log
WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at < $3
  AND user_id IS NOT NULL
  AND event_name IN (%s)`

	qDailyActivePeak = `
SELECT COALESCE(MAX(day_count), 0) AS peak
FROM (
	SELECT COUNT(DISTINCT user_id) AS day_count
	FROM event_log
	WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at < $3
	  AND user_id IS NOT NULL AND event_name = $4
	GROUP BY date_trunc('day', occurred_at)
) AS per_day`

	qManualGrants = `
SELECT COUNT(*) AS count
FROM event_log
WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at < $3
  AND event_name IN (%s)
  AND props->>'manual' = 'true'`

	qReferralStats = `
SELECT COUNT(*) AS new_referrals,
       COUNT(DISTINCT r.referrer_id) AS inviters,
       COUNT(*) FILTER (WHERE EXISTS (
         SELECT 1 FROM event_log e
         WHERE e.tenant_id = r.tenant_id AND e.user_id = r.invitee_id
           AND e.event_name = $4
       )) AS converted
FROM user_referral r
WHERE r.tenant_id = $1 AND r.created_at >= $2 AND r.created_at < $3`

	// 层级由 referrer_id 链上溯派生：1 级 = 邀请人本身不是任何人的被邀人。
	// 递归项不按窗口过滤（跨窗的邀请关系仍要参与上溯），窗口只筛最终结果。
	// depth < 3 让最深生成到 4，即「4 级+」。
	qReferralLevels = `
WITH RECURSIVE chain(user_id, depth) AS (
  SELECT r.invitee_id, 1
  FROM user_referral r
  WHERE r.tenant_id = $1
    AND NOT EXISTS (
      SELECT 1 FROM user_referral p
      WHERE p.tenant_id = $1 AND p.invitee_id = r.referrer_id
    )
  UNION ALL
  SELECT r.invitee_id, chain.depth + 1
  FROM user_referral r
  JOIN chain ON chain.user_id = r.referrer_id
  WHERE r.tenant_id = $1 AND chain.depth < 3
)
SELECT LEAST(chain.depth, 4) AS lvl, COUNT(*) AS count
FROM chain
WHERE chain.user_id IN (
  SELECT invitee_id FROM user_referral
  WHERE tenant_id = $1 AND created_at >= $2 AND created_at < $3
)
GROUP BY lvl
ORDER BY lvl`

	qTopInviters = `
SELECT r.referrer_id AS referrer_id,
       MAX(r.referrer_name) AS name,
       COUNT(*) AS invites,
       COUNT(*) FILTER (WHERE EXISTS (
         SELECT 1 FROM event_log e
         WHERE e.tenant_id = r.tenant_id AND e.user_id = r.invitee_id
           AND e.event_name = $5
       )) AS converted
FROM user_referral r
WHERE r.tenant_id = $1 AND r.created_at >= $2 AND r.created_at < $3
GROUP BY r.referrer_id
ORDER BY invites DESC, r.referrer_id
LIMIT $4`

	// 同期质量：被邀 = 窗口内被邀且本期注册的人；自然 = 窗口内注册但不在被邀名单里的人。
	// 三项比率的分子都是「该人是否发生过某事件」，7 日留存以注册时刻 +7 天为界。
	qQualityCounts = `
WITH reg AS (
  SELECT user_id, MIN(occurred_at) AS first_seen
  FROM event_log
  WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at < $3
    AND user_id IS NOT NULL AND event_name = $4
  GROUP BY user_id
),
invited AS (
  SELECT invitee_id AS user_id
  FROM user_referral
  WHERE tenant_id = $1 AND created_at >= $2 AND created_at < $3
),
pop AS (
  SELECT reg.user_id AS user_id, reg.first_seen AS first_seen,
         (invited.user_id IS NOT NULL) AS is_invited
  FROM reg
  LEFT JOIN invited ON invited.user_id = reg.user_id
)
SELECT pop.is_invited AS is_invited,
       COUNT(*) AS total,
       COUNT(*) FILTER (WHERE EXISTS (
         SELECT 1 FROM event_log e
         WHERE e.tenant_id = $1 AND e.user_id = pop.user_id AND e.event_name = $5
       )) AS tested,
       COUNT(*) FILTER (WHERE EXISTS (
         SELECT 1 FROM event_log e
         WHERE e.tenant_id = $1 AND e.user_id = pop.user_id AND e.event_name IN ($6, $7)
       )) AS paid,
       COUNT(*) FILTER (WHERE EXISTS (
         SELECT 1 FROM event_log e
         WHERE e.tenant_id = $1 AND e.user_id = pop.user_id AND e.event_name = $8
           AND e.occurred_at >= pop.first_seen + INTERVAL '7 days'
       )) AS retained7
FROM pop
GROUP BY pop.is_invited`

	qRecentActivity = `
SELECT occurred_at, event_name, COALESCE(props->>'text', '') AS content
FROM event_log
WHERE tenant_id = $1 AND occurred_at >= $2 AND occurred_at < $3
ORDER BY occurred_at DESC, id DESC
LIMIT $4`

	qDailyRollups = `
SELECT pa.date AS date, pa.new_registrations AS new_registrations, pa.completed_tests AS completed_tests
FROM platform_analytics pa
WHERE pa.tenant_id = $1 AND pa.date >= $2 AND pa.date < $3
ORDER BY pa.date`
)

// Repo ----------------------------------------------------------------------------------------------------------------

type dashboardRP struct {
	shared.RP
}

func NewDashboardRP(logger *slog.Logger, p *platform.Platform) DashboardRP {
	return &dashboardRP{RP: shared.NewRP(logger.With("module", "module/dashboard_rp"), p)}
}

// Method --------------------------------------------------------------------------------------------------------------

func (rp *dashboardRP) EventCounts(ctx context.Context, from, to time.Time) (map[string]int64, error) {
	var rows []struct {
		EventName string `json:"event_name"`
		Count     int64  `json:"count"`
	}
	if err := rp.scan(ctx, &rows, qEventCounts, tenant(ctx), from, to); err != nil {
		return nil, err
	}
	out := make(map[string]int64, len(rows))
	for _, r := range rows {
		out[r.EventName] = r.Count
	}
	return out, nil
}

func (rp *dashboardRP) DistinctUsers(ctx context.Context, eventNames []string, from, to time.Time) (int64, error) {
	args := append([]any{tenant(ctx), from, to}, stringsToAny(eventNames)...)
	var rows []struct {
		Count int64 `json:"count"`
	}
	if err := rp.scan(ctx, &rows, fmt.Sprintf(qDistinctUsers, placeholders(4, len(eventNames))), args...); err != nil {
		return 0, err
	}
	return firstCount(rows), nil
}

func (rp *dashboardRP) DailyActivePeak(ctx context.Context, from, to time.Time) (int64, error) {
	var rows []struct {
		Peak int64 `json:"peak"`
	}
	if err := rp.scan(ctx, &rows, qDailyActivePeak, tenant(ctx), from, to, EventLogin); err != nil {
		return 0, err
	}
	return firstPeak(rows), nil
}

func (rp *dashboardRP) ManualGrantCount(ctx context.Context, from, to time.Time) (int64, error) {
	names := []string{EventPaySuccess, EventUnlockSuccess}
	args := append([]any{tenant(ctx), from, to}, stringsToAny(names)...)
	var rows []struct {
		Count int64 `json:"count"`
	}
	if err := rp.scan(ctx, &rows, fmt.Sprintf(qManualGrants, placeholders(4, len(names))), args...); err != nil {
		return 0, err
	}
	return firstCount(rows), nil
}

func (rp *dashboardRP) ReferralStats(ctx context.Context, from, to time.Time) (*ReferralStats, error) {
	var rows []struct {
		NewReferrals int64 `json:"new_referrals"`
		Inviters     int64 `json:"inviters"`
		Converted    int64 `json:"converted"`
	}
	if err := rp.scan(ctx, &rows, qReferralStats, tenant(ctx), from, to, EventCompleteTest); err != nil {
		return nil, err
	}
	out := &ReferralStats{}
	if len(rows) > 0 {
		out.NewReferrals, out.Inviters, out.Converted = rows[0].NewReferrals, rows[0].Inviters, rows[0].Converted
	}
	return out, nil
}

func (rp *dashboardRP) ReferralLevels(ctx context.Context, from, to time.Time) ([]*LevelCount, error) {
	var rows []struct {
		Level int32 `json:"lvl"`
		Count int64 `json:"count"`
	}
	if err := rp.scan(ctx, &rows, qReferralLevels, tenant(ctx), from, to); err != nil {
		return nil, err
	}
	out := make([]*LevelCount, 0, len(rows))
	for _, r := range rows {
		out = append(out, &LevelCount{Level: r.Level, Count: r.Count})
	}
	return out, nil
}

func (rp *dashboardRP) TopInviters(ctx context.Context, from, to time.Time, limit int) ([]*Inviter, error) {
	var rows []struct {
		ReferrerID string `json:"referrer_id"`
		Name       string `json:"name"`
		Invites    int64  `json:"invites"`
		Converted  int64  `json:"converted"`
	}
	if err := rp.scan(ctx, &rows, qTopInviters, tenant(ctx), from, to, limit, EventCompleteTest); err != nil {
		return nil, err
	}
	out := make([]*Inviter, 0, len(rows))
	for _, r := range rows {
		out = append(out, &Inviter{
			UserID:         r.ReferrerID,
			Name:           r.Name,
			Invites:        r.Invites,
			CompletionRate: percent(r.Converted, r.Invites),
		})
	}
	return out, nil
}

func (rp *dashboardRP) QualityCounts(ctx context.Context, from, to time.Time) (*QualityCounts, error) {
	var rows []struct {
		IsInvited bool  `json:"is_invited"`
		Total     int64 `json:"total"`
		Tested    int64 `json:"tested"`
		Paid      int64 `json:"paid"`
		Retained7 int64 `json:"retained7"`
	}
	args := []any{tenant(ctx), from, to, EventRegister, EventCompleteTest, EventPaySuccess, EventUnlockSuccess, EventLogin}
	if err := rp.scan(ctx, &rows, qQualityCounts, args...); err != nil {
		return nil, err
	}
	out := &QualityCounts{}
	for _, r := range rows {
		g := &GroupQuality{Total: r.Total, Tested: r.Tested, Paid: r.Paid, Retained7: r.Retained7}
		if r.IsInvited {
			out.Invited = g
		} else {
			out.Organic = g
		}
	}
	return out, nil
}

func (rp *dashboardRP) RecentActivity(ctx context.Context, from, to time.Time, limit int) ([]*Activity, error) {
	var rows []struct {
		OccurredAt time.Time `json:"occurred_at"`
		EventName  string    `json:"event_name"`
		Content    string    `json:"content"`
	}
	if err := rp.scan(ctx, &rows, qRecentActivity, tenant(ctx), from, to, limit); err != nil {
		return nil, err
	}
	out := make([]*Activity, 0, len(rows))
	for _, r := range rows {
		out = append(out, &Activity{OccurredAt: r.OccurredAt, Type: r.EventName, Content: r.Content})
	}
	return out, nil
}

func (rp *dashboardRP) DailyRollups(ctx context.Context, from, to time.Time) ([]*DailyRollup, error) {
	var rows []struct {
		Date             time.Time `json:"date"`
		NewRegistrations int64     `json:"new_registrations"`
		CompletedTests   int64     `json:"completed_tests"`
	}
	if err := rp.scan(ctx, &rows, qDailyRollups, tenant(ctx), from, to); err != nil {
		return nil, err
	}
	out := make([]*DailyRollup, 0, len(rows))
	for _, r := range rows {
		out = append(out, &DailyRollup{Date: r.Date, NewRegistrations: r.NewRegistrations, CompletedTests: r.CompletedTests})
	}
	return out, nil
}

// Private -------------------------------------------------------------------------------------------------------------

func (rp *dashboardRP) scan(ctx context.Context, dest any, query string, args ...any) error {
	rows, err := rp.Platform.GetClient(ctx).QueryContext(ctx, query, args...)
	if err != nil {
		return rp.Platform.HandleEntError(err)
	}
	defer func() { _ = rows.Close() }()
	if err := entsql.ScanSlice(rows, dest); err != nil {
		return rp.Platform.HandleEntError(err)
	}
	return nil
}

func tenant(ctx context.Context) string {
	return security.TenantFromCtx(ctx)
}

// placeholders 生成从 start 开始的 n 个占位符，如 placeholders(3, 2) → "$3, $4"。
// 只插入占位符本身，取值仍然全部绑定。
func placeholders(start, n int) string {
	out := make([]string, 0, n)
	for i := 0; i < n; i++ {
		out = append(out, fmt.Sprintf("$%d", start+i))
	}
	return strings.Join(out, ", ")
}

func stringsToAny(in []string) []any {
	out := make([]any, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}
	return out
}

func firstCount(rows []struct {
	Count int64 `json:"count"`
}) int64 {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].Count
}

func firstPeak(rows []struct {
	Peak int64 `json:"peak"`
}) int64 {
	if len(rows) == 0 {
		return 0
	}
	return rows[0].Peak
}
