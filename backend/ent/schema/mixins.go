package schema

import (
	"time"

	"github.com/google/uuid"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// campoID returns the UUID primary key used by every entity.
//
// NOTE: Ent silently ignores an `id` field declared inside a mixin (it always
// falls back to an auto-increment int), so each schema must call this helper
// explicitly as the first entry of its Fields() list.
func campoID() ent.Field {
	return field.UUID("id", uuid.UUID{}).
		Default(uuid.New).
		Immutable()
}

// MixinTimestamps adds the audit columns shared by almost every table.
// `updated_at` is maintained by Ent (UpdateDefault) instead of by database
// triggers, which removes the inconsistency found in the legacy schema where
// only two tables had `moddatetime` triggers.
type MixinTimestamps struct{ mixin.Schema }

func (MixinTimestamps) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// MixinApenasCriadoEm is used by append-only tables that never update a row
// (for example `transparency_documents` and `contact_rate_limit`).
type MixinApenasCriadoEm struct{ mixin.Schema }

func (MixinApenasCriadoEm) Fields() []ent.Field {
	return []ent.Field{
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}
