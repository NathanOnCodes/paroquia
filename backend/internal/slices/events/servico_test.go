package events

import "testing"

func TestSlugificar(t *testing.T) {
	casos := []struct {
		entrada  string
		esperado string
	}{
		{"Missa de Natal", "missa-de-natal"},
		{"  Vigília de Ano Novo  ", "viglia-de-ano-novo"},
		// Acentos caem fora: o charset da regra é [a-z0-9-].
		{"Café com Pão — 2026", "caf-com-po-2026"},
		{"Festa Junina!!!", "festa-junina"},
		{"reunião-dos-jovens", "reunio-dos-jovens"},
		{"", ""},
	}

	for _, caso := range casos {
		if obtido := Slugificar(caso.entrada); obtido != caso.esperado {
			t.Fatalf("Slugificar(%q) = %q, esperado %q", caso.entrada, obtido, caso.esperado)
		}
	}
}

func TestSlugificarRespeitaOCharsetDaRegra(t *testing.T) {
	// O slug só pode conter caixa-baixa, dígitos e hífens — é o mesmo charset
	// imposto pela validação `slug` do contrato HTTP.
	obtido := Slugificar("Encontro de Pastoral 2026 / 2ª Etapa")
	for _, r := range obtido {
		valido := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-'
		if !valido {
			t.Fatalf("caractere %q fora do charset permitido em %q", r, obtido)
		}
	}
}
