package pjsk

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
		field.String("im_id").MaxLen(30),
		field.String("platform").MaxLen(20),
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
		index.Fields("im_id", "platform", "server").Unique(),
	}
}
