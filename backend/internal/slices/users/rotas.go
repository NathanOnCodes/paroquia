package users

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Registrar monta as rotas de gestão de usuários.
//
// Exige papel admin: criar administradores é a operação mais sensível do
// sistema, exatamente como `canManageUsers` no código legado.
//
//	GET   /api/admin/usuarios               admin
//	POST  /api/admin/usuarios               admin
//	GET   /api/admin/usuarios/:id           admin
//	PUT   /api/admin/usuarios/:id           admin
//	PATCH /api/admin/usuarios/:id/status    admin
//	POST  /api/admin/usuarios/:id/senha     admin
func Registrar(api fiber.Router, h *Manipulador, autenticacao *authx.Autenticacao) {
	admin := api.Group("/admin/usuarios",
		autenticacao.ExigirAutenticacao(),
		autenticacao.ExigirPapel(authx.PapelAdmin),
		autenticacao.ExigirSenhaAtualizada(),
	)
	admin.Get("/", h.Listar)
	admin.Post("/", h.Criar)
	admin.Get("/:id", h.Obter)
	admin.Put("/:id", h.Atualizar)
	admin.Patch("/:id/status", h.DefinirStatus)
	admin.Post("/:id/senha", h.RedefinirSenha)
}
