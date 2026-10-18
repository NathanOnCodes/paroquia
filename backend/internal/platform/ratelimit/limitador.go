// Package ratelimit implements the fixed-window limiter that protects the
// public endpoints.
//
// O estado vive na tabela `contact_rate_limit`, como no sistema legado
// system, so the limit survives restarts and works across replicas. The
// granularity is one row per attempt: counting rows is the window.
package ratelimit

import (
	"context"
	"time"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/limitetaxacontato"
)

// Decision is the outcome of a check.
type DecisaoDeLimite struct {
	// Allowed reports whether the request may proceed.
	Permitido bool
	// Remaining is how many attempts are left in the current window.
	Restantes int
	// RetryAfter is how long to wait before trying again. Zero when allowed.
	RepetirDepoisDe time.Duration
}

// Limiter is the fixed-window rate limiter.
type Limitador struct {
	client *ent.Client
	now    func() time.Time
}

// New builds a Limiter. The clock is injectable for tests.
func New(client *ent.Client, clock func() time.Time) *Limitador {
	if clock == nil {
		clock = time.Now
	}
	return &Limitador{client: client, now: clock}
}

// Check evaluates the window for (kind, key) without recording an attempt.
func (l *Limitador) Verificar(ctx context.Context, kind, key string, max int, window time.Duration) (DecisaoDeLimite, error) {
	if max <= 0 {
		return DecisaoDeLimite{Permitido: true}, nil
	}

	since := l.now().Add(-window)
	count, err := l.client.LimiteTaxaContato.Query().
		Where(
			limitetaxacontato.Kind(kind),
			limitetaxacontato.IPHash(key),
			limitetaxacontato.CreatedAtGTE(since),
		).
		Count(ctx)
	if err != nil {
		return DecisaoDeLimite{}, err
	}

	if count >= max {
		retry := window - l.now().Sub(since)
		if retry < 0 {
			retry = 0
		}
		return DecisaoDeLimite{Permitido: false, Restantes: 0, RepetirDepoisDe: window}, nil
	}
	return DecisaoDeLimite{Permitido: true, Restantes: max - count}, nil
}

// Allow records an attempt and reports whether it fits in the window.
//
// É propositalmente "grava e depois conta": num rajada duas requisições
// pass, which is acceptable for a contact form and keeps the implementation
// free of locks and advisory transactions.
func (l *Limitador) Permitir(ctx context.Context, kind, key string, max int, window time.Duration) (DecisaoDeLimite, error) {
	decision, err := l.Verificar(ctx, kind, key, max, window)
	if err != nil {
		return DecisaoDeLimite{}, err
	}
	if !decision.Permitido {
		return decision, nil
	}

	if err := l.client.LimiteTaxaContato.Create().
		SetKind(kind).
		SetIPHash(key).
		Exec(ctx); err != nil {
		return DecisaoDeLimite{}, err
	}
	return decision, nil
}

// Prune removes entries older than the given retention. It is meant to be run
// periodically so the table does not grow without bound.
func (l *Limitador) LimparRegistrosAntigos(ctx context.Context, olderThan time.Duration) (int, error) {
	cutoff := l.now().Add(-olderThan)
	return l.client.LimiteTaxaContato.Delete().
		Where(limitetaxacontato.CreatedAtLT(cutoff)).
		Exec(ctx)
}

// Tipos usados pela aplicação. Espelham a coluna `kind` legada.
const (
	TipoContato           = "contact"
	TipoLinkDeRecorrencia = "recurring_link"
	TipoLogin             = "login"
	TipoPagamento         = "payment"
)
