package authx

import (
	"strings"
	"testing"
)

func TestGerarHashDeSenhaESempreDiferente(t *testing.T) {
	const senha = "senha-forte-123"

	primeiro, err := GerarHashDeSenha(senha)
	if err != nil {
		t.Fatalf("GerarHashDeSenha: %v", err)
	}
	segundo, err := GerarHashDeSenha(senha)
	if err != nil {
		t.Fatalf("GerarHashDeSenha: %v", err)
	}

	if primeiro == segundo {
		t.Fatal("dois hashes da mesma senha não podem ser iguais: o salt deve variar")
	}
	if !strings.HasPrefix(primeiro, "$argon2id$") {
		t.Fatalf("hash deve seguir o formato PHC, veio %q", primeiro)
	}
	if strings.Contains(primeiro, senha) {
		t.Fatal("o hash não pode conter a senha em texto puro")
	}
}

func TestVerificarSenha(t *testing.T) {
	const senha = "senha-forte-123"

	hash, err := GerarHashDeSenha(senha)
	if err != nil {
		t.Fatalf("GerarHashDeSenha: %v", err)
	}

	if !VerificarSenha(hash, senha) {
		t.Fatal("senha correta foi rejeitada")
	}
	if VerificarSenha(hash, senha+"x") {
		t.Fatal("senha incorreta foi aceita")
	}
	if VerificarSenha("hash-quebrado", senha) {
		t.Fatal("hash malformado não deveria validar nada")
	}
	if VerificarSenha("", senha) {
		t.Fatal("hash vazio não deveria validar nada")
	}
}

func TestGerarHashDeSenhaRejeitaVazia(t *testing.T) {
	if _, err := GerarHashDeSenha(""); err == nil {
		t.Fatal("senha vazia deveria ser rejeitada")
	}
}

func TestHashDoTokenNaoGuardaOOriginal(t *testing.T) {
	const bruto = "abcdef0123456789abcdef0123456789"

	hash := HashDoToken(bruto)
	if hash == bruto {
		t.Fatal("o hash do token não pode ser o próprio token")
	}
	if len(hash) != 64 {
		t.Fatalf("SHA-256 em hex tem 64 caracteres, veio %d", len(hash))
	}
	if !TokensIguais(bruto, hash) {
		t.Fatal("o token original deveria conferir com o hash guardado")
	}
	if TokensIguais("outro-token", hash) {
		t.Fatal("token diferente não pode conferir")
	}
}

func TestTokenAleatorioTemTamanhoEsperado(t *testing.T) {
	a := TokenAleatorio(32)
	b := TokenAleatorio(32)

	if len(a) != 64 {
		t.Fatalf("32 bytes em hex dão 64 caracteres, veio %d", len(a))
	}
	if a == b {
		t.Fatal("tokens aleatórios não podem coincidir")
	}
}

func TestPrecisaRegerarHashRejeitaFormatoDesconhecido(t *testing.T) {
	if !PrecisaRegerarHash("$2y$10$algumhashbcrypt") {
		t.Fatal("hash que não é argon2id deveria pedir regeração")
	}

	hash, err := GerarHashDeSenha("senha-forte-123")
	if err != nil {
		t.Fatalf("GerarHashDeSenha: %v", err)
	}
	if PrecisaRegerarHash(hash) {
		t.Fatal("hash argon2id válido não deveria pedir regeração")
	}
}
