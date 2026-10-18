package authx

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// ErrExpiredSession tells the caller that the access token is gone but the
// refresh token may still be valid: the frontend should refresh and retry.
var ErrSessaoExpirada = errors.New("sessão expirada")

// CookieWriter issues and clears the session cookies. It is constructed once
// with the transport settings so handlers never touch cookie attributes.
type GravadorDeCookies struct {
	accessName  string
	refreshName string
	secure      bool
	accessTTL   time.Duration
	refreshTTL  time.Duration
}

// CookieOptions configures the writer.
type OpcoesDeCookie struct {
	AccessName          string
	RefreshName         string
	Secure              bool
	TempoDeVidaDoAcesso time.Duration
	RefreshTTL          time.Duration
}

// NewCookieWriter builds a CookieWriter with sane defaults.
func NovoGravadorDeCookies(opts OpcoesDeCookie) *GravadorDeCookies {
	if opts.AccessName == "" {
		opts.AccessName = "paroquia_access"
	}
	if opts.RefreshName == "" {
		opts.RefreshName = "paroquia_refresh"
	}
	return &GravadorDeCookies{
		accessName:  opts.AccessName,
		refreshName: opts.RefreshName,
		secure:      opts.Secure,
		accessTTL:   opts.TempoDeVidaDoAcesso,
		refreshTTL:  opts.RefreshTTL,
	}
}

// SetSession writes both cookies. The refresh cookie is scoped to the refresh
// path so it is never sent to ordinary endpoints.
func (w *GravadorDeCookies) DefinirSessao(c fiber.Ctx, access, refresh string) {
	w.set(c, w.accessName, access, "/", int(w.accessTTL.Seconds()))
	w.set(c, w.refreshName, refresh, "/api/auth", int(w.refreshTTL.Seconds()))
}

// SetAccess renews only the short-lived cookie after a refresh.
func (w *GravadorDeCookies) DefinirAcesso(c fiber.Ctx, access string) {
	w.set(c, w.accessName, access, "/", int(w.accessTTL.Seconds()))
}

// Clear removes both cookies.
func (w *GravadorDeCookies) Limpar(c fiber.Ctx) {
	w.set(c, w.accessName, "", "/", -1)
	w.set(c, w.refreshName, "", "/api/auth", -1)
}

// AccessToken returns the bearer stored in the cookie, if any.
func (w *GravadorDeCookies) TokenDeAcesso(c fiber.Ctx) string { return c.Cookies(w.accessName) }

// RefreshToken returns the refresh bearer stored in the cookie, if any.
func (w *GravadorDeCookies) TokenDeAtualizacao(c fiber.Ctx) string { return c.Cookies(w.refreshName) }

func (w *GravadorDeCookies) set(c fiber.Ctx, name, value, path string, maxAge int) {
	c.Cookie(&fiber.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		HTTPOnly: true,
		Secure:   w.secure,
		SameSite: "Lax",
		MaxAge:   maxAge,
	})
}

// Auth is the middleware bundle used by route registration.
type Autenticacao struct {
	tokens  *GerenciadorDeTokens
	cookies *GravadorDeCookies
}

// NewAuth wires the middleware bundle.
func NovaAutenticacao(tokens *GerenciadorDeTokens, cookies *GravadorDeCookies) *Autenticacao {
	return &Autenticacao{tokens: tokens, cookies: cookies}
}

// Tokens exposes the token manager for handlers that issue tokens directly.
func (a *Autenticacao) Tokens() *GerenciadorDeTokens { return a.tokens }

// Cookies exposes the cookie writer.
func (a *Autenticacao) Cookies() *GravadorDeCookies { return a.cookies }

// Authenticate resolves the caller from the access cookie or the
// Authorization header and stores the claims on the context. It never fails:
// public routes simply see no identity.
func (a *Autenticacao) Autenticar() fiber.Handler {
	return func(c fiber.Ctx) error {
		raw := a.cookies.TokenDeAcesso(c)
		if raw == "" {
			raw = bearerDoCabecalho(c)
		}
		if raw == "" {
			return c.Next()
		}
		claims, err := a.tokens.InterpretarAcesso(raw)
		if err != nil {
			// An invalid token is treated as an anonymous request rather than
			// a hard failure, so expired sessions degrade gracefully on public
			// pages. Protected routes still get a 401 from RequireAuth.
			return c.Next()
		}
		ComReivindicacoes(c, claims)
		return c.Next()
	}
}

// RequireAuth rejects anonymous callers.
func (a *Autenticacao) ExigirAutenticacao() fiber.Handler {
	return func(c fiber.Ctx) error {
		if DoContexto(c) == nil {
			return httpx.ErroNaoAutenticado("Não autenticado")
		}
		return c.Next()
	}
}

// RequireRole restricts a route group to the given roles, honouring the same
// (comentário removido: texto original indisponível)
func (a *Autenticacao) ExigirPapel(roles ...string) fiber.Handler {
	return func(c fiber.Ctx) error {
		claims := DoContexto(c)
		if claims == nil {
			return httpx.ErroNaoAutenticado("Não autenticado")
		}
		for _, r := range roles {
			if claims.TemPapel(r) {
				return c.Next()
			}
		}
		return httpx.ErroSemPermissao("Sem permissão para esta operação")
	}
}

// RequireFreshPassword forces a password change before anything else. It is
// applied to every route except the password-change endpoint itself.
func (a *Autenticacao) ExigirSenhaAtualizada() fiber.Handler {
	return func(c fiber.Ctx) error {
		claims := DoContexto(c)
		if claims == nil {
			return httpx.ErroNaoAutenticado("Não autenticado")
		}
		if claims.DeveTrocarASenha {
			return httpx.ErroGenerico(fiber.StatusForbidden, "password_change_required",
				"É necessário alterar a senha antes de continuar")
		}
		return c.Next()
	}
}

func bearerDoCabecalho(c fiber.Ctx) string {
	const prefix = "Bearer "
	header := c.Get(fiber.HeaderAuthorization)
	if len(header) <= len(prefix) || header[:len(prefix)] != prefix {
		return ""
	}
	return header[len(prefix):]
}
