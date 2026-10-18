package payments

import (
	"encoding/json"
	"strconv"
	"strings"
	"testing"
	"time"
)

func cargaDeExemplo(t *testing.T) []byte {
	t.Helper()
	carga, err := json.Marshal(map[string]any{
		"id":   "evt_teste_1",
		"type": "payment_intent.succeeded",
		"data": map[string]any{
			"object": map[string]any{"id": "pi_1", "amount": 2500},
		},
	})
	if err != nil {
		t.Fatalf("montando carga: %v", err)
	}
	return carga
}

func novoCliente(segredo string) *ClienteStripe {
	return NovoClienteStripe(ConfiguracaoDoStripe{
		ChaveSecreta:     "sk_teste",
		SegredoDoWebhook: segredo,
	})
}

func TestInterpretarWebhookAceitaAssinaturaValida(t *testing.T) {
	const segredo = "whsec_secreto_de_teste_com_tamanho_suficiente"
	cliente := novoCliente(segredo)

	carga := cargaDeExemplo(t)
	agora := time.Now()
	cabecalho := CabecalhoDeAssinaturaPara(segredo, carga, agora)

	evento, err := cliente.InterpretarWebhook(carga, cabecalho)
	if err != nil {
		t.Fatalf("InterpretarWebhook: %v", err)
	}
	if evento.ID != "evt_teste_1" {
		t.Fatalf("id do evento divergiu: %q", evento.ID)
	}
	if evento.Type != "payment_intent.succeeded" {
		t.Fatalf("tipo do evento divergiu: %q", evento.Type)
	}
}

func TestInterpretarWebhookRejeitaAssinaturaInvalida(t *testing.T) {
	const segredo = "whsec_secreto_de_teste_com_tamanho_suficiente"
	cliente := novoCliente(segredo)

	carga := cargaDeExemplo(t)
	cabecalho := CabecalhoDeAssinaturaPara("whsec_outro_secreto_completamente_diferente!", carga, time.Now())

	if _, err := cliente.InterpretarWebhook(carga, cabecalho); err == nil {
		t.Fatal("assinatura de outro segredo deveria ser rejeitada")
	}
}

func TestInterpretarWebhookRejeitaCargaAlterada(t *testing.T) {
	const segredo = "whsec_secreto_de_teste_com_tamanho_suficiente"
	cliente := novoCliente(segredo)

	carga := cargaDeExemplo(t)
	cabecalho := CabecalhoDeAssinaturaPara(segredo, carga, time.Now())

	// Mesmo cabeçalho, carga adulterada.
	alterada := []byte(strings.Replace(string(carga), "2500", "9999", 1))
	if _, err := cliente.InterpretarWebhook(alterada, cabecalho); err == nil {
		t.Fatal("carga adulterada deveria ser rejeitada")
	}
}

func TestInterpretarWebhookRejeitaForaDaJanela(t *testing.T) {
	const segredo = "whsec_secreto_de_teste_com_tamanho_suficiente"
	cliente := novoCliente(segredo)

	carga := cargaDeExemplo(t)
	antigo := time.Now().Add(-30 * time.Minute)
	cabecalho := CabecalhoDeAssinaturaPara(segredo, carga, antigo)

	if _, err := cliente.InterpretarWebhook(carga, cabecalho); err == nil {
		t.Fatal("assinatura fora da janela de tolerância deveria ser rejeitada")
	}
}

func TestInterpretarWebhookExigeCabecalho(t *testing.T) {
	cliente := novoCliente("whsec_secreto_de_teste_com_tamanho_suficiente")

	for _, cabecalho := range []string{"", "garbage", "t=abc,v1=xyz"} {
		if _, err := cliente.InterpretarWebhook(cargaDeExemplo(t), cabecalho); err == nil {
			t.Fatalf("cabeçalho %q deveria ser rejeitado", cabecalho)
		}
	}
}

func TestInterpretarWebhookAceitaVariasAssinaturas(t *testing.T) {
	// Durante a troca de segredo a Stripe envia mais de um v1; qualquer um
	// válido deve ser aceito.
	const antigo = "whsec_segredo_antigo_com_tamanho_suficiente_aqui"
	const novo = "whsec_segredo_novo_tambem_com_tamanho_suficiente_aqui"
	cliente := novoCliente(antigo)

	carga := cargaDeExemplo(t)
	agora := time.Now()
	v1Antigo := extrairV1(CabecalhoDeAssinaturaPara(antigo, carga, agora))
	v1Novo := extrairV1(CabecalhoDeAssinaturaPara(novo, carga, agora))
	cabecalho := "t=" + strconv.FormatInt(agora.UTC().Unix(), 10) + ",v1=" + v1Novo + ",v1=" + v1Antigo

	if _, err := cliente.InterpretarWebhook(carga, cabecalho); err != nil {
		t.Fatalf("assinatura em rotação deveria ser aceita: %v", err)
	}
}

func TestInterpretarWebhookExigeSegredoConfigurado(t *testing.T) {
	cliente := NovoClienteStripe(ConfiguracaoDoStripe{ChaveSecreta: "sk_teste"})
	carga := cargaDeExemplo(t)

	if _, err := cliente.InterpretarWebhook(carga, CabecalhoDeAssinaturaPara("qualquer", carga, time.Now())); err == nil {
		t.Fatal("sem segredo de webhook nada deveria ser aceito")
	}
}

func extrairV1(cabecalho string) string {
	for _, parte := range strings.Split(cabecalho, ",") {
		if strings.HasPrefix(parte, "v1=") {
			return strings.TrimPrefix(parte, "v1=")
		}
	}
	return ""
}
