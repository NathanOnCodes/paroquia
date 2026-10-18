package transparency

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Manipulador é a superfície HTTP da transparência.
type Manipulador struct{ svc Servico }

// NovoManipulador monta o manipulador de transparência.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// ListarPublicos responde GET /api/public/transparency/periods.
func (h *Manipulador) ListarPublicos(c fiber.Ctx) error {
	saida, err := h.svc.ListarPublicos(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// ObterPublico responde GET /api/public/transparency/periods/:id.
func (h *Manipulador) ObterPublico(c fiber.Ctx) error {
	saida, err := h.svc.ObterPublico(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// ListarAdmin responde GET /api/admin/transparency/periods.
func (h *Manipulador) ListarAdmin(c fiber.Ctx) error {
	saida, err := h.svc.ListarAdmin(c.Context())
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// ObterAdmin responde GET /api/admin/transparency/periods/:id.
func (h *Manipulador) ObterAdmin(c fiber.Ctx) error {
	saida, err := h.svc.ObterAdmin(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// CriarPeriodo responde POST /api/admin/transparency/periods.
func (h *Manipulador) CriarPeriodo(c fiber.Ctx) error {
	var req RequisicaoDePeriodo
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.CriarPeriodo(c.Context(), authx.UsuarioObrigatorio(c), req)
	if err != nil {
		return err
	}
	return httpx.ResponderCriado(c, saida)
}

// AtualizarPeriodo responde PUT /api/admin/transparency/periods/:id.
func (h *Manipulador) AtualizarPeriodo(c fiber.Ctx) error {
	var req RequisicaoDePeriodo
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.AtualizarPeriodo(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// MudarStatus responde PATCH /api/admin/transparency/periods/:id/status.
func (h *Manipulador) MudarStatus(c fiber.Ctx) error {
	var req RequisicaoDeStatus
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.MudarStatus(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req.Status)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// AdicionarLancamento responde POST /api/admin/transparency/periods/:id/entries.
func (h *Manipulador) AdicionarLancamento(c fiber.Ctx) error {
	var req RequisicaoDeLancamento
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	saida, err := h.svc.AdicionarLancamento(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), req)
	if err != nil {
		return err
	}
	return httpx.ResponderCriado(c, saida)
}

// RemoverLancamento responde DELETE /api/admin/transparency/entries/:id.
func (h *Manipulador) RemoverLancamento(c fiber.Ctx) error {
	if err := h.svc.RemoverLancamento(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id")); err != nil {
		return err
	}
	return httpx.ResponderConfirmado(c)
}

// AnexarDocumento responde POST /api/admin/transparency/periods/:id/documents.
func (h *Manipulador) AnexarDocumento(c fiber.Ctx) error {
	arquivo, err := c.FormFile("file")
	if err != nil {
		return httpx.ErroDeValidacao("Arquivo obrigatório",
			httpx.ErroDeCampo{Campo: "file", Mensagem: "campo obrigatório"})
	}
	saida, err := h.svc.AnexarDocumento(c.Context(), authx.UsuarioObrigatorio(c), c.Params("id"), arquivo)
	if err != nil {
		return err
	}
	return httpx.ResponderCriado(c, saida)
}

// BaixarDocumento responde GET /api/files/transparency/:id.
//
// Endpoint público, mas o download só acontece quando o período está
// publicado — a mesma regra da rota legada.
func (h *Manipulador) BaixarDocumento(c fiber.Ctx) error {
	stream, nome, tipo, err := h.svc.AbrirDocumento(c.Context(), c.Params("id"))
	if err != nil {
		return err
	}
	defer func() { _ = stream.Close() }()

	c.Set(fiber.HeaderContentType, tipo)
	c.Set(fiber.HeaderContentDisposition, `attachment; filename="`+nome+`"`)
	return c.SendStream(stream)
}

// Registrar monta as rotas de transparência.
//
//	GET   /api/public/transparency/periods                público
//	GET   /api/public/transparency/periods/:id            público
//	GET   /api/files/transparency/:id                     público, se publicado
//	GET   /api/admin/transparency/periods                 assistente+
//	POST  /api/admin/transparency/periods                 assistente+
//	GET   /api/admin/transparency/periods/:id             assistente+
//	PUT   /api/admin/transparency/periods/:id             assistente+
//	PATCH /api/admin/transparency/periods/:id/status      assistente+ (admin para publicar)
//	POST  /api/admin/transparency/periods/:id/entries     assistente+
//	POST  /api/admin/transparency/periods/:id/documents   assistente+
//	DELETE /api/admin/transparency/entries/:id            assistente+
func Registrar(api fiber.Router, h *Manipulador, autenticacao *authx.Autenticacao) {
	api.Get("/public/transparency/periods", h.ListarPublicos)
	api.Get("/public/transparency/periods/:id", h.ObterPublico)
	api.Get("/files/transparency/:id", h.BaixarDocumento)

	admin := api.Group("/admin/transparency",
		autenticacao.ExigirAutenticacao(),
		autenticacao.ExigirPapel(authx.PapelAssistente),
		autenticacao.ExigirSenhaAtualizada(),
	)
	admin.Get("/periods", h.ListarAdmin)
	admin.Post("/periods", h.CriarPeriodo)
	admin.Get("/periods/:id", h.ObterAdmin)
	admin.Put("/periods/:id", h.AtualizarPeriodo)
	admin.Patch("/periods/:id/status", h.MudarStatus)
	admin.Post("/periods/:id/entries", h.AdicionarLancamento)
	admin.Post("/periods/:id/documents", h.AnexarDocumento)
	admin.Delete("/entries/:id", h.RemoverLancamento)
}
