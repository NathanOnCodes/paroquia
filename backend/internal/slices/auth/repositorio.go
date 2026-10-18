// Package auth implementa a autenticação administrativa: login com senha,
// sessão rotativa por refresh token e troca de senha.
//
// Substitui o `supabase.auth` legado (signInWithPassword, signOut,
// exchangeCodeForSession) por credenciais próprias armazenadas em `users`.
package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/tokenatualizacao"
	"github.com/NathanOnCodes/paroquia/backend/ent/usuario"
)

var (
	// ErrCredenciaisInvalidas é devolvido para qualquer falha de autenticação.
	// A mensagem é genérica de propósito: o login não pode revelar se o e-mail
	// existe.
	ErrCredenciaisInvalidas = errors.New("e-mail ou senha inválidos")
	// ErrSessaoInvalida é devolvido quando o refresh token não pode ser usado.
	ErrSessaoInvalida = errors.New("sessão inválida ou expirada")
	// ErrUsuarioInativo é devolvido para contas desativadas.
	ErrUsuarioInativo = errors.New("usuário inativo")
	// ErrNaoEncontrado é devolvido quando o usuário não existe.
	ErrNaoEncontrado = errors.New("usuário não encontrado")
	// ErrSenhaIncorreta é devolvido quando a senha atual não confere.
	ErrSenhaIncorreta = errors.New("senha atual incorreta")
	// ErrTrocaDeSenhaPendente indica que a senha temporária ainda não foi
	// substituída.
	ErrTrocaDeSenhaPendente = errors.New("é necessário alterar a senha")
)

// RepositorioDeUsuarios é o contrato de persistência dos usuários.
type RepositorioDeUsuarios interface {
	ObterPorEmail(ctx context.Context, email string) (*ent.Usuario, error)
	ObterPorID(ctx context.Context, id uuid.UUID) (*ent.Usuario, error)
	// RegistrarAcesso grava o carimbo do último login.
	RegistrarAcesso(ctx context.Context, id uuid.UUID) error
	// DefinirSenha troca o hash e marca a senha como já alterada.
	DefinirSenha(ctx context.Context, id uuid.UUID, hash string) error
}

// EntradaDeSessao descreve um refresh token a ser registrado.
type EntradaDeSessao struct {
	HashDoToken string
	IDDoUsuario uuid.UUID
	IDDaFamilia uuid.UUID
	ExpiraEm    time.Time
}

// RepositorioDeSessoes é o contrato de persistência dos refresh tokens.
type RepositorioDeSessoes interface {
	Criar(ctx context.Context, entrada EntradaDeSessao) (*ent.TokenAtualizacao, error)
	ObterPorHash(ctx context.Context, hash string) (*ent.TokenAtualizacao, error)
	// Revogar invalida um token específico.
	Revogar(ctx context.Context, id uuid.UUID) error
	// RevogarFamilia invalida todos os tokens derivados da mesma emissão. É o
	// que acontece quando um token vazado é reutilizado.
	RevogarFamilia(ctx context.Context, idDaFamilia uuid.UUID) error
}

// ---------------------------------------------------------------------------
// Implementação Ent
//
// Os nomes de métodos e constantes abaixo espelham as colunas do banco, que
// permanecem em inglês por serem o contrato com os dados legados.
// ---------------------------------------------------------------------------

type repositorioDeUsuarios struct{ client *ent.Client }

// NovoRepositorioDeUsuarios monta o repositório Ent dos usuários.
func NovoRepositorioDeUsuarios(client *ent.Client) RepositorioDeUsuarios {
	return &repositorioDeUsuarios{client: client}
}

func (r *repositorioDeUsuarios) ObterPorEmail(ctx context.Context, email string) (*ent.Usuario, error) {
	linha, err := r.client.Usuario.Query().
		Where(usuario.Email(normalizarEmail(email))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrCredenciaisInvalidas
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeUsuarios) ObterPorID(ctx context.Context, id uuid.UUID) (*ent.Usuario, error) {
	linha, err := r.client.Usuario.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeUsuarios) RegistrarAcesso(ctx context.Context, id uuid.UUID) error {
	agora := time.Now().UTC()
	return r.client.Usuario.UpdateOneID(id).
		SetLastLoginAt(agora).
		Exec(ctx)
}

func (r *repositorioDeUsuarios) DefinirSenha(ctx context.Context, id uuid.UUID, hash string) error {
	agora := time.Now().UTC()
	return r.client.Usuario.UpdateOneID(id).
		SetPasswordHash(hash).
		SetPasswordChangedAt(agora).
		SetMustChangePassword(false).
		Exec(ctx)
}

type repositorioDeSessoes struct{ client *ent.Client }

// NovoRepositorioDeSessoes monta o repositório Ent das sessões.
func NovoRepositorioDeSessoes(client *ent.Client) RepositorioDeSessoes {
	return &repositorioDeSessoes{client: client}
}

func (r *repositorioDeSessoes) Criar(ctx context.Context, entrada EntradaDeSessao) (*ent.TokenAtualizacao, error) {
	return r.client.TokenAtualizacao.Create().
		SetTokenHash(entrada.HashDoToken).
		SetUserID(entrada.IDDoUsuario).
		SetFamilyID(entrada.IDDaFamilia).
		SetExpiresAt(entrada.ExpiraEm).
		Save(ctx)
}

func (r *repositorioDeSessoes) ObterPorHash(ctx context.Context, hash string) (*ent.TokenAtualizacao, error) {
	linha, err := r.client.TokenAtualizacao.Query().
		Where(tokenatualizacao.TokenHash(hash)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrSessaoInvalida
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeSessoes) Revogar(ctx context.Context, id uuid.UUID) error {
	agora := time.Now().UTC()
	return r.client.TokenAtualizacao.UpdateOneID(id).
		SetRevokedAt(agora).
		Exec(ctx)
}

func (r *repositorioDeSessoes) RevogarFamilia(ctx context.Context, idDaFamilia uuid.UUID) error {
	agora := time.Now().UTC()
	return r.client.TokenAtualizacao.Update().
		Where(tokenatualizacao.FamilyID(idDaFamilia)).
		SetRevokedAt(agora).
		Exec(ctx)
}

func normalizarEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
