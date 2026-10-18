package recurring

// RequisicaoDeLink pede o envio do link mágico de gerenciamento.
type RequisicaoDeLink struct {
	Email string `json:"email" validate:"required,email,max=200"`
}

// RespostaDeLink é sempre a mesma, para não revelar se o e-mail existe.
type RespostaDeLink struct {
	OK       bool   `json:"ok"`
	Mensagem string `json:"message"`
}

// RequisicaoDeCancelamento encerra uma recorrência.
type RequisicaoDeCancelamento struct {
	IDAssinatura string `json:"subscription_id" validate:"required,min=1"`
}

// RespostaDeRecorrencia é o formato em trânsito de uma recorrência.
type RespostaDeRecorrencia struct {
	ID                     string  `json:"id"`
	ValorEmCentavos        int     `json:"amount_cents"`
	Moeda                  string  `json:"currency"`
	Intervalo              string  `json:"interval"`
	Status                 string  `json:"status"`
	CancelarNoFimDoPeriodo bool    `json:"cancel_at_period_end"`
	IniciadaEm             *string `json:"started_at"`
	FimDoPeriodoAtual      *string `json:"current_period_end"`
	PedidoDeCancelamentoEm *string `json:"cancellation_requested_at"`
	CanceladaEm            *string `json:"canceled_at"`
	CriadoEm               string  `json:"created_at"`
}

// RespostaDeGerenciamento é o que a página do fiel recebe.
type RespostaDeGerenciamento struct {
	DoadorID     string                  `json:"donor_id"`
	Recorrencias []RespostaDeRecorrencia `json:"subscriptions"`
}

// MensagemDeLinkPadrao é a resposta enviada em qualquer caso.
const MensagemDeLinkPadrao = "Se o e-mail estiver cadastrado, você receberá um link de acesso."
