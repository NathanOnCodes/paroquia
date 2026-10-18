package contact

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Register mounts the contact routes.
//
//	POST  /api/contact                      público, com rate limit
//	GET   /api/admin/contacts               assistente+
//	PATCH /api/admin/contacts/:id/status    assistente+
func Registrar(api fiber.Router, h *Manipulador, auth *authx.Autenticacao) {
	api.Post("/contact", h.Enviar)

	admin := api.Group("/admin/contacts",
		auth.ExigirAutenticacao(),
		auth.ExigirPapel(authx.PapelAssistente),
		auth.ExigirSenhaAtualizada(),
	)
	admin.Get("/", h.ListarAdmin)
	admin.Patch("/:id/status", h.DefinirStatus)
}
