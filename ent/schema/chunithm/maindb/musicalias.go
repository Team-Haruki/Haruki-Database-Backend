package maindb

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
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

func (ChunithmMusicAlias) Edges() []ent.Edge {
	return nil
}
