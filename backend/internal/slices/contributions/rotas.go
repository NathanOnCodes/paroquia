package contributions

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Register mounts the contributions routes.
//
//	GET /api/admin/contributions                    assistente+
//	GET /api/admin/contributions/by-donor/:donorId  assistente+
func Registrar(api fiber.Router, h *Manipulador, auth *authx.Autenticacao) {
	admin := api.Group("/admin/contributions",
		auth.ExigirAutenticacao(),
		auth.ExigirPapel(authx.PapelAssistente),
		auth.ExigirSenhaAtualizada(),
	)
	admin.Get("/", h.Listar)
	admin.Get("/by-donor/:donorId", h.ListarPorDoador)
}
