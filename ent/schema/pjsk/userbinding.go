package pjsk

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type UserBinding struct {
	ent.Schema
}

func (UserBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id"),
		field.String("platform").MaxLen(20),
		field.String("im_id").MaxLen(30),
		field.String("user_id").MaxLen(30),
		field.String("server").MaxLen(2),
		field.Bool("visible").Default(true),
	}
}

func (UserBinding) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("default_refs", UserDefaultBinding.Type),
	}
}

func (UserBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "im_id", "server", "user_id").Unique(),
	}
}
