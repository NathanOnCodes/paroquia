// Package mail sends transactional e-mail.
//
// The interface is minimal on purpose: the slices only ever ask for "send this
// message", and the concrete transport (Resend today) is injected at startup.
package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Message is one outgoing e-mail. Text is optional: when empty it is derived
// from the HTML body so plain-text clients still get readable content.
type MensagemDeEmail struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

// Sender delivers messages.
type RemetenteDeEmail interface {
	Enviar(ctx context.Context, msg MensagemDeEmail) error
}

// ErrNotConfigured is returned when no transport is available.
var ErrNaoConfigurado = errors.New("transporte de e-mail não configurado")

// Config tunes the Resend transport.
type Configuracao struct {
	APIKey string
	From   string
	// MaxAttempts is the number of tries before giving up (default 3).
	MaximoDeTentativas int
	// RetryBackoff is the base delay between tries (default 2s).
	EsperaEntreTentativas time.Duration
	// Timeout bounds each HTTP call (default 10s).
	Timeout time.Duration
	// Endpoint overrides the API base, for tests.
	Endpoint string
}

// Resend is the Resend.com transport.
type ClienteResend struct {
	cfg    Configuracao
	client *http.Client
}

// NewResend builds the transport. When APIKey is empty the returned Sender is
// disabled and Send reports ErrNotConfigured, which the callers treat as a
// soft failure — a missing e-mail must never block a donation.
func NovoClienteResend(cfg Configuracao) *ClienteResend {
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://api.resend.com/emails"
	}
	if cfg.MaximoDeTentativas <= 0 {
		cfg.MaximoDeTentativas = 3
	}
	if cfg.EsperaEntreTentativas <= 0 {
		cfg.EsperaEntreTentativas = 2 * time.Second
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 10 * time.Second
	}
	return &ClienteResend{cfg: cfg, client: &http.Client{Timeout: cfg.Timeout}}
}

// Enabled reports whether the transport can actually deliver.
func (r *ClienteResend) Habilitado() bool { return r.cfg.APIKey != "" }

type cargaDoResend struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html,omitempty"`
	Text    string   `json:"text,omitempty"`
}

// Send delivers the message with retries and exponential backoff.
func (r *ClienteResend) Enviar(ctx context.Context, msg MensagemDeEmail) error {
	if !r.Habilitado() {
		return ErrNaoConfigurado
	}
	if strings.TrimSpace(msg.To) == "" {
		return errors.New("destinatário vazio")
	}
	if msg.Text == "" {
		msg.Text = RemoverHTML(msg.HTML)
	}

	body, err := json.Marshal(cargaDoResend{
		From:    r.cfg.From,
		To:      []string{msg.To},
		Subject: msg.Subject,
		HTML:    msg.HTML,
		Text:    msg.Text,
	})
	if err != nil {
		return fmt.Errorf("serializando e-mail: %w", err)
	}

	var ultimoErro error
	for attempt := 1; attempt <= r.cfg.MaximoDeTentativas; attempt++ {
		if attempt > 1 {
			backoff := r.cfg.EsperaEntreTentativas * time.Duration(1<<(attempt-2))
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.Endpoint, bytes.NewReader(body))
		if err != nil {
			return fmt.Errorf("montando requisição: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+r.cfg.APIKey)
		req.Header.Set("Content-Type", "application/json")

		resp, err := r.client.Do(req)
		if err != nil {
			ultimoErro = fmt.Errorf("enviando e-mail: %w", err)
			continue
		}

		payload, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		_ = resp.Body.Close()

		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return nil
		}

		ultimoErro = fmt.Errorf("resend respondeu %d: %s", resp.StatusCode, strings.TrimSpace(string(payload)))
		// 4xx (exceto 429) não melhora com retry.
		if resp.StatusCode >= 400 && resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
			return ultimoErro
		}
	}
	return ultimoErro
}

// Nop is a Sender that discards everything. Used in tests and in local demos.
type Nulo struct{ Enviadas []MensagemDeEmail }

func (n *Nulo) Enviar(_ context.Context, msg MensagemDeEmail) error {
	n.Enviadas = append(n.Enviadas, msg)
	return nil
}

// StripHTML turns an HTML body into readable plain text. It is intentionally
// crude: the result is only used as the fallback for text-only mail clients.
func RemoverHTML(html string) string {
	if html == "" {
		return ""
	}
	var b strings.Builder
	inTag := false
	for _, r := range html {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			b.WriteRune(r)
		}
	}
	text := b.String()
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&lt;", "<")
	text = strings.ReplaceAll(text, "&gt;", ">")

	lines := strings.Split(text, "\n")
	out := lines[:0]
	for _, ln := range lines {
		if strings.TrimSpace(ln) != "" {
			out = append(out, strings.TrimRight(ln, " \t"))
		}
	}
	return strings.TrimSpace(strings.Join(out, "\n"))
}

// EscapeHTML escapes the five XML entities. Used by every template before a
// user-supplied value is interpolated.
func EscaparHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#39;",
	)
	return replacer.Replace(s)
}
