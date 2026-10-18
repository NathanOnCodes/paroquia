package schema

import (
	"time"

	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// MensagemContato is a message sent through the public contact form.
type MensagemContato struct {
	ent.Schema
}

func (MensagemContato) Annotations() []schema.Annotation {
	return []schema.Annotation{entsql.Annotation{Table: "contact_messages"}}
}

func (MensagemContato) Mixin() []ent.Mixin {
	return []ent.Mixin{MixinTimestamps{}}
}

func (MensagemContato) Fields() []ent.Field {
	return []ent.Field{
		campoID(),
		field.String("name").
			NotEmpty(),
		field.String("email").
			NotEmpty(),
		field.String("phone").
			Optional().
			Nillable(),
		field.String("message").
			NotEmpty(),
		field.Time("privacy_consent_at").
			Default(time.Now),
		field.String("ip_hash").
			Optional().
			Nillable(),
		field.Enum("status").
			Values("new", "in_progress", "resolved", "archived").
			Default("new"),
		field.UUID("assigned_to", uuid.UUID{}).
			Optional().
			Nillable(),
	}
}

func (MensagemContato) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("assignee", Usuario.Type).
			Ref("assigned_contacts").
			Field("assigned_to").
			Unique(),
	}
}
