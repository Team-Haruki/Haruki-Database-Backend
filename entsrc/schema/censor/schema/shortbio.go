package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		field.Int("haruki_user_id").
			Optional().
			Nillable(),
		field.String("result").
			MaxLen(10).
			Optional().
			Nillable(),
	}
}

func (ShortBio) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("haruki_user_id"),
	}
}
