package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// RegistroAuditoria records every administrative mutation. Writing here never fails the
// surrounding operation: a broken audit trail must not block a donation.
type RegistroAuditoria struct {
	ent.Schema
}

func (RegistroAuditoria) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "audit_logs"}}
}

func (RegistroAuditoria) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinApenasCriadoEm{}}
}

func (RegistroAuditoria) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.UUID("actor_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("action").
			NotEmpty(),
		field.String("entity_type").
			NotEmpty(),
		field.String("entity_id").
			Optional().
			Nillable(),
		field.JSON("metadata", map[string]any{}).
			Optional(),
	}
}
