package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ChunithmBinding struct {
	ent.Schema
}

func (ChunithmBinding) Fields() []ent.Field {
	return []ent.Field{
		field.Int("haruki_user_id").Comment("Reference to users table"),
		field.String("server").MaxLen(10),
		field.String("aime_id").MaxLen(50),
	}
}

func (ChunithmBinding) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("haruki_user_id", "server").Unique(),
	}
}

func (ChunithmBinding) Edges() []ent.Edge {
	return nil
}
