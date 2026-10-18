// Package authx owns authentication and authorization: password hashing,
// token issuing/parsing, cookie handling and the RBAC middlewares.
//
// It is deliberately free of any database or framework dependency beyond
// Fiber: the caller supplies the user lookup, which keeps this package
// testable and reusable.
package authx

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters. They follow the OWASP password storage guidance for
// interactive logins (19 MiB of memory, 2 iterations, parallelism 1) and are
// stored in the PHC string format so they can be upgraded later without
// invalidating existing hashes.
const (
	argonTime    uint32 = 2
	argonMemory  uint32 = 19 * 1024 // KiB
	argonThreads uint8  = 1
	argonKeyLen  uint32 = 32
	argonSaltLen        = 16
)

// HashPassword derives an argon2id hash and returns it as a PHC string:
//
//	$argon2id$v=19$m=19456,t=2,p=1$<salt b64>$<hash b64>
func GerarHashDeSenha(plain string) (string, error) {
	if plain == "" {
		return "", errors.New("senha vazia")
	}
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("gerando salt: %w", err)
	}
	sum := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, argonKeyLen)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonThreads,
		codificarB64(salt), codificarB64(sum),
	), nil
}

// VerifyPassword reports whether plain matches the PHC-encoded hash. The
// comparison is constant time.
func VerificarSenha(encoded, plain string) bool {
	salt, want, err := decodificarPHC(encoded)
	if err != nil {
		return false
	}
	got := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	return subtle.ConstantTimeCompare(got, want) == 1
}

// NeedsRehash reports whether a stored hash uses weaker parameters than the
// current ones and should be replaced after a successful login.
func PrecisaRegerarHash(encoded string) bool {
	_, _, err := decodificarPHC(encoded)
	return err != nil
}

func decodificarPHC(encoded string) (salt, sum []byte, err error) {
	// $argon2id$v=19$m=19456,t=2,p=1$SALT$HASH
	parts := dividirPHC(encoded)
	if len(parts) != 6 || parts[1] != "argon2id" {
		return nil, nil, errors.New("formato de hash inválido")
	}
	if salt, err = decodificarB64(parts[4]); err != nil {
		return nil, nil, fmt.Errorf("salt inválido: %w", err)
	}
	if sum, err = decodificarB64(parts[5]); err != nil {
		return nil, nil, fmt.Errorf("hash inválido: %w", err)
	}
	return salt, sum, nil
}

// splitPHC splits on '$' while keeping the leading empty element produced by
// the mandatory leading '$'.
func dividirPHC(s string) []string {
	out := make([]string, 0, 6)
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '$' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	return append(out, s[start:])
}

// RandomToken returns a cryptographically random hex string of n bytes
// (2n characters). It is the bearer of magic links and refresh tokens.
func TokenAleatorio(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		// crypto/rand.Read never fails on supported platforms; panic loudly
		// rather than silently issuing a predictable token.
		panic(fmt.Sprintf("crypto/rand falhou: %v", err))
	}
	return hex.EncodeToString(b)
}

// HashToken returns the SHA-256 hex digest of a bearer token. Only this value
// (comentário removido: texto original indisponível)
func HashDoToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// TokenEqual compares a raw bearer against a stored hash in constant time.
func TokensIguais(raw, storedHash string) bool {
	got := HashDoToken(raw)
	return subtle.ConstantTimeCompare([]byte(got), []byte(storedHash)) == 1
}

// b64Encode/b64Decode use the unpadded standard base64 alphabet required by
// the PHC string format.
func codificarB64(b []byte) string { return base64.RawStdEncoding.EncodeToString(b) }

func decodificarB64(s string) ([]byte, error) { return base64.RawStdEncoding.DecodeString(s) }
