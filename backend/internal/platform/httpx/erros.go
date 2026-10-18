// Package httpx detém o contrato HTTP compartilhado por todas as slices: o
// envelope de erro, a vinculação com validação e os helpers de resposta.
//
// Nada aqui conhece o domínio. As slices traduzem seus erros para este
// envelope por meio dos erros sentinela e dos construtores abaixo.
package httpx

import (
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// ErroDeCampo aponta um campo inválido.
type ErroDeCampo struct {
	Campo    string `json:"field"`
	Mensagem string `json:"message"`
}

// ErroDeResposta é a forma única dos erros devolvidos pela API:
//
//	{ "error": { "code": "...", "message": "...", "details": [...] } }
type ErroDeResposta struct {
	Codigo   string        `json:"code"`
	Mensagem string        `json:"message"`
	Detalhes []ErroDeCampo `json:"details,omitempty"`
}

// corpoDaResposta envolve o erro para que o formato em trânsito seja exatamente
// { "error": {...} }.
type corpoDaResposta struct {
	Erro ErroDeResposta `json:"error"`
}

// Códigos usados pela API. Mantenha-os estáveis: os clientes fazem branch neles.
const (
	CodigoDeValidacao    = "validation_error"
	CodigoNaoAutenticado = "unauthenticated"
	CodigoSemPermissao   = "forbidden"
	CodigoNaoEncontrado  = "not_found"
	CodigoDeConflito     = "conflict"
	CodigoLimiteExcedido = "rate_limited"
	CodigoCargaGrande    = "payload_too_large"
	CodigoErroDeOrigem   = "upstream_error"
	CodigoErroInterno    = "internal_error"
)

// Erros sentinela que as slices devolvem para escolher um status sem expor
// detalhes de transporte na regra de negócio.
var (
	ErrNaoAutenticado = errors.New("não autenticado")
	ErrSemPermissao   = errors.New("sem permissão")
	ErrNaoEncontrado  = errors.New("não encontrado")
	ErrConflito       = errors.New("conflito")
	ErrLimiteExcedido = errors.New("limite excedido")
)

// ErroAPI é um erro de domínio que carrega status HTTP, código e uma mensagem
// segura. As mensagens são escritas para o usuário final e podem ser expostas.
type ErroAPI struct {
	Status   int
	Codigo   string
	Mensagem string
	Detalhes []ErroDeCampo
	// Causa é o erro subjacente; é registrado em log mas nunca serializado.
	Causa error
}

func (e *ErroAPI) Error() string {
	if e.Causa != nil {
		return fmt.Sprintf("%s: %v", e.Mensagem, e.Causa)
	}
	return e.Mensagem
}

// Unwrap expõe a causa para errors.Is/As.
func (e *ErroAPI) Unwrap() error { return e.Causa }

// Envolver preserva a causa para o log enquanto expõe uma mensagem segura.
func (e *ErroAPI) Envolver(causa error) *ErroAPI {
	e.Causa = causa
	return e
}

// Construtores de conveniência para os casos comuns.
func ErroDeValidacao(msg string, detalhes ...ErroDeCampo) *ErroAPI {
	return &ErroAPI{Status: fiber.StatusBadRequest, Codigo: CodigoDeValidacao, Mensagem: msg, Detalhes: detalhes}
}

func ErroNaoAutenticado(msg string) *ErroAPI {
	return &ErroAPI{Status: fiber.StatusUnauthorized, Codigo: CodigoNaoAutenticado, Mensagem: msg}
}

func ErroSemPermissao(msg string) *ErroAPI {
	return &ErroAPI{Status: fiber.StatusForbidden, Codigo: CodigoSemPermissao, Mensagem: msg}
}

func ErroNaoEncontrado(msg string) *ErroAPI {
	return &ErroAPI{Status: fiber.StatusNotFound, Codigo: CodigoNaoEncontrado, Mensagem: msg}
}

func ErroConflito(msg string) *ErroAPI {
	return &ErroAPI{Status: fiber.StatusConflict, Codigo: CodigoDeConflito, Mensagem: msg}
}

func ErroLimiteExcedido(msg string) *ErroAPI {
	return &ErroAPI{Status: fiber.StatusTooManyRequests, Codigo: CodigoLimiteExcedido, Mensagem: msg}
}

func ErroInterno(msg string, causa error) *ErroAPI {
	return &ErroAPI{Status: fiber.StatusInternalServerError, Codigo: CodigoErroInterno, Mensagem: msg, Causa: causa}
}

// ErroGenerico monta um ErroAPI com status e código explícitos.
func ErroGenerico(status int, codigo, mensagem string) *ErroAPI {
	return &ErroAPI{Status: status, Codigo: codigo, Mensagem: mensagem}
}

// ResponderJSON grava uma carga de sucesso com o status informado.
func ResponderJSON(c fiber.Ctx, status int, carga any) error {
	return c.Status(status).JSON(carga)
}

// ResponderOK grava 200 com uma carga.
func ResponderOK(c fiber.Ctx, carga any) error {
	return ResponderJSON(c, fiber.StatusOK, carga)
}

// ResponderCriado grava 201 com uma carga.
func ResponderCriado(c fiber.Ctx, carga any) error {
	return ResponderJSON(c, fiber.StatusCreated, carga)
}

// ResponderSemConteudo grava 204.
func ResponderSemConteudo(c fiber.Ctx) error { return c.SendStatus(fiber.StatusNoContent) }

// confirmado é a resposta padrão de uma mutação sem nada mais a dizer.
type confirmado struct {
	OK bool `json:"ok"`
}

// ResponderConfirmado devolve { "ok": true }.
func ResponderConfirmado(c fiber.Ctx) error { return ResponderOK(c, confirmado{OK: true}) }

// EscreverErro renderiza err como o envelope compartilhado. É o único lugar em
// que um erro vira bytes, de modo que nenhum handler consegue vazar uma mensagem
// interna por acidente.
func EscreverErro(c fiber.Ctx, err error) error {
	// Já é um ErroAPI.
	var erroAPI *ErroAPI
	if errors.As(err, &erroAPI) {
		return c.Status(erroAPI.Status).JSON(corpoDaResposta{Erro: ErroDeResposta{
			Codigo:   erroAPI.Codigo,
			Mensagem: erroAPI.Mensagem,
			Detalhes: erroAPI.Detalhes,
		}})
	}

	// Sentinelas declarados neste pacote.
	switch {
	case errors.Is(err, ErrNaoAutenticado):
		return EscreverErro(c, ErroNaoAutenticado("Não autenticado"))
	case errors.Is(err, ErrSemPermissao):
		return EscreverErro(c, ErroSemPermissao("Sem permissão para esta operação"))
	case errors.Is(err, ErrNaoEncontrado):
		return EscreverErro(c, ErroNaoEncontrado("Registro não encontrado"))
	case errors.Is(err, ErrConflito):
		return EscreverErro(c, ErroConflito("Operação em conflito com o estado atual"))
	case errors.Is(err, ErrLimiteExcedido):
		return EscreverErro(c, ErroLimiteExcedido("Muitas solicitações. Tente novamente mais tarde."))
	}

	// Falhas de validação produzidas pelo vinculador.
	var errosDeValidacao validator.ValidationErrors
	if errors.As(err, &errosDeValidacao) {
		return EscreverErro(c, ErroDeValidacao("Dados inválidos", traduzirValidacao(errosDeValidacao)...))
	}

	var erroDeVinculacao *ErroDeVinculacao
	if errors.As(err, &erroDeVinculacao) {
		return EscreverErro(c, ErroDeValidacao(erroDeVinculacao.Error(), erroDeVinculacao.Campos...))
	}

	// Qualquer outra coisa é um bug: responda de forma opaca.
	return EscreverErro(c, ErroInterno("Erro inesperado", err))
}

// traduzirValidacao converte falhas do validador na lista compartilhada de
// ErroDeCampo.
func traduzirValidacao(falhas validator.ValidationErrors) []ErroDeCampo {
	saida := make([]ErroDeCampo, 0, len(falhas))
	for _, f := range falhas {
		saida = append(saida, ErroDeCampo{
			Campo:    nomeJSON(f),
			Mensagem: mensagemPara(f),
		})
	}
	return saida
}

func nomeJSON(f validator.FieldError) string {
	nome := f.Field()
	if espaco := f.StructNamespace(); espaco != "" {
		partes := strings.Split(espaco, ".")
		nome = strings.Join(partes[1:], ".")
	}
	return paraCaixaBaixa(nome)
}

func mensagemPara(f validator.FieldError) string {
	switch f.Tag() {
	case "required":
		return "campo obrigatório"
	case "email":
		return "e-mail inválido"
	case "min":
		return "valor abaixo do mínimo permitido"
	case "max":
		return "valor acima do máximo permitido"
	case "len":
		return "tamanho inválido"
	case "oneof":
		return "valor fora do conjunto permitido"
	case "uuid":
		return "identificador inválido"
	case "gte":
		return "valor abaixo do mínimo permitido"
	case "lte":
		return "valor acima do máximo permitido"
	case "gt":
		return "valor deve ser positivo"
	case "datetime":
		return "data inválida"
	case "excludes":
		return "valor não permitido"
	case "startswith":
		return "formato inválido"
	case "donation_amount":
		return "valor da doação fora do intervalo permitido"
	case "entry_amount":
		return "valor do lançamento fora do intervalo permitido"
	case "hexcolor6":
		return "cor inválida"
	case "slug":
		return "formato inválido"
	case "phone":
		return "telefone inválido"
	default:
		return "valor inválido"
	}
}

// paraCaixaBaixa converte nomes de campos Go exportados para a forma em
// caixa-baixa usada no JSON.
func paraCaixaBaixa(s string) string {
	if s == "" {
		return s
	}
	var b strings.Builder
	for i, r := range s {
		if ehMaiuscula(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(paraMinuscula(r))
	}
	return b.String()
}

func ehMaiuscula(r rune) bool { return r >= 'A' && r <= 'Z' }

func paraMinuscula(r rune) rune {
	if ehMaiuscula(r) {
		return r + ('a' - 'A')
	}
	return r
}
