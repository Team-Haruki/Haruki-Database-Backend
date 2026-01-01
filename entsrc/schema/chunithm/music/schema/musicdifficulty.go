package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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

func (ChunithmMusicDifficulty) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("music_id", "version").Unique(),
	}
}

func (ChunithmMusicDifficulty) Edges() []ent.Edge {
	return nil
}

func (ChunithmMusicDifficulty) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "music_difficulties"},
	}
}
