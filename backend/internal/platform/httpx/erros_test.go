package httpx

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
)

func novoApp() *fiber.App {
	app := fiber.New(fiber.Config{ErrorHandler: func(c fiber.Ctx, err error) error {
		return EscreverErro(c, err)
	}})
	return app
}

func TestEscreverErroUsaEnvelopeUnico(t *testing.T) {
	app := novoApp()
	app.Get("/validacao", func(c fiber.Ctx) error {
		return ErroDeValidacao("Dados inválidos",
			ErroDeCampo{Campo: "amount_cents", Mensagem: "valor abaixo do mínimo permitido"})
	})

	_, corpo := executar(t, app, "/validacao")

	for _, esperado := range []string{`"error"`, `"validation_error"`, `"amount_cents"`, "valor abaixo do mínimo permitido"} {
		if !strings.Contains(corpo, esperado) {
			t.Fatalf("resposta deveria conter %q, veio %s", esperado, corpo)
		}
	}
}

func TestEscreverErroNaoVazaMensagemInterna(t *testing.T) {
	app := novoApp()
	app.Get("/interno", func(c fiber.Ctx) error {
		return ErroInterno("Falha ao processar", errors.New("senha do banco é postgres://usuario:segredo@host"))
	})

	status, corpo := executar(t, app, "/interno")

	if status != 500 {
		t.Fatalf("status esperado 500, veio %d", status)
	}
	if strings.Contains(corpo, "segredo") || strings.Contains(corpo, "postgres://") {
		t.Fatalf("a causa interna vazou para o cliente: %s", corpo)
	}
	if !strings.Contains(corpo, "Erro inesperado") && !strings.Contains(corpo, "Falha ao processar") {
		t.Fatalf("mensagem segura ausente: %s", corpo)
	}
}

func TestEscreverErroConverteSentinelas(t *testing.T) {
	casos := []struct {
		nome   string
		erro   error
		status int
		codigo string
	}{
		{"não autenticado", ErrNaoAutenticado, 401, CodigoNaoAutenticado},
		{"sem permissão", ErrSemPermissao, 403, CodigoSemPermissao},
		{"não encontrado", ErrNaoEncontrado, 404, CodigoNaoEncontrado},
		{"conflito", ErrConflito, 409, CodigoDeConflito},
		{"limite excedido", ErrLimiteExcedido, 429, CodigoLimiteExcedido},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			app := novoApp()
			app.Get("/", func(c fiber.Ctx) error { return caso.erro })

			status, corpo := executar(t, app, "/")
			if status != caso.status {
				t.Fatalf("status esperado %d, veio %d", caso.status, status)
			}
			if !strings.Contains(corpo, caso.codigo) {
				t.Fatalf("código %q ausente em %s", caso.codigo, corpo)
			}
		})
	}
}

func TestEscreverErroDesembrulhaCausa(t *testing.T) {
	app := novoApp()
	app.Get("/", func(c fiber.Ctx) error {
		return ErroNaoEncontrado("Dizimista não encontrado").Envolver(ErrNaoEncontrado)
	})

	status, corpo := executar(t, app, "/")
	if status != 404 {
		t.Fatalf("status esperado 404, veio %d", status)
	}
	if !strings.Contains(corpo, CodigoNaoEncontrado) {
		t.Fatalf("código ausente: %s", corpo)
	}
}

func TestRegrasDeValidacaoProprias(t *testing.T) {
	validador := Validador()

	type doacao struct {
		Valor int64  `validate:"donation_amount"`
		Email string `validate:"email"`
		Fone  string `validate:"phone"`
	}

	// Regra de negócio: doação entre R$ 1,00 e R$ 100.000,00.
	for _, valor := range []int64{0, 99, 10_000_001, -100} {
		if err := validador.Struct(doacao{Valor: valor}); err == nil {
			t.Fatalf("valor %d deveria ser recusado", valor)
		}
	}
	for _, valor := range []int64{100, 2500, 10_000_000} {
		if err := validador.Struct(doacao{Valor: valor, Email: "a@b.co", Fone: "11 99999-9999"}); err != nil {
			t.Fatalf("valor %d deveria ser aceito: %v", valor, err)
		}
	}

	// Telefone no padrão legado ^[0-9()\s-]{8,20}$.
	if err := validador.Struct(doacao{Valor: 100, Email: "a@b.co", Fone: "abc"}); err == nil {
		t.Fatal("telefone com letras deveria ser recusado")
	}
}

func executar(t *testing.T, app *fiber.App, caminho string) (int, string) {
	t.Helper()
	defer func() { _ = app.Shutdown() }()

	req := httptest.NewRequest("GET", caminho, nil)
	resposta, err := app.Test(req)
	if err != nil {
		t.Fatalf("executando requisição: %v", err)
	}
	defer func() { _ = resposta.Body.Close() }()

	var b strings.Builder
	if _, err := b.WriteString(""); err != nil {
		t.Fatalf("lendo corpo: %v", err)
	}
	tampon := make([]byte, 4096)
	for {
		n, rerr := resposta.Body.Read(tampon)
		b.Write(tampon[:n])
		if rerr != nil {
			break
		}
	}
	return resposta.StatusCode, b.String()
}
