package auth

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// CredenciaisEmitidas é o par de tokens produzido por uma emissão de sessão.
// O access token é curto e vai para um cookie; o refresh token é opaco, vive
// em outro cookie e nunca aparece no corpo das respostas.
type CredenciaisEmitidas struct {
	Acesso      string
	Atualizacao string
}

// Servico é a superfície de negócio da autenticação.
type Servico interface {
	// Login valida as credenciais e emite a sessão.
	Login(ctx context.Context, req RequisicaoDeLogin) (RespostaDeSessao, CredenciaisEmitidas, error)
	// Renovar troca um refresh token válido por um par novo (rotação).
	Renovar(ctx context.Context, tokenDeAtualizacao string) (RespostaDeSessao, CredenciaisEmitidas, error)
	// Encerrar revoga a sessão atual.
	Encerrar(ctx context.Context, tokenDeAtualizacao string) error
	// ObterPerfil devolve a identidade do usuário autenticado.
	ObterPerfil(ctx context.Context, idDoUsuario string) (RespostaDeUsuario, error)
	// TrocarSenha substitui a senha do próprio usuário.
	TrocarSenha(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDeTrocaDeSenha) error
}

// Dependencias agrupa as injeções do serviço.
type Dependencias struct {
	Usuarios  RepositorioDeUsuarios
	Sessoes   RepositorioDeSessoes
	Tokens    *authx.GerenciadorDeTokens
	Auditoria audit.RegistradorDeAuditoria
	Log       *slog.Logger
	// TempoDeVidaDaAtualizacao define a validade do refresh token.
	TempoDeVidaDaAtualizacao time.Duration
	// Relogio é injetável para testes.
	Relogio func() time.Time
}

type servico struct {
	usuarios       RepositorioDeUsuarios
	sessoes        RepositorioDeSessoes
	tokens         *authx.GerenciadorDeTokens
	auditoria      audit.RegistradorDeAuditoria
	log            *slog.Logger
	ttlAtualizacao time.Duration
	agora          func() time.Time
}

// NovoServico monta o serviço de autenticação.
func NovoServico(d Dependencias) Servico {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.Relogio == nil {
		d.Relogio = time.Now
	}
	if d.TempoDeVidaDaAtualizacao <= 0 {
		d.TempoDeVidaDaAtualizacao = 30 * 24 * time.Hour
	}
	return &servico{
		usuarios:       d.Usuarios,
		sessoes:        d.Sessoes,
		tokens:         d.Tokens,
		auditoria:      d.Auditoria,
		log:            d.Log,
		ttlAtualizacao: d.TempoDeVidaDaAtualizacao,
		agora:          d.Relogio,
	}
}

func (s *servico) Login(ctx context.Context, req RequisicaoDeLogin) (RespostaDeSessao, CredenciaisEmitidas, error) {
	usuario, err := s.usuarios.ObterPorEmail(ctx, req.Email)
	if err != nil {
		// Mesmo erro para "e-mail não existe" e "senha errada": o login não
		// pode servir de oráculo para descobrir contas.
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrCredenciaisInvalidas.Error())
	}

	if !usuario.IsActive {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrUsuarioInativo.Error())
	}
	if !authx.VerificarSenha(usuario.PasswordHash, req.Senha) {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrCredenciaisInvalidas.Error())
	}

	// Hash com parâmetros antigos: aproveita o login para atualizar.
	if authx.PrecisaRegerarHash(usuario.PasswordHash) {
		if hash, herr := authx.GerarHashDeSenha(req.Senha); herr == nil {
			if err := s.usuarios.DefinirSenha(ctx, usuario.ID, hash); err != nil {
				s.log.WarnContext(ctx, "falha ao atualizar hash da senha",
					slog.String("erro", err.Error()))
			}
		}
	}

	sessao, bruto, err := s.emitirSessao(ctx, usuario.ID)
	if err != nil {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, err
	}
	if err := s.usuarios.RegistrarAcesso(ctx, usuario.ID); err != nil {
		s.log.WarnContext(ctx, "falha ao registrar acesso", slog.String("erro", err.Error()))
	}
	s.auditar(ctx, "login", usuario.ID.String(), usuario.ID.String())
	return sessao, bruto, nil
}

func (s *servico) Renovar(ctx context.Context, tokenDeAtualizacao string) (RespostaDeSessao, CredenciaisEmitidas, error) {
	if tokenDeAtualizacao == "" {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrSessaoInvalida.Error())
	}

	registro, err := s.sessoes.ObterPorHash(ctx, authx.HashDoToken(tokenDeAtualizacao))
	if err != nil {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrSessaoInvalida.Error())
	}

	agora := s.agora().UTC()

	// Reuso de um token já trocado é tratado como roubo: revoga a família
	// inteira e obriga um novo login.
	if registro.ReusedAt != nil {
		if err := s.sessoes.RevogarFamilia(ctx, registro.FamilyID); err != nil {
			s.log.ErrorContext(ctx, "falha ao revogar família de sessões",
				slog.String("erro", err.Error()))
		}
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrSessaoInvalida.Error())
	}
	if registro.RevokedAt != nil || registro.ExpiresAt.Before(agora) {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrSessaoInvalida.Error())
	}

	usuario, err := s.usuarios.ObterPorID(ctx, registro.UserID)
	if err != nil || !usuario.IsActive {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroNaoAutenticado(ErrSessaoInvalida.Error())
	}

	// Rotaciona: o token antigo deixa de valer e um novo entra na mesma
	// família, tornando a reutilização do antigo detectável.
	if err := s.sessoes.Revogar(ctx, registro.ID); err != nil {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroInterno("Falha ao renovar sessão", err)
	}

	return s.emitirSessao(ctx, usuario.ID)
}

func (s *servico) Encerrar(ctx context.Context, tokenDeAtualizacao string) error {
	if tokenDeAtualizacao == "" {
		return nil
	}
	registro, err := s.sessoes.ObterPorHash(ctx, authx.HashDoToken(tokenDeAtualizacao))
	if err != nil {
		return nil
	}
	if err := s.sessoes.RevogarFamilia(ctx, registro.FamilyID); err != nil {
		s.log.WarnContext(ctx, "falha ao encerrar sessão", slog.String("erro", err.Error()))
	}
	return nil
}

func (s *servico) ObterPerfil(ctx context.Context, idDoUsuario string) (RespostaDeUsuario, error) {
	id, err := uuid.Parse(idDoUsuario)
	if err != nil {
		return RespostaDeUsuario{}, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}
	usuario, err := s.usuarios.ObterPorID(ctx, id)
	if err != nil {
		return RespostaDeUsuario{}, traduzir(err)
	}
	return paraResposta(usuario), nil
}

func (s *servico) TrocarSenha(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDeTrocaDeSenha) error {
	if req.SenhaAtual == req.NovaSenha {
		return httpx.ErroDeValidacao("A nova senha deve ser diferente da atual",
			httpx.ErroDeCampo{Campo: "nova_senha", Mensagem: "deve ser diferente da atual"})
	}

	id, err := uuid.Parse(ator.ID)
	if err != nil {
		return httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}

	usuario, err := s.usuarios.ObterPorID(ctx, id)
	if err != nil {
		return traduzir(err)
	}
	if !authx.VerificarSenha(usuario.PasswordHash, req.SenhaAtual) {
		return httpx.ErroDeValidacao(ErrSenhaIncorreta.Error(),
			httpx.ErroDeCampo{Campo: "senha_atual", Mensagem: "incorreta"})
	}

	hash, err := authx.GerarHashDeSenha(req.NovaSenha)
	if err != nil {
		return httpx.ErroInterno("Falha ao processar nova senha", err)
	}
	if err := s.usuarios.DefinirSenha(ctx, usuario.ID, hash); err != nil {
		return httpx.ErroInterno("Falha ao alterar senha", err)
	}

	s.auditar(ctx, "change_password", usuario.ID.String(), usuario.ID.String())
	return nil
}

// emitirSessao cria o refresh token e assina o access token.
func (s *servico) emitirSessao(ctx context.Context, idDoUsuario uuid.UUID) (RespostaDeSessao, CredenciaisEmitidas, error) {
	bruto := authx.TokenAleatorio(32)

	if _, err := s.sessoes.Criar(ctx, EntradaDeSessao{
		HashDoToken: authx.HashDoToken(bruto),
		IDDoUsuario: idDoUsuario,
		IDDaFamilia: uuid.New(),
		ExpiraEm:    s.agora().UTC().Add(s.ttlAtualizacao),
	}); err != nil {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroInterno("Falha ao abrir sessão", err)
	}

	usuario, err := s.usuarios.ObterPorID(ctx, idDoUsuario)
	if err != nil {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, traduzir(err)
	}

	assinado, err := s.tokens.EmitirAcesso(authx.UsuarioAutenticado{
		ID:               usuario.ID.String(),
		Email:            usuario.Email,
		FullName:         usuario.FullName,
		Role:             string(usuario.Role),
		IsActive:         usuario.IsActive,
		DeveTrocarASenha: usuario.MustChangePassword,
	})
	if err != nil {
		return RespostaDeSessao{}, CredenciaisEmitidas{}, httpx.ErroInterno("Falha ao emitir credencial", err)
	}

	return RespostaDeSessao{Usuario: paraResposta(usuario)},
		CredenciaisEmitidas{Acesso: assinado, Atualizacao: bruto}, nil
}

// auditar registra uma ação sem nunca falhar a operação principal.
func (s *servico) auditar(ctx context.Context, acao, idDaEntidade, idDoAtor string) {
	if s.auditoria == nil {
		return
	}
	s.auditoria.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID:    idDoAtor,
		Action:     acao,
		EntityType: "usuario",
		EntityID:   idDaEntidade,
	})
}

func traduzir(err error) error {
	switch {
	case errors.Is(err, ErrNaoEncontrado):
		return httpx.ErroNaoEncontrado("Usuário não encontrado")
	case errors.Is(err, ErrCredenciaisInvalidas):
		return httpx.ErroNaoAutenticado(ErrCredenciaisInvalidas.Error())
	default:
		return httpx.ErroInterno(fmt.Sprintf("falha de autenticação: %v", err), err)
	}
}
