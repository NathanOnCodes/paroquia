package payments

import (
	"github.com/gofiber/fiber/v3"
)

// Registrar monta as rotas de pagamento.
//
//	POST /api/payments/create-payment     público, formulário de doação
//	POST /api/payments/create-recurring   público, formulário de doação
//	POST /api/payments/create-subscription público, chamado pelo navegador
//	POST /api/stripe/webhook              Stripe, com verificação de assinatura
func Registrar(api fiber.Router, h *Manipulador) {
	api.Post("/payments/create-payment", h.CriarPagamento)
	api.Post("/payments/create-recurring", h.CriarConfiguracao)
	api.Post("/payments/create-subscription", h.CriarAssinatura)
	api.Post("/stripe/webhook", h.Webhook)
}
