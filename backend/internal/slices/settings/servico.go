package settings

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Service is the business surface of the settings slice.
type Servico interface {
	// GetPublic returns the settings, falling back to the defaults when the
	// row is missing or unreadable. The public site must never fail because of
	// a configuration problem.
	ObterPublico(ctx context.Context) Resposta
	// Update rewrites the settings row and records the audit entry.
	Atualizar(ctx context.Context, actor authx.UsuarioAutenticado, req RequisicaoDeAtualizacao) (Resposta, error)
}

type servico struct {
	repo  Repositorio
	audit audit.RegistradorDeAuditoria
	log   *slog.Logger
}

// NewService wires the settings service.
func NovoServico(repo Repositorio, auditor audit.RegistradorDeAuditoria, log *slog.Logger) Servico {
	if log == nil {
		log = slog.Default()
	}
	return &servico{repo: repo, audit: auditor, log: log}
}

func (s *servico) ObterPublico(ctx context.Context) Resposta {
	row, err := s.repo.Obter(ctx)
	if err != nil {
		if !errors.Is(err, ErrNaoEncontrado) {
			s.log.WarnContext(ctx, "configurações ilegíveis; usando padrões",
				slog.String("error", err.Error()))
		}
		return Padroes
	}
	return paraResposta(row)
}

func (s *servico) Atualizar(ctx context.Context, actor authx.UsuarioAutenticado, req RequisicaoDeAtualizacao) (Resposta, error) {
	in, err := req.paraEntrada()
	if err != nil {
		return Resposta{}, err
	}

	row, err := s.repo.Gravar(ctx, in)
	if err != nil {
		return Resposta{}, httpx.ErroInterno("Falha ao salvar configurações", err)
	}

	s.audit.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID:    actor.ID,
		Action:     "update_site_settings",
		EntityType: "site_settings",
		EntityID:   "1",
	})

	return paraResposta(row), nil
}

// toInput normalises the request: blanks become NULL in the Pix and home
// fields, and the Pix key is validated as a whole when present.
func (r RequisicaoDeAtualizacao) paraEntrada() (EntradaDeGravacao, error) {
	pixKey := strings.TrimSpace(r.PixKey)
	var pixKeyPtr *string
	if pixKey != "" {
		normalised, ok := validarChavePix(pixKey)
		if !ok {
			return EntradaDeGravacao{}, httpx.ErroDeValidacao("Chave Pix inválida",
				httpx.ErroDeCampo{Campo: "pix_key", Mensagem: "formato de chave Pix não reconhecido"})
		}
		if strings.TrimSpace(r.PixReceiverName) == "" || strings.TrimSpace(r.PixReceiverCity) == "" {
			return EntradaDeGravacao{}, httpx.ErroDeValidacao("Chave Pix exige nome e cidade do recebedor",
				httpx.ErroDeCampo{Campo: "pix_receiver_name", Mensagem: "obrigatório quando há chave Pix"},
				httpx.ErroDeCampo{Campo: "pix_receiver_city", Mensagem: "obrigatório quando há chave Pix"})
		}
		pixKeyPtr = &normalised
	}

	return EntradaDeGravacao{
		SiteName:        strings.TrimSpace(r.SiteName),
		SiteTitle:       strings.TrimSpace(r.SiteTitle),
		City:            strings.TrimSpace(r.City),
		Description:     strings.TrimSpace(r.Description),
		SecondaryColor:  strings.TrimSpace(r.SecondaryColor),
		PixKey:          pixKeyPtr,
		PixReceiverName: opcionalSemEspacos(r.PixReceiverName),
		PixReceiverCity: opcionalSemEspacos(r.PixReceiverCity),
		HomeTitle:       opcionalSemEspacos(r.HomeTitle),
		HomeDescription: opcionalSemEspacos(r.HomeDescription),
	}, nil
}

func paraResposta(row *ent.ConfiguracaoSite) Resposta {
	return Resposta{
		SiteName:        row.SiteName,
		SiteTitle:       row.SiteTitle,
		City:            row.City,
		Description:     row.Description,
		SecondaryColor:  row.SecondaryColor,
		PixKey:          row.PixKey,
		PixReceiverName: row.PixReceiverName,
		PixReceiverCity: row.PixReceiverCity,
		HomeTitle:       row.HomeTitle,
		HomeDescription: row.HomeDescription,
	}
}

// nullableTrim returns nil for blank input so the column stores NULL.
func opcionalSemEspacos(s string) *string {
	t := strings.TrimSpace(s)
	if t == "" {
		return nil
	}
	return &t
}
