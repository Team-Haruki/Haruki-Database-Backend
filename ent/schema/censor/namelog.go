package censor

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type NameLog struct {
	ent.Schema
}

func (NameLog) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "name_log"},
	}
}

func (NameLog) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable(),
		field.String("user_id").
			MaxLen(30).
			Optional().
			Nillable(),
		field.String("name").
			MaxLen(300).
			Optional().
			Nillable(),
		field.String("im_user_id").
			MaxLen(30).
			Optional().
			Nillable(),
		field.Time("time").
			SchemaType(map[string]string{
				"mysql": "timestamp",
			}).
			Optional().
			Nillable(),
		field.String("result").
			MaxLen(10).
			Optional().
			Nillable(),
	}
}
