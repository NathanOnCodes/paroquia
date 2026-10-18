// Package donors manages the faithful who contribute.
//
// Além do CRUD administrativo, ela detém a regra de reconhecer um doador
// rule used by the payment flow: a donor is keyed by e-mail and their profile
// is refreshed on every donation.
package donors

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/doador"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// ErrNotFound is returned for an unknown donor.
var ErrNaoEncontrado = errors.New("dizimista não encontrado")

// ErrEmailTaken is returned when another donor already uses the e-mail.
var ErrEmailTaken = errors.New("já existe um dizimista com este e-mail")

// Repository is the persistence contract.
type Repositorio interface {
	Listar(ctx context.Context, search string, limit int) ([]*ent.Doador, error)
	Obter(ctx context.Context, id uuid.UUID) (*ent.Doador, error)
	ObterPorEmail(ctx context.Context, email string) (*ent.Doador, error)
	Criar(ctx context.Context, in EntradaDeEscrita) (*ent.Doador, error)
	Atualizar(ctx context.Context, id uuid.UUID, in EntradaDeEscrita) (*ent.Doador, error)
	DefinirAtivo(ctx context.Context, id uuid.UUID, active bool) (*ent.Doador, error)
	// UpsertByEmail reproduz o `upsertDonor` legado: atualiza quando o e-mail
	// known, insert otherwise.
	GravarPorEmail(ctx context.Context, in EntradaDeEscrita) (*ent.Doador, error)
	EmailExiste(ctx context.Context, email string, exclude uuid.UUID) (bool, error)
}

// WriteInput is the editable surface of a donor.
type EntradaDeEscrita struct {
	FullName       string
	Email          string
	Phone          *string
	City           string
	Neighborhood   string
	CommunityID    *uuid.UUID
	CommunityName  *string
	PrivacyConsent bool
	IsActive       bool
}

type repositorio struct{ client *ent.Client }

// NewRepository builds the Ent-backed repository.
func NovoRepositorio(client *ent.Client) Repositorio { return &repositorio{client: client} }

func (r *repositorio) Listar(ctx context.Context, search string, limit int) ([]*ent.Doador, error) {
	q := r.client.Doador.Query().
		WithCommunity().
		Order(ent.Desc(doador.FieldCreatedAt))
	if limit > 0 {
		q = q.Limit(limit)
	}
	if s := strings.TrimSpace(search); s != "" {
		q = q.Where(doador.Or(
			doador.FullNameContainsFold(s),
			doador.EmailContainsFold(s),
			doador.PhoneContainsFold(s),
		))
	}
	return q.All(ctx)
}

func (r *repositorio) Obter(ctx context.Context, id uuid.UUID) (*ent.Doador, error) {
	row, err := r.client.Doador.Query().
		Where(doador.ID(id)).
		WithCommunity().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) ObterPorEmail(ctx context.Context, email string) (*ent.Doador, error) {
	row, err := r.client.Doador.Query().
		Where(doador.Email(strings.ToLower(strings.TrimSpace(email)))).
		WithCommunity().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) Criar(ctx context.Context, in EntradaDeEscrita) (*ent.Doador, error) {
	create := r.client.Doador.Create().
		SetFullName(in.FullName).
		SetEmail(normalizarEmail(in.Email)).
		SetCity(in.City).
		SetNeighborhood(in.Neighborhood).
		SetIsActive(in.IsActive)
	create = aplicarEscrita(create, in)

	row, err := create.Save(ctx)
	if err != nil {
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) Atualizar(ctx context.Context, id uuid.UUID, in EntradaDeEscrita) (*ent.Doador, error) {
	update := r.client.Doador.UpdateOneID(id).
		SetFullName(in.FullName).
		SetEmail(normalizarEmail(in.Email)).
		SetCity(in.City).
		SetNeighborhood(in.Neighborhood).
		SetIsActive(in.IsActive)
	update = aplicarEscritaNaAtualizacao(update, in)

	row, err := update.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrEmailTaken
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) DefinirAtivo(ctx context.Context, id uuid.UUID, active bool) (*ent.Doador, error) {
	row, err := r.client.Doador.UpdateOneID(id).SetIsActive(active).Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) GravarPorEmail(ctx context.Context, in EntradaDeEscrita) (*ent.Doador, error) {
	existing, err := r.ObterPorEmail(ctx, in.Email)
	switch {
	case err == nil:
		return r.Atualizar(ctx, existing.ID, in)
	case errors.Is(err, ErrNaoEncontrado):
		return r.Criar(ctx, in)
	default:
		return nil, err
	}
}

func (r *repositorio) EmailExiste(ctx context.Context, email string, exclude uuid.UUID) (bool, error) {
	q := r.client.Doador.Query().
		Where(doador.Email(normalizarEmail(email)))
	if exclude != uuid.Nil {
		q = q.Where(doador.IDNEQ(exclude))
	}
	return q.Exist(ctx)
}

// applyWrite sets the nullable columns on a create.
//
// The community rule is the legacy one: a donor points at a registered
// community OR names a free-form one, never both. The SQL CHECK constraint
// enforces the same invariant.
func aplicarEscrita(create *ent.DoadorCreate, in EntradaDeEscrita) *ent.DoadorCreate {
	if in.Phone != nil {
		create.SetPhone(*in.Phone)
	}
	if in.CommunityID != nil {
		create.SetCommunityID(*in.CommunityID)
	}
	if in.CommunityName != nil {
		create.SetCommunityName(*in.CommunityName)
	}
	if in.PrivacyConsent {
		now := agoraUTC()
		create.SetPrivacyConsentAt(now)
	}
	return create
}

func aplicarEscritaNaAtualizacao(update *ent.DoadorUpdateOne, in EntradaDeEscrita) *ent.DoadorUpdateOne {
	if in.Phone != nil {
		update.SetPhone(*in.Phone)
	} else {
		update.ClearPhone()
	}
	if in.CommunityID != nil {
		update.SetCommunityID(*in.CommunityID)
	} else {
		update.ClearCommunityID()
	}
	if in.CommunityName != nil {
		update.SetCommunityName(*in.CommunityName)
	} else {
		update.ClearCommunityName()
	}
	if in.PrivacyConsent {
		now := agoraUTC()
		update.SetPrivacyConsentAt(now)
	}
	return update
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

// nowUTC is a seam: it can be replaced in tests to make consent timestamps
// deterministic.
var agoraUTC = func() time.Time { return time.Now().UTC() }
