package auth

import (
	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Manipulador é a superfície HTTP da autenticação.
type Manipulador struct {
	svc     Servico
	cookies *authx.GravadorDeCookies
}

// NovoManipulador monta o manipulador de autenticação.
func NovoManipulador(svc Servico, cookies *authx.GravadorDeCookies) *Manipulador {
	return &Manipulador{svc: svc, cookies: cookies}
}

// Entrar responde POST /api/auth/login.
func (h *Manipulador) Entrar(c fiber.Ctx) error {
	var req RequisicaoDeLogin
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}

	sessao, credenciais, err := h.svc.Login(c.Context(), req)
	if err != nil {
		return err
	}

	h.cookies.DefinirSessao(c, credenciais.Acesso, credenciais.Atualizacao)
	return httpx.ResponderOK(c, sessao)
}

// Renovar responde POST /api/auth/refresh.
func (h *Manipulador) Renovar(c fiber.Ctx) error {
	sessao, credenciais, err := h.svc.Renovar(c.Context(), h.cookies.TokenDeAtualizacao(c))
	if err != nil {
		return err
	}
	// O access é renovado sempre; o refresh é rotacionado pelo serviço.
	h.cookies.DefinirSessao(c, credenciais.Acesso, credenciais.Atualizacao)
	return httpx.ResponderOK(c, sessao)
}

// Sair responde POST /api/auth/logout.
func (h *Manipulador) Sair(c fiber.Ctx) error {
	if err := h.svc.Encerrar(c.Context(), h.cookies.TokenDeAtualizacao(c)); err != nil {
		return err
	}
	h.cookies.Limpar(c)
	return httpx.ResponderConfirmado(c)
}

// Perfil responde GET /api/auth/eu.
func (h *Manipulador) Perfil(c fiber.Ctx) error {
	reivindicacoes := authx.DoContexto(c)
	if reivindicacoes == nil {
		return httpx.ErroNaoAutenticado("Não autenticado")
	}
	perfil, err := h.svc.ObterPerfil(c.Context(), reivindicacoes.UserID)
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, perfil)
}

// TrocarSenha responde POST /api/auth/senha.
func (h *Manipulador) TrocarSenha(c fiber.Ctx) error {
	var req RequisicaoDeTrocaDeSenha
	if err := httpx.Vincular(c, &req); err != nil {
		return err
	}
	if err := h.svc.TrocarSenha(c.Context(), authx.UsuarioObrigatorio(c), req); err != nil {
		return err
	}
	return httpx.ResponderConfirmado(c)
}
