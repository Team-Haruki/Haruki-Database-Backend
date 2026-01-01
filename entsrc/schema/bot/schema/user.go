package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "user"},
	}
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("owner_user_id").
			Comment("Owner user ID"),
		field.Int("bot_id").
			Unique().
			Comment("Bot ID, primary key"),
		field.String("credential").
			MaxLen(512).
			Optional().
			Comment("Bot credential"),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_user_id").Unique(),
	}
}
