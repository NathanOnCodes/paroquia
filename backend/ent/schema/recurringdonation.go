package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DoacaoRecorrente is the local mirror of a Stripe subscription.
//
// The lifecycle is driven by the Stripe webhook: `incomplete` right after the
// subscription is created, `active` once the first invoice is paid, `past_due`
// when an invoice fails, `cancel_at_period_end` when the donor asks to stop and
// `canceled` when Stripe finally deletes it.
type DoacaoRecorrente struct {
	ent.Schema
}

func (DoacaoRecorrente) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "recurring_donations"}}
}

func (DoacaoRecorrente) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (DoacaoRecorrente) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.UUID("donor_id", uuid.UUID{}),
		field.UUID("community_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Int("amount_cents").
			Positive(),
		field.String("currency").
			Default("brl"),
		field.String("interval").
			Default("month"),
		field.String("stripe_customer_id").
			NotEmpty(),
		field.String("stripe_subscription_id").
			NotEmpty().
			Unique(),
		field.Enum("status").
			Values("incomplete", "active", "past_due", "cancel_at_period_end", "canceled").
			Default("incomplete"),
		field.Time("started_at").
			Optional().
			Nillable(),
		field.Bool("cancel_at_period_end").
			Default(false),
		field.Time("cancellation_requested_at").
			Optional().
			Nillable(),
		field.Time("canceled_at").
			Optional().
			Nillable(),
		field.Time("current_period_end").
			Optional().
			Nillable(),
	}
}

func (DoacaoRecorrente) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("donor", Doador.Type).
			Ref("recurring_donations").
			Field("donor_id").
			Unique().
			Required(),
		edge.From("community", Comunidade.Type).
			Ref("recurring_donations").
			Field("community_id").
			Unique(),
		edge.To("contribution_payments", PagamentoContribuicao.Type),
	}
}
