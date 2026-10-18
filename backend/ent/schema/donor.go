package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Doador is a faithful who contributes. The email is the natural key used to
// recognise a returning donor, exactly as the legacy `upsertDonor` did.
//
// A donor either points at a registered community (`community_id`) or names a
// free-form one (`community_name`); never both at the same time.
type Doador struct {
	ent.Schema
}

func (Doador) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "donors"}}
}

func (Doador) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (Doador) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("full_name").
			NotEmpty(),
		field.String("email").
			NotEmpty(),
		field.String("phone").
			Optional().
			Nillable(),
		field.String("city").
			Default(""),
		field.String("neighborhood").
			Default(""),
		field.UUID("community_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("community_name").
			Optional().
			Nillable(),
		field.Time("privacy_consent_at").
			Optional().
			Nillable(),
		field.Bool("is_active").
			Default(true),
	}
}

func (Doador) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("community", Comunidade.Type).
			Ref("donors").
			Field("community_id").
			Unique(),
		edge.To("recurring_donations", DoacaoRecorrente.Type),
		edge.To("contribution_payments", PagamentoContribuicao.Type),
	}
}
