// Package audit records administrative mutations.
//
// Writing an audit entry must never break the operation it describes: a
// failure to log is reported to the logger and swallowed. That is exactly the
// (comentário removido: texto original indisponível)
package audit

import (
	"context"
	"log/slog"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/registroauditoria"
)

// Entry describes one audited action.
type RegistroDeAuditoria struct {
	// ActorID is the user responsible. Empty for donor-initiated actions.
	ActorID string
	// (comentário removido: texto original indisponível)
	Action string
	// (comentário removido: texto original indisponível)
	EntityType string
	// EntityID is the identifier of the aggregate.
	EntityID string
	// Metadata holds arbitrary context; it is stored as jsonb.
	Metadata map[string]any
}

// Logger writes audit entries. It is an interface so slices can be tested with
// an in-memory implementation.
type RegistradorDeAuditoria interface {
	Registrar(ctx context.Context, entry RegistroDeAuditoria)
}

// entLogger persists entries through the Ent client.
type registradorEnt struct {
	client *ent.Client
	log    *slog.Logger
}

// New builds the Ent-backed audit logger.
func New(client *ent.Client, log *slog.Logger) RegistradorDeAuditoria {
	return &registradorEnt{client: client, log: log}
}

// Log stores the entry. Errors are logged and discarded on purpose.
func (l *registradorEnt) Registrar(ctx context.Context, entry RegistroDeAuditoria) {
	if entry.Action == "" || entry.EntityType == "" {
		l.log.WarnContext(ctx, "audit entry incompleta ignorada",
			slog.String("action", entry.Action),
			slog.String("entity_type", entry.EntityType),
		)
		return
	}

	create := l.client.RegistroAuditoria.Create().
		SetAction(entry.Action).
		SetEntityType(entry.EntityType)

	if entry.ActorID != "" {
		actorID, err := uuid.Parse(entry.ActorID)
		if err != nil {
			l.log.WarnContext(ctx, "audit actor_id inválido ignorado",
				slog.String("actor_id", entry.ActorID),
			)
		} else {
			create.SetActorID(actorID)
		}
	}
	if entry.EntityID != "" {
		create.SetEntityID(entry.EntityID)
	}
	if entry.Metadata != nil {
		create.SetMetadata(entry.Metadata)
	}

	if err := create.Exec(ctx); err != nil {
		l.log.ErrorContext(ctx, "falha ao gravar auditoria",
			slog.String("action", entry.Action),
			slog.String("entity_type", entry.EntityType),
			slog.String("error", err.Error()),
		)
	}
}

// Nop is a logger that discards everything. Useful in tests and in commands
// that do not need an audit trail.
type Nulo struct{}

func (Nulo) Registrar(context.Context, RegistroDeAuditoria) {}

// Recorder keeps entries in memory. Test helper.
type RegistradorEmMemoria struct{ Registros []RegistroDeAuditoria }

func (r *RegistradorEmMemoria) Registrar(_ context.Context, entry RegistroDeAuditoria) {
	r.Registros = append(r.Registros, entry)
}

// NewQuery returns a read helper for the audit trail.
func NovaConsultaDeAuditoria(client *ent.Client) *ent.RegistroAuditoriaQuery {
	return client.RegistroAuditoria.Query().Order(ent.Desc(registroauditoria.FieldCreatedAt))
}
