package contact

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/config"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/mail"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/ratelimit"
)

// Service is the business surface of the contact slice.
type Servico interface {
	// Submit stores a public message and notifies the parish office.
	Enviar(ctx context.Context, req RequisicaoDeCriacao, clientIP string) error
	// ListAdmin returns the inbox.
	ListarAdmin(ctx context.Context) ([]Resposta, error)
	// SetStatus moves a message through the workflow.
	DefinirStatus(ctx context.Context, actor authx.UsuarioAutenticado, id string, status string) (Resposta, error)
}

type servico struct {
	repo     Repositorio
	audit    audit.RegistradorDeAuditoria
	mailer   mail.RemetenteDeEmail
	limit    *ratelimit.Limitador
	cfg      config.Limites
	mailCfg  config.ConfiguracaoEmail
	siteName string
	log      *slog.Logger
}

// Deps bundles the constructor dependencies of the contact service.
type Dependencias struct {
	Repo              Repositorio
	Audit             audit.RegistradorDeAuditoria
	Mailer            mail.RemetenteDeEmail
	Limitador         *ratelimit.Limitador
	Limites           config.Limites
	ConfiguracaoEmail config.ConfiguracaoEmail
	SiteName          string
	Registrar         *slog.Logger
}

// NewService wires the contact service.
func NovoServico(d Dependencias) Servico {
	if d.Registrar == nil {
		d.Registrar = slog.Default()
	}
	if d.SiteName == "" {
		d.SiteName = "Paróquia"
	}
	return &servico{
		repo:     d.Repo,
		audit:    d.Audit,
		mailer:   d.Mailer,
		limit:    d.Limitador,
		cfg:      d.Limites,
		mailCfg:  d.ConfiguracaoEmail,
		siteName: d.SiteName,
		log:      d.Registrar,
	}
}

func (s *servico) Enviar(ctx context.Context, req RequisicaoDeCriacao, clientIP string) error {
	// The privacy consent is mandatory: it is the legal basis of the storage.
	if !req.PrivacyConsent {
		return httpx.ErroDeValidacao("É necessário aceitar o aviso de privacidade",
			httpx.ErroDeCampo{Campo: "privacy_consent", Mensagem: "consentimento obrigatório"})
	}

	decision, err := s.limit.Permitir(ctx, ratelimit.TipoContato, authx.HashDoToken(clientIP),
		s.cfg.MaximoDeContatos, s.cfg.JanelaDeContatos)
	if err != nil {
		// A limiter failure must not block the faithful; log and continue.
		s.log.WarnContext(ctx, "rate limit indisponível", slog.String("error", err.Error()))
	} else if !decision.Permitido {
		return httpx.ErroLimiteExcedido("Muitas solicitações. Tente novamente mais tarde.")
	}

	in := EntradaDeCriacao{
		Name:     req.Name,
		Email:    req.Email,
		Mensagem: req.Mensagem,
	}
	if req.Phone != "" {
		phone := req.Phone
		in.Phone = &phone
	}
	if clientIP != "" {
		hash := authx.HashDoToken(clientIP)
		in.IPHash = &hash
	}

	row, err := s.repo.Criar(ctx, in)
	if err != nil {
		return httpx.ErroInterno("Falha ao enviar mensagem", err)
	}

	s.notificarAdmin(ctx, row)
	return nil
}

func (s *servico) ListarAdmin(ctx context.Context) ([]Resposta, error) {
	rows, err := s.repo.Listar(ctx, 200)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar mensagens", err)
	}
	return paraRespostas(rows), nil
}

func (s *servico) DefinirStatus(ctx context.Context, actor authx.UsuarioAutenticado, id string, status string) (Resposta, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return Resposta{}, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}

	row, err := s.repo.DefinirStatus(ctx, uid, status)
	if err != nil {
		if errors.Is(err, ErrNaoEncontrado) {
			return Resposta{}, httpx.ErroNaoEncontrado("Mensagem não encontrada")
		}
		return Resposta{}, httpx.ErroInterno("Falha ao atualizar mensagem", err)
	}

	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: actor.ID, Action: "update_contact_status",
		EntityType: "contact_message", EntityID: row.ID.String(),
		Metadata: map[string]any{"status": status},
	})
	return paraResposta(row), nil
}

// notifyAdmin sends the notification e-mail. Failure is logged and swallowed:
// the message is already stored and the parish can still see it in the inbox.
func (s *servico) notificarAdmin(ctx context.Context, row *ent.MensagemContato) {
	if !s.mailCfg.Habilitado || s.mailCfg.NotificarAdminEm == "" {
		return
	}

	phone := "Não informado"
	if row.Phone != nil && *row.Phone != "" {
		phone = *row.Phone
	}

	html := fmt.Sprintf(`
<h2>Nova mensagem no site da %s</h2>
<p><strong>Nome:</strong> %s</p>
<p><strong>E-mail:</strong> %s</p>
<p><strong>Telefone:</strong> %s</p>
<p><strong>Mensagem:</strong></p>
<p>%s</p>
`, mail.EscaparHTML(s.siteName), mail.EscaparHTML(row.Name),
		mail.EscaparHTML(row.Email), mail.EscaparHTML(phone),
		mail.EscaparHTML(row.Message))

	msg := mail.MensagemDeEmail{
		To:      s.mailCfg.NotificarAdminEm,
		Subject: fmt.Sprintf("Nova mensagem no site da %s", s.siteName),
		HTML:    html,
	}

	// Send on a detached context so a slow e-mail never delays the answer.
	sendCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 15*time.Second)
	defer cancel()
	if err := s.mailer.Enviar(sendCtx, msg); err != nil {
		s.log.WarnContext(ctx, "falha ao notificar administrador",
			slog.String("error", err.Error()))
	}
}
