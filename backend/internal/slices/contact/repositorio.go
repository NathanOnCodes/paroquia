// Package contact handles the public contact form and the inbox that the
// parish staff works from.
//
// The public endpoint is rate limited and requires an explicit privacy
// consent, exactly like the legacy system.
package contact

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/mensagemcontato"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// ErrNotFound is returned for an unknown message.
var ErrNaoEncontrado = errors.New("mensagem não encontrada")

// Repository is the persistence contract.
type Repositorio interface {
	Criar(ctx context.Context, in EntradaDeCriacao) (*ent.MensagemContato, error)
	Listar(ctx context.Context, limit int) ([]*ent.MensagemContato, error)
	DefinirStatus(ctx context.Context, id uuid.UUID, status string) (*ent.MensagemContato, error)
}

// CreateInput is one incoming message.
type EntradaDeCriacao struct {
	Name     string
	Email    string
	Phone    *string
	Mensagem string
	IPHash   *string
}

type repositorio struct{ client *ent.Client }

// NewRepository builds the Ent-backed repository.
func NovoRepositorio(client *ent.Client) Repositorio { return &repositorio{client: client} }

func (r *repositorio) Criar(ctx context.Context, in EntradaDeCriacao) (*ent.MensagemContato, error) {
	create := r.client.MensagemContato.Create().
		SetName(in.Name).
		SetEmail(in.Email).
		SetMessage(in.Mensagem)
	if in.Phone != nil {
		create.SetPhone(*in.Phone)
	}
	if in.IPHash != nil {
		create.SetIPHash(*in.IPHash)
	}
	return create.Save(ctx)
}

func (r *repositorio) Listar(ctx context.Context, limit int) ([]*ent.MensagemContato, error) {
	return r.client.MensagemContato.Query().
		WithAssignee().
		Order(ent.Desc(mensagemcontato.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}

func (r *repositorio) DefinirStatus(ctx context.Context, id uuid.UUID, status string) (*ent.MensagemContato, error) {
	row, err := r.client.MensagemContato.UpdateOneID(id).
		SetStatus(mensagemcontato.Status(status)).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) || database.EhViolacaoDeChaveEstrangeira(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return row, nil
}
