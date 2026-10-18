// Command server sobe a API da paróquia.
//
// A construção é linear e explícita: configurar, abrir o banco, montar a
// aplicação, servir. Nada é resolvido por reflexão ou injeção mágica.
package main

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/NathanOnCodes/paroquia/backend/internal/app"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/config"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/storage"
)

func main() {
	if err := executar(); err != nil {
		slog.Error("encerrando com erro", slog.String("erro", err.Error()))
		os.Exit(1)
	}
}

func executar() error {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: nivelDeLog(),
	}))

	cfg, err := config.Carregar()
	if err != nil {
		return err
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	// Banco: conecta e aplica as migrations versionadas embutidas.
	cliente, err := database.Open(ctx, cfg.BancoDeDados, database.Opcoes{
		Migrations:             database.MigrationsSQL,
		Dir:                    "migrations",
		RegistradorDeAuditoria: log,
	})
	if err != nil {
		return err
	}
	defer func() {
		if err := cliente.Close(); err != nil {
			log.Warn("falha ao fechar o banco", slog.String("erro", err.Error()))
		}
	}()

	// Armazenamento de arquivos em volume local.
	armazenamento, err := storage.NovoSistemaDeArquivos(storage.ConfiguracaoDoSistemaDeArquivos{
		Root:                      cfg.Armazenamento.Dir,
		TamanhoMaximoArquivoBytes: cfg.TamanhoMaximoArquivoBytes(),
	})
	if err != nil {
		return err
	}

	aplicacao, err := app.Novo(ctx, app.Dependencias{
		Config:        cfg,
		Log:           log,
		Cliente:       cliente,
		Armazenamento: armazenamento,
	})
	if err != nil {
		return err
	}

	// Servir em segundo plano e tratar o encerramento gracioso.
	erroDeServidor := make(chan error, 1)
	go func() {
		log.Info("servidor iniciando", slog.String("endereco", cfg.ServidorHTTP.Endereco))
		if err := aplicacao.Fiber.Listen(cfg.ServidorHTTP.Endereco); err != nil &&
			!errors.Is(err, context.Canceled) {
			erroDeServidor <- err
		}
	}()

	select {
	case err := <-erroDeServidor:
		return err
	case <-ctx.Done():
	}

	log.Info("encerrando servidor")
	aguardar := 10 * time.Second
	return aplicacao.Fiber.ShutdownWithTimeout(aguardar)
}

func nivelDeLog() slog.Level {
	switch os.Getenv("LOG_LEVEL") {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
