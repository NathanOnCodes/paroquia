// Package storage abstracts the file volume behind a small interface so the
// slices never depend on the concrete backend.
//
// The reference implementation writes to the local filesystem. The interface
// is intentionally S3-shaped (Save/Open/Delete/URL) so an object store can be
// dropped in later without touching any slice.
package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// Common error values returned by every implementation.
var (
	ErrNaoEncontrado       = errors.New("arquivo não encontrado")
	ErrCaminhoInvalido     = errors.New("caminho de arquivo inválido")
	ErrArquivoGrandeDemais = errors.New("arquivo excede o tamanho máximo permitido")
	ErrTipoInvalido        = errors.New("tipo de arquivo não permitido")
)

// Object is the metadata of a stored file.
type ObjetoArmazenado struct {
	Key      string
	Size     int64
	Modified time.Time
}

// Storer is the storage contract used by the slices.
type ArmazenamentoDeArquivos interface {
	// Save writes the reader to key and returns the number of bytes written.
	Salvar(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error)
	// Open returns a reader for key.
	Open(ctx context.Context, key string) (io.ReadCloser, ObjetoArmazenado, error)
	// Delete removes key. Missing files are not an error: deletion is used for
	// rollback and must stay idempotent.
	Remover(ctx context.Context, key string) error
	// Exists reports whether key is present.
	Existe(ctx context.Context, key string) (bool, error)
}

// ConfiguracaoDoSistemaDeArquivos configures the filesystem implementation.
type ConfiguracaoDoSistemaDeArquivos struct {
	// Root is the absolute or relative path of the volume.
	Root string
	// MaxFileSizeBytes rejects oversized uploads.
	TamanhoMaximoArquivoBytes int64
}

// FS stores objects on the local filesystem, under Root.
//
// Keys are always slash-separated and relative; the implementation rejects
// anything that could escape the root.
type SistemaDeArquivos struct {
	root          string
	maximoDeBytes int64
}

// NewFS constructs the filesystem storage and makes sure the root exists.
func NovoSistemaDeArquivos(cfg ConfiguracaoDoSistemaDeArquivos) (*SistemaDeArquivos, error) {
	if cfg.Root == "" {
		return nil, errors.New("diretório de storage não configurado")
	}
	abs, err := filepath.Abs(cfg.Root)
	if err != nil {
		return nil, fmt.Errorf("resolvendo diretório de storage: %w", err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("criando diretório de storage: %w", err)
	}
	return &SistemaDeArquivos{root: abs, maximoDeBytes: cfg.TamanhoMaximoArquivoBytes}, nil
}

// Save writes the stream to key. Writes are atomic: the content goes to a
// temporary file first and is renamed into place, so a crash cannot leave a
// half-written object behind.
func (f *SistemaDeArquivos) Salvar(ctx context.Context, key string, r io.Reader, size int64, contentType string) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if f.maximoDeBytes > 0 && size > f.maximoDeBytes {
		return 0, fmt.Errorf("%w: %d > %d", ErrArquivoGrandeDemais, size, f.maximoDeBytes)
	}

	full, err := f.resolver(key)
	if err != nil {
		return 0, err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return 0, fmt.Errorf("criando diretório: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(full), ".upload-*")
	if err != nil {
		return 0, fmt.Errorf("criando arquivo temporário: %w", err)
	}
	nomeTemporario := tmp.Name()

	// LimitedReader enforces the ceiling even when the caller lies about size.
	var src io.Reader = r
	if f.maximoDeBytes > 0 {
		src = io.LimitReader(r, f.maximoDeBytes+1)
	}

	written, cerr := io.Copy(tmp, src)
	closeErr := tmp.Close()
	if cerr != nil {
		_ = os.Remove(nomeTemporario)
		return 0, fmt.Errorf("gravando arquivo: %w", cerr)
	}
	if closeErr != nil {
		_ = os.Remove(nomeTemporario)
		return 0, fmt.Errorf("fechando arquivo: %w", closeErr)
	}
	if f.maximoDeBytes > 0 && written > f.maximoDeBytes {
		_ = os.Remove(nomeTemporario)
		return 0, fmt.Errorf("%w: %d > %d", ErrArquivoGrandeDemais, written, f.maximoDeBytes)
	}

	if err := os.Rename(nomeTemporario, full); err != nil {
		_ = os.Remove(nomeTemporario)
		return 0, fmt.Errorf("publicando arquivo: %w", err)
	}
	return written, nil
}

// Open returns a reader over the stored object.
func (f *SistemaDeArquivos) Open(ctx context.Context, key string) (io.ReadCloser, ObjetoArmazenado, error) {
	if err := ctx.Err(); err != nil {
		return nil, ObjetoArmazenado{}, err
	}
	full, err := f.resolver(key)
	if err != nil {
		return nil, ObjetoArmazenado{}, err
	}
	info, err := os.Stat(full)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ObjetoArmazenado{}, ErrNaoEncontrado
		}
		return nil, ObjetoArmazenado{}, fmt.Errorf("lendo metadados: %w", err)
	}
	fh, err := os.Open(full)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, ObjetoArmazenado{}, ErrNaoEncontrado
		}
		return nil, ObjetoArmazenado{}, fmt.Errorf("abrindo arquivo: %w", err)
	}
	return fh, ObjetoArmazenado{Key: key, Size: info.Size(), Modified: info.ModTime()}, nil
}

// Delete removes key; a missing file is reported as success.
func (f *SistemaDeArquivos) Remover(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	full, err := f.resolver(key)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return fmt.Errorf("removendo arquivo: %w", err)
	}
	return nil
}

// Exists reports whether key is present.
func (f *SistemaDeArquivos) Existe(ctx context.Context, key string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	full, err := f.resolver(key)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(full)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, fs.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("verificando arquivo: %w", err)
}

// resolve maps a logical key to an absolute path, refusing traversal.
func (f *SistemaDeArquivos) resolver(key string) (string, error) {
	if key == "" {
		return "", ErrCaminhoInvalido
	}
	clean := path.Clean("/" + key)
	if clean == "/" || strings.Contains(clean, "..") {
		return "", ErrCaminhoInvalido
	}
	full := filepath.Join(f.root, filepath.FromSlash(clean))
	// Belt and braces: the result must stay inside the root.
	if !strings.HasPrefix(full, f.root+string(os.PathSeparator)) {
		return "", ErrCaminhoInvalido
	}
	return full, nil
}
