package events

import (
	"time"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
)

// RequisicaoDeEscrita é o corpo de criação e edição de evento.
type RequisicaoDeEscrita struct {
	Titulo    string `json:"title" validate:"required,min=3,max=200"`
	Slug      string `json:"slug" validate:"omitempty,slug"`
	Descricao string `json:"description" validate:"max=5000"`
	IniciaEm  string `json:"starts_at" validate:"required,datetime=2006-01-02T15:04:05Z07:00"`
	TerminaEm string `json:"ends_at" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Local     string `json:"location" validate:"max=300"`
	Status    string `json:"status" validate:"required,oneof=draft published archived"`
}

// RequisicaoDeStatus alterna o status de publicação.
type RequisicaoDeStatus struct {
	Status string `json:"status" validate:"required,oneof=draft published archived"`
}

// Resposta é o formato em trânsito de um evento.
type Resposta struct {
	ID           string  `json:"id"`
	Titulo       string  `json:"title"`
	Slug         string  `json:"slug"`
	Descricao    string  `json:"description"`
	Imagem       *string `json:"image_path"`
	ImagemURL    *string `json:"image_url"`
	IniciaEm     string  `json:"starts_at"`
	TerminaEm    *string `json:"ends_at"`
	Local        string  `json:"location"`
	Status       string  `json:"status"`
	PublicadoEm  *string `json:"published_at"`
	CriadoEm     string  `json:"created_at"`
	AtualizadoEm string  `json:"updated_at"`
}

// paraResposta converte o registro persistido, resolvendo a URL da imagem.
func paraResposta(linha *ent.Evento, urlDaImagem func(string) string) Resposta {
	resposta := Resposta{
		ID:           linha.ID.String(),
		Titulo:       linha.Title,
		Slug:         linha.Slug,
		Descricao:    linha.Description,
		Imagem:       linha.ImagePath,
		IniciaEm:     entx.DataParaTexto(linha.StartsAt),
		TerminaEm:    entx.DataParaTextoOpcional(linha.EndsAt),
		Local:        linha.Location,
		Status:       string(linha.Status),
		PublicadoEm:  entx.DataParaTextoOpcional(linha.PublishedAt),
		CriadoEm:     entx.DataParaTexto(linha.CreatedAt),
		AtualizadoEm: entx.DataParaTexto(linha.UpdatedAt),
	}
	if linha.ImagePath != nil && urlDaImagem != nil {
		if url := urlDaImagem(*linha.ImagePath); url != "" {
			resposta.ImagemURL = &url
		}
	}
	return resposta
}

func paraRespostas(linhas []*ent.Evento, urlDaImagem func(string) string) []Resposta {
	saida := make([]Resposta, 0, len(linhas))
	for _, l := range linhas {
		saida = append(saida, paraResposta(l, urlDaImagem))
	}
	return saida
}

// interpretarDatas converte os campos textuais em data/hora.
func (r RequisicaoDeEscrita) interpretarDatas() (time.Time, *time.Time, error) {
	inicio, err := time.Parse(time.RFC3339, r.IniciaEm)
	if err != nil {
		return time.Time{}, nil, err
	}
	if r.TerminaEm == "" {
		return inicio, nil, nil
	}
	fim, err := time.Parse(time.RFC3339, r.TerminaEm)
	if err != nil {
		return time.Time{}, nil, err
	}
	return inicio, &fim, nil
}
