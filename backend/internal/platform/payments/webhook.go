package payments

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// webhookTolerance is how far a timestamp may be from now before the signature
// is rejected. It matches Stripe's recommended window.
const toleranciaDoWebhook = 5 * time.Minute

// InterpretarWebhook verifica o cabeçalho `Stripe-Signature` contra a carga
// and decodes the event envelope.
//
// O cabeçalho traz um ou mais pares `t=<unix>,v1=<hmac>`. A assinatura é
// um HMAC-SHA256 de "<timestamp>.<payload>" com o segredo do endpoint.
// When several v1 signatures are present, any of them matching is enough:
// Stripe sends multiple during secret rotation.
func (s *ClienteStripe) InterpretarWebhook(payload []byte, signature string) (*EventoDePagamento, error) {
	if s.cfg.SegredoDoWebhook == "" {
		return nil, fmt.Errorf("%w: segredo de webhook não configurado", ErrAssinaturaInvalida)
	}
	if len(signature) == 0 {
		return nil, fmt.Errorf("%w: cabeçalho ausente", ErrAssinaturaInvalida)
	}

	timestamp, sigs := interpretarCabecalhoDeAssinatura(signature)
	if timestamp.IsZero() || len(sigs) == 0 {
		return nil, fmt.Errorf("%w: cabeçalho malformado", ErrAssinaturaInvalida)
	}

	if delta := time.Since(timestamp); delta > toleranciaDoWebhook || delta < -toleranciaDoWebhook {
		return nil, fmt.Errorf("%w: timestamp fora da tolerância (%v)", ErrAssinaturaInvalida, delta)
	}

	expected := calcularAssinatura(s.cfg.SegredoDoWebhook, timestamp.Unix(), payload)
	if !algumConfere(expected, sigs) {
		return nil, fmt.Errorf("%w: assinatura não confere", ErrAssinaturaInvalida)
	}

	var corpoDaResposta struct {
		ID   string `json:"id"`
		Type string `json:"type"`
		Data struct {
			ObjetoArmazenado json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &corpoDaResposta); err != nil {
		return nil, fmt.Errorf("decodificando evento: %w", err)
	}
	if corpoDaResposta.ID == "" || corpoDaResposta.Type == "" {
		return nil, fmt.Errorf("%w: evento sem id ou type", ErrAssinaturaInvalida)
	}

	return &EventoDePagamento{
		ID:   corpoDaResposta.ID,
		Type: corpoDaResposta.Type,
		Data: corpoDaResposta.Data.ObjetoArmazenado,
	}, nil
}

// calcularAssinatura devolve o HMAC em hex de "<ts>.<payload>".
func calcularAssinatura(secret string, ts int64, payload []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(ts, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

// interpretarCabecalhoDeAssinatura separa `t=...,v1=...,v1=...`.
func interpretarCabecalhoDeAssinatura(header string) (time.Time, []string) {
	var ts time.Time
	var sigs []string

	for _, part := range strings.Split(header, ",") {
		kv := strings.SplitN(strings.TrimSpace(part), "=", 2)
		if len(kv) != 2 {
			continue
		}
		switch kv[0] {
		case "t":
			secs, err := strconv.ParseInt(kv[1], 10, 64)
			if err != nil {
				continue
			}
			ts = time.Unix(secs, 0).UTC()
		case "v1":
			sigs = append(sigs, kv[1])
		}
	}
	return ts, sigs
}

// anyMatch compares in constant time.
func algumConfere(expected string, candidates []string) bool {
	match := false
	for _, c := range candidates {
		if hmac.Equal([]byte(expected), []byte(c)) {
			match = true
		}
	}
	return match
}

// CabecalhoDeAssinaturaPara monta um `Stripe-Signature` válido. É usado
// the tests and by any local harness that replays webhook payloads.
func CabecalhoDeAssinaturaPara(secret string, payload []byte, at time.Time) string {
	ts := at.UTC().Unix()
	return fmt.Sprintf("t=%d,v1=%s", ts, calcularAssinatura(secret, ts, payload))
}
