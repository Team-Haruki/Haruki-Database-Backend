package music

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ChunithmChartData struct {
	ent.Schema
}

func (ChunithmChartData) Fields() []ent.Field {
	return []ent.Field{
		field.Int("music_id"),
		field.Int("difficulty"),
		field.String("creator").MaxLen(50).Optional().Nillable(),
		field.Float("bpm").Optional().Nillable(),
		field.Int("tap_count").Optional().Nillable(),
		field.Int("hold_count").Optional().Nillable(),
		field.Int("slide_count").Optional().Nillable(),
		field.Int("air_count").Optional().Nillable(),
		field.Int("flick_count").Optional().Nillable(),
		field.Int("total_count").Optional().Nillable(),
	}
}

func (ChunithmChartData) Edges() []ent.Edge {
	return nil
}
