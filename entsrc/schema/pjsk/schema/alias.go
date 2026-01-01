package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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

func (Alias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("alias_type", "alias_type_id", "alias").Unique(),
	}
}

func (Alias) Edges() []ent.Edge {
	return nil
}
