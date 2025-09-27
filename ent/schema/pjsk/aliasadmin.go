package pjsk

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type AliasAdmin struct {
	ent.Schema
}

func (AliasAdmin) Fields() []ent.Field {
	return []ent.Field{
		field.String("platform").MaxLen(20),
		field.String("im_id").MaxLen(100),
		field.String("name").MaxLen(100),
	}
}

func (AliasAdmin) Edges() []ent.Edge {
	return nil
}
