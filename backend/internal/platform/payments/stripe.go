// Package payments wraps the payment provider behind a narrow interface.
//
// The reference implementation talks to the Stripe REST API directly instead
// of pulling the official SDK: the surface actually used is small (five
// endpoints) and keeping it hand-written makes the payload shapes explicit,
// which matters because the webhook handler depends on them.
//
// Everything the slices see is expressed in domain terms; nothing leaks the
// provider's vocabulary.
package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Common errors surfaced to the domain.
var (
	ErrNaoConfigurado     = errors.New("gateway de pagamento não configurado")
	ErrAssinaturaInvalida = errors.New("assinatura de webhook inválida")
	ErrEventoDesconhecido = errors.New("evento de webhook desconhecido")
)

// Customer is the payer as known by the provider.
type ClienteDePagamento struct {
	ID    string
	Email string
	Name  string
	Phone string
}

// PaymentIntent is the setup of a single charge.
type IntencaoDePagamento struct {
	ID       string
	Secret   string
	Amount   int64
	Currency string
	Status   string
}

// SetupIntent prepares a payment method for future recurring charges.
type IntencaoDeConfiguracao struct {
	ID     string
	Secret string
	Status string
}

// Subscription is a recurring arrangement.
type Assinatura struct {
	ID         string
	Status     string
	CustomerID string
	// CurrentPeriodStart/End bound the running billing cycle.
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CancelAtPeriodEnd  bool
	CanceledAt         *time.Time
	Metadata           map[string]string
}

// Invoice is one billing attempt of a subscription.
type Fatura struct {
	ID         string
	Status     string
	AmountDue  int64
	AmountPaid int64
	Assinatura string
	// Metadata comes from the subscription that produced the invoice.
	Metadata map[string]string
}

// Charge is the money movement behind a payment intent.
type Cobranca struct {
	ID                  string
	IntencaoDePagamento string
	Refunded            bool
	AmountRefunded      int64
}

// Event is a webhook notification. Data holds the raw object so the handler
// can decode only what it needs.
type EventoDePagamento struct {
	ID   string
	Type string
	Data json.RawMessage
}

// Gateway is the payment provider contract used by the slices.
type MeioDePagamento interface {
	// CreateCustomer registers (or reuses) a payer.
	CriarCliente(ctx context.Context, in EntradaParaCriarCliente) (*ClienteDePagamento, error)
	// CreatePaymentIntent prepares a one-off charge and returns the
	// client secret the browser needs to confirm it.
	CriarIntencaoDePagamento(ctx context.Context, in EntradaDeIntencaoDePagamento) (*IntencaoDePagamento, error)
	// CreateSetupIntent prepares a card/Pix for future recurring charges.
	CriarIntencaoDeConfiguracao(ctx context.Context, in EntradaDeIntencaoDeConfiguracao) (*IntencaoDeConfiguracao, error)
	// CreateSubscription starts a monthly recurring donation.
	CriarAssinatura(ctx context.Context, in EntradaDeAssinatura) (*Assinatura, error)
	// CancelAtPeriodEnd asks the provider to stop renewing at period end.
	CancelAtPeriodEnd(ctx context.Context, subscriptionID string) error
	// ParseWebhook verifies the signature and decodes the event.
	InterpretarWebhook(payload []byte, signature string) (*EventoDePagamento, error)
}

// Input types.

type EntradaParaCriarCliente struct {
	Email    string
	Name     string
	Phone    string
	Metadata map[string]string
}

type EntradaDeIntencaoDePagamento struct {
	Amount       int64
	Currency     string
	CustomerID   string
	ReceiptEmail string
	Metadata     map[string]string
}

type EntradaDeIntencaoDeConfiguracao struct {
	CustomerID string
	Metadata   map[string]string
}

type EntradaDeAssinatura struct {
	CustomerID      string
	Amount          int64
	Currency        string
	PaymentMethodID string
	Metadata        map[string]string
	// ProductName labels the recurring product shown on the invoice.
	ProductName string
}

// Provider errors carry the HTTP status so callers can tell a bad request from
// an outage.
type ErroDoProvedor struct {
	Status   int
	Codigo   string
	Mensagem string
}

func (e *ErroDoProvedor) Error() string {
	return fmt.Sprintf("provedor respondeu %d (%s): %s", e.Status, e.Codigo, e.Mensagem)
}

// ---------------------------------------------------------------------------
// Stripe implementation
// ---------------------------------------------------------------------------

const URLBaseDoStripe = "https://api.stripe.com"

// StripeConfig configures the Stripe client.
type ConfiguracaoDoStripe struct {
	ChaveSecreta     string
	SegredoDoWebhook string
	Currency         string
	// APIVersion pins the payload shapes the webhook handler expects.
	VersaoDaAPI string
	// Timeout bounds every call (default 20s).
	Timeout time.Duration
	// BaseURL overrides the API root, for tests.
	BaseURL string
}

// Stripe implements Gateway against the Stripe REST API.
type ClienteStripe struct {
	cfg    ConfiguracaoDoStripe
	client *http.Client
	now    func() time.Time
}

// NewStripe builds the client. An empty SecretKey yields a disabled client
// whose methods all report ErrNotConfigured.
func NovoClienteStripe(cfg ConfiguracaoDoStripe) *ClienteStripe {
	if cfg.BaseURL == "" {
		cfg.BaseURL = URLBaseDoStripe
	}
	if cfg.Currency == "" {
		cfg.Currency = "brl"
	}
	if cfg.Timeout <= 0 {
		cfg.Timeout = 20 * time.Second
	}
	return &ClienteStripe{
		cfg:    cfg,
		client: &http.Client{Timeout: cfg.Timeout},
		now:    time.Now,
	}
}

// Enabled reports whether the gateway can be used.
func (s *ClienteStripe) Habilitado() bool { return s.cfg.ChaveSecreta != "" }

// SetClock injects a clock, for tests.
func (s *ClienteStripe) DefinirRelogio(clock func() time.Time) {
	if clock != nil {
		s.now = clock
	}
}

func (s *ClienteStripe) CriarCliente(ctx context.Context, in EntradaParaCriarCliente) (*ClienteDePagamento, error) {
	form := url.Values{}
	form.Set("email", in.Email)
	if in.Name != "" {
		form.Set("name", in.Name)
	}
	if in.Phone != "" {
		form.Set("phone", in.Phone)
	}
	for k, v := range in.Metadata {
		form.Set("metadata["+k+"]", v)
	}

	var out struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Phone string `json:"phone"`
	}
	if err := s.post(ctx, "/v1/customers", form, &out); err != nil {
		return nil, err
	}
	return &ClienteDePagamento{ID: out.ID, Email: out.Email, Name: out.Name, Phone: out.Phone}, nil
}

func (s *ClienteStripe) CriarIntencaoDePagamento(ctx context.Context, in EntradaDeIntencaoDePagamento) (*IntencaoDePagamento, error) {
	form := url.Values{}
	form.Set("amount", strconv.FormatInt(in.Amount, 10))
	form.Set("currency", primeiroNaoVazio(in.Currency, s.cfg.Currency))
	form.Set("automatic_payment_methods[enabled]", "true")
	if in.CustomerID != "" {
		form.Set("customer", in.CustomerID)
	}
	if in.ReceiptEmail != "" {
		form.Set("receipt_email", in.ReceiptEmail)
	}
	for k, v := range in.Metadata {
		form.Set("metadata["+k+"]", v)
	}

	var out struct {
		ID           string `json:"id"`
		ClientSecret string `json:"client_secret"`
		Amount       int64  `json:"amount"`
		Currency     string `json:"currency"`
		Status       string `json:"status"`
	}
	if err := s.post(ctx, "/v1/payment_intents", form, &out); err != nil {
		return nil, err
	}
	return &IntencaoDePagamento{
		ID:       out.ID,
		Secret:   out.ClientSecret,
		Amount:   out.Amount,
		Currency: out.Currency,
		Status:   out.Status,
	}, nil
}

func (s *ClienteStripe) CriarIntencaoDeConfiguracao(ctx context.Context, in EntradaDeIntencaoDeConfiguracao) (*IntencaoDeConfiguracao, error) {
	form := url.Values{}
	form.Set("customer", in.CustomerID)
	form.Set("usage", "off_session")
	form.Set("payment_method_types[0]", "card")
	form.Set("payment_method_types[1]", "pix")
	for k, v := range in.Metadata {
		form.Set("metadata["+k+"]", v)
	}

	var out struct {
		ID           string `json:"id"`
		ClientSecret string `json:"client_secret"`
		Status       string `json:"status"`
	}
	if err := s.post(ctx, "/v1/setup_intents", form, &out); err != nil {
		return nil, err
	}
	return &IntencaoDeConfiguracao{ID: out.ID, Secret: out.ClientSecret, Status: out.Status}, nil
}

func (s *ClienteStripe) CriarAssinatura(ctx context.Context, in EntradaDeAssinatura) (*Assinatura, error) {
	// A reusable price is looked up by nickname first; when it does not exist a
	// new product+price pair is created and re-used by every later call. This
	// fixes the legacy behaviour that created a fresh price per subscription.
	priceID, err := s.garantirPrecoRecorrente(ctx, in.Amount, in.ProductName)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("customer", in.CustomerID)
	form.Set("items[0][price]", priceID)
	form.Set("payment_behavior", "default_incomplete")
	form.Set("default_payment_method", in.PaymentMethodID)
	form.Set("payment_settings[save_default_payment_method]", "on_subscription")
	form.Set("payment_settings[payment_method_types][0]", "card")
	form.Set("payment_settings[payment_method_types][1]", "pix")
	for k, v := range in.Metadata {
		form.Set("metadata["+k+"]", v)
	}

	var out cargaDaAssinatura
	if err := s.post(ctx, "/v1/subscriptions", form, &out); err != nil {
		return nil, err
	}
	return out.paraAssinatura(), nil
}

func (s *ClienteStripe) CancelAtPeriodEnd(ctx context.Context, subscriptionID string) error {
	form := url.Values{}
	form.Set("cancel_at_period_end", "true")
	return s.post(ctx, "/v1/subscriptions/"+url.PathEscape(subscriptionID), form, nil)
}

// ensureRecurringPrice finds or creates the shared recurring price for the
// given amount, so subscriptions do not litter the account with prices.
func (s *ClienteStripe) garantirPrecoRecorrente(ctx context.Context, amount int64, product string) (string, error) {
	nickname := fmt.Sprintf("recorrente-%d-%s", amount, s.cfg.Currency)

	list := url.Values{}
	list.Set("active", "true")
	list.Set("type", "recurring")
	list.Set("limit", "100")

	var found struct {
		Data []struct {
			ID       string `json:"id"`
			Nickname string `json:"nickname"`
		} `json:"data"`
	}
	if err := s.get(ctx, "/v1/prices", list, &found); err == nil {
		for _, p := range found.Data {
			if p.Nickname == nickname {
				return p.ID, nil
			}
		}
	}

	form := url.Values{}
	form.Set("currency", s.cfg.Currency)
	form.Set("unit_amount", strconv.FormatInt(amount, 10))
	form.Set("recurring[interval]", "month")
	form.Set("nickname", nickname)
	form.Set("product_data[name]", primeiroNaoVazio(product, "Contribuição recorrente mensal"))

	var created struct {
		ID string `json:"id"`
	}
	if err := s.post(ctx, "/v1/prices", form, &created); err != nil {
		return "", err
	}
	return created.ID, nil
}

// ---------------------------------------------------------------------------
// HTTP plumbing
// ---------------------------------------------------------------------------

func (s *ClienteStripe) post(ctx context.Context, path string, form url.Values, out any) error {
	return s.do(ctx, http.MethodPost, path, strings.NewReader(form.Encode()), out)
}

func (s *ClienteStripe) get(ctx context.Context, path string, query url.Values, out any) error {
	return s.do(ctx, http.MethodGet, path+"?"+query.Encode(), nil, out)
}

func (s *ClienteStripe) do(ctx context.Context, method, path string, body io.Reader, out any) error {
	if !s.Habilitado() {
		return ErrNaoConfigurado
	}
	req, err := http.NewRequestWithContext(ctx, method, s.cfg.BaseURL+path, body)
	if err != nil {
		return fmt.Errorf("montando requisição: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.cfg.ChaveSecreta)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if s.cfg.VersaoDaAPI != "" {
		req.Header.Set("Stripe-Version", s.cfg.VersaoDaAPI)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("falha de rede com o provedor: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	payload, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("lendo resposta do provedor: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return interpretarErroDoProvedor(resp.StatusCode, payload)
	}
	if out == nil {
		return nil
	}
	if err := json.Unmarshal(payload, out); err != nil {
		return fmt.Errorf("decodificando resposta do provedor: %w", err)
	}
	return nil
}

func interpretarErroDoProvedor(status int, payload []byte) error {
	var body struct {
		Erro struct {
			Codigo   string `json:"code"`
			Mensagem string `json:"message"`
			Type     string `json:"type"`
		} `json:"error"`
	}
	_ = json.Unmarshal(payload, &body)
	msg := body.Erro.Mensagem
	if msg == "" {
		msg = strings.TrimSpace(string(payload))
	}
	return &ErroDoProvedor{Status: status, Codigo: body.Erro.Codigo, Mensagem: msg}
}

func primeiroNaoVazio(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// Payload shapes
// ---------------------------------------------------------------------------

type cargaDaAssinatura struct {
	ID                 string `json:"id"`
	Status             string `json:"status"`
	ClienteDePagamento string `json:"customer"`
	CancelAtPeriodEnd  bool   `json:"cancel_at_period_end"`
	CanceledAt         *int64 `json:"canceled_at"`
	Items              struct {
		Data []struct {
			CurrentPeriodStart int64 `json:"current_period_start"`
			CurrentPeriodEnd   int64 `json:"current_period_end"`
		} `json:"data"`
	} `json:"items"`
	Metadata map[string]string `json:"metadata"`
}

func (p *cargaDaAssinatura) paraAssinatura() *Assinatura {
	sub := &Assinatura{
		ID:                p.ID,
		Status:            p.Status,
		CustomerID:        p.ClienteDePagamento,
		CancelAtPeriodEnd: p.CancelAtPeriodEnd,
		Metadata:          p.Metadata,
	}
	if len(p.Items.Data) > 0 {
		sub.CurrentPeriodStart = time.Unix(p.Items.Data[0].CurrentPeriodStart, 0).UTC()
		sub.CurrentPeriodEnd = time.Unix(p.Items.Data[0].CurrentPeriodEnd, 0).UTC()
	}
	if p.CanceledAt != nil && *p.CanceledAt > 0 {
		t := time.Unix(*p.CanceledAt, 0).UTC()
		sub.CanceledAt = &t
	}
	return sub
}
