package maindb

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ChunithmDefaultServer struct {
	ent.Schema
}

func (ChunithmDefaultServer) Fields() []ent.Field {
	return []ent.Field{
		field.String("im_id").MaxLen(30),
		field.String("platform").MaxLen(20),
		field.String("server").MaxLen(10),
	}
}

func (ChunithmDefaultServer) Edges() []ent.Edge {
	return nil
}
