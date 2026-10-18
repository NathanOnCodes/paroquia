// Package contributions is the read-only ledger the parish office works from.
//
// Rows are created by the payment webhook; this slice never writes them.
package contributions

import (
	"context"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/pagamentocontribuicao"
)

// Repository is the persistence contract.
type Repositorio interface {
	// List returns the most recent contributions.
	Listar(ctx context.Context, limit int) ([]*ent.PagamentoContribuicao, error)
	// ListByDonor returns the contribution history of one donor.
	ListarPorDoador(ctx context.Context, donorID uuid.UUID, limit int) ([]*ent.PagamentoContribuicao, error)
	// Get returns one contribution.
	Obter(ctx context.Context, id uuid.UUID) (*ent.PagamentoContribuicao, error)
}

type repositorio struct{ client *ent.Client }

// NewRepository builds the Ent-backed repository.
func NovoRepositorio(client *ent.Client) Repositorio { return &repositorio{client: client} }

func (r *repositorio) Listar(ctx context.Context, limit int) ([]*ent.PagamentoContribuicao, error) {
	return r.client.PagamentoContribuicao.Query().
		WithDonor().
		WithCommunity().
		Order(ent.Desc(pagamentocontribuicao.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

func (r *repositorio) ListarPorDoador(ctx context.Context, donorID uuid.UUID, limit int) ([]*ent.PagamentoContribuicao, error) {
	q := r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.DonorID(donorID)).
		WithDonor().
		WithCommunity().
		Order(ent.Desc(pagamentocontribuicao.FieldCreatedAt))
	if limit > 0 {
		q = q.Limit(limit)
	}
	return q.All(ctx)
}

func (r *repositorio) Obter(ctx context.Context, id uuid.UUID) (*ent.PagamentoContribuicao, error) {
	return r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.ID(id)).
		WithDonor().
		WithCommunity().
		Only(ctx)
}
