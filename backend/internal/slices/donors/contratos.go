package donors

import (
	"strings"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// WriteRequest is the admin payload for create and update.
type RequisicaoDeEscrita struct {
	FullName      string `json:"full_name" validate:"required,min=3,max=200"`
	Email         string `json:"email" validate:"required,email,max=200"`
	Phone         string `json:"phone" validate:"omitempty,phone"`
	City          string `json:"city" validate:"max=120"`
	Neighborhood  string `json:"neighborhood" validate:"max=120"`
	CommunityID   string `json:"community_id" validate:"omitempty,uuid"`
	CommunityName string `json:"community_name" validate:"max=160"`
	IsActive      bool   `json:"is_active"`
}

// StatusRequest toggles a donor's active flag.
type RequisicaoDeStatus struct {
	IsActive bool `json:"is_active"`
}

// Response is the wire shape of a donor.
type Resposta struct {
	ID            string  `json:"id"`
	FullName      string  `json:"full_name"`
	Email         string  `json:"email"`
	Phone         *string `json:"phone"`
	City          string  `json:"city"`
	Neighborhood  string  `json:"neighborhood"`
	CommunityID   *string `json:"community_id"`
	CommunityName *string `json:"community_name"`
	// CommunityLabel is the resolved community name, whatever its origin.
	CommunityLabel   *string `json:"community_label"`
	PrivacyConsentAt *string `json:"privacy_consent_at"`
	IsActive         bool    `json:"is_active"`
	CreatedAt        string  `json:"created_at"`
	UpdatedAt        string  `json:"updated_at"`
}

func (r RequisicaoDeEscrita) paraEntrada() EntradaDeEscrita {
	in := EntradaDeEscrita{
		FullName:     strings.TrimSpace(r.FullName),
		Email:        r.Email,
		City:         strings.TrimSpace(r.City),
		Neighborhood: strings.TrimSpace(r.Neighborhood),
		IsActive:     r.IsActive,
		// The caller is the parish office, so recording consent now is correct.
		PrivacyConsent: true,
	}
	if p := strings.TrimSpace(r.Phone); p != "" {
		in.Phone = &p
	}
	if id := strings.TrimSpace(r.CommunityID); id != "" {
		if parsed, err := uuid.Parse(id); err == nil {
			in.CommunityID = entx.UUIDOpcional(parsed)
		}
	}
	// The free-form name is only kept when no registered community was chosen,
	// mirroring the legacy rule.
	if in.CommunityID == nil {
		if n := strings.TrimSpace(r.CommunityName); n != "" {
			in.CommunityName = &n
		}
	}
	return in
}

func paraResposta(row *ent.Doador) Resposta {
	out := Resposta{
		ID:               row.ID.String(),
		FullName:         row.FullName,
		Email:            row.Email,
		Phone:            row.Phone,
		City:             row.City,
		Neighborhood:     row.Neighborhood,
		CommunityID:      entx.UUIDParaTextoOpcional(row.CommunityID),
		CommunityName:    row.CommunityName,
		PrivacyConsentAt: entx.DataParaTextoOpcional(row.PrivacyConsentAt),
		IsActive:         row.IsActive,
		CreatedAt:        entx.DataParaTexto(row.CreatedAt),
		UpdatedAt:        entx.DataParaTexto(row.UpdatedAt),
	}
	switch {
	case row.Edges.Community != nil:
		name := row.Edges.Community.Name
		out.CommunityLabel = &name
	case row.CommunityName != nil:
		out.CommunityLabel = row.CommunityName
	}
	return out
}

func paraRespostas(rows []*ent.Doador) []Resposta {
	out := make([]Resposta, 0, len(rows))
	for _, r := range rows {
		out = append(out, paraResposta(r))
	}
	return out
}
