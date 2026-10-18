package events

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Registrar monta as rotas de eventos.
//
//	GET   /api/public/events           público, somente publicados
//	GET   /api/public/events/:slug     público
//	GET   /api/admin/events            assistente+
//	POST  /api/admin/events            assistente+
//	GET   /api/admin/events/:id        assistente+
//	PUT   /api/admin/events/:id        assistente+
//	PATCH /api/admin/events/:id/status assistente+
func Registrar(api fiber.Router, h *Manipulador, autenticacao *authx.Autenticacao) {
	api.Get("/public/events", h.ListarPublicas)
	api.Get("/public/events/:slug", h.ObterPublico)

	admin := api.Group("/admin/events",
		autenticacao.ExigirAutenticacao(),
		autenticacao.ExigirPapel(authx.PapelAssistente),
		autenticacao.ExigirSenhaAtualizada(),
	)
	admin.Get("/", h.ListarAdmin)
	admin.Post("/", h.Criar)
	admin.Get("/:id", h.Obter)
	admin.Put("/:id", h.Atualizar)
	admin.Patch("/:id/status", h.DefinirStatus)
}
