package pjsk

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type UserPreference struct {
	ent.Schema
}

func (UserPreference) Fields() []ent.Field {
	return []ent.Field{
		field.String("im_id").MaxLen(30),
		field.String("platform").MaxLen(20),
		field.String("option").MaxLen(50),
		field.String("value").MaxLen(50),
	}
}

func (UserPreference) Edges() []ent.Edge {
	return nil
}
