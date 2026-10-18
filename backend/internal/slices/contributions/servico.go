package contributions

import (
	"context"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Service is the business surface of the contributions slice.
type Servico interface {
	// ListAdmin returns the ledger for the admin screen.
	ListarAdmin(ctx context.Context) ([]Resposta, error)
	// ListByDonor returns the history of one donor.
	ListarPorDoador(ctx context.Context, donorID string) ([]Resposta, error)
}

type servico struct{ repo Repositorio }

// NewService wires the contributions service.
func NovoServico(repo Repositorio) Servico { return &servico{repo: repo} }

func (s *servico) ListarAdmin(ctx context.Context) ([]Resposta, error) {
	rows, err := s.repo.Listar(ctx, 300)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar contribuições", err)
	}
	return paraRespostas(rows), nil
}

func (s *servico) ListarPorDoador(ctx context.Context, donorID string) ([]Resposta, error) {
	uid, err := uuid.Parse(donorID)
	if err != nil {
		return nil, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "donor_id", Mensagem: "identificador inválido"})
	}
	rows, err := s.repo.ListarPorDoador(ctx, uid, 300)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar contribuições", err)
	}
	return paraRespostas(rows), nil
}
