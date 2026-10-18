package database

import (
	"context"
	"errors"
	"fmt"
	"strings"

	entsql "entgo.io/ent/dialect/sql"

	"github.com/NathanOnCodes/paroquia/backend/ent"
)

// Actor identifies who is performing a database operation. It is published to
// (comentário removido: texto original indisponível)
// (comentário removido: texto original indisponível)
//
// Publishing the actor is what keeps the database-level guards meaningful:
// (comentário removido: texto original indisponível)
// forgets to declare its actor cannot publish a financial period or change a
// user's role by accident.
type Ator struct {
	// (comentário removido: texto original indisponível)
	ID string
}

// System is the actor used by background work that legitimately bypasses the
// guards (seeds, maintenance, data migration).
var Sistema = Ator{}

// actorKey is the context key holding the Actor.
type chaveDoAtor struct{}

// WithActor returns a context that publishes the actor to every statement run
// with it.
func ComAtor(ctx context.Context, actor Ator) context.Context {
	if actor.ID == "" {
		return ctx
	}
	if err := validarIDDoAtor(actor.ID); err != nil {
		// A malformed id must never end up interpolated into a SET statement.
		panic(fmt.Sprintf("database.WithActor: %v", err))
	}
	ctx = entsql.WithVar(ctx, "app.actor_id", actor.ID)
	return context.WithValue(ctx, chaveDoAtor{}, actor)
}

// ActorFrom returns the actor carried by ctx.
func AtorDe(ctx context.Context) (Ator, bool) {
	a, ok := ctx.Value(chaveDoAtor{}).(Ator)
	return a, ok
}

// validateActorID accepts only UUID-ish literals: the value is interpolated
// into a SET statement by the Ent driver, so it must never contain a quote.
func validarIDDoAtor(id string) error {
	if len(id) != 36 {
		return fmt.Errorf("id de ator inválido: %q", id)
	}
	for i, r := range id {
		isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		if !isHex && r != '-' {
			return fmt.Errorf("id de ator inválido: %q", id)
		}
		if r == '-' && i != 8 && i != 13 && i != 18 && i != 23 {
			return fmt.Errorf("id de ator inválido: %q", id)
		}
	}
	if strings.ContainsAny(id, "'\";\\") {
		return errors.New("id de ator contém caracteres proibidos")
	}
	return nil
}

// WithTx runs fn inside a transaction, publishing the actor to the session so
// the database triggers can enforce their rules. It rolls back on error or
// panic.
//
// Every write that touches a guarded table (users.role, financial_periods
// .status) MUST go through this function with a real Actor; otherwise the
// trigger rejects the statement.
func ComTransacao(ctx context.Context, client *ent.Client, actor Ator, fn func(ctx context.Context, tx *ent.Tx) error) error {
	tx, err := client.Tx(ComAtor(ctx, actor))
	if err != nil {
		return fmt.Errorf("abrindo transação: %w", err)
	}

	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()

	if err := fn(ComAtor(ctx, actor), tx); err != nil {
		if rerr := tx.Rollback(); rerr != nil {
			return fmt.Errorf("%w; rollback falhou: %v", err, rerr)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("confirmando transação: %w", err)
	}
	return nil
}
