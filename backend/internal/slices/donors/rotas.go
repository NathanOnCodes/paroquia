package donors

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Register mounts the donors routes.
//
//	GET   /api/admin/donors               assistente+
//	POST  /api/admin/donors               assistente+
//	GET   /api/admin/donors/:id           assistente+
//	PUT   /api/admin/donors/:id           assistente+
//	PATCH /api/admin/donors/:id/status    assistente+
func Registrar(api fiber.Router, h *Manipulador, auth *authx.Autenticacao) {
	admin := api.Group("/admin/donors",
		auth.ExigirAutenticacao(),
		auth.ExigirPapel(authx.PapelAssistente),
		auth.ExigirSenhaAtualizada(),
	)
	admin.Get("/", h.Listar)
	admin.Post("/", h.Criar)
	admin.Get("/:id", h.Obter)
	admin.Put("/:id", h.Atualizar)
	admin.Patch("/:id/status", h.DefinirStatus)
}
