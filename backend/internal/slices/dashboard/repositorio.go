// Package dashboard aggregates the numbers the parish office watches.
//
// The aggregation happens in the database rather than in memory: the legacy
// implementation pulled every paid contribution and summed it in Node, which
// does not survive a growing ledger.
package dashboard

import (
	"context"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/doacaorecorrente"
	"github.com/NathanOnCodes/paroquia/backend/ent/lancamentofinanceiro"
	"github.com/NathanOnCodes/paroquia/backend/ent/pagamentocontribuicao"
)

// Repository is the persistence contract.
type Repositorio interface {
	// Totals returns the headline numbers.
	Totais(ctx context.Context) (Totais, error)
	// CommunityRanking returns the total raised per community, best first.
	CommunityRanking(ctx context.Context, limit int) ([]TotalPorComunidade, error)
}

// Totals holds the headline numbers of the dashboard.
type Totais struct {
	TotalRaisedCents  int64
	TotalExpenseCents int64
	TotalDonors       int
	PendingPayments   int
	FailedPayments    int
	CanceledRecurring int
	ActiveRecurring   int
}

type repositorio struct{ client *ent.Client }

// NewRepository builds the Ent-backed repository.
func NovoRepositorio(client *ent.Client) Repositorio { return &repositorio{client: client} }

func (r *repositorio) Totais(ctx context.Context) (Totais, error) {
	var out Totais
	var err error

	if out.TotalRaisedCents, err = r.somarPagas(ctx); err != nil {
		return out, err
	}
	if out.TotalExpenseCents, err = r.somarDespesasManuais(ctx); err != nil {
		return out, err
	}
	if out.TotalDonors, err = r.client.Doador.Query().Count(ctx); err != nil {
		return out, err
	}
	if out.PendingPayments, err = r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.StatusEQ(pagamentocontribuicao.StatusPending)).
		Count(ctx); err != nil {
		return out, err
	}
	if out.FailedPayments, err = r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.StatusEQ(pagamentocontribuicao.StatusFailed)).
		Count(ctx); err != nil {
		return out, err
	}
	if out.CanceledRecurring, err = r.client.DoacaoRecorrente.Query().
		Where(doacaorecorrente.StatusEQ(doacaorecorrente.StatusCanceled)).
		Count(ctx); err != nil {
		return out, err
	}
	if out.ActiveRecurring, err = r.client.DoacaoRecorrente.Query().
		Where(doacaorecorrente.StatusEQ(doacaorecorrente.StatusActive)).
		Count(ctx); err != nil {
		return out, err
	}
	return out, nil
}

// sumPaid returns the total of every settled contribution.
func (r *repositorio) somarPagas(ctx context.Context) (int64, error) {
	var rows []struct {
		Total int64 `json:"total"`
	}
	err := r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.StatusEQ(pagamentocontribuicao.StatusPaid)).
		Aggregate(ent.As(ent.Sum(pagamentocontribuicao.FieldAmountCents), "total")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Total, nil
}

// sumManualExpenses returns the total of the manually recorded expenses.
func (r *repositorio) somarDespesasManuais(ctx context.Context) (int64, error) {
	var rows []struct {
		Total int64 `json:"total"`
	}
	err := r.client.LancamentoFinanceiro.Query().
		Where(
			lancamentofinanceiro.TypeEQ(lancamentofinanceiro.TypeExpense),
			lancamentofinanceiro.SourceEQ(lancamentofinanceiro.SourceManual),
		).
		Aggregate(ent.As(ent.Sum(lancamentofinanceiro.FieldAmountCents), "total")).
		Scan(ctx, &rows)
	if err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return rows[0].Total, nil
}

func (r *repositorio) CommunityRanking(ctx context.Context, limit int) ([]TotalPorComunidade, error) {
	var rows []struct {
		CommunityID string `json:"community_id"`
		Total       int64  `json:"total"`
		Count       int    `json:"count"`
	}
	err := r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.StatusEQ(pagamentocontribuicao.StatusPaid)).
		GroupBy(pagamentocontribuicao.FieldCommunityID).
		Aggregate(
			ent.As(ent.Sum(pagamentocontribuicao.FieldAmountCents), "total"),
			ent.As(ent.Count(), "count"),
		).
		Scan(ctx, &rows)
	if err != nil {
		return nil, err
	}

	// Resolve names in a second query: the group-by keeps the first query
	// simple and the community list is small.
	names := map[string]string{}
	comms, err := r.client.Comunidade.Query().All(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range comms {
		names[c.ID.String()] = c.Name
	}

	out := make([]TotalPorComunidade, 0, len(rows))
	for _, rw := range rows {
		name := names[rw.CommunityID]
		if name == "" {
			name = "Comunidade livre"
		}
		out = append(out, TotalPorComunidade{
			CommunityID:   rw.CommunityID,
			CommunityName: name,
			AmountCents:   rw.Total,
			Contributions: rw.Count,
		})
	}

	// Sort descending by amount; the slice is tiny.
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j].AmountCents > out[j-1].AmountCents; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}
