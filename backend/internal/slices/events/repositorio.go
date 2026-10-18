// Package events publica a agenda da paróquia e cuida do upload de imagens.
package events

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/evento"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// Erros de domínio.
var (
	ErrNaoEncontrado = errors.New("evento não encontrado")
	ErrSlugEmUso     = errors.New("já existe um evento com este endereço")
)

// Repositorio é o contrato de persistência dos eventos.
type Repositorio interface {
	// ListarPublicas devolve os eventos publicados e futuros, como a policy
	// `events_public_read_published` fazia.
	ListarPublicas(ctx context.Context, aPartirDe time.Time, limite int) ([]*ent.Evento, error)
	// ObterPublicoPorSlug devolve um evento publicado pelo endereço.
	ObterPublicoPorSlug(ctx context.Context, slug string) (*ent.Evento, error)

	ListarAdmin(ctx context.Context, limite int) ([]*ent.Evento, error)
	Obter(ctx context.Context, id uuid.UUID) (*ent.Evento, error)
	Criar(ctx context.Context, entrada EntradaDeEscrita) (*ent.Evento, error)
	Atualizar(ctx context.Context, id uuid.UUID, entrada EntradaDeEscrita) (*ent.Evento, error)
	DefinirStatus(ctx context.Context, id uuid.UUID, status string, publicadoEm *time.Time) (*ent.Evento, error)
	// SlugExiste garante o endereço único, ignorando o próprio evento.
	SlugExiste(ctx context.Context, slug string, ignorar uuid.UUID) (bool, error)
}

// EntradaDeEscrita é a superfície editável de um evento.
type EntradaDeEscrita struct {
	Titulo          string
	Slug            string
	Descricao       string
	CaminhoDaImagem *string
	IniciaEm        time.Time
	TerminaEm       *time.Time
	Local           string
	Status          string
	IDAutor         *uuid.UUID
	PublicadoEm     *time.Time
}

type repositorio struct{ client *ent.Client }

// NovoRepositorio monta o repositório Ent dos eventos.
func NovoRepositorio(client *ent.Client) Repositorio {
	return &repositorio{client: client}
}

func (r *repositorio) ListarPublicas(ctx context.Context, aPartirDe time.Time, limite int) ([]*ent.Evento, error) {
	return r.client.Evento.Query().
		Where(
			evento.StatusEQ(evento.StatusPublished),
			evento.StartsAtGTE(aPartirDe),
		).
		Order(ent.Asc(evento.FieldStartsAt)).
		Limit(limite).
		All(ctx)
}

func (r *repositorio) ObterPublicoPorSlug(ctx context.Context, slug string) (*ent.Evento, error) {
	linha, err := r.client.Evento.Query().
		Where(
			evento.Slug(slug),
			evento.StatusEQ(evento.StatusPublished),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) ListarAdmin(ctx context.Context, limite int) ([]*ent.Evento, error) {
	return r.client.Evento.Query().
		Order(ent.Desc(evento.FieldStartsAt)).
		Limit(limite).
		All(ctx)
}

func (r *repositorio) Obter(ctx context.Context, id uuid.UUID) (*ent.Evento, error) {
	linha, err := r.client.Evento.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) Criar(ctx context.Context, entrada EntradaDeEscrita) (*ent.Evento, error) {
	criar := r.client.Evento.Create().
		SetTitle(entrada.Titulo).
		SetSlug(entrada.Slug).
		SetDescription(entrada.Descricao).
		SetStartsAt(entrada.IniciaEm).
		SetLocation(entrada.Local).
		SetStatus(evento.Status(entrada.Status))

	criar.SetNillableImagePath(entrada.CaminhoDaImagem)
	criar.SetNillableEndsAt(entrada.TerminaEm)
	criar.SetNillableCreatedBy(entrada.IDAutor)
	criar.SetNillablePublishedAt(entrada.PublicadoEm)

	linha, err := criar.Save(ctx)
	if err != nil {
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrSlugEmUso
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) Atualizar(ctx context.Context, id uuid.UUID, entrada EntradaDeEscrita) (*ent.Evento, error) {
	atualizar := r.client.Evento.UpdateOneID(id).
		SetTitle(entrada.Titulo).
		SetSlug(entrada.Slug).
		SetDescription(entrada.Descricao).
		SetStartsAt(entrada.IniciaEm).
		SetLocation(entrada.Local).
		SetStatus(evento.Status(entrada.Status))

	atualizar.SetNillableImagePath(entrada.CaminhoDaImagem)
	atualizar.SetNillableEndsAt(entrada.TerminaEm)
	atualizar.SetNillableCreatedBy(entrada.IDAutor)
	atualizar.SetNillablePublishedAt(entrada.PublicadoEm)

	linha, err := atualizar.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrSlugEmUso
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) DefinirStatus(ctx context.Context, id uuid.UUID, status string, publicadoEm *time.Time) (*ent.Evento, error) {
	atualizar := r.client.Evento.UpdateOneID(id).
		SetStatus(evento.Status(status))
	if publicadoEm != nil {
		atualizar.SetPublishedAt(*publicadoEm)
	} else {
		atualizar.ClearPublishedAt()
	}

	linha, err := atualizar.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) SlugExiste(ctx context.Context, slug string, ignorar uuid.UUID) (bool, error) {
	consulta := r.client.Evento.Query().Where(evento.Slug(slug))
	if ignorar != uuid.Nil {
		consulta = consulta.Where(evento.IDNEQ(ignorar))
	}
	return consulta.Exist(ctx)
}

// SlugificarProduz o endereço de um evento a partir do título, usando apenas
// caixa-baixa, dígitos e hífens — o mesmo charset do `slug` legado.
func Slugificar(texto string) string {
	var b strings.Builder
	anteriorEhHifen := false
	for _, r := range strings.ToLower(strings.TrimSpace(texto)) {
		switch {
		case (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9'):
			b.WriteRune(r)
			anteriorEhHifen = false
		case r == ' ' || r == '-' || r == '_':
			if !anteriorEhHifen && b.Len() > 0 {
				b.WriteByte('-')
				anteriorEhHifen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
