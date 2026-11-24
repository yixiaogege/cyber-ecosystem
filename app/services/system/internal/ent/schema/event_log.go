package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// Append-only C-end event stream: inserts only, never updated, never deleted.
// tenant is a plain column rather than the subject-derived TenantMixin because
// anonymous events (user_id NULL, before registration) carry no subject.
type EventLog struct {
	ent.Schema
}

func (EventLog) Fields() []ent.Field {
	return []ent.Field{
		field.String("event_name").NotEmpty().MaxRuneLen(64).Immutable().
			Comment("事件词表名；聚合口径见 module/dashboard 的常量"),
		field.String("user_id").Optional().Nillable().MaxRuneLen(20).Immutable().
			Comment("登录用户；注册前的匿名行为为 NULL"),
		field.String("anon_id").Optional().Nillable().MaxRuneLen(64).Immutable().
			Comment("匿名设备标识；匿名行归因为用户由上游负责"),
		field.JSON("props", map[string]any{}).Optional().
			Comment("事件附加属性；聚合只按 event_name 计数，不依赖其结构"),
		field.String("channel").Optional().Nillable().MaxRuneLen(20).Immutable().
			Comment("来源渠道；未采集为 NULL"),
		field.Time("occurred_at").Immutable().
			Comment("事件发生时刻（业务时钟）"),
		field.Time("received_at").Immutable().
			Comment("入库时刻（采集时钟）；与 occurred_at 的差值即上报延迟"),
		field.String("tenant_id").Default("").MaxRuneLen(20).Immutable().
			Comment("所属租户；匿名事件为空"),
	}
}

func (EventLog) Edges() []ent.Edge {
	return nil
}

func (EventLog) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

func (EventLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("event_name", "occurred_at"),
		index.Fields("user_id", "occurred_at"),
	}
}

func (EventLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "event_log"},
	}
}
