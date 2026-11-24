package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// Daily rollup behind the trend chart: one row per tenant per day, upserted by
// the (tenant_id, date) unique index. Weekly / monthly / yearly views are
// aggregations over this table rather than re-scans of event_log.
type PlatformAnalytics struct {
	ent.Schema
}

func (PlatformAnalytics) Fields() []ent.Field {
	return []ent.Field{
		field.Time("date").Immutable().
			Comment("统计日；日切后去时刻，写入前统一到租户时区的零点"),
		field.Int("new_registrations").Default(0).
			Comment("当日新增注册数"),
		field.Int("completed_tests").Default(0).
			Comment("当日完成测评数"),
		field.String("tenant_id").Default("").MaxRuneLen(20).Immutable().
			Comment("所属租户"),
	}
}

func (PlatformAnalytics) Edges() []ent.Edge {
	return nil
}

func (PlatformAnalytics) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (PlatformAnalytics) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "date").Unique(),
	}
}

func (PlatformAnalytics) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "platform_analytics"},
	}
}
