// Package audit expõe a trilha de auditoria para a equipe.
//
// O registro em si é feito por `platform/audit`; esta slice apenas consulta.
package audit

import (
	"context"

	"github.com/gofiber/fiber/v3"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/registroauditoria"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
)

// Repositorio é o contrato de leitura da trilha.
type Repositorio interface {
	Listar(ctx context.Context, limite int) ([]*ent.RegistroAuditoria, error)
	ListarPorEntidade(ctx context.Context, tipo, id string, limite int) ([]*ent.RegistroAuditoria, error)
}

type repositorio struct{ client *ent.Client }

// NovoRepositorio monta o repositório Ent da trilha de auditoria.
func NovoRepositorio(client *ent.Client) Repositorio {
	return &repositorio{client: client}
}

func (r *repositorio) Listar(ctx context.Context, limite int) ([]*ent.RegistroAuditoria, error) {
	return r.client.RegistroAuditoria.Query().
		Order(ent.Desc(registroauditoria.FieldCreatedAt)).
		Limit(limite).
		All(ctx)
}

func (r *repositorio) ListarPorEntidade(ctx context.Context, tipo, id string, limite int) ([]*ent.RegistroAuditoria, error) {
	return r.client.RegistroAuditoria.Query().
		Where(
			registroauditoria.EntityType(tipo),
			registroauditoria.EntityID(id),
		).
		Order(ent.Desc(registroauditoria.FieldCreatedAt)).
		Limit(limite).
		All(ctx)
}

// Resposta é o formato em trânsito de um registro de auditoria.
type Resposta struct {
	ID           string         `json:"id"`
	IDDoAtor     *string        `json:"actor_id"`
	Acao         string         `json:"action"`
	Tipo         string         `json:"entity_type"`
	IDDaEntidade *string        `json:"entity_id"`
	Metadados    map[string]any `json:"metadata"`
	CriadoEm     string         `json:"created_at"`
}

// Servico é a superfície de negócio da trilha de auditoria.
type Servico interface {
	Listar(ctx context.Context, tipo, id string) ([]Resposta, error)
}

type servico struct{ repo Repositorio }

// NovoServico monta o serviço de auditoria.
func NovoServico(repo Repositorio) Servico { return &servico{repo: repo} }

func (s *servico) Listar(ctx context.Context, tipo, id string) ([]Resposta, error) {
	linhas, err := s.repo.Listar(ctx, 300)
	if tipo != "" && id != "" {
		linhas, err = s.repo.ListarPorEntidade(ctx, tipo, id, 300)
	}
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar auditoria", err)
	}

	saida := make([]Resposta, 0, len(linhas))
	for _, l := range linhas {
		saida = append(saida, Resposta{
			ID:           l.ID.String(),
			IDDoAtor:     entx.UUIDParaTextoOpcional(l.ActorID),
			Acao:         l.Action,
			Tipo:         l.EntityType,
			IDDaEntidade: l.EntityID,
			Metadados:    l.Metadata,
			CriadoEm:     entx.DataParaTexto(l.CreatedAt),
		})
	}
	return saida, nil
}

// Manipulador é a superfície HTTP da trilha de auditoria.
type Manipulador struct{ svc Servico }

// NovoManipulador monta o manipulador de auditoria.
func NovoManipulador(svc Servico) *Manipulador { return &Manipulador{svc: svc} }

// Listar responde GET /api/admin/audit-logs.
func (h *Manipulador) Listar(c fiber.Ctx) error {
	saida, err := h.svc.Listar(c.Context(), c.Query("entity_type"), c.Query("entity_id"))
	if err != nil {
		return err
	}
	return httpx.ResponderOK(c, saida)
}

// Registrar monta as rotas da trilha de auditoria.
//
//	GET /api/admin/audit-logs   admin
func Registrar(api fiber.Router, h *Manipulador, autenticacao *authx.Autenticacao) {
	api.Get("/admin/audit-logs",
		autenticacao.ExigirAutenticacao(),
		autenticacao.ExigirPapel(authx.PapelAdmin),
		autenticacao.ExigirSenhaAtualizada(),
		h.Listar,
	)
}
