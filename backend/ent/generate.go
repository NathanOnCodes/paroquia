package ent

// The `sql/upsert` feature generates ON CONFLICT builders. It is required by
// the settings singleton and, more importantly, by the Stripe webhook
// idempotency ledger, which relies on INSERT ... ON CONFLICT DO NOTHING.
//
//go:generate go run -mod=mod entgo.io/ent/cmd/ent generate ./schema --feature sql/upsert,squash
