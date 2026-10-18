package dashboard

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Handler is the HTTP surface of the dashboard slice.
type Manipulador struct{ svc Servico }

// NewHandler builds the dashboard handler.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// Summary answers GET /api/admin/dashboard.
func (h *Manipulador) Resumo(c fiber.Ctx) error {
	out, err := h.svc.Resumo(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, out)
}

// Register mounts the dashboard routes.
//
//	GET /api/admin/dashboard   assistente+
func Registrar(api fiber.Router, h *Manipulador, auth *authx.Autenticacao) {
	api.Get("/admin/dashboard",
		auth.ExigirAutenticacao(),
		auth.ExigirPapel(authx.PapelAssistente),
		auth.ExigirSenhaAtualizada(),
		h.Resumo,
	)
}
