package schema

import (
	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Evento is a parish event published on the public agenda.
//
// `created_by`/`updated_by` reference staff users but are intentionally kept as
// plain columns instead of edges: they are audit trails, never joined.
type Evento struct {
	ent.Schema
}

func (Evento) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "events"}}
}

func (Evento) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (Evento) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("title").
			NotEmpty(),
		field.String("slug").
			NotEmpty().
			Unique(),
		field.String("description").
			Default(""),
		field.String("image_path").
			Optional().
			Nillable(),
		field.Time("starts_at"),
		field.Time("ends_at").
			Optional().
			Nillable(),
		field.String("location").
			Default(""),
		field.Enum("status").
			Values("draft", "published", "archived").
			Default("draft"),
		field.UUID("created_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.UUID("updated_by", uuid.UUID{}).
			Optional().
			Nillable(),
		field.Time("published_at").
			Optional().
			Nillable(),
	}
}
