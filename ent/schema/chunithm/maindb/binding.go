package maindb

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ChunithmBinding struct {
	ent.Schema
}

func (ChunithmBinding) Fields() []ent.Field {
	return []ent.Field{
		field.String("im_id").MaxLen(30),
		field.String("platform").MaxLen(20),
		field.String("server").MaxLen(10),
		field.String("aime_id").MaxLen(50),
	}
}

func (ChunithmBinding) Edges() []ent.Edge {
	return nil
}
