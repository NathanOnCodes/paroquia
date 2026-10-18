package donors

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Handler is the HTTP surface of the donors slice.
type Manipulador struct {
	svc Servico
}

// NewHandler builds the donors handler.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// List answers GET /api/admin/donors.
func (h *Manipulador) Listar(c fiber.Ctx) error {
	out, err := h.svc.ListarAdmin(c.Context(), c.Query("search"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}

// Get answers GET /api/admin/donors/:id.
func (h *Manipulador) Obter(c fiber.Ctx) error {
	out, err := h.svc.ObterAdmin(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}

// Create answers POST /api/admin/donors.
func (h *Manipulador) Criar(c fiber.Ctx) error {
	var req RequisicaoDeEscrita
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	out, err := h.svc.Criar(c.Context(), authx.UsuarioObrigatorio(c), req)
	if err != nil {
		return err
	}
	return httpx.ResponderCriado(c, out)
}

// Update answers PUT /api/admin/donors/:id.
func (h *Manipulador) Atualizar(c fiber.Ctx) error {
	var req RequisicaoDeEscrita
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	out, err := h.svc.Atualizar(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}

// SetStatus answers PATCH /api/admin/donors/:id/status.
func (h *Manipulador) DefinirStatus(c fiber.Ctx) error {
	var req RequisicaoDeStatus
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	out, err := h.svc.DefinirStatus(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req.IsActive)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}
