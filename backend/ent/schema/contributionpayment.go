package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PagamentoContribuicao is one settled (or attempted) contribution. Rows are
// created by the Stripe webhook, never by the public request handler, which
// keeps the ledger driven by the payment provider's truth.
type PagamentoContribuicao struct {
	ent.Schema
}

func (PagamentoContribuicao) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "contribution_payments"}}
}

func (PagamentoContribuicao) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (PagamentoContribuicao) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.UUID("donor_id", uuid.UUID{}),
		field.UUID("community_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("recurring_donation_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("community_name").
			Optional().
			Nillable(),
		field.Int("amount_cents").
			Positive(),
		field.String("currency").
			Default("brl"),
		field.Enum("payment_type").
			Values("one_time", "recurring"),
		field.Enum("status").
			Values("pending", "paid", "failed", "canceled", "refunded").
			Default("pending"),
		field.String("stripe_payment_intent_id").
			Optional().
			Nillable(),
		field.String("stripe_invoice_id").
			Optional().
			Nillable(),
		field.String("stripe_subscription_id").
			Optional().
			Nillable(),
		field.String("stripe_event_id").
			Optional().
			Nillable(),
		field.Time("paid_at").
			Optional().
			Nillable(),
		field.Time("refunded_at").
			Optional().
			Nillable(),
	}
}

func (PagamentoContribuicao) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("donor", Doador.Type).
			Ref("contribution_payments").
			Field("donor_id").
			Unique().
			Required(),
		edge.From("community", Comunidade.Type).
			Ref("contribution_payments").
			Field("community_id").
			Unique(),
		edge.From("recurring_donation", DoacaoRecorrente.Type).
			Ref("contribution_payments").
			Field("recurring_donation_id").
			Unique(),
	}
}

// Payment status and type values, mirrored by CHECK constraints in SQL.
const (
	PaymentStatusPending  = "pending"
	PaymentStatusPaid     = "paid"
	PaymentStatusFailed   = "failed"
	PaymentStatusCanceled = "canceled"
	PaymentStatusRefunded = "refunded"

	PaymentTypeOneTime   = "one_time"
	PaymentTypeRecurring = "recurring"
)
