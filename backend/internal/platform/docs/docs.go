// Package docs serve a especificação OpenAPI e a interface de referência.
//
// A UI é a do Scalar, carregada da CDN e apontada para o spec versionado em
// `api/openapi.yaml`. Mantê-la fora do binário permite atualizar a
// documentação sem reconstruir a imagem.
package docs

import (
	_ "embed"

	"github.com/gofiber/fiber/v3"
)

//go:embed openapi.yaml
var especificacao []byte

// Registrar monta as rotas de documentação.
//
//	GET /docs        interface interativa (Scalar)
//	GET /openapi.yaml especificação bruta
func Registrar(app fiber.Router) {
	app.Get("/openapi.yaml", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/yaml; charset=utf-8")
		return c.Send(especificacao)
	})

	app.Get("/docs", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "text/html; charset=utf-8")
		return c.SendString(pagina)
	})
}

// pagina é a casca mínima que carrega o Scalar e o aponta para o spec.
const pagina = `<!doctype html>
<html lang="pt-BR">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>API da Paróquia — Documentação</title>
</head>
<body>
  <script id="api-reference" data-url="/openapi.yaml"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>
`
