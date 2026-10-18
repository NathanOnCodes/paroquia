package dashboard

import (
	"context"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Service is the business surface of the dashboard slice.
type Servico interface {
	// Summary returns the headline numbers and the community ranking.
	Resumo(ctx context.Context) (Resposta, error)
}

type servico struct{ repo Repositorio }

// NewService wires the dashboard service.
func NovoServico(repo Repositorio) Servico { return &servico{repo: repo} }

func (s *servico) Resumo(ctx context.Context) (Resposta, error) {
	totals, err := s.repo.Totais(ctx)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao calcular indicadores", err)
	}

	ranking, err := s.repo.CommunityRanking(ctx, 10)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao calcular ranking", err)
	}

	out := Resposta{
		TotalRaisedCents:  totals.TotalRaisedCents,
		TotalExpenseCents: totals.TotalExpenseCents,
		BalanceCents:      totals.TotalRaisedCents - totals.TotalExpenseCents,
		TotalDonors:       totals.TotalDonors,
		PendingPayments:   totals.PendingPayments,
		FailedPayments:    totals.FailedPayments,
		CanceledRecurring: totals.CanceledRecurring,
		ActiveRecurring:   totals.ActiveRecurring,
		CommunityRanking:  make([]TotalPorComunidade, 0, len(ranking)),
	}
	for _, r := range ranking {
		out.CommunityRanking = append(out.CommunityRanking, TotalPorComunidade{
			CommunityID:   r.CommunityID,
			CommunityName: r.CommunityName,
			AmountCents:   r.AmountCents,
			Contributions: r.Contributions,
		})
	}
	return out, nil
}
