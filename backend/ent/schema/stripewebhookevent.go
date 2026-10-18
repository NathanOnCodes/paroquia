package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// EventoWebhookStripe is the idempotency ledger of the payment webhook. The
// unique index on `stripe_event_id` is the real guarantee; the lookup done
// before processing is only an optimisation.
type EventoWebhookStripe struct {
	ent.Schema
}

func (EventoWebhookStripe) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "stripe_webhook_events"}}
}

func (EventoWebhookStripe) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinApenasCriadoEm{}}
}

func (EventoWebhookStripe) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("stripe_event_id").
			NotEmpty().
			Unique(),
		field.String("event_type").
			NotEmpty(),
		field.String("payload_hash").
			NotEmpty(),
		field.Enum("processing_status").
			Values("processed", "failed", "skipped").
			Default("processed"),
		field.String("error_message").
			Optional().
			Nillable(),
	}
}
