// Package users cuida das contas administrativas: criação, edição,
// ativação/desativação e redefinição de senha.
//
// Substitui `auth.admin.createUser`/`deleteUser` do Supabase. A troca de papel
// tem trava no banco (trigger `prevent_role_change`) e aqui no serviço.
package users

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/usuario"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// Erros de domínio.
var (
	ErrNaoEncontrado   = errors.New("usuário não encontrado")
	ErrEmailEmUso      = errors.New("já existe um usuário com este e-mail")
	ErrAutoDesativacao = errors.New("você não pode desativar a própria conta")
	ErrUltimoAdmin     = errors.New("não é possível remover o último administrador")
)

// Repositorio é o contrato de persistência dos usuários administrativos.
type Repositorio interface {
	Listar(ctx context.Context) ([]*ent.Usuario, error)
	Obter(ctx context.Context, id uuid.UUID) (*ent.Usuario, error)
	ObterPorEmail(ctx context.Context, email string) (*ent.Usuario, error)
	Criar(ctx context.Context, entrada EntradaDeEscrita) (*ent.Usuario, error)
	Atualizar(ctx context.Context, id uuid.UUID, entrada EntradaDeEscrita) (*ent.Usuario, error)
	DefinirAtivo(ctx context.Context, id uuid.UUID, ativo bool) (*ent.Usuario, error)
	DefinirSenha(ctx context.Context, id uuid.UUID, hash string) error
	// QuantidadeDeAdmins conta administradores ativos.
	QuantidadeDeAdmins(ctx context.Context) (int, error)
}

// EntradaDeEscrita é a superfície editável de um usuário.
type EntradaDeEscrita struct {
	Email        string
	NomeCompleto string
	HashSenha    string
	Papel        string
	Ativo        bool
	// ExigeTrocaDeSenha marca a conta para trocar a senha no primeiro acesso.
	ExigeTrocaDeSenha bool
}

type repositorio struct{ client *ent.Client }

// NovoRepositorio monta o repositório Ent dos usuários administrativos.
func NovoRepositorio(client *ent.Client) Repositorio {
	return &repositorio{client: client}
}

func (r *repositorio) Listar(ctx context.Context) ([]*ent.Usuario, error) {
	return r.client.Usuario.Query().
		Order(ent.Asc(usuario.FieldFullName)).
		All(ctx)
}

func (r *repositorio) Obter(ctx context.Context, id uuid.UUID) (*ent.Usuario, error) {
	linha, err := r.client.Usuario.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) ObterPorEmail(ctx context.Context, email string) (*ent.Usuario, error) {
	linha, err := r.client.Usuario.Query().
		Where(usuario.Email(normalizarEmail(email))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) Criar(ctx context.Context, entrada EntradaDeEscrita) (*ent.Usuario, error) {
	linha, err := r.client.Usuario.Create().
		SetEmail(normalizarEmail(entrada.Email)).
		SetFullName(entrada.NomeCompleto).
		SetPasswordHash(entrada.HashSenha).
		SetRole(usuario.Role(entrada.Papel)).
		SetIsActive(entrada.Ativo).
		SetMustChangePassword(entrada.ExigeTrocaDeSenha).
		Save(ctx)
	if err != nil {
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrEmailEmUso
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) Atualizar(ctx context.Context, id uuid.UUID, entrada EntradaDeEscrita) (*ent.Usuario, error) {
	// A troca de papel é feita numa transação com o ator publicado, para que o
	// trigger `prevent_role_change` possa validar quem está mudando.
	linha, err := r.client.Usuario.UpdateOneID(id).
		SetFullName(entrada.NomeCompleto).
		SetRole(usuario.Role(entrada.Papel)).
		SetIsActive(entrada.Ativo).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		if database.EhViolacaoDeUnicidade(err) {
			return nil, ErrEmailEmUso
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) DefinirAtivo(ctx context.Context, id uuid.UUID, ativo bool) (*ent.Usuario, error) {
	linha, err := r.client.Usuario.UpdateOneID(id).
		SetIsActive(ativo).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorio) DefinirSenha(ctx context.Context, id uuid.UUID, hash string) error {
	return r.client.Usuario.UpdateOneID(id).
		SetPasswordHash(hash).
		SetMustChangePassword(true).
		Exec(ctx)
}

func (r *repositorio) QuantidadeDeAdmins(ctx context.Context) (int, error) {
	return r.client.Usuario.Query().
		Where(
			usuario.RoleEQ(usuario.RoleAdmin),
			usuario.IsActive(true),
		).
		Count(ctx)
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
