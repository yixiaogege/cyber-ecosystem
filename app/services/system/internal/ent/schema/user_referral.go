package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"cyber-ecosystem/shared-go/orm/ent/mixins"
)

// Append-only invite edges: one row per successful invitation. The referrer
// chain (level 1 = direct invite, level N = invited by a level N-1 invitee) is
// derived by walking referrer_id, so rows are never rewritten.
type UserReferral struct {
	ent.Schema
}

func (UserReferral) Fields() []ent.Field {
	return []ent.Field{
		field.String("referrer_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("邀请人；与 system.User 的逻辑引用，不做外键"),
		field.String("referrer_name").Default("").MaxRuneLen(64).Immutable().
			Comment("邀请时刻的展示名快照；活体身份在 system.User，快照只为让本表独立可渲染"),
		field.String("invitee_id").NotEmpty().MaxRuneLen(20).Immutable().
			Comment("被邀人；一个用户至多被邀一次"),
		field.String("tenant_id").Default("").MaxRuneLen(20).Immutable().
			Comment("所属租户"),
	}
}

func (UserReferral) Edges() []ent.Edge {
	return nil
}

func (UserReferral) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixins.IDStringMixin{},
		mixins.CreatedUpdatedMixin{},
	}
}

// created_at 的索引由 CreatedUpdatedMixin 提供（窗口筛选按 created_at 走它）。
func (UserReferral) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("invitee_id").Unique(),
		index.Fields("referrer_id"),
	}
}

func (UserReferral) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		entsql.Annotation{Table: "user_referral"},
	}
}
