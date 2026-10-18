package contributions

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Handler is the HTTP surface of the contributions slice.
type Manipulador struct {
	svc Servico
}

// NewHandler builds the contributions handler.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// List answers GET /api/admin/contributions.
func (h *Manipulador) Listar(c fiber.Ctx) error {
	out, err := h.svc.ListarAdmin(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}

// ListByDonor answers GET /api/admin/contributions/by-donor/:donorId.
func (h *Manipulador) ListarPorDoador(c fiber.Ctx) error {
	out, err := h.svc.ListarPorDoador(c.Context(), c.Params("donorId"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}
