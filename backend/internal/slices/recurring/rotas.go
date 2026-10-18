package recurring

import (
	"github.com/gofiber/fiber/v3"
)

// Registrar monta as rotas de recorrência do fiel.
//
// Nenhuma exige login administrativo: a identidade vem do link mágico por
// e-mail, guardado em cookie de uso curto.
//
//	POST /api/recurring/request-link  público, com rate limit
//	GET  /api/recurring/session       troca o link por cookie de sessão
//	GET  /api/recurring/gerenciar     lista as recorrências do fiel
//	POST /api/recurring/cancel        cancela uma recorrência
func Registrar(api fiber.Router, h *Manipulador) {
	api.Post("/recurring/request-link", h.SolicitarLink)
	api.Get("/recurring/session", h.AbrirSessao)
	api.Get("/recurring/gerenciar", h.Listar)
	api.Post("/recurring/cancel", h.Cancelar)
}
