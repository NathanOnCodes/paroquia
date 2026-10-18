package payments

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/payments"
)

// Manipulador é a superfície HTTP dos pagamentos.
type Manipulador struct {
	svc Servico
	// MeioDePagamento é usado para verificar a assinatura do webhook.
	MeioDePagamento payments.MeioDePagamento
}

// NovoManipulador monta o manipulador de pagamentos.
func NovoManipulador(svc Servico, meio payments.MeioDePagamento) *Manipulador {
	return &Manipulador{svc: svc, MeioDePagamento: meio}
}

// CriarPagamento responde POST /api/payments/create-payment.
func (h *Manipulador) CriarPagamento(c fiber.Ctx) error {
	var req RequisicaoDeDoacao
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	if err := validarRequisicao(req); err != nil {
		return err
	}
	saida, err := h.svc.CriarPagamentoAvulso(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// CriarConfiguracao responde POST /api/payments/create-recurring.
func (h *Manipulador) CriarConfiguracao(c fiber.Ctx) error {
	var req RequisicaoDeDoacao
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	if err := validarRequisicao(req); err != nil {
		return err
	}
	saida, err := h.svc.CriarConfiguracaoDeRecorrencia(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// CriarAssinatura responde POST /api/payments/create-subscription.
func (h *Manipulador) CriarAssinatura(c fiber.Ctx) error {
	var req RequisicaoDeAssinatura
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.CriarAssinatura(c.Context(), req)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// Webhook responde POST /api/stripe/webhook.
func (h *Manipulador) Webhook(c fiber.Ctx) error {
	carga := c.Body()
	evento, err := h.MeioDePagamento.InterpretarWebhook(carga, c.Get("Stripe-Signature"))
	if err != nil {
		return httpx.ErroDeValidacao("Assinatura inválida",
			httpx.ErroDeCampo{Campo: "Stripe-Signature", Mensagem: "assinatura inválida"})
	}

	resultado, err := h.svc.ProcessarWebhook(c.Context(), evento)
	if err != nil {
		// 500 faz a Stripe reenviar o evento.
		return httpx.ErroInterno("Falha ao processar evento", err)
	}
	return httpx.ResponderOK(c, fiber.Map{
		"received": true,
		"status":   resultado.Status,
	})
}

// validarRequisicao aplica as regras transversais do formulário de doação.
func validarRequisicao(req RequisicaoDeDoacao) error {
	if !req.AceitePrivacidade {
		return httpx.ErroDeValidacao("É necessário aceitar o aviso de privacidade",
			httpx.ErroDeCampo{Campo: "privacy_consent", Mensagem: "consentimento obrigatório"})
	}
	if !req.ValidarComunidade() {
		return httpx.ErroDeValidacao("Informe a comunidade",
			httpx.ErroDeCampo{Campo: "community_id", Mensagem: "informe a comunidade ou o nome dela"})
	}
	return nil
}
