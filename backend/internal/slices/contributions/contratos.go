package contributions

import (
	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// Response is the wire shape of one contribution. It carries both the raw
// community fields (id or free-form name) and a resolved label, so the table
// renders without a second lookup.
type Resposta struct {
	ID                 string  `json:"id"`
	DonorID            string  `json:"donor_id"`
	DonorName          *string `json:"donor_name"`
	DonorEmail         *string `json:"donor_email"`
	CommunityID        *string `json:"community_id"`
	CommunityName      *string `json:"community_name"`
	RecurringID        *string `json:"recurring_donation_id"`
	AmountCents        int     `json:"amount_cents"`
	Currency           string  `json:"currency"`
	PaymentType        string  `json:"payment_type"`
	Status             string  `json:"status"`
	StripePaymentID    *string `json:"stripe_payment_intent_id"`
	StripeInvoiceID    *string `json:"stripe_invoice_id"`
	StripeSubscription *string `json:"stripe_subscription_id"`
	PaidAt             *string `json:"paid_at"`
	RefundedAt         *string `json:"refunded_at"`
	CreatedAt          string  `json:"created_at"`
}

func paraResposta(row *ent.PagamentoContribuicao) Resposta {
	out := Resposta{
		ID:                 row.ID.String(),
		DonorID:            row.DonorID.String(),
		CommunityID:        entx.UUIDParaTextoOpcional(row.CommunityID),
		CommunityName:      row.CommunityName,
		RecurringID:        entx.UUIDParaTextoOpcional(row.RecurringDonationID),
		AmountCents:        row.AmountCents,
		Currency:           row.Currency,
		PaymentType:        string(row.PaymentType),
		Status:             string(row.Status),
		StripePaymentID:    row.StripePaymentIntentID,
		StripeInvoiceID:    row.StripeInvoiceID,
		StripeSubscription: row.StripeSubscriptionID,
		PaidAt:             entx.DataParaTextoOpcional(row.PaidAt),
		RefundedAt:         entx.DataParaTextoOpcional(row.RefundedAt),
		CreatedAt:          entx.DataParaTexto(row.CreatedAt),
	}
	if row.Edges.Donor != nil {
		name := row.Edges.Donor.FullName
		email := row.Edges.Donor.Email
		out.DonorName = &name
		out.DonorEmail = &email
		// The donor's free-form community name is what the webhook stored when
		// no registered community was chosen.
		if out.CommunityName == nil {
			out.CommunityName = row.Edges.Donor.CommunityName
		}
	}
	if row.Edges.Community != nil {
		name := row.Edges.Community.Name
		out.CommunityName = &name
	}
	return out
}

func paraRespostas(rows []*ent.PagamentoContribuicao) []Resposta {
	out := make([]Resposta, 0, len(rows))
	for _, r := range rows {
		out = append(out, paraResposta(r))
	}
	return out
}
