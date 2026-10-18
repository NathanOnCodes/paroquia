package payments

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/mail"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/payments"

	"github.com/NathanOnCodes/paroquia/backend/internal/slices/donors"
)

// ResolucaoDeDoador é o recorte do slice `donors` que este precisa. A
// dependência é injetada por interface, então `payments` não conhece a
// implementação.
type ResolucaoDeDoador interface {
	GravarAPartirDaDoacao(ctx context.Context, entrada donors.IdentidadeDaDoacao) (*ent.Doador, error)
	BuscarPorEmail(ctx context.Context, email string) (*ent.Doador, error)
}

// EntradaDeDoador reproduz o que o fluxo de doação sabe sobre o fiel.
type EntradaDeDoador struct {
	NomeCompleto   string
	Email          string
	Telefone       string
	Cidade         string
	Bairro         string
	IDComunidade   *uuid.UUID
	NomeComunidade string
}

// ConfiguracaoSite é o recorte das configurações usado aqui.
type ConfiguracaoSite interface {
	NomeDoSite(ctx context.Context) string
}

// Servico é a superfície de negócio dos pagamentos.
type Servico interface {
	// CriarPagamentoAvulso prepara uma cobrança única.
	CriarPagamentoAvulso(ctx context.Context, req RequisicaoDeDoacao) (RespostaDeIntencao, error)
	// CriarConfiguracaoDeRecorrencia prepara um método de pagamento para
	// cobranças futuras.
	CriarConfiguracaoDeRecorrencia(ctx context.Context, req RequisicaoDeDoacao) (RespostaDeConfiguracao, error)
	// CriarAssinatura efetiva a recorrência após a confirmação no navegador.
	CriarAssinatura(ctx context.Context, req RequisicaoDeAssinatura) (RespostaDeAssinatura, error)
	// ProcessarWebhook trata um evento do Stripe com idempotência.
	ProcessarWebhook(ctx context.Context, evento *payments.EventoDePagamento) (ResultadoDoWebhook, error)
}

// ResultadoDoWebhook descreve o desfecho do processamento.
type ResultadoDoWebhook struct {
	Status string `json:"status"` // "processed" | "skipped"
}

// Dependencias agrupa as injeções do serviço.
type Dependencias struct {
	Contribuicoes   RepositorioDeContribuicoes
	Recorrencias    RepositorioDeRecorrencias
	EventosWebhook  RepositorioDeEventosWebhook
	Doadores        ResolucaoDeDoador
	Site            ConfiguracaoSite
	MeioDePagamento payments.MeioDePagamento
	Remetente       mail.RemetenteDeEmail
	Auditoria       audit.RegistradorDeAuditoria
	Log             *slog.Logger
	Relogio         func() time.Time
}

type servico struct {
	Dependencias
}

// NovoServico monta o serviço de pagamentos.
func NovoServico(d Dependencias) Servico {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.Relogio == nil {
		d.Relogio = time.Now
	}
	return &servico{Dependencias: d}
}

func (s *servico) CriarPagamentoAvulso(ctx context.Context, req RequisicaoDeDoacao) (RespostaDeIntencao, error) {
	if req.Recurrence {
		return RespostaDeIntencao{}, httpx.ErroDeValidacao("Use o endpoint de recorrência",
			httpx.ErroDeCampo{Campo: "is_recurring", Mensagem: "incompatível com este endpoint"})
	}

	doador, err := s.Doadores.GravarAPartirDaDoacao(ctx, req.paraDoador())
	if err != nil {
		return RespostaDeIntencao{}, httpx.ErroInterno("Falha ao cadastrar dizimista", err)
	}

	intencao, err := s.MeioDePagamento.CriarIntencaoDePagamento(ctx, payments.EntradaDeIntencaoDePagamento{
		Amount:       req.ValorCentavos,
		Currency:     "brl",
		ReceiptEmail: req.Email,
		Metadata:     metadados(doador.ID, idComunidade(req.IDComunidade), req.ValorCentavos, "one_time"),
	})
	if err != nil {
		return RespostaDeIntencao{}, traduzirProvedor(err)
	}

	return RespostaDeIntencao{
		ClientSecret:  intencao.Secret,
		ValorCentavos: req.ValorCentavos,
	}, nil
}

func (s *servico) CriarConfiguracaoDeRecorrencia(ctx context.Context, req RequisicaoDeDoacao) (RespostaDeConfiguracao, error) {
	if !req.Recurrence {
		return RespostaDeConfiguracao{}, httpx.ErroDeValidacao("Use o endpoint de pagamento avulso",
			httpx.ErroDeCampo{Campo: "is_recurring", Mensagem: "incompatível com este endpoint"})
	}

	doador, err := s.Doadores.GravarAPartirDaDoacao(ctx, req.paraDoador())
	if err != nil {
		return RespostaDeConfiguracao{}, httpx.ErroInterno("Falha ao cadastrar dizimista", err)
	}

	// Reusa o cliente do Stripe já conhecido do doador.
	idCliente, err := s.Recorrencias.ObterUltimoCliente(ctx, doador.ID)
	if err != nil {
		return RespostaDeConfiguracao{}, httpx.ErroInterno("Falha ao consultar recorrências", err)
	}
	if idCliente == "" {
		cliente, cerr := s.MeioDePagamento.CriarCliente(ctx, payments.EntradaParaCriarCliente{
			Email: doador.Email,
			Name:  doador.FullName,
			Phone: phoneOrEmpty(doador.Phone),
			Metadata: map[string]string{
				"donor_id": doador.ID.String(),
			},
		})
		if cerr != nil {
			return RespostaDeConfiguracao{}, traduzirProvedor(cerr)
		}
		idCliente = cliente.ID
	}

	configuracao, err := s.MeioDePagamento.CriarIntencaoDeConfiguracao(ctx, payments.EntradaDeIntencaoDeConfiguracao{
		CustomerID: idCliente,
		Metadata:   metadados(doador.ID, idComunidade(req.IDComunidade), req.ValorCentavos, "recurring"),
	})
	if err != nil {
		return RespostaDeConfiguracao{}, traduzirProvedor(err)
	}

	return RespostaDeConfiguracao{
		ClientSecret:    configuracao.Secret,
		IDClienteStripe: idCliente,
		IDDoador:        doador.ID.String(),
	}, nil
}

func (s *servico) CriarAssinatura(ctx context.Context, req RequisicaoDeAssinatura) (RespostaDeAssinatura, error) {
	idDoador, err := uuid.Parse(req.IDDoador)
	if err != nil {
		return RespostaDeAssinatura{}, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "donor_id", Mensagem: "identificador inválido"})
	}

	var nomeDoProduto string
	if s.Site != nil {
		nomeDoProduto = fmt.Sprintf("%s - contribuição recorrente mensal", s.Site.NomeDoSite(ctx))
	}

	assinatura, err := s.MeioDePagamento.CriarAssinatura(ctx, payments.EntradaDeAssinatura{
		CustomerID:      req.IDClienteStripe,
		Amount:          req.ValorCentavos,
		Currency:        "brl",
		PaymentMethodID: req.IDMetodoDePagamento,
		ProductName:     nomeDoProduto,
		Metadata:        metadados(idDoador, idComunidade(req.IDComunidade), req.ValorCentavos, "recurring"),
	})
	if err != nil {
		return RespostaDeAssinatura{}, traduzirProvedor(err)
	}

	if _, err := s.Recorrencias.Criar(ctx, EntradaDeRecorrencia{
		IDDoador:           idDoador,
		IDComunidade:       idComunidade(req.IDComunidade),
		ValorEmCentavos:    int(req.ValorCentavos),
		Moeda:              "brl",
		Intervalo:          "month",
		IDClienteStripe:    req.IDClienteStripe,
		IDAssinaturaStripe: assinatura.ID,
		Status:             "incomplete",
		IniciadaEm:         timePtr(assinatura.CurrentPeriodStart),
		FimDoPeriodoAtual:  timePtr(assinatura.CurrentPeriodEnd),
	}); err != nil {
		return RespostaDeAssinatura{}, httpx.ErroInterno("Falha ao registrar recorrência", err)
	}

	return RespostaDeAssinatura{IDAssinatura: assinatura.ID}, nil
}

func (s *servico) ProcessarWebhook(ctx context.Context, evento *payments.EventoDePagamento) (ResultadoDoWebhook, error) {
	// A idempotência real é o índice único em `stripe_event_id`; este
	// pré-checamento só evita trabalho repetido.
	jaTratado, err := s.EventosWebhook.Existe(ctx, evento.ID)
	if err != nil {
		s.Log.WarnContext(ctx, "falha ao consultar eventos anteriores",
			slog.String("erro", err.Error()))
	} else if jaTratado {
		return ResultadoDoWebhook{Status: "skipped"}, nil
	}

	erroDeProcessamento := s.tratarEvento(ctx, evento)

	status := "processed"
	var mensagem *string
	if erroDeProcessamento != nil {
		status = "failed"
		texto := erroDeProcessamento.Error()
		mensagem = &texto
	}

	if err := s.EventosWebhook.Registrar(ctx, EntradaDeEventoWebhook{
		IDDoEvento:     evento.ID,
		TipoDoEvento:   evento.Type,
		HashDaCarga:    authx.HashDoToken(evento.ID),
		Status:         status,
		MensagemDeErro: mensagem,
	}); err != nil {
		s.Log.ErrorContext(ctx, "falha ao registrar evento do webhook",
			slog.String("erro", err.Error()))
	}

	if erroDeProcessamento != nil {
		return ResultadoDoWebhook{}, erroDeProcessamento
	}
	return ResultadoDoWebhook{Status: "processed"}, nil
}

// traduzirProvedor converte falhas do gateway em erros HTTP seguros.
func traduzirProvedor(err error) error {
	var erroDoProvedor *payments.ErroDoProvedor
	if errors.As(err, &erroDoProvedor) {
		return httpx.ErroInterno("Falha no meio de pagamento", err)
	}
	if errors.Is(err, payments.ErrNaoConfigurado) {
		return httpx.ErroGenerico(http.StatusServiceUnavailable, "upstream_error",
			"Meio de pagamento indisponível no momento")
	}
	return httpx.ErroInterno("Falha no meio de pagamento", err)
}

func metadados(idDoador uuid.UUID, idComunidade *uuid.UUID, valor int64, tipo string) map[string]string {
	m := map[string]string{
		"donor_id":     idDoador.String(),
		"amount_cents": fmt.Sprintf("%d", valor),
		"payment_type": tipo,
	}
	if idComunidade != nil {
		m["community_id"] = idComunidade.String()
	}
	return m
}

// idComunidade converte o identificador textual em ponteiro, ou nil.
func idComunidade(id string) *uuid.UUID {
	if id == "" {
		return nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	return &parsed
}

func timePtr(t time.Time) *time.Time {
	if t.IsZero() {
		return nil
	}
	return &t
}

func phoneOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
