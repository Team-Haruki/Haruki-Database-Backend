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
		field.Bool("pjsk_ban_state").
			Default(false).
			Comment("Whether user is banned from PJSK features"),
		field.String("pjsk_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for PJSK ban"),
		field.Bool("chunithm_ban_state").
			Default(false).
			Comment("Whether user is banned from Chunithm features"),
		field.String("chunithm_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for Chunithm ban"),
		field.Bool("pjsk_main_ban_state").
			Default(false).
			Comment("Whether user is banned from PJSK Main features"),
		field.String("pjsk_main_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for PJSK Main ban"),
		field.Bool("pjsk_ranking_ban_state").
			Default(false).
			Comment("Whether user is banned from PJSK Ranking features"),
		field.String("pjsk_ranking_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for PJSK Ranking ban"),
		field.Bool("pjsk_alias_ban_state").
			Default(false).
			Comment("Whether user is banned from PJSK Alias features"),
		field.String("pjsk_alias_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for PJSK Alias ban"),
		field.Bool("pjsk_mysekai_ban_state").
			Default(false).
			Comment("Whether user is banned from PJSK Mysekai features"),
		field.String("pjsk_mysekai_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for PJSK Mysekai ban"),
		field.Bool("chunithm_main_ban_state").
			Default(false).
			Comment("Whether user is banned from Chunithm Main features"),
		field.String("chunithm_main_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for Chunithm Main ban"),
		field.Bool("chunithm_alias_ban_state").
			Default(false).
			Comment("Whether user is banned from Chunithm Alias features"),
		field.String("chunithm_alias_ban_reason").
			MaxLen(255).
			Optional().
			Comment("Reason for Chunithm Alias ban"),
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
