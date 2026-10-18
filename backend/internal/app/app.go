// Package app monta a aplicação: configuração, banco, middlewares e as
// rotas de cada slice.
//
// Todo recurso é criado por função construtora e injetado. Não há globais nem
// `init()` de ligação: o que existe em `main.go` é a única linha do tempo de
// construção, e ela é legível de cima a baixo.
package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/requestid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/config"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/docs"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/mail"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/payments"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/ratelimit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/storage"

	audit2 "github.com/NathanOnCodes/paroquia/backend/internal/slices/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/auth"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/communities"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/contact"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/contributions"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/dashboard"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/donors"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/events"
	pagamentos "github.com/NathanOnCodes/paroquia/backend/internal/slices/payments"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/recurring"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/settings"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/transparency"
	"github.com/NathanOnCodes/paroquia/backend/internal/slices/users"
)

// Aplicacao reúne o que precisa viver até o encerramento do processo.
type Aplicacao struct {
	Fiber   *fiber.App
	Cliente *ent.Client
	Config  *config.Configuracao
	Log     *slog.Logger
	// Armazenamento fica exposto para o endpoint de arquivos.
	Armazenamento storage.ArmazenamentoDeArquivos
}

// Dependencias permite substituir qualquer peça em testes.
type Dependencias struct {
	Config          *config.Configuracao
	Log             *slog.Logger
	Cliente         *ent.Client
	Armazenamento   storage.ArmazenamentoDeArquivos
	Remetente       mail.RemetenteDeEmail
	MeioDePagamento payments.MeioDePagamento
	Relogio         func() time.Time
}

// Novo constrói a aplicação inteira. Cada `Novo*` abaixo é um construtor:
// nada é instanciado implicitamente.
func Novo(ctx context.Context, d Dependencias) (*Aplicacao, error) {
	if d.Config == nil {
		return nil, fmt.Errorf("configuração obrigatória")
	}
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.Relogio == nil {
		d.Relogio = time.Now
	}

	// --- infraestrutura -----------------------------------------------
	registrador := audit.NovoRegistradorDeAuditoria(d.Cliente, d.Log)
	limitador := ratelimit.NovoLimitador(d.Cliente, d.Relogio)

	if d.Remetente == nil {
		d.Remetente = mail.NovoClienteResend(mail.Configuracao{
			APIKey:                d.Config.ConfiguracaoEmail.APIKey,
			From:                  d.Config.ConfiguracaoEmail.From,
			MaximoDeTentativas:    d.Config.ConfiguracaoEmail.MaximoDeTentativas,
			EsperaEntreTentativas: d.Config.ConfiguracaoEmail.EsperaEntreTentativas,
			Timeout:               d.Config.ConfiguracaoEmail.TempoLimiteDeRequisicao,
		})
	}
	if d.MeioDePagamento == nil {
		d.MeioDePagamento = payments.NovoClienteStripe(payments.ConfiguracaoDoStripe{
			ChaveSecreta:     d.Config.ConfiguracaoPagamentos.ChaveSecreta,
			SegredoDoWebhook: d.Config.ConfiguracaoPagamentos.SegredoDoWebhook,
			Currency:         d.Config.ConfiguracaoPagamentos.Currency,
			VersaoDaAPI:      d.Config.ConfiguracaoPagamentos.VersaoDaAPI,
		})
	}

	// --- autenticação --------------------------------------------------
	gerenciador, err := authx.NovoGerenciadorDeTokens(d.Config.Autenticacao.SegredoJWT, authx.OpcoesDoGerenciadorDeTokens{
		Issuer:              "paroquia",
		TempoDeVidaDoAcesso: d.Config.Autenticacao.TempoDeVidaDoAcesso,
		Clock:               d.Relogio,
	})
	if err != nil {
		return nil, fmt.Errorf("configurando tokens: %w", err)
	}

	gravadorDeCookies := authx.NovoGravadorDeCookies(authx.OpcoesDeCookie{
		AccessName:          d.Config.Autenticacao.CookieDeAcesso,
		RefreshName:         d.Config.Autenticacao.CookieDeAtualizacao,
		Secure:              d.Config.CookiesSeguros(),
		TempoDeVidaDoAcesso: d.Config.Autenticacao.TempoDeVidaDoAcesso,
		RefreshTTL:          d.Config.Autenticacao.TempoDeVidaDaAtualizacao,
	})
	autenticacao := authx.NovaAutenticacao(gerenciador, gravadorDeCookies)

	// --- slices --------------------------------------------------------
	configRepo := settings.NovoRepositorio(d.Cliente)
	configSvc := settings.NovoServico(configRepo, registrador, d.Log)
	configH := settings.NovoManipulador(configSvc)

	comunidadesRepo := communities.NovoRepositorio(d.Cliente)
	comunidadesSvc := communities.NovoServico(comunidadesRepo, registrador, d.Log)
	comunidadesH := communities.NovoManipulador(comunidadesSvc)

	doadorRepo := donors.NovoRepositorio(d.Cliente)
	doadorSvc := donors.NovoServico(doadorRepo, registrador, d.Log)
	doadorH := donors.NovoManipulador(doadorSvc)

	contatoRepo := contact.NovoRepositorio(d.Cliente)
	contatoSvc := contact.NovoServico(contact.Dependencias{
		Repo:              contatoRepo,
		Audit:             registrador,
		Mailer:            d.Remetente,
		Limitador:         limitador,
		Limites:           d.Config.Limites,
		ConfiguracaoEmail: d.Config.ConfiguracaoEmail,
		SiteName:          d.Config.Site.Name,
		Registrar:         d.Log,
	})
	contatoH := contact.NovoManipulador(contatoSvc)

	contribuicoesRepo := contributions.NovoRepositorio(d.Cliente)
	contribuicoesSvc := contributions.NovoServico(contribuicoesRepo)
	contribuicoesH := contributions.NovoManipulador(contribuicoesSvc)

	painelRepo := dashboard.NovoRepositorio(d.Cliente)
	painelSvc := dashboard.NovoServico(painelRepo)
	painelH := dashboard.NovoManipulador(painelSvc)

	usuariosRepo := users.NovoRepositorio(d.Cliente)
	usuariosSvc := users.NovoServico(users.Dependencias{
		Repositorio: usuariosRepo,
		Auditoria:   registrador,
		Log:         d.Log,
	})
	usuariosH := users.NovoManipulador(usuariosSvc)

	authSvc := auth.NovoServico(auth.Dependencias{
		Usuarios:                 auth.NovoRepositorioDeUsuarios(d.Cliente),
		Sessoes:                  auth.NovoRepositorioDeSessoes(d.Cliente),
		Tokens:                   gerenciador,
		Auditoria:                registrador,
		Log:                      d.Log,
		TempoDeVidaDaAtualizacao: d.Config.Autenticacao.TempoDeVidaDaAtualizacao,
		Relogio:                  d.Relogio,
	})
	authH := auth.NovoManipulador(authSvc, gravadorDeCookies)

	pagamentosSvc := pagamentos.NovoServico(pagamentos.Dependencias{
		Contribuicoes:   pagamentos.NovoRepositorioDeContribuicoes(d.Cliente),
		Recorrencias:    pagamentos.NovoRepositorioDeRecorrencias(d.Cliente),
		EventosWebhook:  pagamentos.NovoRepositorioDeEventosWebhook(d.Cliente),
		Doadores:        doadorSvc,
		MeioDePagamento: d.MeioDePagamento,
		Remetente:       d.Remetente,
		Auditoria:       registrador,
		Log:             d.Log,
		Relogio:         d.Relogio,
	})
	pagamentosH := pagamentos.NovoManipulador(pagamentosSvc, d.MeioDePagamento)

	// Imagens de evento: URL pública simples, como o bucket legado.
	urlDaImagem := func(caminho string) string {
		if caminho == "" {
			return ""
		}
		return "/api/files/events/" + caminho
	}

	eventosRepo := events.NovoRepositorio(d.Cliente)
	eventosSvc := events.NovoServico(events.Dependencias{
		Repositorio:   eventosRepo,
		Armazenamento: d.Armazenamento,
		Auditoria:     registrador,
		Log:           d.Log,
		TamanhoMaximo: d.Config.TamanhoMaximoArquivoBytes(),
		URLDaImagem:   urlDaImagem,
		Relogio:       d.Relogio,
	})
	eventosH := events.NovoManipulador(eventosSvc)

	transparenciaSvc := transparency.NovoServico(transparency.Dependencias{
		Periodos:      transparency.NovoRepositorioDePeriodos(d.Cliente),
		Lancamentos:   transparency.NovoRepositorioDeLancamentos(d.Cliente),
		Documentos:    transparency.NovoRepositorioDeDocumentos(d.Cliente),
		Armazenamento: d.Armazenamento,
		Auditoria:     registrador,
		Log:           d.Log,
		TamanhoMaximo: d.Config.TamanhoMaximoArquivoBytes(),
		Relogio:       d.Relogio,
	})
	transparenciaH := transparency.NovoManipulador(transparenciaSvc)

	recorrenciasSvc := recurring.NovoServico(recurring.Dependencias{
		Tokens:       recurring.NovoRepositorioDeTokens(d.Cliente),
		Recorrencias: recurring.NovoRepositorioDeRecorrencias(d.Cliente),
		Remetente:    d.Remetente,
		Limitador:    limitador,
		Limites:      d.Config.Limites,
		URLPublica:   d.Config.ServidorHTTP.URLPublica,
		NomeDoSite:   d.Config.Site.Name,
		TTLDoToken:   d.Config.Autenticacao.TempoDeVidaDoTokenDeRecorrencia,
		Auditoria:    registrador,
		Log:          d.Log,
		Relogio:      d.Relogio,
	})
	recorrenciasH := recurring.NovoManipulador(recorrenciasSvc, d.Config.Autenticacao.CookieDeRecorrencia)

	auditoriaRepo := audit2.NovoRepositorio(d.Cliente)
	auditoriaH := audit2.NovoManipulador(audit2.NovoServico(auditoriaRepo))

	// --- aplicação Fiber ----------------------------------------------
	app := fiber.New(fiber.Config{
		AppName:      d.Config.Site.Name,
		ErrorHandler: httpx.ManipuladorDeErros(d.Log),
	})

	app.Use(requestid.New())
	app.Use(httpx.RecuperarPanico(d.Log))
	if d.Config.Telemetria.RegistrarRequisicoes {
		app.Use(httpx.RegistradorDeRequisicoes(d.Log))
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{d.Config.ServidorHTTP.URLPublica},
		AllowCredentials: true,
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
	}))
	app.Use(autenticacao.Autenticar())

	// Rotas públicas de leitura e escrita.
	api := app.Group("/api")
	settings.Registrar(api, configH, autenticacao)
	communities.Registrar(api, comunidadesH, autenticacao)
	contact.Registrar(api, contatoH, autenticacao)
	auth.Registrar(api, authH, autenticacao)
	pagamentos.Registrar(api, pagamentosH)
	events.Registrar(api, eventosH, autenticacao)
	transparency.Registrar(api, transparenciaH, autenticacao)
	recurring.Registrar(api, recorrenciasH)

	// Área administrativa.
	users.Registrar(api, usuariosH, autenticacao)
	donors.Registrar(api, doadorH, autenticacao)
	contributions.Registrar(api, contribuicoesH, autenticacao)
	dashboard.Registrar(api, painelH, autenticacao)
	audit2.Registrar(api, auditoriaH, autenticacao)

	// Imagens de evento, servidas do volume de armazenamento.
	app.Get("/api/files/events/*", func(c fiber.Ctx) error {
		caminho := "events/" + c.Params("*")
		stream, obj, err := d.Armazenamento.Open(c.Context(), caminho)
		if err != nil {
			return httpx.ErroNaoEncontrado("Imagem não encontrada")
		}
		defer func() { _ = stream.Close() }()
		_ = obj
		return c.SendStream(stream)
	})

	// Documentação OpenAPI com interface interativa.
	docs.Registrar(app)

	// Saúde do processo.
	app.Get("/saude", func(c fiber.Ctx) error {
		if err := database.VerificarConexao(c.Context(), d.Cliente); err != nil {
			return httpx.ErroInterno("Banco de dados indisponível", err)
		}
		return httpx.ResponderOK(c, fiber.Map{"ok": true})
	})

	return &Aplicacao{
		Fiber:         app,
		Cliente:       d.Cliente,
		Config:        d.Config,
		Log:           d.Log,
		Armazenamento: d.Armazenamento,
	}, nil
}
