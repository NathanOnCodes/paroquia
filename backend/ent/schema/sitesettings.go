package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ConfiguracaoSite is a singleton row (id is always 1) holding the branding and
// the Pix data of the parish. A CHECK constraint keeps it single.
type ConfiguracaoSite struct {
	ent.Schema
}

func (ConfiguracaoSite) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "site_settings"}}
}

func (ConfiguracaoSite) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (ConfiguracaoSite) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			Immutable().
			Default(1),
		field.String("site_name").
			Default("Paroquia"),
		field.String("site_title").
			Default("Paroquia"),
		field.String("city").
			Default(""),
		field.String("description").
			Default("Comunidade, fe e acolhimento."),
		field.String("secondary_color").
			Default("#1D4E89"),
		field.String("pix_key").
			Optional().
			Nillable(),
		field.String("pix_receiver_name").
			Optional().
			Nillable(),
		field.String("pix_receiver_city").
			Optional().
			Nillable(),
		field.String("home_title").
			Optional().
			Nillable(),
		field.String("home_description").
			Optional().
			Nillable(),
	}
}
