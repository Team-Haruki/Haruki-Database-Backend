package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UserDefaultBinding struct {
	ent.Schema
}

func (UserDefaultBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.Int("haruki_user_id").Comment("Reference to users table"),
		field.String("server").MaxLen(7),
		field.Int("binding_id"),
	}
}

func (UserDefaultBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("binding", UserBinding.Type).
			Ref("default_refs").
			Field("binding_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (UserDefaultBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("haruki_user_id", "server").Unique(),
	}
}
