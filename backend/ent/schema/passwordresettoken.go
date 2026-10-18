package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TokenRedefinicaoSenha lets a staff member recover access. It follows the same
// design as the legacy magic link: random bearer, SHA-256 at rest, single use.
type TokenRedefinicaoSenha struct {
	ent.Schema
}

func (TokenRedefinicaoSenha) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "password_reset_tokens"}}
}

func (TokenRedefinicaoSenha) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinApenasCriadoEm{}}
}

func (TokenRedefinicaoSenha) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("token_hash").
			NotEmpty().
			Unique().
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.Time("expires_at"),
		field.Time("consumed_at").
			Optional().
			Nillable(),
	}
}

func (TokenRedefinicaoSenha) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", Usuario.Type).
			Ref("password_reset_tokens").
			Field("user_id").
			Unique().
			Required(),
	}
}
