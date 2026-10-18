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

// LancamentoFinanceiro is a single income or expense line inside a period.
type LancamentoFinanceiro struct {
	ent.Schema
}

func (LancamentoFinanceiro) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "financial_entries"}}
}

func (LancamentoFinanceiro) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (LancamentoFinanceiro) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.UUID("financial_period_id", uuid.UUID{}),
		field.Enum("type").
			Values("income", "expense"),
		field.String("category").
			NotEmpty(),
		field.String("description").
			NotEmpty(),
		field.Int("amount_cents").
			Positive(),
		field.Time("entry_date").
			SchemaType(map[string]string{dialect.Postgres: "date"}),
		field.Enum("source").
			Values("online_contribution", "manual", "other").
			Default("manual"),
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("updated_by", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (LancamentoFinanceiro) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("period", PeriodoFinanceiro.Type).
			Ref("entries").
			Field("financial_period_id").
			Unique().
			Required(),
	}
}

// Financial entry type and source values, mirrored by CHECK constraints in SQL.
const (
	EntryTypeIncome  = "income"
	EntryTypeExpense = "expense"

	EntrySourceOnline = "online_contribution"
	EntrySourceManual = "manual"
	EntrySourceOther  = "other"
)
