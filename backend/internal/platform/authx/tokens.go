package authx

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Papéis aceitos pelo sistema. Espelham o CHECK de `users.role`.
const (
	PapelAdmin      = "admin"
	PapelAssistente = "assistente"
)

// ErrInvalidToken is returned for any token that cannot be trusted.
var ErrTokenInvalido = errors.New("token inválido")

// Claims is the payload of an access token. It carries exactly what the
// request handlers need, so no database lookup is required on the hot path.
type Reivindicacoes struct {
	UserID           string `json:"uid"`
	Email            string `json:"email"`
	Role             string `json:"role"`
	DeveTrocarASenha bool   `json:"mcp,omitempty"`
	jwt.RegisteredClaims
}

// (comentário removido: texto original indisponível)
// (comentário removido: texto original indisponível)
func (c *Reivindicacoes) TemPapel(required string) bool {
	switch required {
	case PapelAdmin:
		return c.Role == PapelAdmin
	case PapelAssistente:
		return c.Role == PapelAdmin || c.Role == PapelAssistente
	default:
		return false
	}
}

// IsAdmin reports whether the holder may perform administrative actions.
func (c *Reivindicacoes) EhAdmin() bool { return c.Role == PapelAdmin }

// TokenManager issues and verifies access tokens. It is created by
// NewTokenManager and never reads global state.
type GerenciadorDeTokens struct {
	secret    []byte
	issuer    string
	accessTTL time.Duration
	now       func() time.Time
}

// TokenManagerOptions carries the optional knobs.
type OpcoesDoGerenciadorDeTokens struct {
	Issuer              string
	TempoDeVidaDoAcesso time.Duration
	// Clock is injectable for tests.
	Clock func() time.Time
}

// NewTokenManager builds a TokenManager. The secret must be at least 32 bytes.
func NovoGerenciadorDeTokens(secret string, opts OpcoesDoGerenciadorDeTokens) (*GerenciadorDeTokens, error) {
	if len(secret) < 32 {
		return nil, errors.New("segredo JWT deve ter ao menos 32 caracteres")
	}
	if opts.Issuer == "" {
		opts.Issuer = "paroquia"
	}
	if opts.TempoDeVidaDoAcesso <= 0 {
		opts.TempoDeVidaDoAcesso = 15 * time.Minute
	}
	if opts.Clock == nil {
		opts.Clock = time.Now
	}
	return &GerenciadorDeTokens{
		secret:    []byte(secret),
		issuer:    opts.Issuer,
		accessTTL: opts.TempoDeVidaDoAcesso,
		now:       opts.Clock,
	}, nil
}

// IssueAccess signs a short-lived access token for the given user.
func (m *GerenciadorDeTokens) EmitirAcesso(user UsuarioAutenticado) (string, error) {
	now := m.now()
	claims := Reivindicacoes{
		UserID:           user.ID,
		Email:            user.Email,
		Role:             user.Role,
		DeveTrocarASenha: user.DeveTrocarASenha,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    m.issuer,
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessTTL)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("assinando token: %w", err)
	}
	return signed, nil
}

// ParseAccess verifies the signature and the validity window of a token.
func (m *GerenciadorDeTokens) InterpretarAcesso(raw string) (*Reivindicacoes, error) {
	claims := &Reivindicacoes{}
	token, err := jwt.ParseWithClaims(raw, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: algoritmo inesperado %s", ErrTokenInvalido, t.Header["alg"])
		}
		return m.secret, nil
	},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(30*time.Second),
		// O relógio é injetável, então os testes podem simular vencimento sem
		// depender do tempo real.
		jwt.WithTimeFunc(m.now),
	)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrTokenInvalido, err)
	}
	if !token.Valid {
		return nil, ErrTokenInvalido
	}
	if claims.UserID == "" {
		return nil, fmt.Errorf("%w: subject ausente", ErrTokenInvalido)
	}
	return claims, nil
}

// AccessTTL exposes the configured lifetime, for cookie Max-Age.
func (m *GerenciadorDeTokens) TempoDeVidaDoAcesso() time.Duration { return m.accessTTL }

// AuthUser is the identity carried by a request after authentication.
type UsuarioAutenticado struct {
	ID               string
	Email            string
	FullName         string
	Role             string
	IsActive         bool
	DeveTrocarASenha bool
}
