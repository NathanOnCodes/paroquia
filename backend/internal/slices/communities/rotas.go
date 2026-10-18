package communities

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Register mounts the communities routes.
//
//	GET  /api/public/communities              público (apenas ativas)
//	GET  /api/admin/communities               assistente+
//	POST /api/admin/communities               assistente+
//	GET  /api/admin/communities/:id           assistente+
//	PUT  /api/admin/communities/:id           assistente+
//	PATCH /api/admin/communities/:id/status   assistente+
func Registrar(api fiber.Router, h *Manipulador, auth *authx.Autenticacao) {
	api.Get("/public/communities", h.ListarPublico)

	admin := api.Group("/admin/communities",
		auth.ExigirAutenticacao(),
		auth.ExigirPapel(authx.PapelAssistente),
		auth.ExigirSenhaAtualizada(),
	)
	admin.Get("/", h.ListarAdmin)
	admin.Post("/", h.Criar)
	admin.Get("/:id", h.Obter)
	admin.Put("/:id", h.Atualizar)
	admin.Patch("/:id/status", h.DefinirStatus)
}
