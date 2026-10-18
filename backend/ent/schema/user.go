package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Usuario is the administrative account of the system. It replaces the legacy
// `auth.users` + `profiles` pair: credentials and profile live together now.
//
// Only staff members have accounts; the faithful never log in.
type Usuario struct {
	ent.Schema
}

func (Usuario) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "users"}}
}

func (Usuario) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (Usuario) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("email").
			NotEmpty().
			Unique().
			Immutable(),
		field.String("password_hash").
			NotEmpty(),
		field.String("full_name").
			NotEmpty(),
		field.Enum("role").
			Values("admin", "assistente").
			Default("assistente"),
		field.Bool("is_active").
			Default(true),
		field.Bool("must_change_password").
			Default(false),
		field.Time("password_changed_at").
			Optional().
			Nillable(),
		field.Time("last_login_at").
			Optional().
			Nillable(),
	}
}

func (Usuario) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("refresh_tokens", TokenAtualizacao.Type),
		edge.To("password_reset_tokens", TokenRedefinicaoSenha.Type),
		edge.To("assigned_contacts", MensagemContato.Type),
	}
}

// Roles accepted by the `users.role` column. The same values are enforced by a
