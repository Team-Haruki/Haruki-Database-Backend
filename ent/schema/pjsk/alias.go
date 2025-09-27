package pjsk

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Alias struct {
	ent.Schema
}

func (Alias) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id"),
		field.String("alias_type").MaxLen(20),
		field.Int("alias_type_id"),
		field.String("alias").MaxLen(100),
	}
}

func (Alias) Edges() []ent.Edge {
	return nil
}
