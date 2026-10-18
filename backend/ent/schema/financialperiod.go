package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// PeriodoFinanceiro is a window of financial accountability published to the
// faithful. Only an admin may move it to `published` or `archived`; the rule is
// enforced in the service layer and duplicated by a database trigger.
type PeriodoFinanceiro struct {
	ent.Schema
}

func (PeriodoFinanceiro) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "financial_periods"}}
}

func (PeriodoFinanceiro) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (PeriodoFinanceiro) Fields() []ent.Field {
	date := func(name string) ent.Field {
		return field.Time(name).
			SchemaType(map[string]string{dialect.Postgres: "date"})
	}

	return []ent.Field{
		campoID(),
		field.String("name").
			NotEmpty(),
		date("start_date"),
		date("end_date"),
		field.Enum("status").
			Values("draft", "review", "published", "archived").
			Default("draft"),
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("reviewed_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("published_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("published_at").
			Optional().
			Nillable(),
	}
}

func (PeriodoFinanceiro) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("entries", LancamentoFinanceiro.Type),
		edge.To("documents", DocumentoTransparencia.Type),
	}
}
