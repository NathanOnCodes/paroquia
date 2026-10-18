package donors

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Service is the business surface of the donors slice.
type Servico interface {
	ListarAdmin(ctx context.Context, search string) ([]Resposta, error)
	ObterAdmin(ctx context.Context, id string) (Resposta, error)
	Criar(ctx context.Context, actor authx.UsuarioAutenticado, req RequisicaoDeEscrita) (Resposta, error)
	Atualizar(ctx context.Context, actor authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita) (Resposta, error)
	DefinirStatus(ctx context.Context, actor authx.UsuarioAutenticado, id string, active bool) (Resposta, error)

	// FindByEmail is used by the payment and recurring slices to resolve the
	// donor behind a donation.
	BuscarPorEmail(ctx context.Context, email string) (*ent.Doador, error)
	// UpsertFromDonation recognises a returning donor or creates a new one,
	// recording the privacy consent at that moment.
	GravarAPartirDaDoacao(ctx context.Context, in IdentidadeDaDoacao) (*ent.Doador, error)
}

// DonationIdentity is what the payment flow knows about the payer.
type IdentidadeDaDoacao struct {
	FullName      string
	Email         string
	Phone         string
	City          string
	Neighborhood  string
	CommunityID   *uuid.UUID
	CommunityName string
}

type servico struct {
	repo  Repositorio
	audit audit.RegistradorDeAuditoria
	log   *slog.Logger
}

// NewService wires the donors service.
func NovoServico(repo Repositorio, auditor audit.RegistradorDeAuditoria, log *slog.Logger) Servico {
	if log == nil {
		log = slog.Default()
	}
	return &servico{repo: repo, audit: auditor, log: log}
}

func (s *servico) ListarAdmin(ctx context.Context, search string) ([]Resposta, error) {
	rows, err := s.repo.Listar(ctx, search, 200)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar dizimistas", err)
	}
	return paraRespostas(rows), nil
}

func (s *servico) ObterAdmin(ctx context.Context, id string) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}
	row, err := s.repo.Obter(ctx, uid)
	if err != nil {
		return Resposta{}, traduzir(err)
	}
	return paraResposta(row), nil
}

func (s *servico) Criar(ctx context.Context, actor authx.UsuarioAutenticado, req RequisicaoDeEscrita) (Resposta, error) {
	in := req.paraEntrada()

	exists, err := s.repo.EmailExiste(ctx, in.Email, uuid.Nil)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao verificar dizimista", err)
	}
	if exists {
		return Resposta{}, httpx.ErroConflito("Já existe um dizimista com este e-mail")
	}

	row, err := s.repo.Criar(ctx, in)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: actor.ID, Action: "create_donor",
		EntityType: "donor", EntityID: row.ID.String(),
	})
	return paraResposta(row), nil
}

func (s *servico) Atualizar(ctx context.Context, actor authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}
	in := req.paraEntrada()

	exists, err := s.repo.EmailExiste(ctx, in.Email, uid)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao verificar dizimista", err)
	}
	if exists {
		return Resposta{}, httpx.ErroConflito("Já existe um dizimista com este e-mail")
	}

	row, err := s.repo.Atualizar(ctx, uid, in)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: actor.ID, Action: "update_donor",
		EntityType: "donor", EntityID: row.ID.String(),
	})
	return paraResposta(row), nil
}

func (s *servico) DefinirStatus(ctx context.Context, actor authx.UsuarioAutenticado, id string, active bool) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}

	row, err := s.repo.DefinirAtivo(ctx, uid, active)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	action := "activate_donor"
	if !active {
		action = "deactivate_donor"
	}
	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: actor.ID, Action: action,
		EntityType: "donor", EntityID: row.ID.String(),
	})
	return paraResposta(row), nil
}

func (s *servico) BuscarPorEmail(ctx context.Context, email string) (*ent.Doador, error) {
	row, err := s.repo.ObterPorEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			return nil, nil
		}
		return nil, err
	}
	return row, nil
}

func (s *servico) GravarAPartirDaDoacao(ctx context.Context, in IdentidadeDaDoacao) (*ent.Doador, error) {
	// A registered community and a free-form name are mutually exclusive.
	name := strings.TrimSpace(in.CommunityName)
	if in.CommunityID != nil {
		name = ""
	}

	input := EntradaDeEscrita{
		FullName:       strings.TrimSpace(in.FullName),
		Email:          in.Email,
		City:           strings.TrimSpace(in.City),
		Neighborhood:   strings.TrimSpace(in.Neighborhood),
		CommunityID:    in.CommunityID,
		PrivacyConsent: true,
		IsActive:       true,
	}
	if p := strings.TrimSpace(in.Phone); p != "" {
		input.Phone = &p
	}
	if name != "" {
		input.CommunityName = &name
	}

	row, err := s.repo.GravarPorEmail(ctx, input)
	if err != nil {
		return nil, err
	}
	return row, nil
}

func interpretarID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}
	return uid, nil
}

func traduzir(err error) error {
	switch {
	case errors.Is(err, ErrNaoEncontrado):
		return httpx.ErroNaoEncontrado("Dizimista não encontrado")
	case errors.Is(err, ErrEmailTaken), database.EhViolacaoDeUnicidade(err):
		return httpx.ErroConflito("Já existe um dizimista com este e-mail")
	default:
		return httpx.ErroInterno("Falha ao processar dizimista", err)
	}
}
