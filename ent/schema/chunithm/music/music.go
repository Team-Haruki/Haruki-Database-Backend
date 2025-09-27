package music

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ChunithmMusic struct {
	ent.Schema
}

func (ChunithmMusic) Fields() []ent.Field {
	return []ent.Field{
		field.Int("music_id"),
		field.String("title").MaxLen(255),
		field.String("artist").MaxLen(255),
		field.String("category").MaxLen(50).Optional().Nillable(),
		field.String("version").MaxLen(10).Optional().Nillable(),
		field.Time("release_date").Optional().Nillable(),
		field.Int("is_deleted").Default(0),
		field.String("deleted_version").MaxLen(10).Optional().Nillable(),
	}
}

func (ChunithmMusic) Edges() []ent.Edge {
	return nil
}
