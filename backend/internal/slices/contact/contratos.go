package contact

import (
	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// CreateRequest is the public contact form payload.
type RequisicaoDeCriacao struct {
	Name           string `json:"name" validate:"required,min=3,max=200"`
	Email          string `json:"email" validate:"required,email,max=200"`
	Phone          string `json:"phone" validate:"omitempty,phone"`
	Mensagem       string `json:"message" validate:"required,min=10,max=4000"`
	PrivacyConsent bool   `json:"privacy_consent" validate:"required"`
}

// StatusRequest updates the workflow state of a message.
type RequisicaoDeStatus struct {
	Status string `json:"status" validate:"required,oneof=new in_progress resolved archived"`
}

// Response is the wire shape of a message in the admin inbox.
type Resposta struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	Phone            *string `json:"phone"`
	Mensagem         string  `json:"message"`
	Status           string  `json:"status"`
	AssignedTo       *string `json:"assigned_to"`
	AssignedToName   *string `json:"assigned_to_name"`
	PrivacyConsentAt string  `json:"privacy_consent_at"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

// Accepted is the answer of the public endpoint. It never reveals whether the
// message was stored, so the endpoint cannot be used to probe the system.
type Aceito struct {
	OK bool `json:"ok"`
}

func paraResposta(row *ent.MensagemContato) Resposta {
	out := Resposta{
		ID:               row.ID.String(),
		Name:             row.Name,
		Email:            row.Email,
		Phone:            row.Phone,
		Mensagem:         row.Message,
		Status:           string(row.Status),
		PrivacyConsentAt: entx.DataParaTexto(row.PrivacyConsentAt),
		CreatedAt:        entx.DataParaTexto(row.CreatedAt),
		UpdatedAt:        entx.DataParaTexto(row.UpdatedAt),
	}
	if row.AssignedTo != nil {
		out.AssignedTo = entx.UUIDParaTextoOpcional(row.AssignedTo)
	}
	if row.Edges.Assignee != nil {
		name := row.Edges.Assignee.FullName
		out.AssignedToName = &name
	}
	return out
}

func paraRespostas(rows []*ent.MensagemContato) []Resposta {
	out := make([]Resposta, 0, len(rows))
	for _, r := range rows {
		out = append(out, paraResposta(r))
	}
	return out
}
