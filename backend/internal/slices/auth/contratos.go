package auth

import (
	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// RequisicaoDeLogin é o corpo de POST /api/auth/login.
type RequisicaoDeLogin struct {
	Email string `json:"email" validate:"required,email,max=200"`
	Senha string `json:"senha" validate:"required,min=8,max=200"`
}

// RequisicaoDeTrocaDeSenha é o corpo de POST /api/auth/senha.
type RequisicaoDeTrocaDeSenha struct {
	SenhaAtual string `json:"senha_atual" validate:"required,min=8,max=200"`
	NovaSenha  string `json:"nova_senha" validate:"required,min=8,max=200"`
}

// RespostaDeUsuario é a identidade exposta ao cliente.
type RespostaDeUsuario struct {
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

// RespostaDeSessao é a resposta de login e renovação. O refresh token não
// volta no corpo: ele vive num cookie httpOnly.
type RespostaDeSessao struct {
	Usuario RespostaDeUsuario `json:"user"`
}

// paraResposta converte o registro persistido no formato de trânsito.
func paraResposta(linha *ent.Usuario) RespostaDeUsuario {
	return RespostaDeUsuario{
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
