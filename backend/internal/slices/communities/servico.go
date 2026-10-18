package communities

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Service is the business surface of the communities slice.
type Servico interface {
	// ListPublic returns the active communities, for the donation form.
	ListarPublico(ctx context.Context) ([]Resposta, error)
	// ListAdmin returns every community.
	ListarAdmin(ctx context.Context) ([]Resposta, error)
	// GetAdmin returns one community.
	ObterAdmin(ctx context.Context, id string) (Resposta, error)
	// Create inserts a community.
	Criar(ctx context.Context, actor authx.UsuarioAutenticado, req RequisicaoDeEscrita) (Resposta, error)
	// Update rewrites a community.
	Atualizar(ctx context.Context, actor authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita) (Resposta, error)
	// SetStatus toggles visibility.
	DefinirStatus(ctx context.Context, actor authx.UsuarioAutenticado, id string, active bool) (Resposta, error)
}

type servico struct {
	repo  Repositorio
	audit audit.RegistradorDeAuditoria
	log   *slog.Logger
}

// NewService wires the communities service.
func NovoServico(repo Repositorio, auditor audit.RegistradorDeAuditoria, log *slog.Logger) Servico {
	if log == nil {
		log = slog.Default()
	}
	return &servico{repo: repo, audit: auditor, log: log}
}

func (s *servico) ListarPublico(ctx context.Context) ([]Resposta, error) {
	rows, err := s.repo.ListarAtivas(ctx)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar comunidades", err)
	}
	return paraRespostas(rows), nil
}

func (s *servico) ListarAdmin(ctx context.Context) ([]Resposta, error) {
	rows, err := s.repo.ListarTodas(ctx)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar comunidades", err)
	}
	return paraRespostas(rows), nil
}

func (s *servico) ObterAdmin(ctx context.Context, id string) (Resposta, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return Resposta{}, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}
	row, err := s.repo.Obter(ctx, uid)
	if err != nil {
		return Resposta{}, traduzir(err)
	}
	return paraResposta(row), nil
}

func (s *servico) Criar(ctx context.Context, actor authx.UsuarioAutenticado, req RequisicaoDeEscrita) (Resposta, error) {
	in := req.paraEntrada()

	exists, err := s.repo.NomeExiste(ctx, in.Name, uuid.Nil)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao verificar comunidade", err)
	}
	if exists {
		return Resposta{}, httpx.ErroConflito("Já existe uma comunidade com este nome")
	}

	row, err := s.repo.Criar(ctx, in)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: actor.ID, Action: "create_community",
		EntityType: "community", EntityID: row.ID.String(),
	})
	return paraResposta(row), nil
}

func (s *servico) Atualizar(ctx context.Context, actor authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita) (Resposta, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return Resposta{}, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}
	in := req.paraEntrada()

	exists, err := s.repo.NomeExiste(ctx, in.Name, uid)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao verificar comunidade", err)
	}
	if exists {
		return Resposta{}, httpx.ErroConflito("Já existe uma comunidade com este nome")
	}

	row, err := s.repo.Atualizar(ctx, uid, in)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: actor.ID, Action: "update_community",
		EntityType: "community", EntityID: row.ID.String(),
	})
	return paraResposta(row), nil
}

func (s *servico) DefinirStatus(ctx context.Context, actor authx.UsuarioAutenticado, id string, active bool) (Resposta, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return Resposta{}, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}

	row, err := s.repo.DefinirAtivo(ctx, uid, active)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	action := "activate_community"
	if !active {
		action = "deactivate_community"
	}
	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: actor.ID, Action: action,
		EntityType: "community", EntityID: row.ID.String(),
	})
	return paraResposta(row), nil
}

func (r RequisicaoDeEscrita) paraEntrada() EntradaDeEscrita {
	return EntradaDeEscrita{
		Name:        r.Name,
		Description: r.Description,
		City:        r.City,
		IsActive:    r.IsActive,
	}
}

// translate maps repository errors onto the HTTP contract.
func traduzir(err error) error {
	switch {
	case errors.Is(err, ErrNaoEncontrado):
		return httpx.ErroNaoEncontrado("Comunidade não encontrada")
	case errors.Is(err, ErrNameTaken), database.EhViolacaoDeUnicidade(err):
		return httpx.ErroConflito("Já existe uma comunidade com este nome")
	default:
		return httpx.ErroInterno("Falha ao processar comunidade", err)
	}
}

func paraRespostas(rows []*ent.Comunidade) []Resposta {
	out := make([]Resposta, 0, len(rows))
	for _, r := range rows {
		out = append(out, paraResposta(r))
	}
	return out
}
