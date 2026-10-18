package settings

import (
	"context"
	"errors"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/configuracaosite"
)

// rowID é o único id que uma linha de `site_settings` pode ter. O CHECK
// constraint in the migration enforces the same thing.
const IDDaLinha = 1

// ErrNotFound is returned when the singleton row is missing. The service
// recovers from it by serving the defaults, so the public site never breaks
// because of a missing row.
var ErrNaoEncontrado = errors.New("configurações do site não encontradas")

// Repository is the persistence contract. It lives in the consumer package so
// the service never depends on Ent.
type Repositorio interface {
	// Get returns the singleton row.
	Obter(ctx context.Context) (*ent.ConfiguracaoSite, error)
	// Upsert writes the singleton row.
	Gravar(ctx context.Context, in EntradaDeGravacao) (*ent.ConfiguracaoSite, error)
}

// UpsertInput is the full editable surface of the settings row.
type EntradaDeGravacao struct {
	SiteName        string
	SiteTitle       string
	City            string
	Description     string
	SecondaryColor  string
	PixKey          *string
	PixReceiverName *string
	PixReceiverCity *string
	HomeTitle       *string
	HomeDescription *string
}

type repositorio struct{ client *ent.Client }

// NewRepository builds the Ent-backed repository.
func NovoRepositorio(client *ent.Client) Repositorio { return &repositorio{client: client} }

func (r *repositorio) Obter(ctx context.Context) (*ent.ConfiguracaoSite, error) {
	row, err := r.client.ConfiguracaoSite.Query().
		Where(configuracaosite.ID(IDDaLinha)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return row, nil
}

func (r *repositorio) Gravar(ctx context.Context, in EntradaDeGravacao) (*ent.ConfiguracaoSite, error) {
	create := r.client.ConfiguracaoSite.Create().
		SetID(IDDaLinha).
		SetSiteName(in.SiteName).
		SetSiteTitle(in.SiteTitle).
		SetCity(in.City).
		SetDescription(in.Description).
		SetSecondaryColor(in.SecondaryColor)

	if in.PixKey != nil {
		create.SetNillablePixKey(in.PixKey)
	}
	if in.PixReceiverName != nil {
		create.SetNillablePixReceiverName(in.PixReceiverName)
	}
	if in.PixReceiverCity != nil {
		create.SetNillablePixReceiverCity(in.PixReceiverCity)
	}
	if in.HomeTitle != nil {
		create.SetNillableHomeTitle(in.HomeTitle)
	}
	if in.HomeDescription != nil {
		create.SetNillableHomeDescription(in.HomeDescription)
	}

	// Singleton row: the insert becomes an update when it already exists.
	// The nullable Pix/home fields are set or cleared explicitly, so an admin
	// can wipe the Pix key by leaving the form blank.
	upsert := create.
		OnConflictColumns("id").
		UpdateNewValues()

	if in.PixKey != nil {
		upsert.SetPixKey(*in.PixKey)
	} else {
		upsert.ClearPixKey()
	}
	if in.PixReceiverName != nil {
		upsert.SetPixReceiverName(*in.PixReceiverName)
	} else {
		upsert.ClearPixReceiverName()
	}
	if in.PixReceiverCity != nil {
		upsert.SetPixReceiverCity(*in.PixReceiverCity)
	} else {
		upsert.ClearPixReceiverCity()
	}
	if in.HomeTitle != nil {
		upsert.SetHomeTitle(*in.HomeTitle)
	} else {
		upsert.ClearHomeTitle()
	}
	if in.HomeDescription != nil {
		upsert.SetHomeDescription(*in.HomeDescription)
	} else {
		upsert.ClearHomeDescription()
	}

	if err := upsert.Exec(ctx); err != nil {
		return nil, err
	}
	return r.Obter(ctx)
}
