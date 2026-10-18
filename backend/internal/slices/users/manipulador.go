package users

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Manipulador é a superfície HTTP das contas administrativas.
type Manipulador struct{ svc Servico }

// NovoManipulador monta o manipulador de usuários.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// Listar responde GET /api/admin/usuarios.
func (h *Manipulador) Listar(c fiber.Ctx) error {
	saida, err := h.svc.Listar(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// Obter responde GET /api/admin/usuarios/:id.
func (h *Manipulador) Obter(c fiber.Ctx) error {
	saida, err := h.svc.Obter(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// Criar responde POST /api/admin/usuarios.
func (h *Manipulador) Criar(c fiber.Ctx) error {
	var req RequisicaoDeEscrita
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.Criar(c.Context(), authx.UsuarioObrigatorio(c), req)
	if err != nil {
		return err
	}
	return httpx.ResponderCriado(c, saida)
}

// Atualizar responde PUT /api/admin/usuarios/:id.
func (h *Manipulador) Atualizar(c fiber.Ctx) error {
	var req RequisicaoDeEscrita
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.Atualizar(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// DefinirStatus responde PATCH /api/admin/usuarios/:id/status.
func (h *Manipulador) DefinirStatus(c fiber.Ctx) error {
	var req RequisicaoDeStatus
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.DefinirStatus(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req.Ativo)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// RedefinirSenha responde POST /api/admin/usuarios/:id/senha.
func (h *Manipulador) RedefinirSenha(c fiber.Ctx) error {
	var req RequisicaoDeSenha
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	if err := h.svc.RedefinirSenha(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req); err != nil {
		return err
	}
	return httpx.ResponderConfirmado(c)
}
