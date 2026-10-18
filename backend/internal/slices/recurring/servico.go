package recurring

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/config"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/mail"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/ratelimit"
)

// Servico é a superfície de negócio da recorrência sob a ótica do fiel.
type Servico interface {
	// SolicitarLink envia o link mágico. Sempre responde igual, para não
	// revelar se o e-mail está cadastrado.
	SolicitarLink(ctx context.Context, email, ip string) error
	// AbrirSessao valida o link, o consome e devolve a sessão do doador.
	AbrirSessao(ctx context.Context, token string) (string, error)
	// Listar devolve as recorrências da sessão atual.
	Listar(ctx context.Context, token string) (RespostaDeGerenciamento, error)
	// Cancelar pede o cancelamento ao fim do período.
	Cancelar(ctx context.Context, token, idAssinatura string) error
}

// Dependencias agrupa as injeções do serviço.
type Dependencias struct {
	Tokens       RepositorioDeTokens
	Recorrencias RepositorioDeRecorrencias
	Remetente    mail.RemetenteDeEmail
	Limitador    *ratelimit.Limitador
	Limites      config.Limites
	URLPublica   string
	NomeDoSite   string
	TTLDoToken   time.Duration
	Auditoria    audit.RegistradorDeAuditoria
	Log          *slog.Logger
	Relogio      func() time.Time
}

type servico struct{ Dependencias }

// NovoServico monta o serviço de recorrência do fiel.
func NovoServico(d Dependencias) Servico {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.Relogio == nil {
		d.Relogio = time.Now
	}
	if d.TTLDoToken <= 0 {
		d.TTLDoToken = 24 * time.Hour
	}
	if d.NomeDoSite == "" {
		d.NomeDoSite = "Paróquia"
	}
	return &servico{Dependencias: d}
}

func (s *servico) SolicitarLink(ctx context.Context, email, ip string) error {
	email = normalizarEmail(email)

	// Rate limit por IP e por e-mail, como no sistema legado.
	if ip != "" {
		decisao, err := s.Limitador.Permitir(ctx, ratelimit.TipoLinkDeRecorrencia,
			authx.HashDoToken(ip), s.Limites.MaximoDeRecorrenciasPorIP, s.Limites.JanelaDeRecorrencias)
		if err == nil && !decisao.Permitido {
			return httpx.ErroLimiteExcedido("Muitas solicitações. Tente novamente mais tarde.")
		}
	}
	desde := s.Relogio().Add(-s.Limites.JanelaDeRecorrencias)
	emitidos, err := s.Tokens.QuantosPorEmail(ctx, email, desde)
	if err == nil && emitidos >= s.Limites.MaximoDeRecorrenciasPorEmail {
		return httpx.ErroLimiteExcedido("Muitas solicitações. Tente novamente mais tarde.")
	}

	doador, err := s.Recorrencias.ObterDoadorPorEmail(ctx, email)
	if err != nil {
		// Anti-enumeração: e-mail desconhecido não gera erro nem e-mail.
		if errors.Is(err, ErrNaoEncontrado) {
			return nil
		}
		return httpx.ErroInterno("Falha ao consultar cadastro", err)
	}

	bruto := authx.TokenAleatorio(32)
	if _, err := s.Tokens.Criar(ctx, EntradaDeToken{
		HashDoToken: authx.HashDoToken(bruto),
		Email:       email,
		ExpiraEm:    s.Relogio().Add(s.TTLDoToken),
	}); err != nil {
		return httpx.ErroInterno("Falha ao gerar link de acesso", err)
	}

	link := fmt.Sprintf("%s/api/recurring/session?token=%s", s.URLPublica, bruto)
	horas := int(s.TTLDoToken.Hours())

	html := fmt.Sprintf(`
<h2>Gerenciar sua doação recorrente</h2>
<p>Olá, %s.</p>
<p>Recebemos uma solicitação para acessar suas contribuições recorrentes na %s.</p>
<p><a href="%s">Acessar minhas recorrências</a></p>
<p>Este link é de <strong>uso único</strong> e expira em %d hora(s).</p>
<p>Se você não pediu este acesso, pode ignorar esta mensagem.</p>`,
		mail.EscaparHTML(doador.FullName), mail.EscaparHTML(s.NomeDoSite),
		mail.EscaparHTML(link), horas)

	if err := s.Remetente.Enviar(ctx, mail.MensagemDeEmail{
		To:      email,
		Subject: "Gerenciar sua doação recorrente",
		HTML:    html,
	}); err != nil {
		s.Log.WarnContext(ctx, "falha ao enviar link de recorrência",
			slog.String("erro", err.Error()))
	}
	return nil
}

// AbrirSessao consome o link e devolve o identificador do doador, que passa a
// ser a identidade da sessão de cookie.
func (s *servico) AbrirSessao(ctx context.Context, token string) (string, error) {
	if token == "" {
		return "", ErrTokenInvalido
	}

	registro, err := s.Tokens.ObterPorHash(ctx, authx.HashDoToken(token))
	if err != nil {
		return "", ErrTokenInvalido
	}

	agora := s.Relogio().UTC()
	// O token continua valendo para a sessão que ele abriu; o que se impede é
	// abrir o link de novo.
	if registro.ConsumedAt != nil || registro.ExpiresAt.Before(agora) {
		return "", ErrTokenInvalido
	}

	doador, err := s.Recorrencias.ObterDoadorPorEmail(ctx, registro.Email)
	if err != nil {
		return "", ErrTokenInvalido
	}

	if err := s.Tokens.Consumir(ctx, registro.ID); err != nil {
		s.Log.WarnContext(ctx, "falha ao consumir token", slog.String("erro", err.Error()))
	}
	return doador.ID.String(), nil
}

func (s *servico) Listar(ctx context.Context, token string) (RespostaDeGerenciamento, error) {
	idDoador, err := s.resolverDoador(ctx, token)
	if err != nil {
		// Nunca exponha o motivo: a página mostra estado vazio.
		return RespostaDeGerenciamento{DoadorID: "", Recorrencias: []RespostaDeRecorrencia{}}, nil
	}

	linhas, err := s.Recorrencias.ListarPorDoador(ctx, idDoador)
	if err != nil {
		return RespostaDeGerenciamento{}, httpx.ErroInterno("Falha ao listar recorrências", err)
	}

	resposta := RespostaDeGerenciamento{
		DoadorID:     idDoador.String(),
		Recorrencias: make([]RespostaDeRecorrencia, 0, len(linhas)),
	}
	for _, l := range linhas {
		resposta.Recorrencias = append(resposta.Recorrencias, paraResposta(l))
	}
	return resposta, nil
}

func (s *servico) Cancelar(ctx context.Context, token, idAssinatura string) error {
	idDoador, err := s.resolverDoador(ctx, token)
	if err != nil {
		return httpx.ErroNaoAutenticado("Sessão expirada. Solicite um novo link de acesso.")
	}
	if strings.TrimSpace(idAssinatura) == "" {
		return httpx.ErroDeValidacao("Identificador da recorrência obrigatório",
			httpx.ErroDeCampo{Campo: "subscription_id", Mensagem: "campo obrigatório"})
	}

	// A recorrência precisa pertencer ao doador da sessão.
	linha, err := s.Recorrencias.ObterPorAssinaturaEDoador(ctx, idAssinatura, idDoador)
	if err != nil {
		return httpx.ErroNaoEncontrado("Recorrência não encontrada para este e-mail")
	}
	if linha.Status == "canceled" {
		return nil
	}

	if _, err := s.Recorrencias.CancelarNoFimDoPeriodo(ctx, linha.ID, s.Relogio().UTC()); err != nil {
		return httpx.ErroInterno("Falha ao cancelar a recorrência", err)
	}

	if s.Auditoria != nil {
		s.Auditoria.Registrar(ctx, audit.RegistroDeAuditoria{
			ActorID: idDoador.String(), Action: "donor_cancel_recurring",
			EntityType: "recurring_donation", EntityID: linha.ID.String(),
			Metadata: map[string]any{"source": "public"},
		})
	}
	return nil
}

// resolverDoador devolve o doador por trás do token da sessão.
func (s *servico) resolverDoador(ctx context.Context, token string) (uuid.UUID, error) {
	if token == "" {
		return uuid.Nil, ErrTokenInvalido
	}
	registro, err := s.Tokens.ObterPorHash(ctx, authx.HashDoToken(token))
	if err != nil {
		return uuid.Nil, ErrTokenInvalido
	}
	if registro.ExpiresAt.Before(s.Relogio().UTC()) {
		return uuid.Nil, ErrTokenInvalido
	}
	doador, err := s.Recorrencias.ObterDoadorPorEmail(ctx, registro.Email)
	if err != nil {
		return uuid.Nil, ErrTokenInvalido
	}
	return doador.ID, nil
}

func paraResposta(linha *ent.DoacaoRecorrente) RespostaDeRecorrencia {
	return RespostaDeRecorrencia{
		ID:                     linha.ID.String(),
		ValorEmCentavos:        linha.AmountCents,
		Moeda:                  linha.Currency,
		Intervalo:              linha.Interval,
		Status:                 string(linha.Status),
		CancelarNoFimDoPeriodo: linha.CancelAtPeriodEnd,
		IniciadaEm:             entx.DataParaTextoOpcional(linha.StartedAt),
		FimDoPeriodoAtual:      entx.DataParaTextoOpcional(linha.CurrentPeriodEnd),
		PedidoDeCancelamentoEm: entx.DataParaTextoOpcional(linha.CancellationRequestedAt),
		CanceladaEm:            entx.DataParaTextoOpcional(linha.CanceledAt),
		CriadoEm:               entx.DataParaTexto(linha.CreatedAt),
	}
}
