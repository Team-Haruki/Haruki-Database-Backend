package bot

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
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
