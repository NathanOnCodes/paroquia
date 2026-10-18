package users

import (
	"context"
	"errors"
	"log/slog"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Servico é a superfície de negócio das contas administrativas.
type Servico interface {
	Listar(ctx context.Context) ([]Resposta, error)
	Obter(ctx context.Context, id string) (Resposta, error)
	Criar(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDeEscrita) (Resposta, error)
	Atualizar(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita) (Resposta, error)
	DefinirStatus(ctx context.Context, ator authx.UsuarioAutenticado, id string, ativo bool) (Resposta, error)
	RedefinirSenha(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDeSenha) error
}

// Dependencias agrupa as injeções do serviço.
type Dependencias struct {
	Repositorio Repositorio
	Auditoria   audit.RegistradorDeAuditoria
	Log         *slog.Logger
}

type servico struct {
	repo      Repositorio
	auditoria audit.RegistradorDeAuditoria
	log       *slog.Logger
}

// NovoServico monta o serviço de usuários administrativos.
func NovoServico(d Dependencias) Servico {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	return &servico{repo: d.Repositorio, auditoria: d.Auditoria, log: d.Log}
}

func (s *servico) Listar(ctx context.Context) ([]Resposta, error) {
	linhas, err := s.repo.Listar(ctx)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar usuários", err)
	}
	return paraRespostas(linhas), nil
}

func (s *servico) Obter(ctx context.Context, id string) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}
	linha, err := s.repo.Obter(ctx, uid)
	if err != nil {
		return Resposta{}, traduzir(err)
	}
	return paraResposta(linha), nil
}

func (s *servico) Criar(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDeEscrita) (Resposta, error) {
	if req.Senha == "" {
		return Resposta{}, httpx.ErroDeValidacao("Senha obrigatória na criação",
			httpx.ErroDeCampo{Campo: "password", Mensagem: "campo obrigatório"})
	}

	if existente, err := s.repo.ObterPorEmail(ctx, req.Email); err == nil && existente != nil {
		return Resposta{}, httpx.ErroConflito("Já existe um usuário com este e-mail")
	}

	hash, err := authx.GerarHashDeSenha(req.Senha)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao processar senha", err)
	}

	linha, err := s.repo.Criar(ctx, EntradaDeEscrita{
		Email:             req.Email,
		NomeCompleto:      req.NomeCompleto,
		HashSenha:         hash,
		Papel:             req.Papel,
		Ativo:             req.Ativo,
		ExigeTrocaDeSenha: true,
	})
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	s.auditar(ctx, "create_staff_user", linha.ID.String(), ator.ID)
	return paraResposta(linha), nil
}

func (s *servico) Atualizar(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}

	// Regra legada: ninguém desativa a própria conta por aqui.
	if uid.String() == ator.ID && !req.Ativo {
		return Resposta{}, httpx.ErroDeValidacao(ErrAutoDesativacao.Error(),
			httpx.ErroDeCampo{Campo: "is_active", Mensagem: "você não pode desativar a própria conta"})
	}

	atual, err := s.repo.Obter(ctx, uid)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	// Não deixa a paróquia sem administrador.
	if atual.Role == "admin" && req.Papel != "admin" {
		if err := s.garantirOutroAdmin(ctx); err != nil {
			return Resposta{}, err
		}
	}

	linha, err := s.repo.Atualizar(ctx, uid, EntradaDeEscrita{
		Email:        req.Email,
		NomeCompleto: req.NomeCompleto,
		Papel:        req.Papel,
		Ativo:        req.Ativo,
	})
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	s.auditar(ctx, "update_staff_user", linha.ID.String(), ator.ID)
	return paraResposta(linha), nil
}

func (s *servico) DefinirStatus(ctx context.Context, ator authx.UsuarioAutenticado, id string, ativo bool) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}
	if uid.String() == ator.ID && !ativo {
		return Resposta{}, httpx.ErroDeValidacao(ErrAutoDesativacao.Error(),
			httpx.ErroDeCampo{Campo: "is_active", Mensagem: "você não pode desativar a própria conta"})
	}

	atual, err := s.repo.Obter(ctx, uid)
	if err != nil {
		return Resposta{}, traduzir(err)
	}
	if atual.Role == "admin" && !ativo {
		if err := s.garantirOutroAdmin(ctx); err != nil {
			return Resposta{}, err
		}
	}

	linha, err := s.repo.DefinirAtivo(ctx, uid, ativo)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	acao := "activate_staff_user"
	if !ativo {
		acao = "deactivate_staff_user"
	}
	s.auditar(ctx, acao, linha.ID.String(), ator.ID)
	return paraResposta(linha), nil
}

func (s *servico) RedefinirSenha(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDeSenha) error {
	uid, err := interpretarID(id)
	if err != nil {
		return err
	}
	hash, err := authx.GerarHashDeSenha(req.NovaSenha)
	if err != nil {
		return httpx.ErroInterno("Falha ao processar senha", err)
	}
	if err := s.repo.DefinirSenha(ctx, uid, hash); err != nil {
		return traduzir(err)
	}
	s.auditar(ctx, "reset_staff_password", uid.String(), ator.ID)
	return nil
}

// garantirOutroAdmin recusa a operação se ela deixaria a paróquia sem
// administrador ativo.
func (s *servico) garantirOutroAdmin(ctx context.Context) error {
	admins, err := s.repo.QuantidadeDeAdmins(ctx)
	if err != nil {
		return httpx.ErroInterno("Falha ao verificar administradores", err)
	}
	if admins <= 1 {
		return httpx.ErroConflito(ErrUltimoAdmin.Error())
	}
	return nil
}

func (s *servico) auditar(ctx context.Context, acao, idDaEntidade, idDoAtor string) {
	if s.auditoria == nil {
		return
	}
	s.auditoria.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID:    idDoAtor,
		Action:     acao,
		EntityType: "user",
		EntityID:   idDaEntidade,
	})
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
		return httpx.ErroNaoEncontrado("Usuário não encontrado")
	case errors.Is(err, ErrEmailEmUso), database.EhViolacaoDeUnicidade(err):
		return httpx.ErroConflito("Já existe um usuário com este e-mail")
	default:
		return httpx.ErroInterno("Falha ao processar usuário", err)
	}
}
