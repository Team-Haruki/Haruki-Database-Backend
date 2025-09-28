package censor

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

type Result struct {
	ent.Schema
}

func (Result) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "censor_result"},
	}
}

func (Result) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Immutable(),
		field.String("name").
			MaxLen(300).
			NotEmpty().
			Comment("Name to be censored"),
		field.Int("result").
			Optional().
			Nillable().
			Comment("Censor result (int code)"),
		field.Time("time").
			SchemaType(map[string]string{
				"mysql": "timestamp",
			}).
			Optional().
			Nillable().
			Comment("Censor timestamp"),
	}
}
