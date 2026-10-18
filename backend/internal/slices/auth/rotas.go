package auth

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
)

// Registrar monta as rotas de autenticação.
//
//	POST /api/auth/login     público
//	POST /api/auth/refresh   público (usa o cookie de refresh)
//	POST /api/auth/logout    público
//	GET  /api/auth/eu        autenticado
//	POST /api/auth/senha     autenticado
func Registrar(api fiber.Router, h *Manipulador, autenticacao *authx.Autenticacao) {
	api.Post("/auth/login", h.Entrar)
	api.Post("/auth/refresh", h.Renovar)
	api.Post("/auth/logout", h.Sair)

	privado := api.Group("/auth",
		autenticacao.ExigirAutenticacao(),
	)
	privado.Get("/eu", h.Perfil)
	privado.Post("/senha", h.TrocarSenha)
}
