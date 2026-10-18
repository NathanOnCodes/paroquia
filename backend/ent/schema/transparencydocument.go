package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// DocumentoTransparencia is a DRE (or receipt) attached to a financial period.
// Files live in the storage volume and are only reachable through the file
// endpoint once the parent period is published.
type DocumentoTransparencia struct {
	ent.Schema
}

func (DocumentoTransparencia) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "transparency_documents"}}
}

func (DocumentoTransparencia) Mixin() []ent.Mixin {
	// Append-only table: no `updated_at` in the legacy schema either.
	return []ent.Mixin{MixinApenasCriadoEm{}}
}

func (DocumentoTransparencia) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.UUID("financial_period_id", uuid.UUID{}),
		field.String("file_path").
			NotEmpty(),
		field.String("original_file_name").
			NotEmpty(),
		field.String("mime_type").
			NotEmpty(),
		field.Int("file_size").
			Positive(),
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (DocumentoTransparencia) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("period", PeriodoFinanceiro.Type).
			Ref("documents").
			Field("financial_period_id").
			Unique().
			Required(),
	}
}
