package transparency

import (
	"time"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// RequisicaoDePeriodo é o corpo de criação e edição de período.
type RequisicaoDePeriodo struct {
	Nome     string `json:"name" validate:"required,min=3,max=200"`
	InicioEm string `json:"start_date" validate:"required,datetime=2006-01-02"`
	FimEm    string `json:"end_date" validate:"required,datetime=2006-01-02"`
	Status   string `json:"status" validate:"omitempty,oneof=draft review published archived"`
}

// RequisicaoDeStatus muda a situação do período.
type RequisicaoDeStatus struct {
	Status string `json:"status" validate:"required,oneof=draft review published archived"`
}

// RequisicaoDeLancamento é o corpo de uma receita ou despesa.
type RequisicaoDeLancamento struct {
	Tipo          string `json:"type" validate:"required,oneof=income expense"`
	Categoria     string `json:"category" validate:"required,min=2,max=120"`
	Descricao     string `json:"description" validate:"required,min=3,max=500"`
	ValorCentavos int    `json:"amount_cents" validate:"required,entry_amount"`
	Data          string `json:"entry_date" validate:"required,datetime=2006-01-02"`
}

// RespostaDePeriodo é o formato em trânsito de um período, já com os totais.
type RespostaDePeriodo struct {
	ID           string                 `json:"id"`
	Nome         string                 `json:"name"`
	InicioEm     string                 `json:"start_date"`
	FimEm        string                 `json:"end_date"`
	Status       string                 `json:"status"`
	Receitas     int64                  `json:"income_cents"`
	Despesas     int64                  `json:"expense_cents"`
	Saldo        int64                  `json:"balance_cents"`
	PublicadoEm  *string                `json:"published_at"`
	CriadoEm     string                 `json:"created_at"`
	AtualizadoEm string                 `json:"updated_at"`
	Lancamentos  []RespostaDeLancamento `json:"entries,omitempty"`
	Documentos   []RespostaDeDocumento  `json:"documents,omitempty"`
}

// RespostaDeLancamento é uma linha de receita ou despesa.
type RespostaDeLancamento struct {
	ID            string `json:"id"`
	Tipo          string `json:"type"`
	Categoria     string `json:"category"`
	Descricao     string `json:"description"`
	ValorCentavos int    `json:"amount_cents"`
	Data          string `json:"entry_date"`
	Origem        string `json:"source"`
	CriadoEm      string `json:"created_at"`
}

// RespostaDeDocumento é um anexo do período.
type RespostaDeDocumento struct {
	ID           string `json:"id"`
	NomeOriginal string `json:"original_file_name"`
	TipoMIME     string `json:"mime_type"`
	Tamanho      int    `json:"file_size"`
	CriadoEm     string `json:"created_at"`
}

func paraRespostaDePeriodo(linha *ent.PeriodoFinanceiro, receitas, despesas int64) RespostaDePeriodo {
	return RespostaDePeriodo{
		ID:           linha.ID.String(),
		Nome:         linha.Name,
		InicioEm:     entx.DataSimples(linha.StartDate),
		FimEm:        entx.DataSimples(linha.EndDate),
		Status:       string(linha.Status),
		Receitas:     receitas,
		Despesas:     despesas,
		Saldo:        receitas - despesas,
		PublicadoEm:  entx.DataParaTextoOpcional(linha.PublishedAt),
		CriadoEm:     entx.DataParaTexto(linha.CreatedAt),
		AtualizadoEm: entx.DataParaTexto(linha.UpdatedAt),
	}
}

func paraRespostaDeLancamento(linha *ent.LancamentoFinanceiro) RespostaDeLancamento {
	return RespostaDeLancamento{
		ID:            linha.ID.String(),
		Tipo:          string(linha.Type),
		Categoria:     linha.Category,
		Descricao:     linha.Description,
		ValorCentavos: linha.AmountCents,
		Data:          entx.DataSimples(linha.EntryDate),
		Origem:        string(linha.Source),
		CriadoEm:      entx.DataParaTexto(linha.CreatedAt),
	}
}

func paraRespostaDeDocumento(linha *ent.DocumentoTransparencia) RespostaDeDocumento {
	return RespostaDeDocumento{
		ID:           linha.ID.String(),
		NomeOriginal: linha.OriginalFileName,
		TipoMIME:     linha.MimeType,
		Tamanho:      linha.FileSize,
		CriadoEm:     entx.DataParaTexto(linha.CreatedAt),
	}
}

// interpretarDatas converte as datas do formulário.
func (r RequisicaoDePeriodo) interpretarDatas() (time.Time, time.Time, error) {
	inicio, err := time.Parse("2006-01-02", r.InicioEm)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	fim, err := time.Parse("2006-01-02", r.FimEm)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	return inicio, fim, nil
}

// interpretarData converte a data de um lançamento.
func (r RequisicaoDeLancamento) interpretarData() (time.Time, error) {
	return time.Parse("2006-01-02", r.Data)
}
