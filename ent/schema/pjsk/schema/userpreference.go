package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type UserPreference struct {
	ent.Schema
}

func (UserPreference) Fields() []ent.Field {
	return []ent.Field{
		field.Int("haruki_user_id").Comment("Reference to users table"),
		field.String("option").MaxLen(50),
		field.String("value").MaxLen(50),
	}
}

func (UserPreference) Edges() []ent.Edge {
	return nil
}
