package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type User struct {
	ent.Schema
}

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Unique().
			Comment("User ID, 6-digit random number"),
		field.String("platform").
			MaxLen(20).
			Comment("Platform name"),
		field.String("user_id").
			MaxLen(50).
			Comment("User ID on the platform"),
		field.Bool("ban_state").
			Default(false).
			Comment("Whether user is banned"),
		field.String("ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for ban"),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("platform", "user_id").Unique(),
	}
}

func (User) Edges() []ent.Edge {
	return nil
}
