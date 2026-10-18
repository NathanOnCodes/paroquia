package httpx

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

// ErroDeVinculacao explica por que uma requisição não pôde virar struct. Ele
// sempre carrega mensagens seguras para o usuário final.
type ErroDeVinculacao struct {
	Origem string
	Campos []ErroDeCampo
}

func (e *ErroDeVinculacao) Error() string {
	if e.Origem == "" {
		return "dados inválidos"
	}
	return fmt.Sprintf("%s inválido", e.Origem)
}

// instanciaDoValidador é o validador do processo. As regras próprias do domínio
// são registradas uma vez, por NovoValidador.
var instanciaDoValidador = novoValidador()

// NovoValidador registra as regras de validação do domínio e devolve a
// instância compartilhada. É um construtor de propósito: chamadores podem
// montar a própria instância isolada nos testes.
func novoValidador() *validator.Validate {
	v := validator.New()

	// faixa de doação usada pelo sistema legado: inteiro entre R$ 1,00 e
	// R$ 100.000,00.
	_ = v.RegisterValidation("donation_amount", func(fl validator.FieldLevel) bool {
		n := fl.Field().Int()
		return n >= 100 && n <= 10_000_000
	})

	// faixa de lançamento financeiro: até R$ 10.000.000,00.
	_ = v.RegisterValidation("entry_amount", func(fl validator.FieldLevel) bool {
		n := fl.Field().Int()
		return n >= 1 && n <= 1_000_000_000
	})

	// cor no formato "#RRGGBB".
	_ = v.RegisterValidation("hexcolor6", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if len(s) != 7 || s[0] != '#' {
			return false
		}
		for _, r := range s[1:] {
			ehHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
			if !ehHex {
				return false
			}
		}
		return true
	})

	// slug de evento: caixa-baixa, dígitos e hífens.
	_ = v.RegisterValidation("slug", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if s == "" {
			return true
		}
		for _, r := range s {
			ok := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
			if !ok {
				return false
			}
		}
		return true
	})

	// telefone no padrão legado ^[0-9()\s-]{8,20}$.
	_ = v.RegisterValidation("phone", func(fl validator.FieldLevel) bool {
		s := fl.Field().String()
		if s == "" {
			return true
		}
		if len(s) < 8 || len(s) > 20 {
			return false
		}
		for _, r := range s {
			ok := (r >= '0' && r <= '9') || r == '(' || r == ')' || r == ' ' || r == '-'
			if !ok {
				return false
			}
		}
		return true
	})

	return v
}

// Validador expõe a instância compartilhada, para quem precisa de regras extras.
func Validador() *validator.Validate { return instanciaDoValidador }

// Vincular decodifica e valida o corpo JSON da requisição em destino.
//
// Substitui o parsing ad-hoc no topo de cada handler e garante que uma carga
// malformada nunca alcance a regra de negócio.
func Vincular(c fiber.Ctx, destino any) error {
	if err := c.Bind().JSON(destino); err != nil {
		return &ErroDeVinculacao{
			Origem: "body",
			Campos: []ErroDeCampo{{Campo: "body", Mensagem: "corpo da requisição inválido"}},
		}
	}
	if err := instanciaDoValidador.Struct(destino); err != nil {
		var errosDeValidacao validator.ValidationErrors
		if comoErroDeValidacao(err, &errosDeValidacao) {
			return &ErroDeVinculacao{Origem: "body", Campos: traduzirValidacao(errosDeValidacao)}
		}
		return &ErroDeVinculacao{
			Origem: "body",
			Campos: []ErroDeCampo{{Campo: "body", Mensagem: "dados inválidos"}},
		}
	}
	return nil
}

// VincularConsulta decodifica e valida a string de consulta em destino.
func VincularConsulta(c fiber.Ctx, destino any) error {
	if err := c.Bind().Query(destino); err != nil {
		return &ErroDeVinculacao{
			Origem: "query",
			Campos: []ErroDeCampo{{Campo: "query", Mensagem: "parâmetros de consulta inválidos"}},
		}
	}
	if err := instanciaDoValidador.Struct(destino); err != nil {
		var errosDeValidacao validator.ValidationErrors
		if comoErroDeValidacao(err, &errosDeValidacao) {
			return &ErroDeVinculacao{Origem: "query", Campos: traduzirValidacao(errosDeValidacao)}
		}
		return &ErroDeVinculacao{
			Origem: "query",
			Campos: []ErroDeCampo{{Campo: "query", Mensagem: "parâmetros de consulta inválidos"}},
		}
	}
	return nil
}

func comoErroDeValidacao(err error, alvo *validator.ValidationErrors) bool {
	if v, ok := err.(validator.ValidationErrors); ok {
		*alvo = v
		return true
	}
	return false
}

// PrimeiroCampo devolve a lista de campos citados no erro, para registro em log.
func PrimeiroCampo(err *ErroDeVinculacao) string {
	if err == nil || len(err.Campos) == 0 {
		return ""
	}
	var b strings.Builder
	for i, c := range err.Campos {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(c.Campo)
	}
	return b.String()
}
