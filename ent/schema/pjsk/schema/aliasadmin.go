package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type AliasAdmin struct {
	ent.Schema
}

func (AliasAdmin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("haruki_user_id").Unique().Comment("Reference to users table"),
		field.String("name").MaxLen(100),
	}
}

func (AliasAdmin) Edges() []ent.Edge {
	return nil
}
