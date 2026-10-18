package settings

import (
	"regexp"
	"strings"
)

// Contratos de requisição e resposta. Espelham o `siteSettingsSchema` legado
// existing frontend keeps working unchanged.

// UpdateRequest is the admin payload that rewrites the settings row.
// Every string is optional in the wire sense but validated when present; blank
// Pix and home fields are stored as NULL, exactly like the legacy behaviour.
type RequisicaoDeAtualizacao struct {
	SiteName        string `json:"site_name" validate:"required,min=2,max=120"`
	SiteTitle       string `json:"site_title" validate:"required,min=2,max=120"`
	City            string `json:"city" validate:"max=120"`
	Description     string `json:"description" validate:"required,min=2,max=240"`
	SecondaryColor  string `json:"secondary_color" validate:"required,hexcolor6"`
	PixKey          string `json:"pix_key" validate:"max=254"`
	PixReceiverName string `json:"pix_receiver_name" validate:"max=25"`
	PixReceiverCity string `json:"pix_receiver_city" validate:"max=15"`
	HomeTitle       string `json:"home_title" validate:"max=180"`
	HomeDescription string `json:"home_description" validate:"max=400"`
}

// Response is the public shape of the settings row. It is served verbatim to
// every visitor, so it must never contain anything private.
type Resposta struct {
	SiteName        string  `json:"site_name"`
	SiteTitle       string  `json:"site_title"`
	City            string  `json:"city"`
	Description     string  `json:"description"`
	SecondaryColor  string  `json:"secondary_color"`
	PixKey          *string `json:"pix_key"`
	PixReceiverName *string `json:"pix_receiver_name"`
	PixReceiverCity *string `json:"pix_receiver_city"`
	HomeTitle       *string `json:"home_title"`
	HomeDescription *string `json:"home_description"`
}

// Os padrões espelham os DEFAULTs SQL de `site_settings`; assim uma linha
// renders a coherent page.
var Padroes = Resposta{
	SiteName:       "Paroquia",
	SiteTitle:      "Paroquia",
	City:           "",
	Description:    "Comunidade, fe e acolhimento.",
	SecondaryColor: "#1D4E89",
}

var (
	emailRe = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)
	evpRe   = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
	phoneRe = regexp.MustCompile(`^\+?[0-9]{10,13}$`)
	cpfRe   = regexp.MustCompile(`^[0-9]{11}$`)
	cnpjRe  = regexp.MustCompile(`^[0-9]{14}$`)
)

// validatePixKey accepts one of the five official Pix key kinds: CPF, CNPJ,
// e-mail, phone and EVP (random UUID).
//
// The legacy system validated the key by generating a static Pix payload with
// `pix-utils`; essa biblioteca não tem equivalente em Go, então as regras
// implemented directly. Generating the payload stays a frontend concern.
func validarChavePix(key string) (string, bool) {
	k := strings.TrimSpace(key)
	if k == "" {
		return "", false
	}
	switch {
	case cpfRe.MatchString(k):
		return k, true
	case cnpjRe.MatchString(k):
		return k, true
	case emailRe.MatchString(k):
		return strings.ToLower(k), true
	case phoneRe.MatchString(k):
		return k, true
	case evpRe.MatchString(k):
		return strings.ToLower(k), true
	default:
		return "", false
	}
}
