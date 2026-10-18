package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// TokenAtualizacao stores the opaque half of the session. Only the SHA-256 hash of
// the token is persisted, exactly like the legacy `recurring_magic_tokens`
// table did, so a database leak cannot be replayed as a session.
//
// Tokens are rotated on every refresh; `reused_at` marks a replay attempt and
// revokes the whole family.
type TokenAtualizacao struct {
	ent.Schema
}

func (TokenAtualizacao) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "refresh_tokens"}}
}

func (TokenAtualizacao) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinApenasCriadoEm{}}
}

func (TokenAtualizacao) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("token_hash").
			NotEmpty().
			Unique().
			Immutable(),
		field.UUID("user_id", uuid.UUID{}),
		field.UUID("family_id", uuid.UUID{}),
		field.Time("expires_at"),
		field.Time("revoked_at").
			Optional().
			Nillable(),
		field.Time("reused_at").
			Optional().
			Nillable(),
	}
}

func (TokenAtualizacao) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", Usuario.Type).
			Ref("refresh_tokens").
			Field("user_id").
			Unique().
			Required(),
	}
}
