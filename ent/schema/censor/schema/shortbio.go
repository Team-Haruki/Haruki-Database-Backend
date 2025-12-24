package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type ShortBio struct {
	ent.Schema
}

func (ShortBio) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "short_bio"},
	}
}

func (ShortBio) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable(),
		field.String("user_id").
			MaxLen(30).
			Optional().
			Nillable(),
		field.String("content").
			MaxLen(60).
			Optional().
			Nillable(),
		field.String("im_user_id").
			MaxLen(30).
			Optional().
			Nillable(),
		field.String("result").
			MaxLen(10).
			Optional().
			Nillable(),
	}
}
