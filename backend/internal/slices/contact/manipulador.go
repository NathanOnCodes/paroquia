package contact

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Handler is the HTTP surface of the contact slice.
type Manipulador struct {
	svc Servico
}

// NewHandler builds the contact handler.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// Submit answers POST /api/contact.
func (h *Manipulador) Enviar(c fiber.Ctx) error {
	var req RequisicaoDeCriacao
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	if err := h.svc.Enviar(c.Context(), req, IPDoCliente(c)); err != nil {
		return err
	}
	return httpx.ResponderOK(c, Aceito{OK: true})
}

// ListAdmin answers GET /api/admin/contacts.
func (h *Manipulador) ListarAdmin(c fiber.Ctx) error {
	out, err := h.svc.ListarAdmin(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}

// SetStatus answers PATCH /api/admin/contacts/:id/status.
func (h *Manipulador) DefinirStatus(c fiber.Ctx) error {
	var req RequisicaoDeStatus
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	out, err := h.svc.DefinirStatus(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req.Status)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}

// ClientIP extracts the caller address, honouring the standard proxy headers.
// It is shared with the other public endpoints that need rate limiting.
func IPDoCliente(c fiber.Ctx) string {
	if fwd := c.Get("X-Forwarded-For"); fwd != "" {
		if i := strings.IndexByte(fwd, ','); i >= 0 {
			return strings.TrimSpace(fwd[:i])
		}
		return strings.TrimSpace(fwd)
	}
	return c.IP()
}
