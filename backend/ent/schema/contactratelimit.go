package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// LimiteTaxaContato backs the fixed-window rate limiter for the public
// endpoints. It replaces the legacy table of the same name and now carries a
// `kind` column that the legacy TypeScript types forgot to declare.
type LimiteTaxaContato struct {
	ent.Schema
}

func (LimiteTaxaContato) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "contact_rate_limit"}}
}

func (LimiteTaxaContato) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinApenasCriadoEm{}}
}

func (LimiteTaxaContato) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("kind").
			Default("contact"),
		field.String("ip_hash").
			NotEmpty(),
	}
}
