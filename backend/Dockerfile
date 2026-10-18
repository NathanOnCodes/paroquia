# syntax=docker/dockerfile:1.7

# ==========================================================================
# Build multi-estágio da API da paróquia.
#
# Estágios:
#   deps    - baixa e cacheia as dependências
#   test    - roda os testes (opcional, via --target test)
#   build   - compila o binário estático
#   runtime - imagem final mínima, sem shell e sem usuário root
# ==========================================================================


# --------------------------------------------------------------------------
# deps
# --------------------------------------------------------------------------
FROM golang:1.25-alpine AS deps

WORKDIR /src

ENV CGO_ENABLED=0 \
    GOFLAGS=-mod=mod

# Só o manifesto primeiro: a camada de dependências só muda quando ele muda.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify


# --------------------------------------------------------------------------
# test
# --------------------------------------------------------------------------
FROM deps AS test

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go vet ./... && go test ./...


# --------------------------------------------------------------------------
# build
# --------------------------------------------------------------------------
FROM deps AS build

COPY . .

ARG VERSAO=dev
ARG DATA_BUILD=desconhecida

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath \
      -ldflags="-s -w -X main.versao=${VERSAO} -X main.dataBuild=${DATA_BUILD}" \
      -o /out/server ./cmd/server


# --------------------------------------------------------------------------
# runtime - distroless: sem shell, sem gerenciador de pacotes, usuário não-root
# --------------------------------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot AS runtime

WORKDIR /app

COPY --from=build --chown=nonroot:nonroot /out/server /app/server

# Volume onde ficam imagens de eventos e documentos de transparência.
VOLUME ["/app/data"]

# Porta padrão; sobrescrita por HTTP_ADDR.
EXPOSE 8080

USER nonroot:nonroot

ENTRYPOINT ["/app/server"]
