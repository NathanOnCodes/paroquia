package payments

import (
	"strings"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/internal/slices/donors"
)

// RequisicaoDeDoacao é o corpo dos endpoints de criação de pagamento.
type RequisicaoDeDoacao struct {
	NomeCompleto      string `json:"full_name" validate:"required,min=3,max=200"`
	Email             string `json:"email" validate:"required,email,max=200"`
	Telefone          string `json:"phone" validate:"omitempty,phone"`
	IDComunidade      string `json:"community_id" validate:"omitempty,uuid"`
	NomeComunidade    string `json:"community_name" validate:"max=160"`
	Cidade            string `json:"city" validate:"required,min=2,max=120"`
	Bairro            string `json:"neighborhood" validate:"max=120"`
	ValorCentavos     int64  `json:"amount_cents" validate:"required,donation_amount"`
	Recurrence        bool   `json:"is_recurring"`
	AceitePrivacidade bool   `json:"privacy_consent" validate:"required"`
}

// RequisicaoDeAssinatura efetiva a recorrência depois que o navegador
// confirma o método de pagamento.
type RequisicaoDeAssinatura struct {
	IDDoador            string `json:"donor_id" validate:"required,uuid"`
	IDComunidade        string `json:"community_id" validate:"omitempty,uuid"`
	NomeComunidade      string `json:"community_name" validate:"max=160"`
	ValorCentavos       int64  `json:"amount_cents" validate:"required,donation_amount"`
	IDClienteStripe     string `json:"stripe_customer_id" validate:"required,min=1"`
	IDMetodoDePagamento string `json:"payment_method_id" validate:"required,min=1"`
}

// RespostaDeIntencao devolve o client secret que o navegador usa para
// confirmar uma cobrança avulsa.
type RespostaDeIntencao struct {
	ClientSecret  string `json:"client_secret"`
	ValorCentavos int64  `json:"amount_cents"`
}

// RespostaDeConfiguracao devolve o client secret do setup da recorrência.
type RespostaDeConfiguracao struct {
	ClientSecret    string `json:"client_secret"`
	IDClienteStripe string `json:"stripe_customer_id"`
	IDDoador        string `json:"donor_id"`
}

// RespostaDeAssinatura identifica a recorrência criada.
type RespostaDeAssinatura struct {
	IDAssinatura string `json:"subscription_id"`
}

// paraDoador extrai a identidade do fiel para o cadastro automático.
func (r RequisicaoDeDoacao) paraDoador() donors.IdentidadeDaDoacao {
	entrada := donors.IdentidadeDaDoacao{
		FullName:      strings.TrimSpace(r.NomeCompleto),
		Email:         r.Email,
		City:          strings.TrimSpace(r.Cidade),
		Neighborhood:  strings.TrimSpace(r.Bairro),
		CommunityName: strings.TrimSpace(r.NomeComunidade),
	}
	if t := strings.TrimSpace(r.Telefone); t != "" {
		entrada.Phone = t
	}
	if id := strings.TrimSpace(r.IDComunidade); id != "" {
		if parsed, err := uuid.Parse(id); err == nil {
			entrada.CommunityID = &parsed
			// Regra legada: havendo comunidade cadastrada, o nome livre some.
			entrada.CommunityName = ""
		}
	}
	return entrada
}

// ValidarComunidade exige comunidade cadastrada OU nome livre, como no
// `requireCommunity` do sistema legado.
func (r RequisicaoDeDoacao) ValidarComunidade() bool {
	return r.IDComunidade != "" || strings.TrimSpace(r.NomeComunidade) != ""
}
