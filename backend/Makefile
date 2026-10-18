# ==========================================================================
# Comandos do backend da paróquia
#
#   make ajuda      lista todos os comandos
# ==========================================================================

.PHONY: ajuda dev build test vet lint generate migrate docker docker-test limpar

GO        ?= go
GOBIN     := $(shell $(GO) env GOPATH)/bin
VERSAO    ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
DATA_BUILD?= $(shell date -u +%Y-%m-%dT%H:%M:%SZ)

ajuda: ## Mostra os comandos disponíveis
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-14s\033[0m %s\n", $$1, $$2}'

dev: ## Sobe a API em modo desenvolvimento
	$(GO) run ./cmd/server

build: ## Compila o binário em ./bin/server
	CGO_ENABLED=0 $(GO) build -trimpath -ldflags="-s -w" -o bin/server ./cmd/server

test: ## Roda os testes
	$(GO) test ./...

vet: ## Analisa o código
	$(GO) vet ./...

generate: ## Regenera o código do Ent e dos contratos
	$(GO) generate ./ent
	@echo "Confira se git status está limpo: o código gerado é versionado."

check-generate: ## Garante que o código gerado está em dia (usado no CI)
	$(GO) generate ./ent
	@git diff --exit-code -- ent || (echo "Código gerado desatualizado: rode 'make generate'" && exit 1)

migrate: ## Aplica as migrations no banco definido em DATABASE_URL
	$(GO) run ./cmd/server --apenas-migrar

docker: ## Constrói a imagem de produção
	docker build --target runtime \
		--build-arg VERSAO=$(VERSAO) \
		--build-arg DATA_BUILD=$(DATA_BUILD) \
		-t paroquia-api:$(VERSAO) .

docker-test: ## Roda os testes dentro do Docker
	docker build --target test -t paroquia-api:test .

up: ## Sobe API + PostgreSQL via compose
	docker compose up --build

down: ## Para os containers
	docker compose down

limpar: ## Remove artefatos de build
	rm -rf bin/ coverage.out
