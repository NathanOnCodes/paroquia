package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// TokenMagicoRecorrente lets a donor manage their recurring donations without an
// account. The design is carried over unchanged from the legacy system: a
// random bearer sent by e-mail, stored only as a SHA-256 hash, single use and
// time limited.
type TokenMagicoRecorrente struct {
	ent.Schema
}

func (TokenMagicoRecorrente) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "recurring_magic_tokens"}}
}

func (TokenMagicoRecorrente) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinApenasCriadoEm{}}
}

func (TokenMagicoRecorrente) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("token_hash").
			NotEmpty().
			Unique().
			Immutable(),
		field.String("email").
			NotEmpty(),
		field.Time("expires_at"),
		field.Time("consumed_at").
			Optional().
			Nillable(),
	}
}
