package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ChunithmDefaultServer struct {
	ent.Schema
}

func (ChunithmDefaultServer) Fields() []ent.Field {
	return []ent.Field{
		field.Int("haruki_user_id").Unique().Comment("Reference to users table"),
		field.String("server").MaxLen(10),
	}
}

func (ChunithmDefaultServer) Edges() []ent.Edge {
	return nil
}
