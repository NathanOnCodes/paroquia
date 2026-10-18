package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// Comunidade is a parish community (or chapel) that receives the tithes.
type Comunidade struct {
	ent.Schema
}

func (Comunidade) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "communities"}}
}

func (Comunidade) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (Comunidade) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Default(""),
		field.String("city").
			Default(""),
		field.Bool("is_active").
			Default(true),
	}
}

func (Comunidade) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("donors", Doador.Type),
		edge.To("recurring_donations", DoacaoRecorrente.Type),
		edge.To("contribution_payments", PagamentoContribuicao.Type),
	}
}
