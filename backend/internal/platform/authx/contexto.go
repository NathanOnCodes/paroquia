package authx

import (
	"github.com/gofiber/fiber/v3"
)

// localsKey is the key used to stash the authenticated claims in the Fiber
// request locals. It is unexported so only this package can write it.
const chaveDeLocals = "paroquia:auth:claims"

// WithClaims stores the authenticated claims on the request context.
func ComReivindicacoes(c fiber.Ctx, claims *Reivindicacoes) {
	c.Locals(chaveDeLocals, claims)
}

// FromContext returns the authenticated claims, or nil on a public route.
func DoContexto(c fiber.Ctx) *Reivindicacoes {
	v := c.Locals(chaveDeLocals)
	if v == nil {
		return nil
	}
	claims, _ := v.(*Reivindicacoes)
	return claims
}

// User returns the identity of the caller. It reports false when the route is
// unauthenticated.
func User(c fiber.Ctx) (UsuarioAutenticado, bool) {
	claims := DoContexto(c)
	if claims == nil {
		return UsuarioAutenticado{}, false
	}
	return UsuarioAutenticado{
		ID:               claims.UserID,
		Email:            claims.Email,
		Role:             claims.Role,
		DeveTrocarASenha: claims.DeveTrocarASenha,
	}, true
}

// MustUser is a convenience for handlers that sit behind RequireAuth.
// It returns a zero value instead of panicking so a bug degrades to a 401
// rather than a crash.
func UsuarioObrigatorio(c fiber.Ctx) UsuarioAutenticado {
	u, _ := User(c)
	return u
}
