package users

import (
	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// RequisicaoDeEscrita é o corpo de criação e edição de usuário.
type RequisicaoDeEscrita struct {
	Email        string `json:"email" validate:"required,email,max=200"`
	NomeCompleto string `json:"full_name" validate:"required,min=3,max=200"`
	// Senha é exigida apenas na criação.
	Senha string `json:"password" validate:"omitempty,min=8,max=200"`
	Papel string `json:"role" validate:"required,oneof=admin assistente"`
	Ativo bool   `json:"is_active"`
}

// RequisicaoDeSenha redefine a senha de um usuário (somente admin).
type RequisicaoDeSenha struct {
	NovaSenha string `json:"password" validate:"required,min=8,max=200"`
}

// RequisicaoDeStatus ativa/desativa uma conta.
type RequisicaoDeStatus struct {
	Ativo bool `json:"is_active"`
}

// Resposta é o formato em trânsito de um usuário administrativo.
type Resposta struct {
	ID               string  `json:"id"`
	Email            string  `json:"email"`
	NomeCompleto     string  `json:"full_name"`
	Papel            string  `json:"role"`
	Ativo            bool    `json:"is_active"`
	DeveTrocarASenha bool    `json:"must_change_password"`
	UltimoAcessoEm   *string `json:"last_login_at"`
	CriadoEm         string  `json:"created_at"`
	AtualizadoEm     string  `json:"updated_at"`
}

func paraResposta(linha *ent.Usuario) Resposta {
	return Resposta{
		ID:               linha.ID.String(),
		Email:            linha.Email,
		NomeCompleto:     linha.FullName,
		Papel:            string(linha.Role),
		Ativo:            linha.IsActive,
		DeveTrocarASenha: linha.MustChangePassword,
		UltimoAcessoEm:   entx.DataParaTextoOpcional(linha.LastLoginAt),
		CriadoEm:         entx.DataParaTexto(linha.CreatedAt),
		AtualizadoEm:     entx.DataParaTexto(linha.UpdatedAt),
	}
}

func paraRespostas(linhas []*ent.Usuario) []Resposta {
	saida := make([]Resposta, 0, len(linhas))
	for _, l := range linhas {
		saida = append(saida, paraResposta(l))
	}
	return saida
}
