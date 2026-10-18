// Package database opens the Ent client and applies the versioned SQL
// migrations.
//
// The SQL migrations are the source of truth for the schema: Ent is used for
// typed queries only. Auto-migration from the Ent schemas is intentionally not
// (comentário removido: texto original indisponível)
package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/config"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	migrate_iofs "github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Options tunes the connection pool and migration behaviour.
type Opcoes struct {
	// Migrations is the embedded filesystem holding the .sql files.
	Migrations fs.FS
	// (comentário removido: texto original indisponível)
	Dir string
	// Logger receives migration progress.
	RegistradorDeAuditoria *slog.Logger
	// SkipMigrate disables the migration step (useful in tests).
	IgnorarMigracao bool
}

// Open connects to PostgreSQL, applies the migrations and returns a ready
// Ent client. The caller owns Close.
func Open(ctx context.Context, cfg config.BancoDeDados, opts Opcoes) (*ent.Client, error) {
	if cfg.URL == "" {
		return nil, errors.New("DATABASE_URL vazia")
	}

	db, err := sql.Open("pgx", cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("abrindo conexão: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaximoConexoesAbertas)
	db.SetMaxIdleConns(cfg.MaximoConexoesOciosas)
	db.SetConnMaxLifetime(cfg.TempoMaximoDeVidaDaConexao)

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := db.PingContext(pingCtx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("conectando ao PostgreSQL: %w", err)
	}

	if !opts.IgnorarMigracao && cfg.MigrarAutomaticamente {
		if err := Migrar(cfg.URL, opts); err != nil {
			_ = db.Close()
			return nil, err
		}
	}

	return ent.NewClient(ent.Driver(entsql.OpenDB(dialect.Postgres, db))), nil
}

// Migrate applies every pending SQL migration. It is idempotent.
func Migrar(databaseURL string, opts Opcoes) error {
	if opts.Migrations == nil {
		return errors.New("migrations não fornecidas")
	}
	dir := opts.Dir
	if dir == "" {
		dir = "migrations"
	}

	source, err := migrate_iofs.New(opts.Migrations, dir)
	if err != nil {
		return fmt.Errorf("lendo migrations embutidas: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, databaseURL)
	if err != nil {
		return fmt.Errorf("preparando migrador: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("aplicando migrations: %w", err)
	}

	version, dirty, verr := m.Version()
	if verr == nil && opts.RegistradorDeAuditoria != nil {
		opts.RegistradorDeAuditoria.Info("migrations aplicadas",
			slog.Uint64("version", uint64(version)),
			slog.Bool("dirty", dirty),
		)
	}
	return nil
}

// Ping reports whether the database is reachable by running a trivial
// (comentário removido: texto original indisponível)
// migrations have been applied.
func VerificarConexao(ctx context.Context, client *ent.Client) error {
	_, err := client.Usuario.Query().Exist(ctx)
	return err
}
