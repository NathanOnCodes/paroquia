package recurring

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Manipulador é a superfície HTTP da recorrência do fiel.
type Manipulador struct {
	svc          Servico
	nomeDoCookie string
}

// NovoManipulador monta o manipulador de recorrência.
func NovoManipulador(svc Servico, nomeDoCookie string) *Manipulador {
	if nomeDoCookie == "" {
		nomeDoCookie = "recurring_mgmt"
	}
	return &Manipulador{svc: svc, nomeDoCookie: nomeDoCookie}
}

// SolicitarLink responde POST /api/recurring/request-link.
func (h *Manipulador) SolicitarLink(c fiber.Ctx) error {
	var req RequisicaoDeLink
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	// A resposta é sempre a mesma, mesmo quando o e-mail não existe.
	if err := h.svc.SolicitarLink(c.Context(), req.Email, IPDoCliente(c)); err != nil {
		return err
	}
	return httpx.ResponderOK(c, RespostaDeLink{OK: true, Mensagem: MensagemDeLinkPadrao})
}

// AbrirSessao responde GET /api/recurring/session?token=...
func (h *Manipulador) AbrirSessao(c fiber.Ctx) error {
	base := c.Query("next", "/doacoes/gerenciar")

	token := c.Query("token")
	if token == "" {
		return c.Redirect().To(base)
	}

	sessao, err := h.svc.AbrirSessao(c.Context(), token)
	if err != nil {
		return c.Redirect().To("/?error=invalid")
	}

	// Cookie httpOnly de 1 hora: o token vale 24h, mas a sessão do navegador é
	// curta, como no sistema legado.
	c.Cookie(&fiber.Cookie{
		Name:     h.nomeDoCookie,
		Value:    token,
		Path:     "/",
		HTTPOnly: true,
		SameSite: "Lax",
		MaxAge:   3600,
	})
	_ = sessao
	return c.Redirect().To(base)
}

// Listar responde GET /api/recurring/gerenciar.
func (h *Manipulador) Listar(c fiber.Ctx) error {
	saida, err := h.svc.Listar(c.Context(), c.Cookies(h.nomeDoCookie))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// Cancelar responde POST /api/recurring/cancel.
func (h *Manipulador) Cancelar(c fiber.Ctx) error {
	var req RequisicaoDeCancelamento
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	if err := h.svc.Cancelar(c.Context(), c.Cookies(h.nomeDoCookie), req.IDAssinatura); err != nil {
		return err
	}
	return httpx.ResponderConfirmado(c)
}

// IPDoCliente extrai o endereço real atrás de proxies.
func IPDoCliente(c fiber.Ctx) string {
	if encaminhado := c.Get("X-Forwarded-For"); encaminhado != "" {
		for i := 0; i < len(encaminhado); i++ {
			if encaminhado[i] == ',' {
				return encaminhado[:i]
			}
		}
		return encaminhado
	}
	return c.IP()
}
