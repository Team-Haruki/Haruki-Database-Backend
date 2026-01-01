package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ChunithmChartData struct {
	ent.Schema
}

func (ChunithmChartData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("music_id"),
		field.Int("difficulty"),
		field.String("creator").MaxLen(100).Optional().Nillable(),
		field.Float("bpm").Optional().Nillable(),
		field.Int("tap_count").Optional().Nillable(),
		field.Int("hold_count").Optional().Nillable(),
		field.Int("slide_count").Optional().Nillable(),
		field.Int("air_count").Optional().Nillable(),
		field.Int("flick_count").Optional().Nillable(),
		field.Int("total_count").Optional().Nillable(),
	}
}

func (ChunithmChartData) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("music_id", "difficulty").Unique(),
	}
}

func (ChunithmChartData) Edges() []ent.Edge {
	return nil
}

func (ChunithmChartData) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "chart_data"},
	}
}
