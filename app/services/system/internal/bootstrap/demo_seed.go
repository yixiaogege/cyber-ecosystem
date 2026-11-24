package bootstrap

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"slices"
	"time"

	"github.com/rs/xid"

	"cyber-ecosystem/shared-go/kratos/security"

	"cyber-ecosystem/app/services/system/internal/conf"
	"cyber-ecosystem/app/services/system/internal/ent"
	"cyber-ecosystem/app/services/system/internal/platform"
)

const (
	// 演示数据：只在 seed.demo = true（裸机 dev）时执行，容器部署默认关闭。
	// 上游业务域（测评 / 支付 / 用户画像）本轮不存在，因此这里按同一套日序列造出
	// 自洽的 event_log / user_referral / platform_analytics，让运营总览可用。
	// 关键约束：指标卡取自 event_log，趋势取自 platform_analytics，两者由同一条
	// 日序列生成，所以「新增注册」卡与趋势近 30 天合计必然相等。
	demoSeed      = 20261007 // 固定随机源，同一份数据每次启动都一样
	rollupDays    = 1200     // platform_analytics 覆盖 3 年 + 余量（趋势年视图取 3 年）
	eventDays     = 30       // event_log 只写最长窗口（30 天）内
	baseRegsToday = 60.0     // 今日新增注册基数
	dailyDecay    = 1.001    // 每天回退的衰减因子 → 3 年前约为今天的 1/3
	testRatio     = 0.82     // 完成测评 / 新增注册
	payRatio      = 0.10     // 付费 / 完成测评
	bulkSize      = 300
	maxLevel      = 4   // 「4 级+」：与 module/dashboard 的上溯上限一致
	organicSkip   = 0.5 // 自然流量被抽中后丢弃的概率 → 被邀人群的分母质量更高
)

var funnelRates = []float64{
	// 漏斗相邻阶段比率（完成测评 → 打开能力报告 → 查看报告中心 → 发起分享 → 注册会员/解锁）。
	0.764, 0.500, 0.702, 0.636,
}

var inviterNames = []string{
	// 邀请人昵称池：referrer_name 是邀请时刻的展示名快照。
	"龘好久", "选手乙", "王小明", "陆明明", "赵一鸣", "阿强", "萌新丁", "路人丙",
	"林小北", "陈默", "周野", "苏念", "何一鸣", "郑好", "黄小满", "夏至",
}

var eventTexts = map[string]string{
	// 最近动态的事件文案：event_log.props.text 是日志原文，前端只翻译类型标签。
	"member_activate": "开通会员 31 天",
	"pay_success":     "解锁 认知报告（单次）",
	"unlock_success":  "解锁 能力报告",
	"report_view":     "完成 认知报告 8 维",
	"share":           "发起分享 · 邀请好友",
	"complete_test":   "完成测评 · 认知能力",
}

type DemoSeed struct {
	log  *slog.Logger
	plat *platform.Platform
	cfg  *conf.Seed
}

func NewDemoSeed(cfg *conf.Seed, p *platform.Platform, logger *slog.Logger) (*DemoSeed, error) {
	return &DemoSeed{log: logger, plat: p, cfg: cfg}, nil
}

// Run -----------------------------------------------------------------------------------------------------------------

func (s *DemoSeed) Run(ctx context.Context) error {
	if s.cfg == nil || !s.cfg.GetDemo() {
		return nil
	}
	client := s.plat.GetClient(ctx)
	existing, err := client.EventLog.Query().Count(ctx)
	if err != nil {
		return s.plat.HandleEntError(fmt.Errorf("seed: count event_log: %w", err))
	}
	if existing > 0 {
		s.log.Info("seed: demo data already present; skipping", "event_log", existing)
		return nil
	}

	tenant := security.DefaultTenant
	today := startOfDay(time.Now())
	rng := rand.New(rand.NewSource(demoSeed))
	days := buildDays(rng, today)
	if err := s.seedRollups(ctx, client, tenant, days); err != nil {
		return err
	}
	roster := buildRoster(days)
	edges := buildReferrals(rng, days, roster)
	if err := s.seedReferrals(ctx, client, tenant, edges); err != nil {
		return err
	}
	invited := invitedSet(edges)
	if err := s.seedEvents(ctx, client, tenant, days, roster, invited, rng); err != nil {
		return err
	}
	s.log.Info("seed: demo data written",
		"rollup_days", len(days), "event_days", eventDays, "users", len(roster), "referrals", len(edges))
	return nil
}

// Private -------------------------------------------------------------------------------------------------------------

type demoDay struct {
	// 一条日序列：指标卡与趋势共用同一份来源，因此两边必然一致。
	dayAgo int
	date   time.Time
	regs   int64
	tests  int64
}

func buildDays(rng *rand.Rand, today time.Time) []*demoDay {
	out := make([]*demoDay, 0, rollupDays)
	for i := rollupDays - 1; i >= 0; i-- {
		regs := int64(math.Round(baseRegsToday / math.Pow(dailyDecay, float64(i))))
		tests := int64(math.Round(float64(regs) * testRatio * (0.94 + 0.12*rng.Float64())))
		out = append(out, &demoDay{dayAgo: i, date: today.AddDate(0, 0, -i), regs: regs, tests: tests})
	}
	return out
}

func (s *DemoSeed) seedRollups(ctx context.Context, client *ent.Client, tenant string, days []*demoDay) error {
	creates := make([]*ent.PlatformAnalyticsCreate, 0, len(days))
	for _, d := range days {
		creates = append(creates, client.PlatformAnalytics.Create().
			SetTenantID(tenant).
			SetDate(d.date).
			SetNewRegistrations(int(d.regs)).
			SetCompletedTests(int(d.tests)))
	}
	for chunk := range slices.Chunk(creates, bulkSize) {
		if err := client.PlatformAnalytics.CreateBulk(chunk...).Exec(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: insert platform_analytics: %w", err))
		}
	}
	return nil
}

type demoUser struct {
	// 窗口内的一个人：id 取自 xid（与 ent 主键同构），dayAgo 为注册日。
	// 演示数据不复用 system.User：本种子要能独立成形（超管账户由 IAM seed 负责），
	// user_id 因此只是逻辑引用，不建外键。
	id     string
	dayAgo int
	name   string // 仅在本人成为邀请人时赋值（referrer_name 快照）
}

func buildRoster(days []*demoDay) []*demoUser {
	out := make([]*demoUser, 0, 2048)
	for _, d := range days {
		if d.dayAgo >= eventDays {
			continue
		}
		for i := int64(0); i < d.regs; i++ {
			out = append(out, &demoUser{id: xid.New().String(), dayAgo: d.dayAgo})
		}
	}
	return out
}

type demoEdge struct {
	// 一条邀请关系；level 由 referrer 的层级推出（1 级 = 邀请人本身无人邀请）。
	referrer *demoUser
	invitee  *demoUser
	at       time.Time
}

var levelWeights = []float64{
	// 目标裂变层级分布（1 级 65% / 2 级 24% / 3 级 9% / 4 级+ 2%），
	// 对齐 manage 演示值的金字塔形状：层级由 referrer 自身的层级决定，因此聚合时
	// 用 referrer_id 链上溯算出的深度与这里赋值的一致。
	0.65, 0.24, 0.09, 0.02,
}

func buildReferrals(rng *rand.Rand, days []*demoDay, roster []*demoUser) []*demoEdge {
	// 按天造边（dayAgo 从大到小）：被邀人是当天注册队列里的一员
	// （邀请 → 当天注册），referrer 只从「注册更早且层级正好低一级」的人里挑，
	// 于是层级天然形成金字塔，且不会出现环。
	byDay := make(map[int][]*demoUser, eventDays)
	for _, u := range roster {
		byDay[u.dayAgo] = append(byDay[u.dayAgo], u)
	}
	// byLevel 只装「已注册且早于当天」的人；level 0 = 未被任何人邀请（自然流量）。
	byLevel := make(map[int32][]*demoUser, maxLevel)
	level := make(map[string]int32, len(roster))
	picked := make(map[string]bool, 512)
	out := make([]*demoEdge, 0, 256)

	for _, d := range days {
		if d.dayAgo >= eventDays {
			continue
		}
		invitees := byDay[d.dayAgo]
		edges := int(math.Round(6 + 6*float64(eventDays-d.dayAgo)/float64(eventDays)))
		if edges > len(invitees) {
			edges = len(invitees)
		}
		for i := 0; i < edges; i++ {
			target := rollLevel(rng)
			referrer := pickReferrer(rng, byLevel[target-1], picked)
			if referrer == nil {
				target, referrer = 1, pickReferrer(rng, byLevel[0], picked)
			}
			if referrer == nil {
				break
			}
			if referrer.name == "" {
				referrer.name = inviterNames[int(rng.Intn(len(inviterNames)))]
			}
			picked[referrer.id] = true
			level[invitees[i].id] = target
			out = append(out, &demoEdge{referrer: referrer, invitee: invitees[i], at: timeAt(d.dayAgo, rng)})
		}
		// 当天注册的人到下一轮才可能成为邀请人。
		for _, u := range invitees {
			byLevel[level[u.id]] = append(byLevel[level[u.id]], u)
		}
	}
	return out
}

func rollLevel(rng *rand.Rand) int32 {
	r := rng.Float64()
	var acc float64
	for i, w := range levelWeights {
		acc += w
		if r < acc {
			return int32(i + 1)
		}
	}
	return int32(len(levelWeights))
}

func pickReferrer(rng *rand.Rand, pool []*demoUser, picked map[string]bool) *demoUser {
	// 用平方偏置抽取：少数人贡献多数邀请，于是「去重邀请人数」远小于
	// 「邀请关系数」（真实裂变的长尾），同时 80% 的邀请落在还没邀请过人的候选上。
	if len(pool) == 0 {
		return nil
	}
	if rng.Float64() < 0.8 {
		fresh := make([]*demoUser, 0, len(pool))
		for _, u := range pool {
			if !picked[u.id] {
				fresh = append(fresh, u)
			}
		}
		if len(fresh) > 0 {
			pool = fresh
		}
	}
	return pool[int(math.Pow(rng.Float64(), 2)*float64(len(pool)))]
}

func invitedSet(edges []*demoEdge) map[string]bool {
	out := make(map[string]bool, len(edges))
	for _, e := range edges {
		out[e.invitee.id] = true
	}
	return out
}

func (s *DemoSeed) seedReferrals(ctx context.Context, client *ent.Client, tenant string, edges []*demoEdge) error {
	creates := make([]*ent.UserReferralCreate, 0, len(edges))
	for _, e := range edges {
		creates = append(creates, client.UserReferral.Create().
			SetTenantID(tenant).
			SetReferrerID(e.referrer.id).
			SetReferrerName(e.referrer.name).
			SetInviteeID(e.invitee.id).
			SetCreatedAt(e.at))
	}
	for chunk := range slices.Chunk(creates, bulkSize) {
		if err := client.UserReferral.CreateBulk(chunk...).Exec(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: insert user_referral: %w", err))
		}
	}
	return nil
}

func (s *DemoSeed) seedEvents(ctx context.Context, client *ent.Client, tenant string, days []*demoDay, roster []*demoUser, invited map[string]bool, rng *rand.Rand) error {
	byDay := make(map[int][]*demoUser, eventDays)
	for _, u := range roster {
		byDay[u.dayAgo] = append(byDay[u.dayAgo], u)
	}
	pool := make([]*demoUser, 0, len(roster))
	batch := make([]*ent.EventLogCreate, 0, 4096)

	for _, d := range days {
		if d.dayAgo >= eventDays {
			continue
		}
		// 注册 + 注册当天的登录
		for _, u := range byDay[d.dayAgo] {
			batch = append(batch,
				eventCreate(client, tenant, u.id, "register", timeAt(d.dayAgo, rng), ""),
				eventCreate(client, tenant, u.id, "login", timeAt(d.dayAgo, rng), ""),
			)
		}
		pool = append(pool, byDay[d.dayAgo]...)

		// 活跃：从已注册人群里抽样，被邀用户回访概率更高（质量差值由此产生）
		batch = append(batch, sampleLogins(client, tenant, pool, invited, d.dayAgo, rng)...)

		// 完成测评：按日序列的完成数分配给已有用户，被邀用户更容易被抽中
		batch = append(batch, sampleEvents(client, tenant, pool, invited, d.dayAgo, d.tests, "complete_test", rng)...)

		// 漏斗后续阶段按固定比率衰减
		prev := d.tests
		for _, stage := range []struct {
			name string
			rate float64
		}{
			{"report_view", funnelRates[0]},
			{"report_center_view", funnelRates[1]},
			{"share", funnelRates[2]},
			{"member_activate", funnelRates[3]},
		} {
			n := int64(math.Round(float64(prev) * stage.rate))
			batch = append(batch, sampleEvents(client, tenant, pool, invited, d.dayAgo, n, stage.name, rng)...)
			prev = n
		}
		// 进入测评（漏斗口径注的分母）
		batch = append(batch, sampleEvents(client, tenant, pool, invited, d.dayAgo, int64(math.Round(float64(d.tests)/0.76)), "enter_test", rng)...)
		// 付费 / 解锁：末阶段的一部分
		batch = append(batch, sampleEvents(client, tenant, pool, invited, d.dayAgo, int64(math.Round(float64(d.tests)*payRatio)), "pay_success", rng)...)
	}
	// 一条手动开通：指标卡「新增付费·解锁」的副标签据此计数
	if last := byDay[0]; len(last) > 0 {
		batch = append(batch, eventCreate(client, tenant, last[0].id, "pay_success", timeAt(0, rng), "").
			SetProps(map[string]any{"manual": true}))
	}
	return s.flush(ctx, client, batch)
}

func (s *DemoSeed) flush(ctx context.Context, client *ent.Client, batch []*ent.EventLogCreate) error {
	for chunk := range slices.Chunk(batch, bulkSize) {
		if err := client.EventLog.CreateBulk(chunk...).Exec(ctx); err != nil {
			return s.plat.HandleEntError(fmt.Errorf("seed: insert event_log: %w", err))
		}
	}
	return nil
}

func sampleLogins(client *ent.Client, tenant string, pool []*demoUser, invited map[string]bool, dayAgo int, rng *rand.Rand) []*ent.EventLogCreate {
	// 抽样当天的活跃用户，当天每人只产生一条登录事件。自然流量被抽中后
	// 有 organicSkip 的概率被丢弃，被邀用户因此获得更高的回访率（质量部件的留存差值）。
	target := int(math.Round(180 / math.Pow(dailyDecay, float64(dayAgo))))
	seen := make(map[string]bool, target)
	out := make([]*ent.EventLogCreate, 0, target)
	for tries := 0; len(out) < target && tries < target*20; tries++ {
		u := pool[rng.Intn(len(pool))]
		if seen[u.id] || u.dayAgo == dayAgo {
			continue // 注册当天的登录已在上面写过
		}
		if !invited[u.id] && rng.Float64() < organicSkip {
			continue
		}
		seen[u.id] = true
		out = append(out, eventCreate(client, tenant, u.id, "login", timeAt(dayAgo, rng), ""))
	}
	return out
}

func sampleEvents(client *ent.Client, tenant string, pool []*demoUser, invited map[string]bool, dayAgo int, n int64, eventName string, rng *rand.Rand) []*ent.EventLogCreate {
	// 抽样 n 条同名事件；同样按 organicSkip 偏向被邀用户。
	out := make([]*ent.EventLogCreate, 0, n)
	for i := int64(0); i < n; i++ {
		u := pool[rng.Intn(len(pool))]
		if !invited[u.id] && rng.Float64() < organicSkip {
			u = pool[rng.Intn(len(pool))]
		}
		out = append(out, eventCreate(client, tenant, u.id, eventName, timeAt(dayAgo, rng), eventTexts[eventName]))
	}
	return out
}

func eventCreate(client *ent.Client, tenant, userID, eventName string, at time.Time, text string) *ent.EventLogCreate {
	ec := client.EventLog.Create().
		SetTenantID(tenant).
		SetEventName(eventName).
		SetUserID(userID).
		SetOccurredAt(at).
		SetReceivedAt(at)
	if text != "" {
		ec.SetProps(map[string]any{"text": text})
	}
	return ec
}

func timeAt(dayAgo int, rng *rand.Rand) time.Time {
	// 把 dayAgo 换算成当天的随机时刻；「今天」只取到当前小时，避免写未来时间。
	now := time.Now()
	day := startOfDay(now).AddDate(0, 0, -dayAgo)
	span := 24
	if dayAgo == 0 {
		span = now.Hour() + 1
	}
	return day.Add(time.Duration(rng.Intn(span*60)) * time.Minute)
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
