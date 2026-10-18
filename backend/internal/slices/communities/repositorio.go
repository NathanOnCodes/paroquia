// Package communities manages the parish communities that receive tithes.
//
// The public list is restricted to active communities — the equivalent of the
// policy `communities_public_read` legada — a equipe vê e edita todas.
package communities

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/comunidade"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// ErrNotFound is returned when a community does not exist.
var ErrNaoEncontrado = errors.New("comunidade não encontrada")

// ErrNameTaken is returned when a name collides with another community.
var ErrNameTaken = errors.New("já existe uma comunidade com este nome")

// Repository is the persistence contract.
type Repositorio interface {
	// ListActive returns the communities the faithful may pick from.
	ListarAtivas(ctx context.Context) ([]*ent.Comunidade, error)
	// ListAll returns every community, for the admin screen.
	ListarTodas(ctx context.Context) ([]*ent.Comunidade, error)
	// Get returns one community by id.
	Obter(ctx context.Context, id uuid.UUID) (*ent.Comunidade, error)
	// Create inserts a community.
	Criar(ctx context.Context, in EntradaDeEscrita) (*ent.Comunidade, error)
	// Update rewrites a community.
	Atualizar(ctx context.Context, id uuid.UUID, in EntradaDeEscrita) (*ent.Comunidade, error)
	// SetActive toggles the community's visibility.
	DefinirAtivo(ctx context.Context, id uuid.UUID, active bool) (*ent.Comunidade, error)
	// NameExists reports whether another community already uses the name.
	NomeExiste(ctx context.Context, name string, exclude uuid.UUID) (bool, error)
}

// WriteInput is the editable surface of a community.
type EntradaDeEscrita struct {
	Name        string
	Description string
	City        string
	IsActive    bool
}

type repositorio struct{ client *ent.Client }

// NewRepository builds the Ent-backed repository.
func NovoRepositorio(client *ent.Client) Repositorio { return &repositorio{client: client} }

func (r *repositorio) ListarAtivas(ctx context.Context) ([]*ent.Comunidade, error) {
	return r.client.Comunidade.Query().
		Where(comunidade.IsActive(true)).
		Order(ent.Asc(comunidade.FieldName)).
		All(ctx)
}

func (r *repositorio) ListarTodas(ctx context.Context) ([]*ent.Comunidade, error) {
	return r.client.Comunidade.Query().
		Order(ent.Asc(comunidade.FieldName)).
		All(ctx)
}

func (r *repositorio) Obter(ctx context.Context, id uuid.UUID) (*ent.Comunidade, error) {
	row, err := r.client.Comunidade.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) Criar(ctx context.Context, in EntradaDeEscrita) (*ent.Comunidade, error) {
	row, err := r.client.Comunidade.Create().
		SetName(in.Name).
		SetDescription(in.Description).
		SetCity(in.City).
		SetIsActive(in.IsActive).
		Save(ctx)
	if err != nil {
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrNameTaken
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) Atualizar(ctx context.Context, id uuid.UUID, in EntradaDeEscrita) (*ent.Comunidade, error) {
	row, err := r.client.Comunidade.UpdateOneID(id).
		SetName(in.Name).
		SetDescription(in.Description).
		SetCity(in.City).
		SetIsActive(in.IsActive).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrNameTaken
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) DefinirAtivo(ctx context.Context, id uuid.UUID, active bool) (*ent.Comunidade, error) {
	row, err := r.client.Comunidade.UpdateOneID(id).
		SetIsActive(active).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) NomeExiste(ctx context.Context, name string, exclude uuid.UUID) (bool, error) {
	q := r.client.Comunidade.Query().Where(comunidade.Name(name))
	if exclude != uuid.Nil {
		q = q.Where(comunidade.IDNEQ(exclude))
	}
	return q.Exist(ctx)
}
