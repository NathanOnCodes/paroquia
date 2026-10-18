// Package recurring cuida da contribuição recorrente sob a ótica do fiel:
// o link mágico por e-mail, a listagem e o cancelamento sem senha.
package recurring

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/doacaorecorrente"
	"github.com/NathanOnCodes/paroquia/backend/ent/doador"
	"github.com/NathanOnCodes/paroquia/backend/ent/tokenmagicorecorrente"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// Erros de domínio.
var (
	ErrTokenInvalido = errors.New("token inválido ou expirado")
	ErrNaoEncontrado = errors.New("recorrência não encontrada")
)

// EntradaDeToken descreve um link mágico a ser registrado.
type EntradaDeToken struct {
	HashDoToken string
	Email       string
	ExpiraEm    time.Time
}

// RepositorioDeTokens é o contrato dos links mágicos.
type RepositorioDeTokens interface {
	Criar(ctx context.Context, entrada EntradaDeToken) (*ent.TokenMagicoRecorrente, error)
	ObterPorHash(ctx context.Context, hash string) (*ent.TokenMagicoRecorrente, error)
	// Consumir marca o link como usado. Ele continua valendo para a sessão de
	// cookie que abriu, mas não pode ser aberto de novo.
	Consumir(ctx context.Context, id uuid.UUID) error
	// QuantosPorEmail conta os links emitidos para um e-mail na janela.
	QuantosPorEmail(ctx context.Context, email string, desde time.Time) (int, error)
}

// RepositorioDeRecorrencias é o contrato das consultas feitas pelo fiel.
type RepositorioDeRecorrencias interface {
	ObterDoadorPorEmail(ctx context.Context, email string) (*ent.Doador, error)
	ListarPorDoador(ctx context.Context, idDoador uuid.UUID) ([]*ent.DoacaoRecorrente, error)
	ObterPorAssinaturaEDoador(ctx context.Context, idAssinatura string, idDoador uuid.UUID) (*ent.DoacaoRecorrente, error)
	// CancelarNoFimDoPeriodo pede o cancelamento efetivo.
	CancelarNoFimDoPeriodo(ctx context.Context, id uuid.UUID, pedidoEm time.Time) (*ent.DoacaoRecorrente, error)
}

// ---------------------------------------------------------------------------
// Implementação Ent
// ---------------------------------------------------------------------------

type repositorioDeTokens struct{ client *ent.Client }

// NovoRepositorioDeTokens monta o repositório Ent dos links mágicos.
func NovoRepositorioDeTokens(client *ent.Client) RepositorioDeTokens {
	return &repositorioDeTokens{client: client}
}

func (r *repositorioDeTokens) Criar(ctx context.Context, entrada EntradaDeToken) (*ent.TokenMagicoRecorrente, error) {
	return r.client.TokenMagicoRecorrente.Create().
		SetTokenHash(entrada.HashDoToken).
		SetEmail(entrada.Email).
		SetExpiresAt(entrada.ExpiraEm).
		Save(ctx)
}

func (r *repositorioDeTokens) ObterPorHash(ctx context.Context, hash string) (*ent.TokenMagicoRecorrente, error) {
	linha, err := r.client.TokenMagicoRecorrente.Query().
		Where(tokenmagicorecorrente.TokenHash(hash)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrTokenInvalido
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeTokens) Consumir(ctx context.Context, id uuid.UUID) error {
	agora := time.Now().UTC()
	return r.client.TokenMagicoRecorrente.UpdateOneID(id).
		SetConsumedAt(agora).
		Exec(ctx)
}

func (r *repositorioDeTokens) QuantosPorEmail(ctx context.Context, email string, desde time.Time) (int, error) {
	return r.client.TokenMagicoRecorrente.Query().
		Where(
			tokenmagicorecorrente.Email(email),
			tokenmagicorecorrente.CreatedAtGTE(desde),
		).
		Count(ctx)
}

type repositorioDeRecorrencias struct{ client *ent.Client }

// NovoRepositorioDeRecorrencias monta o repositório Ent das consultas do fiel.
func NovoRepositorioDeRecorrencias(client *ent.Client) RepositorioDeRecorrencias {
	return &repositorioDeRecorrencias{client: client}
}

func (r *repositorioDeRecorrencias) ObterDoadorPorEmail(ctx context.Context, email string) (*ent.Doador, error) {
	linha, err := r.client.Doador.Query().
		Where(doador.Email(normalizarEmail(email))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeRecorrencias) ListarPorDoador(ctx context.Context, idDoador uuid.UUID) ([]*ent.DoacaoRecorrente, error) {
	return r.client.DoacaoRecorrente.Query().
		Where(doacaorecorrente.DonorID(idDoador)).
		Order(ent.Desc(doacaorecorrente.FieldCreatedAt)).
		All(ctx)
}

func (r *repositorioDeRecorrencias) ObterPorAssinaturaEDoador(ctx context.Context, idAssinatura string, idDoador uuid.UUID) (*ent.DoacaoRecorrente, error) {
	linha, err := r.client.DoacaoRecorrente.Query().
		Where(
			doacaorecorrente.StripeSubscriptionID(idAssinatura),
			doacaorecorrente.DonorID(idDoador),
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

func (r *repositorioDeRecorrencias) CancelarNoFimDoPeriodo(ctx context.Context, id uuid.UUID, pedidoEm time.Time) (*ent.DoacaoRecorrente, error) {
	linha, err := r.client.DoacaoRecorrente.UpdateOneID(id).
		SetCancelAtPeriodEnd(true).
		SetStatus(doacaorecorrente.StatusCancelAtPeriodEnd).
		SetCancellationRequestedAt(pedidoEm).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		if database.EhViolacaoDeChaveEstrangeira(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
