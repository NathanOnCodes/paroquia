package events

import (
	"mime/multipart"

	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Manipulador é a superfície HTTP dos eventos.
type Manipulador struct{ svc Servico }

// NovoManipulador monta o manipulador de eventos.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// ListarPublicas responde GET /api/public/events.
func (h *Manipulador) ListarPublicas(c fiber.Ctx) error {
	saida, err := h.svc.ListarPublicas(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// ObterPublico responde GET /api/public/events/:slug.
func (h *Manipulador) ObterPublico(c fiber.Ctx) error {
	saida, err := h.svc.ObterPublico(c.Context(), c.Params("slug"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// ListarAdmin responde GET /api/admin/events.
func (h *Manipulador) ListarAdmin(c fiber.Ctx) error {
	saida, err := h.svc.ListarAdmin(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// Obter responde GET /api/admin/events/:id.
func (h *Manipulador) Obter(c fiber.Ctx) error {
	saida, err := h.svc.Obter(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// Criar responde POST /api/admin/events.
func (h *Manipulador) Criar(c fiber.Ctx) error {
	var req RequisicaoDeEscrita
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	imagem, err := obterImagem(c)
	if err != nil {
		return err
	}
	saida, err := h.svc.Criar(c.Context(), authx.UsuarioObrigatorio(c), req, imagem)
	if err != nil {
		return err
	}
	return httpx.ResponderCriado(c, saida)
}

// Atualizar responde PUT /api/admin/events/:id.
func (h *Manipulador) Atualizar(c fiber.Ctx) error {
	var req RequisicaoDeEscrita
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	imagem, err := obterImagem(c)
	if err != nil {
		return err
	}
	saida, err := h.svc.Atualizar(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req, imagem)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// DefinirStatus responde PATCH /api/admin/events/:id/status.
func (h *Manipulador) DefinirStatus(c fiber.Ctx) error {
	var req RequisicaoDeStatus
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.DefinirStatus(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req.Status)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// obterImagem extrai o arquivo opcional do formulário multipart.
func obterImagem(c fiber.Ctx) (*multipart.FileHeader, error) {
	arquivo, err := c.FormFile("image")
	if err != nil {
		// Campo ausente não é erro: a imagem é opcional.
		return nil, nil
	}
	return arquivo, nil
}
