package authx

import (
	"testing"
	"time"
)

func TestEmitirEInterpretarAcesso(t *testing.T) {
	gerenciador, err := NovoGerenciadorDeTokens("segredo-de-teste-com-pelo-menos-32-bytes!", OpcoesDoGerenciadorDeTokens{
		Issuer:              "paroquia-teste",
		TempoDeVidaDoAcesso: 15 * time.Minute,
	})
	if err != nil {
		t.Fatalf("NovoGerenciadorDeTokens: %v", err)
	}

	original := UsuarioAutenticado{
		ID:               "11111111-1111-1111-1111-111111111111",
		Email:            "admin@paroquia.org",
		FullName:         "Pároco",
		Role:             PapelAdmin,
		IsActive:         true,
		DeveTrocarASenha: false,
	}

	assinado, err := gerenciador.EmitirAcesso(original)
	if err != nil {
		t.Fatalf("EmitirAcesso: %v", err)
	}

	reivindicacoes, err := gerenciador.InterpretarAcesso(assinado)
	if err != nil {
		t.Fatalf("InterpretarAcesso: %v", err)
	}
	if reivindicacoes.UserID != original.ID {
		t.Fatalf("id do usuário divergiu: %q", reivindicacoes.UserID)
	}
	if reivindicacoes.Role != PapelAdmin {
		t.Fatalf("papel divergiu: %q", reivindicacoes.Role)
	}
}

func TestInterpretarAcessoRejeitaOutroSegredo(t *testing.T) {
	origem, _ := NovoGerenciadorDeTokens("segredo-de-teste-com-pelo-menos-32-bytes!", OpcoesDoGerenciadorDeTokens{})
	outro, _ := NovoGerenciadorDeTokens("outro-segredo-tambem-com-32-bytes-ok!!", OpcoesDoGerenciadorDeTokens{})

	assinado, err := origem.EmitirAcesso(UsuarioAutenticado{ID: "11111111-1111-1111-1111-111111111111", Role: PapelAdmin})
	if err != nil {
		t.Fatalf("EmitirAcesso: %v", err)
	}

	if _, err := outro.InterpretarAcesso(assinado); err == nil {
		t.Fatal("token assinado com outro segredo deveria ser rejeitado")
	}
}

func TestInterpretarAcessoRejeitaTokenExpirado(t *testing.T) {
	agora := time.Now()
	relogio := &relogioMutavel{atual: agora}

	gerenciador, err := NovoGerenciadorDeTokens("segredo-de-teste-com-pelo-menos-32-bytes!", OpcoesDoGerenciadorDeTokens{
		TempoDeVidaDoAcesso: 5 * time.Minute,
		Clock:               relogio.Agora,
	})
	if err != nil {
		t.Fatalf("NovoGerenciadorDeTokens: %v", err)
	}

	assinado, err := gerenciador.EmitirAcesso(UsuarioAutenticado{ID: "11111111-1111-1111-1111-111111111111"})
	if err != nil {
		t.Fatalf("EmitirAcesso: %v", err)
	}

	// Ainda válido dentro da janela de vida.
	if _, err := gerenciador.InterpretarAcesso(assinado); err != nil {
		t.Fatalf("token recente deveria ser aceito: %v", err)
	}

	// Depois de vencido (além dos 30s de tolerância), é rejeitado.
	relogio.Avancar(10 * time.Minute)
	if _, err := gerenciador.InterpretarAcesso(assinado); err == nil {
		t.Fatal("token vencido deveria ser rejeitado")
	}
}

// relogioMutavel permite simular a passagem do tempo nos testes.
type relogioMutavel struct{ atual time.Time }

func (r *relogioMutavel) Agora() time.Time      { return r.atual }
func (r *relogioMutavel) Avancar(d time.Duration) { r.atual = r.atual.Add(d) }

func TestNovoGerenciadorDeTokensExigeSegredoLongo(t *testing.T) {
	if _, err := NovoGerenciadorDeTokens("curto", OpcoesDoGerenciadorDeTokens{}); err == nil {
		t.Fatal("segredo curto deveria ser rejeitado")
	}
}

func TestHierarquiaDePapeis(t *testing.T) {
	admin := Reivindicacoes{Role: PapelAdmin}
	assistente := Reivindicacoes{Role: PapelAssistente}

	if !admin.TemPapel(PapelAdmin) || !admin.TemPapel(PapelAssistente) {
		t.Fatal("admin deve satisfazer qualquer exigência")
	}
	if !assistente.TemPapel(PapelAssistente) {
		t.Fatal("assistente deve satisfazer a própria exigência")
	}
	if assistente.TemPapel(PapelAdmin) {
		t.Fatal("assistente não pode satisfazer exigência de admin")
	}
	convidado := Reivindicacoes{Role: "convidado"}
	if convidado.TemPapel(PapelAssistente) {
		t.Fatal("papel desconhecido não deve satisfazer nada")
	}

	if !admin.EhAdmin() || assistente.EhAdmin() {
		t.Fatal("EhAdmin deve refletir exatamente o papel admin")
	}
}
