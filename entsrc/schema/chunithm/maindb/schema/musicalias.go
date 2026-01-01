package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type ChunithmMusicAlias struct {
	ent.Schema
}

func (ChunithmMusicAlias) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").Unique().Immutable(),
		field.Int("music_id"),
		field.String("alias").MaxLen(100),
	}
}

func (ChunithmMusicAlias) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("music_id", "alias").Unique(),
	}
}

func (ChunithmMusicAlias) Edges() []ent.Edge {
	return nil
}
