package music

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type ChunithmMusicDifficulty struct {
	ent.Schema
}

func (ChunithmMusicDifficulty) Fields() []ent.Field {
	return []ent.Field{
		field.Int("music_id"),
		field.String("version").MaxLen(10),
		field.Float("diff0_const").Optional().Nillable(),
		field.Float("diff1_const").Optional().Nillable(),
		field.Float("diff2_const").Optional().Nillable(),
		field.Float("diff3_const").Optional().Nillable(),
		field.Float("diff4_const").Optional().Nillable(),
	}
}

func (ChunithmMusicDifficulty) Edges() []ent.Edge {
	return nil
}
