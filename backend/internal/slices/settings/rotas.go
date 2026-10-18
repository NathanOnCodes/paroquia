package settings

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Register mounts the settings routes.
//
//	GET /api/public/settings   público
//	GET /api/admin/settings    admin
//	PUT /api/admin/settings    admin
func Registrar(api fiber.Router, h *Manipulador, auth *authx.Autenticacao) {
	api.Get("/public/settings", h.ObterPublico)

	admin := api.Group("/admin/settings",
		auth.ExigirAutenticacao(),
		auth.ExigirPapel(authx.PapelAdmin),
		auth.ExigirSenhaAtualizada(),
	)
	admin.Get("/", h.ObterAdmin)
	admin.Put("/", h.Atualizar)
}
