// Package transparency publica a prestação de contas da paróquia e gerencia
// os períodos, lançamentos e documentos que a compõem.
//
// Só o admin publica ou arquiva um período — regra garantida no serviço e,
// em segunda instância, pelo trigger `check_transparency_publish` no banco.
package transparency

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/documentotransparencia"
	"github.com/NathanOnCodes/paroquia/backend/ent/lancamentofinanceiro"
	"github.com/NathanOnCodes/paroquia/backend/ent/periodofinanceiro"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/database"
)

// Erros de domínio.
var (
	ErrNaoEncontrado     = errors.New("registro não encontrado")
	ErrPeriodoFechado    = errors.New("o período não aceita alterações")
	ErrDataFinalAnterior = errors.New("a data final deve ser posterior à inicial")
)

// RepositorioDePeriodos é o contrato dos períodos.
type RepositorioDePeriodos interface {
	ListarPublicos(ctx context.Context) ([]*ent.PeriodoFinanceiro, error)
	ObterPublico(ctx context.Context, id uuid.UUID) (*ent.PeriodoFinanceiro, error)
	ListarAdmin(ctx context.Context) ([]*ent.PeriodoFinanceiro, error)
	Obter(ctx context.Context, id uuid.UUID) (*ent.PeriodoFinanceiro, error)
	Criar(ctx context.Context, entrada EntradaDePeriodo) (*ent.PeriodoFinanceiro, error)
	Atualizar(ctx context.Context, id uuid.UUID, entrada EntradaDePeriodo) (*ent.PeriodoFinanceiro, error)
	// MudarStatus publica/arquiva. Exige que o ator seja admin: o banco também
	// confere.
	MudarStatus(ctx context.Context, id uuid.UUID, status string,
		publicadoPor *uuid.UUID, publicadoEm *time.Time, revisadoPor *uuid.UUID) (*ent.PeriodoFinanceiro, error)
}

// EntradaDePeriodo é a superfície editável de um período.
type EntradaDePeriodo struct {
	Nome      string
	InicioEm  time.Time
	FimEm     time.Time
	Status    string
	CriadoPor *uuid.UUID
}

// EntradaDeLancamento descreve uma linha de receita ou despesa.
type EntradaDeLancamento struct {
	IDDoPeriodo   uuid.UUID
	Tipo          string
	Categoria     string
	Descricao     string
	ValorCentavos int
	Data          time.Time
	Origem        string
	CriadoPor     *uuid.UUID
}

// RepositorioDeLancamentos é o contrato dos lançamentos.
type RepositorioDeLancamentos interface {
	ListarPorPeriodo(ctx context.Context, idDoPeriodo uuid.UUID) ([]*ent.LancamentoFinanceiro, error)
	Criar(ctx context.Context, entrada EntradaDeLancamento) (*ent.LancamentoFinanceiro, error)
	Remover(ctx context.Context, id uuid.UUID) error
	// SomarPorPeriodo devolve receitas e despesas sem trazer as linhas.
	SomarPorPeriodo(ctx context.Context, idDoPeriodo uuid.UUID) (receitas, despesas int64, err error)
}

// EntradaDeDocumento descreve um anexo do período.
type EntradaDeDocumento struct {
	IDDoPeriodo      uuid.UUID
	CaminhoDoArquivo string
	NomeOriginal     string
	TipoMIME         string
	Tamanho          int64
	CriadoPor        *uuid.UUID
}

// RepositorioDeDocumentos é o contrato dos anexos.
type RepositorioDeDocumentos interface {
	ListarPorPeriodo(ctx context.Context, idDoPeriodo uuid.UUID) ([]*ent.DocumentoTransparencia, error)
	Obter(ctx context.Context, id uuid.UUID) (*ent.DocumentoTransparencia, error)
	Criar(ctx context.Context, entrada EntradaDeDocumento) (*ent.DocumentoTransparencia, error)
	// ObterComPeriodo devolve o documento com o status do período, que decide
	// se o arquivo pode ser baixado.
	ObterComPeriodo(ctx context.Context, id uuid.UUID) (*ent.DocumentoTransparencia, error)
}

// ---------------------------------------------------------------------------
// Implementação Ent
// ---------------------------------------------------------------------------

type repositorioDePeriodos struct{ client *ent.Client }

// NovoRepositorioDePeriodos monta o repositório Ent dos períodos.
func NovoRepositorioDePeriodos(client *ent.Client) RepositorioDePeriodos {
	return &repositorioDePeriodos{client: client}
}

func (r *repositorioDePeriodos) ListarPublicos(ctx context.Context) ([]*ent.PeriodoFinanceiro, error) {
	return r.client.PeriodoFinanceiro.Query().
		Where(periodofinanceiro.StatusEQ(periodofinanceiro.StatusPublished)).
		Order(ent.Desc(periodofinanceiro.FieldStartDate)).
		All(ctx)
}

func (r *repositorioDePeriodos) ObterPublico(ctx context.Context, id uuid.UUID) (*ent.PeriodoFinanceiro, error) {
	linha, err := r.client.PeriodoFinanceiro.Query().
		Where(
			periodofinanceiro.ID(id),
			periodofinanceiro.StatusEQ(periodofinanceiro.StatusPublished),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDePeriodos) ListarAdmin(ctx context.Context) ([]*ent.PeriodoFinanceiro, error) {
	return r.client.PeriodoFinanceiro.Query().
		Order(ent.Desc(periodofinanceiro.FieldStartDate)).
		All(ctx)
}

func (r *repositorioDePeriodos) Obter(ctx context.Context, id uuid.UUID) (*ent.PeriodoFinanceiro, error) {
	linha, err := r.client.PeriodoFinanceiro.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDePeriodos) Criar(ctx context.Context, entrada EntradaDePeriodo) (*ent.PeriodoFinanceiro, error) {
	criar := r.client.PeriodoFinanceiro.Create().
		SetName(entrada.Nome).
		SetStartDate(entrada.InicioEm).
		SetEndDate(entrada.FimEm).
		SetStatus(periodofinanceiro.Status(entrada.Status))
	criar.SetNillableCreatedBy(entrada.CriadoPor)
	return criar.Save(ctx)
}

func (r *repositorioDePeriodos) Atualizar(ctx context.Context, id uuid.UUID, entrada EntradaDePeriodo) (*ent.PeriodoFinanceiro, error) {
	// Nome e datas apenas: o status muda exclusivamente por MudarStatus.
	linha, err := r.client.PeriodoFinanceiro.UpdateOneID(id).
		SetName(entrada.Nome).
		SetStartDate(entrada.InicioEm).
		SetEndDate(entrada.FimEm).
		Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDePeriodos) MudarStatus(ctx context.Context, id uuid.UUID, status string,
	publicadoPor *uuid.UUID, publicadoEm *time.Time, revisadoPor *uuid.UUID) (*ent.PeriodoFinanceiro, error) {

	atualizar := r.client.PeriodoFinanceiro.UpdateOneID(id).
		SetStatus(periodofinanceiro.Status(status))
	atualizar.SetNillablePublishedBy(publicadoPor)
	if publicadoEm != nil {
		atualizar.SetPublishedAt(*publicadoEm)
	} else {
		atualizar.ClearPublishedAt()
	}
	atualizar.SetNillableReviewedBy(revisadoPor)

	linha, err := atualizar.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

type repositorioDeLancamentos struct{ client *ent.Client }

// NovoRepositorioDeLancamentos monta o repositório Ent dos lançamentos.
func NovoRepositorioDeLancamentos(client *ent.Client) RepositorioDeLancamentos {
	return &repositorioDeLancamentos{client: client}
}

func (r *repositorioDeLancamentos) ListarPorPeriodo(ctx context.Context, idDoPeriodo uuid.UUID) ([]*ent.LancamentoFinanceiro, error) {
	return r.client.LancamentoFinanceiro.Query().
		Where(lancamentofinanceiro.FinancialPeriodID(idDoPeriodo)).
		Order(ent.Asc(lancamentofinanceiro.FieldEntryDate)).
		All(ctx)
}

func (r *repositorioDeLancamentos) Criar(ctx context.Context, entrada EntradaDeLancamento) (*ent.LancamentoFinanceiro, error) {
	criar := r.client.LancamentoFinanceiro.Create().
		SetFinancialPeriodID(entrada.IDDoPeriodo).
		SetType(lancamentofinanceiro.Type(entrada.Tipo)).
		SetCategory(entrada.Categoria).
		SetDescription(entrada.Descricao).
		SetAmountCents(entrada.ValorCentavos).
		SetEntryDate(entrada.Data).
		SetSource(lancamentofinanceiro.Source(entrada.Origem))
	criar.SetNillableCreatedBy(entrada.CriadoPor)
	return criar.Save(ctx)
}

func (r *repositorioDeLancamentos) Remover(ctx context.Context, id uuid.UUID) error {
	err := r.client.LancamentoFinanceiro.DeleteOneID(id).Exec(ctx)
	if err != nil && ent.IsNotFound(err) {
		return ErrNaoEncontrado
	}
	return err
}

func (r *repositorioDeLancamentos) SomarPorPeriodo(ctx context.Context, idDoPeriodo uuid.UUID) (int64, int64, error) {
	linhas, err := r.client.LancamentoFinanceiro.Query().
		Where(lancamentofinanceiro.FinancialPeriodID(idDoPeriodo)).
		All(ctx)
	if err != nil {
		return 0, 0, err
	}
	var receitas, despesas int64
	for _, l := range linhas {
		switch l.Type {
		case lancamentofinanceiro.TypeIncome:
			receitas += int64(l.AmountCents)
		case lancamentofinanceiro.TypeExpense:
			despesas += int64(l.AmountCents)
		}
	}
	return receitas, despesas, nil
}

type repositorioDeDocumentos struct{ client *ent.Client }

// NovoRepositorioDeDocumentos monta o repositório Ent dos anexos.
func NovoRepositorioDeDocumentos(client *ent.Client) RepositorioDeDocumentos {
	return &repositorioDeDocumentos{client: client}
}

func (r *repositorioDeDocumentos) ListarPorPeriodo(ctx context.Context, idDoPeriodo uuid.UUID) ([]*ent.DocumentoTransparencia, error) {
	return r.client.DocumentoTransparencia.Query().
		Where(documentotransparencia.FinancialPeriodID(idDoPeriodo)).
		Order(ent.Desc(documentotransparencia.FieldCreatedAt)).
		All(ctx)
}

func (r *repositorioDeDocumentos) Obter(ctx context.Context, id uuid.UUID) (*ent.DocumentoTransparencia, error) {
	linha, err := r.client.DocumentoTransparencia.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeDocumentos) ObterComPeriodo(ctx context.Context, id uuid.UUID) (*ent.DocumentoTransparencia, error) {
	linha, err := r.client.DocumentoTransparencia.Query().
		Where(documentotransparencia.ID(id)).
		WithPeriod().
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}

func (r *repositorioDeDocumentos) Criar(ctx context.Context, entrada EntradaDeDocumento) (*ent.DocumentoTransparencia, error) {
	criar := r.client.DocumentoTransparencia.Create().
		SetFinancialPeriodID(entrada.IDDoPeriodo).
		SetFilePath(entrada.CaminhoDoArquivo).
		SetOriginalFileName(entrada.NomeOriginal).
		SetMimeType(entrada.TipoMIME).
		SetFileSize(int(entrada.Tamanho))
	criar.SetNillableCreatedBy(entrada.CriadoPor)
	linha, err := criar.Save(ctx)
	if err != nil {
		if database.EhViolacaoDeChaveEstrangeira(err) {
			return nil, ErrNaoEncontrado
		}
		return nil, err
	}
	return linha, nil
}
