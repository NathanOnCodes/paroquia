package settings

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Handler is the HTTP surface of the settings slice. It does nothing but
// translate between the wire format and the service.
type Manipulador struct {
	svc Servico
}

// NewHandler builds the settings handler.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// GetPublic answers GET /api/public/settings.
func (h *Manipulador) ObterPublico(c fiber.Ctx) error {
	return httpx.ResponderOK(c, h.svc.ObterPublico(c.Context()))
}

// GetAdmin answers GET /api/admin/settings. It returns the same payload as the
// public endpoint today; it exists so the admin UI has a stable place to read
// from if the two ever diverge.
func (h *Manipulador) ObterAdmin(c fiber.Ctx) error {
	return httpx.ResponderOK(c, h.svc.ObterPublico(c.Context()))
}

// Update answers PUT /api/admin/settings.
func (h *Manipulador) Atualizar(c fiber.Ctx) error {
	var req RequisicaoDeAtualizacao
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}

	out, err := h.svc.Atualizar(c.Context(), authx.UsuarioObrigatorio(c), req)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}
