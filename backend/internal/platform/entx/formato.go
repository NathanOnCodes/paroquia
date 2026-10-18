// Package entx holds small conversion helpers shared by the slices when they
// turn Ent entities into wire DTOs. Keeping them in one place avoids every
// slice re-implementing (and diverging on) null and time formatting.
package entx

import (
	"time"

	"github.com/google/uuid"
)

// TimeFormat is the canonical timestamp format of the API. It matches the
// RFC3339 shape the legacy TypeScript client already expects.
const FormatoDeData = time.RFC3339

// UUIDString renders a UUID for the wire.
func UUIDParaTexto(id uuid.UUID) string {
	if id == uuid.Nil {
		return ""
	}
	return id.String()
}

// UUIDStringPtr renders a nullable UUID.
func UUIDParaTextoOpcional(id *uuid.UUID) *string {
	if id == nil {
		return nil
	}
	s := id.String()
	return &s
}

// TimeString renders a timestamp for the wire.
func DataParaTexto(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(FormatoDeData)
}

// TimeStringPtr renders a nullable timestamp. Zero values become NULL.
func DataParaTextoOpcional(t *time.Time) *string {
	if t == nil || t.IsZero() {
		return nil
	}
	s := t.Format(FormatoDeData)
	return &s
}

// DateString renders a date-only column (no time component).
func DataSimples(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02")
}

// DatePtr returns a pointer to t, for nullable Ent fields.
func DataOpcional(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// TimePtr returns a pointer to t, for nullable Ent fields.
func DataHoraOpcional(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

// UUIDPtr returns a pointer to id, for nullable Ent fields.
func UUIDOpcional(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
