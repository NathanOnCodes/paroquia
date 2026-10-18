package communities

import (
	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// Request and response contracts of the communities slice.

// WriteRequest is the admin payload for create and update.
type RequisicaoDeEscrita struct {
	Name        string `json:"name" validate:"required,min=3,max=120"`
	Description string `json:"description" validate:"max=500"`
	City        string `json:"city" validate:"max=120"`
	IsActive    bool   `json:"is_active"`
}

// StatusRequest toggles visibility.
type RequisicaoDeStatus struct {
	IsActive bool `json:"is_active"`
}

// Response is the wire shape of a community.
type Resposta struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	City        string `json:"city"`
	IsActive    bool   `json:"is_active"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

func paraResposta(row *ent.Comunidade) Resposta {
	return Resposta{
		ID:          row.ID.String(),
		Name:        row.Name,
		Description: row.Description,
		City:        row.City,
		IsActive:    row.IsActive,
		CreatedAt:   entx.DataParaTexto(row.CreatedAt),
		UpdatedAt:   entx.DataParaTexto(row.UpdatedAt),
	}
}
