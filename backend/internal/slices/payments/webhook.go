package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/payments"
)

// tratarEvento desencaminha um evento do Stripe. Cada handler é idempotente:
// rodar duas vezes produz o mesmo estado.
func (s *servico) tratarEvento(ctx context.Context, evento *payments.EventoDePagamento) error {
	switch evento.Type {
	case "payment_intent.succeeded", "payment_intent.payment_failed", "payment_intent.canceled":
		return s.tratarIntencao(ctx, evento)

	case "charge.refunded":
		return s.tratarEstorno(ctx, evento)

	case "invoice.payment_succeeded", "invoice.payment_failed":
		return s.tratarFatura(ctx, evento)

	case "customer.subscription.created", "customer.subscription.updated":
		return s.sincronizarAssinatura(ctx, evento)

	case "customer.subscription.deleted":
		return s.tratarAssinaturaExcluida(ctx, evento)

	default:
		// Eventos não tratados são registrados como processados: é o
		// comportamento do sistema legado.
		return nil
	}
}

func (s *servico) tratarIntencao(ctx context.Context, evento *payments.EventoDePagamento) error {
	var pi struct {
		ID       string            `json:"id"`
		Amount   int64             `json:"amount"`
		Currency string            `json:"currency"`
		Metadata map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(evento.Data, &pi); err != nil {
		return fmt.Errorf("decodificando payment_intent: %w", err)
	}

	idDoador := pi.Metadata["donor_id"]
	if idDoador == "" {
		// Sem doador não há nada a lançar no livro-razão.
		return nil
	}
	uid, err := uuid.Parse(idDoador)
	if err != nil {
		return fmt.Errorf("donor_id inválido: %w", err)
	}

	status := "paid"
	pagoEm := timePtr(s.Relogio().UTC())
	switch evento.Type {
	case "payment_intent.payment_failed":
		status, pagoEm = "failed", nil
	case "payment_intent.canceled":
		status, pagoEm = "canceled", nil
	}

	idIntencao := pi.ID
	_, err = s.Contribuicoes.Gravar(ctx, EntradaDeContribuicao{
		IDDoador:         uid,
		IDComunidade:     idComunidadeDe(pi.Metadata),
		NomeComunidade:   nomeComunidadeDe(pi.Metadata),
		ValorEmCentavos:  int(pi.Amount),
		Moeda:            defaultIfEmpty(pi.Currency, "brl"),
		TipoDePagamento:  defaultIfEmpty(pi.Metadata["payment_type"], "one_time"),
		Status:           status,
		IDIntencaoStripe: &idIntencao,
		IDEventoStripe:   &idIntencao,
		PagoEm:           pagoEm,
	})
	return err
}

func (s *servico) tratarEstorno(ctx context.Context, evento *payments.EventoDePagamento) error {
	var charge struct {
		PaymentIntent string `json:"payment_intent"`
		Refunded      bool   `json:"refunded"`
	}
	if err := json.Unmarshal(evento.Data, &charge); err != nil {
		return fmt.Errorf("decodificando charge: %w", err)
	}
	if charge.PaymentIntent == "" {
		return nil
	}

	existente, err := s.Contribuicoes.ObterPorIntencao(ctx, charge.PaymentIntent)
	if err != nil {
		// Sem contribuição correspondente não há o que estornar.
		return nil
	}

	estornado := s.Relogio().UTC()
	_, err = s.Contribuicoes.Gravar(ctx, EntradaDeContribuicao{
		IDDoador:           existente.DonorID,
		IDComunidade:       existente.CommunityID,
		IDDoacaoRecorrente: existente.RecurringDonationID,
		NomeComunidade:     existente.CommunityName,
		ValorEmCentavos:    existente.AmountCents,
		Moeda:              existente.Currency,
		TipoDePagamento:    string(existente.PaymentType),
		Status:             "refunded",
		IDIntencaoStripe:   existente.StripePaymentIntentID,
		EstornadoEm:        &estornado,
	})
	return err
}

func (s *servico) tratarFatura(ctx context.Context, evento *payments.EventoDePagamento) error {
	var fatura struct {
		ID         string `json:"id"`
		Status     string `json:"status"`
		AmountDue  int64  `json:"amount_due"`
		AmountPaid int64  `json:"amount_paid"`
		Currency   string `json:"currency"`
		Parent     struct {
			SubscriptionDetails struct {
				Subscription string            `json:"subscription"`
				Metadata     map[string]string `json:"metadata"`
			} `json:"subscription_details"`
		} `json:"parent"`
	}
	if err := json.Unmarshal(evento.Data, &fatura); err != nil {
		return fmt.Errorf("decodificando invoice: %w", err)
	}

	idAssinatura := fatura.Parent.SubscriptionDetails.Subscription
	var idDoador uuid.UUID
	var idComunidade *uuid.UUID
	var nomeComunidade *string
	var idRecorrencia *uuid.UUID

	if idAssinatura != "" {
		if recorrente, err := s.Recorrencias.ObterPorAssinatura(ctx, idAssinatura); err == nil && recorrente != nil {
			idDoador = recorrente.DonorID
			idComunidade = recorrente.CommunityID
			idRecorrencia = &recorrente.ID
			if doador, derr := s.Doadores.BuscarPorEmail(ctx, ""); derr == nil && doador != nil {
				nomeComunidade = doador.CommunityName
			}
		}
	}

	if idDoador == uuid.Nil {
		meta := fatura.Parent.SubscriptionDetails.Metadata
		if meta["donor_id"] == "" {
			return nil
		}
		parsed, err := uuid.Parse(meta["donor_id"])
		if err != nil {
			return fmt.Errorf("donor_id inválido na fatura: %w", err)
		}
		idDoador = parsed
		idComunidade = idComunidadeDe(meta)
		nomeComunidade = nomeComunidadeDe(meta)
	}

	valor := fatura.AmountPaid
	if evento.Type == "invoice.payment_failed" {
		valor = fatura.AmountDue
	}

	idFatura := fatura.ID
	idEvento := fatura.ID
	_, err := s.Contribuicoes.Gravar(ctx, EntradaDeContribuicao{
		IDDoador:           idDoador,
		IDComunidade:       idComunidade,
		IDDoacaoRecorrente: idRecorrencia,
		NomeComunidade:     nomeComunidade,
		ValorEmCentavos:    int(valor),
		Moeda:              defaultIfEmpty(fatura.Currency, "brl"),
		TipoDePagamento:    "recurring",
		Status:             "paid",
		IDFaturaStripe:     &idFatura,
		IDAssinaturaStripe: &idAssinatura,
		IDEventoStripe:     &idEvento,
		PagoEm:             timePtr(s.Relogio().UTC()),
	})
	if err != nil {
		return err
	}

	if idRecorrencia == nil {
		return nil
	}

	status := "active"
	if evento.Type == "invoice.payment_failed" {
		status = "past_due"
	}
	return s.Recorrencias.AtualizarStatus(ctx, *idRecorrencia, status, false, nil, nil)
}

func (s *servico) sincronizarAssinatura(ctx context.Context, evento *payments.EventoDePagamento) error {
	var assinatura struct {
		ID                string `json:"id"`
		Status            string `json:"status"`
		CancelAtPeriodEnd bool   `json:"cancel_at_period_end"`
		CanceledAt        *int64 `json:"canceled_at"`
		Items             struct {
			Data []struct {
				CurrentPeriodEnd int64 `json:"current_period_end"`
			} `json:"data"`
		} `json:"items"`
	}
	if err := json.Unmarshal(evento.Data, &assinatura); err != nil {
		return fmt.Errorf("decodificando subscription: %w", err)
	}

	local, err := s.Recorrencias.ObterPorAssinatura(ctx, assinatura.ID)
	if err != nil {
		// Assinatura desconhecida: nada a sincronizar.
		return nil
	}

	status := mapearStatus(assinatura.Status)
	var canceladaEm, fimDoPeriodo *time.Time
	if assinatura.CanceledAt != nil && *assinatura.CanceledAt > 0 {
		canceladaEm = timePtr(time.Unix(*assinatura.CanceledAt, 0).UTC())
	}
	if len(assinatura.Items.Data) > 0 && assinatura.Items.Data[0].CurrentPeriodEnd > 0 {
		fimDoPeriodo = timePtr(time.Unix(assinatura.Items.Data[0].CurrentPeriodEnd, 0).UTC())
	}

	return s.Recorrencias.AtualizarStatus(ctx, local.ID, status,
		assinatura.CancelAtPeriodEnd, canceladaEm, fimDoPeriodo)
}

func (s *servico) tratarAssinaturaExcluida(ctx context.Context, evento *payments.EventoDePagamento) error {
	var assinatura struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal(evento.Data, &assinatura); err != nil {
		return fmt.Errorf("decodificando subscription: %w", err)
	}
	local, err := s.Recorrencias.ObterPorAssinatura(ctx, assinatura.ID)
	if err != nil {
		return nil
	}
	agora := s.Relogio().UTC()
	return s.Recorrencias.AtualizarStatus(ctx, local.ID, "canceled", false, &agora, nil)
}

// mapearStatus traduz o ciclo de vida da Stripe para o nosso.
func mapearStatus(status string) string {
	switch status {
	case "canceled":
		return "canceled"
	case "cancel_at_period_end":
		return "cancel_at_period_end"
	case "past_due", "incomplete", "incomplete_expired", "unpaid":
		return "past_due"
	default:
		return "active"
	}
}

func idComunidadeDe(meta map[string]string) *uuid.UUID {
	if meta["community_id"] == "" {
		return nil
	}
	id, err := uuid.Parse(meta["community_id"])
	if err != nil {
		return nil
	}
	return &id
}

func nomeComunidadeDe(meta map[string]string) *string {
	if meta["community_name"] == "" {
		return nil
	}
	n := meta["community_name"]
	return &n
}

func defaultIfEmpty(valor, padrao string) string {
	if valor == "" {
		return padrao
	}
	return valor
}
