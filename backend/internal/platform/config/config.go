// Package config loads and validates the runtime configuration.
//
// Nothing in this package reads the environment outside of Load: every
// constructor receives a validated Config value, which keeps the rest of the
// codebase free of os.Getenv calls and trivially testable.
package config

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the complete, validated runtime configuration.
type Configuracao struct {
	Env                    string
	ServidorHTTP           ServidorHTTP
	BancoDeDados           BancoDeDados
	Autenticacao           Autenticacao
	Armazenamento          Armazenamento
	ConfiguracaoEmail      ConfiguracaoEmail
	ConfiguracaoPagamentos ConfiguracaoPagamentos
	Site                   Site
	Limites                Limites
	Telemetria             Telemetria
}

type ServidorHTTP struct {
	Endereco string
	// PublicURL is the externally visible origin. It is used to build links
	// sent by e-mail and to decide whether cookies must be Secure.
	URLPublica string
}

type BancoDeDados struct {
	URL                        string
	MaximoConexoesAbertas      int
	MaximoConexoesOciosas      int
	TempoMaximoDeVidaDaConexao time.Duration
	// AutoMigrate applies the embedded SQL migrations on startup.
	MigrarAutomaticamente bool
}

type Autenticacao struct {
	// AccessTokenTTL is the lifetime of the short-lived JWT.
	TempoDeVidaDoAcesso time.Duration
	// RefreshTokenTTL is the lifetime of the rotating refresh token.
	TempoDeVidaDaAtualizacao time.Duration
	// JWTSecret signs the access token. Must be at least 32 bytes.
	SegredoJWT string
	// CookieName is the cookie carrying the access token.
	CookieDeAcesso      string
	CookieDeAtualizacao string
	// RecurringCookie carries the donor's magic-link session.
	CookieDeRecorrencia string
	// RecurringTokenTTL mirrors the legacy TTL_HORAS_TOKEN_RECORRENTE.
	TempoDeVidaDoTokenDeRecorrencia time.Duration
}

type Armazenamento struct {
	// Dir is the root of the upload volume.
	Dir string
	// MaxFileSizeMB limits every upload.
	TamanhoMaximoArquivoMB int64
}

type ConfiguracaoEmail struct {
	APIKey                  string
	From                    string
	NotificarAdminEm        string
	Habilitado              bool
	MaximoDeTentativas      int
	EsperaEntreTentativas   time.Duration
	TempoLimiteDeRequisicao time.Duration
}

type ConfiguracaoPagamentos struct {
	ChaveSecreta     string
	SegredoDoWebhook string
	ChavePublicavel  string
	Currency         string
	// MinAmountCents mirrors STRIPE_MIN_AMOUNT_CENTS (R$ 1,00).
	ValorMinimoEmCentavos int
	// MaxAmountCents mirrors the legacy validation ceiling (R$ 100.000,00).
	ValorMaximoEmCentavos int
	VersaoDaAPI           string
	Habilitado            bool
}

type Site struct {
	Name string
}

type Limites struct {
	// Contact: 3 requests per 5 minutes per IP (legacy behaviour).
	MaximoDeContatos int
	JanelaDeContatos time.Duration
	// RecurringLink: 3 per 5 minutes per IP and 2 per 5 minutes per e-mail.
	MaximoDeRecorrenciasPorIP    int
	MaximoDeRecorrenciasPorEmail int
	JanelaDeRecorrencias         time.Duration
}

type Telemetria struct {
	// RequestLog enables the structured request logger.
	RegistrarRequisicoes bool
	NivelDeRegistro      string
}

// Load reads the environment, applies documented defaults and validates the
// result. It returns an error listing every problem at once so that a
// misconfigured deployment can be fixed in a single pass.
func Carregar() (*Configuracao, error) {
	prod := strings.EqualFold(obter("APP_ENV", "development"), "production")

	cfg := &Configuracao{
		Env: obter("APP_ENV", "development"),
		ServidorHTTP: ServidorHTTP{
			Endereco:   obter("HTTP_ADDR", ":8080"),
			URLPublica: obter("PUBLIC_URL", "http://localhost:8080"),
		},
		BancoDeDados: BancoDeDados{
			URL:                        obter("DATABASE_URL", "postgres://paroquia:paroquia@127.0.0.1:5432/paroquia?sslmode=disable"),
			MaximoConexoesAbertas:      obterInteiro("DATABASE_MAX_OPEN_CONNS", 25),
			MaximoConexoesOciosas:      obterInteiro("DATABASE_MAX_IDLE_CONNS", 10),
			TempoMaximoDeVidaDaConexao: obterDuracao("DATABASE_CONN_MAX_LIFETIME", time.Hour),
			MigrarAutomaticamente:      obterBool("DATABASE_AUTO_MIGRATE", true),
		},
		Autenticacao: Autenticacao{
			TempoDeVidaDoAcesso:             obterDuracao("AUTH_ACCESS_TTL", 15*time.Minute),
			TempoDeVidaDaAtualizacao:        obterDuracao("AUTH_REFRESH_TTL", 30*24*time.Hour),
			SegredoJWT:                      obter("AUTH_JWT_SECRET", ""),
			CookieDeAcesso:                  obter("AUTH_ACCESS_COOKIE", "paroquia_access"),
			CookieDeAtualizacao:             obter("AUTH_REFRESH_COOKIE", "paroquia_refresh"),
			CookieDeRecorrencia:             obter("AUTH_RECURRING_COOKIE", "recurring_mgmt"),
			TempoDeVidaDoTokenDeRecorrencia: obterDuracao("TTL_HORAS_TOKEN_RECORRENTE", 24*time.Hour),
		},
		Armazenamento: Armazenamento{
			Dir:                    obter("STORAGE_DIR", "./data/uploads"),
			TamanhoMaximoArquivoMB: int64(obterInteiro("TAMANHO_MAXIMO_ARQUIVO_MB", 10)),
		},
		ConfiguracaoEmail: ConfiguracaoEmail{
			APIKey:                  obter("CHAVE_API_RESEND", ""),
			From:                    obter("EMAIL_REMETENTE", "Paróquia <nao-responder@paroquia.local>"),
			NotificarAdminEm:        obter("EMAIL_NOTIFICACAO_ADMIN", ""),
			MaximoDeTentativas:      obterInteiro("EMAIL_MAX_ATTEMPTS", 3),
			EsperaEntreTentativas:   obterDuracao("EMAIL_RETRY_BACKOFF", 2*time.Second),
			TempoLimiteDeRequisicao: obterDuracao("EMAIL_TIMEOUT", 10*time.Second),
		},
		ConfiguracaoPagamentos: ConfiguracaoPagamentos{
			ChaveSecreta:          obter("CHAVE_SECRETA_STRIPE", ""),
			SegredoDoWebhook:      obter("SEGREDO_WEBHOOK_STRIPE", ""),
			ChavePublicavel:       obter("NEXT_PUBLIC_CHAVE_PUBLICA_STRIPE", ""),
			Currency:              obter("STRIPE_CURRENCY", "brl"),
			ValorMinimoEmCentavos: obterInteiro("STRIPE_MIN_AMOUNT_CENTS", 100),
			ValorMaximoEmCentavos: obterInteiro("STRIPE_MAX_AMOUNT_CENTS", 10_000_000),
			VersaoDaAPI:           obter("STRIPE_API_VERSION", "2026-07-29.dahlia"),
		},
		Site: Site{Name: obter("SITE_NAME", "Paróquia")},
		Limites: Limites{
			MaximoDeContatos:             obterInteiro("LIMIT_CONTACT_MAX", 3),
			JanelaDeContatos:             obterDuracao("LIMIT_CONTACT_WINDOW", 5*time.Minute),
			MaximoDeRecorrenciasPorIP:    obterInteiro("LIMIT_RECURRING_IP_MAX", 3),
			MaximoDeRecorrenciasPorEmail: obterInteiro("LIMIT_RECURRING_EMAIL_MAX", 2),
			JanelaDeRecorrencias:         obterDuracao("LIMIT_RECURRING_WINDOW", 5*time.Minute),
		},
		Telemetria: Telemetria{
			RegistrarRequisicoes: obterBool("HTTP_REQUEST_LOG", true),
			NivelDeRegistro:      obter("LOG_LEVEL", "info"),
		},
	}

	cfg.ConfiguracaoEmail.Habilitado = cfg.ConfiguracaoEmail.APIKey != ""
	cfg.ConfiguracaoPagamentos.Habilitado = cfg.ConfiguracaoPagamentos.ChaveSecreta != ""

	var errs []string
	add := func(format string, args ...any) { errs = append(errs, fmt.Sprintf(format, args...)) }

	if cfg.Autenticacao.SegredoJWT == "" {
		if prod {
			add("AUTH_JWT_SECRET é obrigatória em produção")
		} else {
			cfg.Autenticacao.SegredoJWT = "local-dev-secret-do-not-use-in-production-0123456789"
		}
	}
	if len(cfg.Autenticacao.SegredoJWT) < 32 {
		add("AUTH_JWT_SECRET deve ter ao menos 32 caracteres")
	}
	if _, err := url.Parse(cfg.ServidorHTTP.URLPublica); err != nil {
		add("PUBLIC_URL inválida: %v", err)
	}
	if cfg.BancoDeDados.URL == "" {
		add("DATABASE_URL é obrigatória")
	}
	if cfg.Armazenamento.TamanhoMaximoArquivoMB < 1 {
		add("TAMANHO_MAXIMO_ARQUIVO_MB deve ser >= 1")
	}
	if cfg.Autenticacao.TempoDeVidaDoAcesso <= 0 || cfg.Autenticacao.TempoDeVidaDaAtualizacao <= 0 {
		add("TTLs de autenticação devem ser positivos")
	}
	if cfg.Autenticacao.TempoDeVidaDoTokenDeRecorrencia < time.Hour {
		add("TTL_HORAS_TOKEN_RECORRENTE deve ser ao menos 1 hora")
	}
	if prod {
		if cfg.ConfiguracaoEmail.APIKey == "" {
			add("CHAVE_API_RESEND é obrigatória em produção")
		}
		if cfg.ConfiguracaoPagamentos.ChaveSecreta == "" {
			add("CHAVE_SECRETA_STRIPE é obrigatória em produção")
		}
		if cfg.ConfiguracaoPagamentos.SegredoDoWebhook == "" {
			add("SEGREDO_WEBHOOK_STRIPE é obrigatória em produção")
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("configuração inválida:\n  - %s", strings.Join(errs, "\n  - "))
	}
	return cfg, nil
}

// SecureCookies reports whether cookies must carry the Secure attribute.
func (c *Configuracao) CookiesSeguros() bool {
	return strings.HasPrefix(strings.ToLower(c.ServidorHTTP.URLPublica), "https://")
}

// MaxFileSizeBytes returns the upload ceiling in bytes.
func (c *Configuracao) TamanhoMaximoArquivoBytes() int64 {
	return c.Armazenamento.TamanhoMaximoArquivoMB * 1024 * 1024
}

func obter(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok && v != "" {
		return v
	}
	return fallback
}

func obterInteiro(key string, fallback int) int {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func obterBool(key string, fallback bool) bool {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func obterDuracao(key string, fallback time.Duration) time.Duration {
	v, ok := os.LookupEnv(key)
	if !ok || v == "" {
		return fallback
	}
	// (comentário removido: texto original indisponível)
	// compatibility with the legacy TTL_HORAS_TOKEN_RECORRENTE variable.
	if d, err := time.ParseDuration(v); err == nil {
		return d
	}
	if n, err := strconv.Atoi(v); err == nil {
		return time.Duration(n) * time.Hour
	}
	return fallback
}
