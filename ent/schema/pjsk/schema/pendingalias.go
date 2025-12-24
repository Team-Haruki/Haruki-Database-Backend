package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type PendingAlias struct {
	ent.Schema
}

func (PendingAlias) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("alias_type").MaxLen(20),
		field.Int("alias_type_id"),
		field.String("alias").MaxLen(100),
		field.String("submitted_by").MaxLen(100),
		field.Time("submitted_at"),
	}
}

func (PendingAlias) Edges() []ent.Edge {
	return nil
}
