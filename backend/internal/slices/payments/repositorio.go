// Package payments expõe a criação de contribuições (avulsas e recorrentes) e
// o webhook do Stripe, que é a única fonte de verdade do livro-razão.
package payments

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/doacaorecorrente"
	"github.com/NathanOnCodes/paroquia/backend/ent/eventowebhookstripe"
	"github.com/NathanOnCodes/paroquia/backend/ent/pagamentocontribuicao"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// Erros de domínio.
var (
	ErrNaoEncontrado = errors.New("registro não encontrado")
	ErrJaProcessado  = errors.New("evento já processado")
)

// EntradaDeContribuicao descreve uma linha do livro-razão.
type EntradaDeContribuicao struct {
	IDDoador           uuid.UUID
	IDComunidade       *uuid.UUID
	IDDoacaoRecorrente *uuid.UUID
	NomeComunidade     *string
	ValorEmCentavos    int
	Moeda              string
	TipoDePagamento    string
	Status             string
	IDIntencaoStripe   *string
	IDFaturaStripe     *string
	IDAssinaturaStripe *string
	IDEventoStripe     *string
	PagoEm             *time.Time
	EstornadoEm        *time.Time
}

// RepositorioDeContribuicoes é o contrato do livro-razão.
type RepositorioDeContribuicoes interface {
	// Gravar insere ou atualiza a linha identificada pelo intento/fatura.
	Gravar(ctx context.Context, entrada EntradaDeContribuicao) (*ent.PagamentoContribuicao, error)
	ObterPorIntencao(ctx context.Context, id string) (*ent.PagamentoContribuicao, error)
	ObterPorFatura(ctx context.Context, id string) (*ent.PagamentoContribuicao, error)
}

// EntradaDeRecorrencia descreve uma recorrência local.
type EntradaDeRecorrencia struct {
	IDDoador           uuid.UUID
	IDComunidade       *uuid.UUID
	ValorEmCentavos    int
	Moeda              string
	Intervalo          string
	IDClienteStripe    string
	IDAssinaturaStripe string
	Status             string
	IniciadaEm         *time.Time
	FimDoPeriodoAtual  *time.Time
}

// RepositorioDeRecorrencias é o contrato das recorrências.
type RepositorioDeRecorrencias interface {
	Criar(ctx context.Context, entrada EntradaDeRecorrencia) (*ent.DoacaoRecorrente, error)
	ObterPorAssinatura(ctx context.Context, id string) (*ent.DoacaoRecorrente, error)
	ObterPorDoador(ctx context.Context, idDoador uuid.UUID) ([]*ent.DoacaoRecorrente, error)
	ObterUltimoCliente(ctx context.Context, idDoador uuid.UUID) (string, error)
	AtualizarStatus(ctx context.Context, id uuid.UUID, status string,
		cancelarNoFim bool, canceladaEm, fimDoPeriodo *time.Time) error
}

// EntradaDeEventoWebhook registra o resultado do processamento.
type EntradaDeEventoWebhook struct {
	IDDoEvento     string
	TipoDoEvento   string
	HashDaCarga    string
	Status         string
	MensagemDeErro *string
}

// RepositorioDeEventosWebhook é o livro de idempotência do webhook.
type RepositorioDeEventosWebhook interface {
	// Existe informa se o evento já foi tratado.
	Existe(ctx context.Context, idDoEvento string) (bool, error)
	// Registrar grava o resultado com `ON CONFLICT DO NOTHING`, que é a
	// garantia real de idempotência.
	Registrar(ctx context.Context, entrada EntradaDeEventoWebhook) error
}

// ---------------------------------------------------------------------------
// Implementação Ent
// ---------------------------------------------------------------------------

type repositorioDeContribuicoes struct{ client *ent.Client }

// NovoRepositorioDeContribuicoes monta o repositório Ent do livro-razão.
func NovoRepositorioDeContribuicoes(client *ent.Client) RepositorioDeContribuicoes {
	return &repositorioDeContribuicoes{client: client}
}

func (r *repositorioDeContribuicoes) ObterPorIntencao(ctx context.Context, id string) (*ent.PagamentoContribuicao, error) {
	linha, err := r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.StripePaymentIntentID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeContribuicoes) ObterPorFatura(ctx context.Context, id string) (*ent.PagamentoContribuicao, error) {
	linha, err := r.client.PagamentoContribuicao.Query().
		Where(pagamentocontribuicao.StripeInvoiceID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeContribuicoes) Gravar(ctx context.Context, entrada EntradaDeContribuicao) (*ent.PagamentoContribuicao, error) {
	// O índice único parcial em `stripe_payment_intent_id`/`stripe_invoice_id`
	// é a garantia final; aqui fazemos lookup e depois escrita, e uma corrida
	// resolvida pelo ON CONFLICT abaixo.
	existente, err := r.localizar(ctx, entrada)
	if err != nil && !errors.Is(err, ErrNaoEncontrado) {
		return nil, err
	}

	if existente != nil {
		return r.atualizar(ctx, existente, entrada)
	}

	criar := r.client.PagamentoContribuicao.Create().
		SetDonorID(entrada.IDDoador).
		SetAmountCents(entrada.ValorEmCentavos).
		SetCurrency(entrada.Moeda).
		SetPaymentType(pagamentocontribuicao.PaymentType(entrada.TipoDePagamento)).
		SetStatus(pagamentocontribuicao.Status(entrada.Status))

	criar.SetNillableCommunityID(entrada.IDComunidade)
	criar.SetNillableRecurringDonationID(entrada.IDDoacaoRecorrente)
	criar.SetNillableCommunityName(entrada.NomeComunidade)
	criar.SetNillableStripePaymentIntentID(entrada.IDIntencaoStripe)
	criar.SetNillableStripeInvoiceID(entrada.IDFaturaStripe)
	criar.SetNillableStripeSubscriptionID(entrada.IDAssinaturaStripe)
	criar.SetNillableStripeEventID(entrada.IDEventoStripe)
	criar.SetNillablePaidAt(entrada.PagoEm)
	criar.SetNillableRefundedAt(entrada.EstornadoEm)

	linha, err := criar.Save(ctx)
	if err != nil {
		// Corrida perdida contra o índice único: devolve a linha vencedora.
		if database.EhViolacaoDeUnicidade(err) {
			if outra, lerr := r.localizar(ctx, entrada); lerr == nil && outra != nil {
				return r.atualizar(ctx, outra, entrada)
			}
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeContribuicoes) localizar(ctx context.Context, entrada EntradaDeContribuicao) (*ent.PagamentoContribuicao, error) {
	if entrada.IDIntencaoStripe != nil && *entrada.IDIntencaoStripe != "" {
		return r.ObterPorIntencao(ctx, *entrada.IDIntencaoStripe)
	}
	if entrada.IDFaturaStripe != nil && *entrada.IDFaturaStripe != "" {
		return r.ObterPorFatura(ctx, *entrada.IDFaturaStripe)
	}
	return nil, ErrNaoEncontrado
}

func (r *repositorioDeContribuicoes) atualizar(ctx context.Context,
	existente *ent.PagamentoContribuicao, entrada EntradaDeContribuicao) (*ent.PagamentoContribuicao, error) {

	atualizar := r.client.PagamentoContribuicao.UpdateOneID(existente.ID).
		SetAmountCents(entrada.ValorEmCentavos).
		SetStatus(pagamentocontribuicao.Status(entrada.Status))

	atualizar.SetNillableCommunityID(entrada.IDComunidade)
	atualizar.SetNillableRecurringDonationID(entrada.IDDoacaoRecorrente)
	atualizar.SetNillableCommunityName(entrada.NomeComunidade)
	atualizar.SetNillableStripeSubscriptionID(entrada.IDAssinaturaStripe)
	atualizar.SetNillableStripeEventID(entrada.IDEventoStripe)
	atualizar.SetNillablePaidAt(entrada.PagoEm)
	atualizar.SetNillableRefundedAt(entrada.EstornadoEm)

	return atualizar.Save(ctx)
}

type repositorioDeRecorrencias struct{ client *ent.Client }

// NovoRepositorioDeRecorrencias monta o repositório Ent das recorrências.
func NovoRepositorioDeRecorrencias(client *ent.Client) RepositorioDeRecorrencias {
	return &repositorioDeRecorrencias{client: client}
}

func (r *repositorioDeRecorrencias) Criar(ctx context.Context, entrada EntradaDeRecorrencia) (*ent.DoacaoRecorrente, error) {
	criar := r.client.DoacaoRecorrente.Create().
		SetDonorID(entrada.IDDoador).
		SetAmountCents(entrada.ValorEmCentavos).
		SetCurrency(entrada.Moeda).
		SetInterval(entrada.Intervalo).
		SetStripeCustomerID(entrada.IDClienteStripe).
		SetStripeSubscriptionID(entrada.IDAssinaturaStripe).
		SetStatus(doacaorecorrente.Status(entrada.Status))

	criar.SetNillableCommunityID(entrada.IDComunidade)
	criar.SetNillableStartedAt(entrada.IniciadaEm)
	criar.SetNillableCurrentPeriodEnd(entrada.FimDoPeriodoAtual)

	return criar.Save(ctx)
}

func (r *repositorioDeRecorrencias) ObterPorAssinatura(ctx context.Context, id string) (*ent.DoacaoRecorrente, error) {
	linha, err := r.client.DoacaoRecorrente.Query().
		Where(doacaorecorrente.StripeSubscriptionID(id)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeRecorrencias) ObterPorDoador(ctx context.Context, idDoador uuid.UUID) ([]*ent.DoacaoRecorrente, error) {
	return r.client.DoacaoRecorrente.Query().
		Where(doacaorecorrente.DonorID(idDoador)).
		Order(ent.Desc(doacaorecorrente.FieldCreatedAt)).
		All(ctx)
}

func (r *repositorioDeRecorrencias) ObterUltimoCliente(ctx context.Context, idDoador uuid.UUID) (string, error) {
	linha, err := r.client.DoacaoRecorrente.Query().
		Where(doacaorecorrente.DonorID(idDoador)).
		Order(ent.Desc(doacaorecorrente.FieldCreatedAt)).
		First(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", err
	}
	return linha.StripeCustomerID, nil
}

func (r *repositorioDeRecorrencias) AtualizarStatus(ctx context.Context, id uuid.UUID, status string,
	cancelarNoFim bool, canceladaEm, fimDoPeriodo *time.Time) error {

	atualizar := r.client.DoacaoRecorrente.UpdateOneID(id).
		SetStatus(doacaorecorrente.Status(status)).
		SetCancelAtPeriodEnd(cancelarNoFim)

	atualizar.SetNillableCanceledAt(canceladaEm)
	atualizar.SetNillableCurrentPeriodEnd(fimDoPeriodo)

	return atualizar.Exec(ctx)
}

type repositorioDeEventosWebhook struct{ client *ent.Client }

// NovoRepositorioDeEventosWebhook monta o livro de idempotência do webhook.
func NovoRepositorioDeEventosWebhook(client *ent.Client) RepositorioDeEventosWebhook {
	return &repositorioDeEventosWebhook{client: client}
}

func (r *repositorioDeEventosWebhook) Existe(ctx context.Context, idDoEvento string) (bool, error) {
	return r.client.EventoWebhookStripe.Query().
		Where(eventowebhookstripe.StripeEventID(idDoEvento)).
		Exist(ctx)
}

func (r *repositorioDeEventosWebhook) Registrar(ctx context.Context, entrada EntradaDeEventoWebhook) error {
	criar := r.client.EventoWebhookStripe.Create().
		SetStripeEventID(entrada.IDDoEvento).
		SetEventType(entrada.TipoDoEvento).
		SetPayloadHash(entrada.HashDaCarga).
		SetProcessingStatus(eventowebhookstripe.ProcessingStatus(entrada.Status))
	criar.SetNillableErrorMessage(entrada.MensagemDeErro)

	// ON CONFLICT DO NOTHING: se outro processo já registrou o evento, esta
	// chamada é silenciosa e correta.
	return criar.
		OnConflictColumns("stripe_event_id").
		DoNothing().
		Exec(ctx)
}

func normalizar(s string) string { return strings.TrimSpace(s) }
