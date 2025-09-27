package pjsk

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type RejectedAlias struct {
	ent.Schema
}

func (RejectedAlias) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("alias_type").MaxLen(20),
		field.Int("alias_type_id"),
		field.String("alias").MaxLen(100),
		field.String("reviewed_by").MaxLen(100),
		field.String("reason").MaxLen(255),
		field.Time("reviewed_at"),
	}
}

func (RejectedAlias) Edges() []ent.Edge {
	return nil
}
