package pjsk

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type GroupAlias struct {
	ent.Schema
}

func (GroupAlias) Fields() []ent.Field {
	return []ent.Field{
		field.String("platform").MaxLen(20),
		field.String("group_id").MaxLen(50),
		field.String("alias_type").MaxLen(20),
		field.Int("alias_type_id"),
		field.String("alias").MaxLen(100),
	}
}

func (GroupAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "group_id", "alias_type", "alias_type_id", "alias").Unique(),
	}
}

func (GroupAlias) Edges() []ent.Edge {
	return nil
}
