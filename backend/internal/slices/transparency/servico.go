package transparency

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/ent"
	"github.com/NathanOnCodes/paroquia/backend/ent/documentotransparencia"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/storage"
)

// Tipos de documento aceitos, como no sistema legado.
var mimeAceitos = map[string]string{
	"application/pdf": ".pdf",
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
}

// Servico é a superfície de negócio da transparência.
type Servico interface {
	// Leitura pública: somente períodos publicados.
	ListarPublicos(ctx context.Context) ([]RespostaDePeriodo, error)
	ObterPublico(ctx context.Context, id string) (RespostaDePeriodo, error)

	ListarAdmin(ctx context.Context) ([]RespostaDePeriodo, error)
	ObterAdmin(ctx context.Context, id string) (RespostaDePeriodo, error)
	CriarPeriodo(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDePeriodo) (RespostaDePeriodo, error)
	AtualizarPeriodo(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDePeriodo) (RespostaDePeriodo, error)
	MudarStatus(ctx context.Context, ator authx.UsuarioAutenticado, id, status string) (RespostaDePeriodo, error)

	AdicionarLancamento(ctx context.Context, ator authx.UsuarioAutenticado, idDoPeriodo string, req RequisicaoDeLancamento) (RespostaDeLancamento, error)
	RemoverLancamento(ctx context.Context, ator authx.UsuarioAutenticado, id string) error

	AnexarDocumento(ctx context.Context, ator authx.UsuarioAutenticado, idDoPeriodo string, arquivo *multipart.FileHeader) (RespostaDeDocumento, error)
	// AbrirDocumento devolve o stream e o nome do arquivo, ou recusa quando o
	// período não está publicado.
	AbrirDocumento(ctx context.Context, id string) (io.ReadCloser, string, string, error)
}

// Dependencias agrupa as injeções do serviço.
type Dependencias struct {
	Periodos      RepositorioDePeriodos
	Lancamentos   RepositorioDeLancamentos
	Documentos    RepositorioDeDocumentos
	Armazenamento storage.ArmazenamentoDeArquivos
	Auditoria     audit.RegistradorDeAuditoria
	Log           *slog.Logger
	TamanhoMaximo int64
	Relogio       func() time.Time
}

type servico struct{ Dependencias }

// NovoServico monta o serviço de transparência.
func NovoServico(d Dependencias) Servico {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.Relogio == nil {
		d.Relogio = time.Now
	}
	return &servico{Dependencias: d}
}

func (s *servico) ListarPublicos(ctx context.Context) ([]RespostaDePeriodo, error) {
	periodos, err := s.Periodos.ListarPublicos(ctx)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar períodos", err)
	}
	saida := make([]RespostaDePeriodo, 0, len(periodos))
	for _, p := range periodos {
		receitas, despesas, _ := s.Lancamentos.SomarPorPeriodo(ctx, p.ID)
		saida = append(saida, paraRespostaDePeriodo(p, receitas, despesas))
	}
	return saida, nil
}

func (s *servico) ObterPublico(ctx context.Context, id string) (RespostaDePeriodo, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return RespostaDePeriodo{}, err
	}
	periodo, err := s.Periodos.ObterPublico(ctx, uid)
	if err != nil {
		return RespostaDePeriodo{}, traduzir(err)
	}
	return s.completar(ctx, periodo, true)
}

func (s *servico) ListarAdmin(ctx context.Context) ([]RespostaDePeriodo, error) {
	periodos, err := s.Periodos.ListarAdmin(ctx)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar períodos", err)
	}
	saida := make([]RespostaDePeriodo, 0, len(periodos))
	for _, p := range periodos {
		receitas, despesas, _ := s.Lancamentos.SomarPorPeriodo(ctx, p.ID)
		saida = append(saida, paraRespostaDePeriodo(p, receitas, despesas))
	}
	return saida, nil
}

func (s *servico) ObterAdmin(ctx context.Context, id string) (RespostaDePeriodo, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return RespostaDePeriodo{}, err
	}
	periodo, err := s.Periodos.Obter(ctx, uid)
	if err != nil {
		return RespostaDePeriodo{}, traduzir(err)
	}
	return s.completar(ctx, periodo, true)
}

func (s *servico) CriarPeriodo(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDePeriodo) (RespostaDePeriodo, error) {
	inicio, fim, err := req.interpretarDatas()
	if err != nil {
		return RespostaDePeriodo{}, httpx.ErroDeValidacao("Data inválida",
			httpx.ErroDeCampo{Campo: "start_date", Mensagem: "data inválida"})
	}
	if fim.Before(inicio) {
		return RespostaDePeriodo{}, httpx.ErroDeValidacao(ErrDataFinalAnterior.Error(),
			httpx.ErroDeCampo{Campo: "end_date", Mensagem: "deve ser posterior a start_date"})
	}

	status := req.Status
	if status == "" {
		status = "draft"
	}

	autorID, _ := uuid.Parse(ator.ID)
	periodo, err := s.Periodos.Criar(ctx, EntradaDePeriodo{
		Nome: strings.TrimSpace(req.Nome), InicioEm: inicio, FimEm: fim,
		Status: status, CriadoPor: entx.UUIDOpcional(autorID),
	})
	if err != nil {
		return RespostaDePeriodo{}, traduzir(err)
	}
	s.auditar(ctx, "create_financial_period", periodo.ID.String(), ator.ID)
	return paraRespostaDePeriodo(periodo, 0, 0), nil
}

func (s *servico) AtualizarPeriodo(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDePeriodo) (RespostaDePeriodo, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return RespostaDePeriodo{}, err
	}
	inicio, fim, err := req.interpretarDatas()
	if err != nil {
		return RespostaDePeriodo{}, httpx.ErroDeValidacao("Data inválida",
			httpx.ErroDeCampo{Campo: "start_date", Mensagem: "data inválida"})
	}
	if fim.Before(inicio) {
		return RespostaDePeriodo{}, httpx.ErroDeValidacao(ErrDataFinalAnterior.Error(),
			httpx.ErroDeCampo{Campo: "end_date", Mensagem: "deve ser posterior a start_date"})
	}

	periodo, err := s.Periodos.Atualizar(ctx, uid, EntradaDePeriodo{
		Nome: strings.TrimSpace(req.Nome), InicioEm: inicio, FimEm: fim,
	})
	if err != nil {
		return RespostaDePeriodo{}, traduzir(err)
	}
	s.auditar(ctx, "update_financial_period", periodo.ID.String(), ator.ID)
	return s.completar(ctx, periodo, false)
}

func (s *servico) MudarStatus(ctx context.Context, ator authx.UsuarioAutenticado, id, status string) (RespostaDePeriodo, error) {
	// Publicar ou arquivar é ato exclusivo do admin, como no
	// `canPublishTransparency` legado. O trigger no banco reforça.
	if (status == "published" || status == "archived") && ator.Role != authx.PapelAdmin {
		return RespostaDePeriodo{}, httpx.ErroSemPermissao("Apenas o admin pode publicar ou arquivar períodos")
	}

	uid, err := interpretarID(id)
	if err != nil {
		return RespostaDePeriodo{}, err
	}

	autorID, _ := uuid.Parse(ator.ID)
	var publicadoPor, revisadoPor *uuid.UUID
	var publicadoEm *time.Time
	switch status {
	case "published":
		agora := s.Relogio().UTC()
		publicadoPor, publicadoEm = entx.UUIDOpcional(autorID), &agora
	case "review":
		revisadoPor = entx.UUIDOpcional(autorID)
	}

	periodo, err := s.Periodos.MudarStatus(ctx, uid, status, publicadoPor, publicadoEm, revisadoPor)
	if err != nil {
		return RespostaDePeriodo{}, traduzir(err)
	}
	s.auditar(ctx, "change_financial_period_status", periodo.ID.String(), ator.ID)
	return s.completar(ctx, periodo, false)
}

func (s *servico) AdicionarLancamento(ctx context.Context, ator authx.UsuarioAutenticado, idDoPeriodo string, req RequisicaoDeLancamento) (RespostaDeLancamento, error) {
	uid, err := interpretarID(idDoPeriodo)
	if err != nil {
		return RespostaDeLancamento{}, err
	}
	data, err := req.interpretarData()
	if err != nil {
		return RespostaDeLancamento{}, httpx.ErroDeValidacao("Data inválida",
			httpx.ErroDeCampo{Campo: "entry_date", Mensagem: "data inválida"})
	}

	periodo, err := s.Periodos.Obter(ctx, uid)
	if err != nil {
		return RespostaDeLancamento{}, traduzir(err)
	}
	if periodo.Status == "published" || periodo.Status == "archived" {
		return RespostaDeLancamento{}, httpx.ErroConflito(ErrPeriodoFechado.Error())
	}

	autorID, _ := uuid.Parse(ator.ID)
	linha, err := s.Lancamentos.Criar(ctx, EntradaDeLancamento{
		IDDoPeriodo: uid, Tipo: req.Tipo, Categoria: strings.TrimSpace(req.Categoria),
		Descricao: strings.TrimSpace(req.Descricao), ValorCentavos: req.ValorCentavos,
		Data: data, Origem: "manual", CriadoPor: entx.UUIDOpcional(autorID),
	})
	if err != nil {
		return RespostaDeLancamento{}, traduzir(err)
	}
	s.auditar(ctx, "add_financial_entry", linha.ID.String(), ator.ID)
	return paraRespostaDeLancamento(linha), nil
}

func (s *servico) RemoverLancamento(ctx context.Context, ator authx.UsuarioAutenticado, id string) error {
	uid, err := interpretarID(id)
	if err != nil {
		return err
	}
	if err := s.Lancamentos.Remover(ctx, uid); err != nil {
		return traduzir(err)
	}
	s.auditar(ctx, "delete_financial_entry", uid.String(), ator.ID)
	return nil
}

func (s *servico) AnexarDocumento(ctx context.Context, ator authx.UsuarioAutenticado, idDoPeriodo string, arquivo *multipart.FileHeader) (RespostaDeDocumento, error) {
	if arquivo == nil {
		return RespostaDeDocumento{}, httpx.ErroDeValidacao("Arquivo obrigatório",
			httpx.ErroDeCampo{Campo: "file", Mensagem: "campo obrigatório"})
	}
	uid, err := interpretarID(idDoPeriodo)
	if err != nil {
		return RespostaDeDocumento{}, err
	}

	tipo := strings.ToLower(strings.TrimSpace(arquivo.Header.Get("Content-Type")))
	extensao, ok := mimeAceitos[tipo]
	if !ok {
		return RespostaDeDocumento{}, httpx.ErroDeValidacao("Tipo de arquivo não permitido",
			httpx.ErroDeCampo{Campo: "file", Mensagem: "use PDF, JPEG, PNG ou WebP"})
	}
	if s.TamanhoMaximo > 0 && arquivo.Size > s.TamanhoMaximo {
		return RespostaDeDocumento{}, httpx.ErroDeValidacao("Arquivo acima do tamanho permitido",
			httpx.ErroDeCampo{Campo: "file", Mensagem: "arquivo grande demais"})
	}

	origem, err := arquivo.Open()
	if err != nil {
		return RespostaDeDocumento{}, httpx.ErroInterno("Falha ao ler arquivo", err)
	}
	defer func() { _ = origem.Close() }()

	caminho := fmt.Sprintf("periods/%s/%s%s", uid.String(), uuid.NewString(), extensao)
	if _, err := s.Armazenamento.Salvar(ctx, caminho, origem, arquivo.Size, tipo); err != nil {
		return RespostaDeDocumento{}, httpx.ErroInterno("Falha ao armazenar arquivo", err)
	}

	autorID, _ := uuid.Parse(ator.ID)
	registro, err := s.Documentos.Criar(ctx, EntradaDeDocumento{
		IDDoPeriodo: uid, CaminhoDoArquivo: caminho,
		NomeOriginal: arquivo.Filename, TipoMIME: tipo,
		Tamanho: arquivo.Size, CriadoPor: entx.UUIDOpcional(autorID),
	})
	if err != nil {
		// Rollback: sem registro, o arquivo órfão não fica.
		if rerr := s.Armazenamento.Remover(ctx, caminho); rerr != nil {
			s.Log.WarnContext(ctx, "falha ao remover arquivo órfão",
				slog.String("erro", rerr.Error()))
		}
		return RespostaDeDocumento{}, traduzir(err)
	}

	s.auditar(ctx, "upload_transparency_document", registro.ID.String(), ator.ID)
	return paraRespostaDeDocumento(registro), nil
}

func (s *servico) AbrirDocumento(ctx context.Context, id string) (io.ReadCloser, string, string, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return nil, "", "", err
	}

	registro, err := s.Documentos.ObterComPeriodo(ctx, uid)
	if err != nil {
		return nil, "", "", traduzir(err)
	}
	// Somente períodos publicados liberam o download.
	if registro.Edges.Period == nil || registro.Edges.Period.Status != "published" {
		return nil, "", "", httpx.ErroSemPermissao("Documento indisponível")
	}

	stream, _, err := s.Armazenamento.Open(ctx, registro.FilePath)
	if err != nil {
		return nil, "", "", httpx.ErroInterno("Falha ao abrir documento", err)
	}
	return stream, registro.OriginalFileName, registro.MimeType, nil
}

// completar anexa lançamentos e documentos quando solicitado.
func (s *servico) completar(ctx context.Context, periodo *ent.PeriodoFinanceiro, detalhar bool) (RespostaDePeriodo, error) {
	receitas, despesas, _ := s.Lancamentos.SomarPorPeriodo(ctx, periodo.ID)
	resposta := paraRespostaDePeriodo(periodo, receitas, despesas)

	if !detalhar {
		return resposta, nil
	}

	if lancamentos, err := s.Lancamentos.ListarPorPeriodo(ctx, periodo.ID); err == nil {
		resposta.Lancamentos = make([]RespostaDeLancamento, 0, len(lancamentos))
		for _, l := range lancamentos {
			resposta.Lancamentos = append(resposta.Lancamentos, paraRespostaDeLancamento(l))
		}
	}
	if documentos, err := s.Documentos.ListarPorPeriodo(ctx, periodo.ID); err == nil {
		resposta.Documentos = make([]RespostaDeDocumento, 0, len(documentos))
		for _, d := range documentos {
			resposta.Documentos = append(resposta.Documentos, paraRespostaDeDocumento(d))
		}
	}
	return resposta, nil
}

func (s *servico) auditar(ctx context.Context, acao, idDaEntidade, idDoAtor string) {
	if s.Auditoria == nil {
		return
	}
	s.Auditoria.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: idDoAtor, Action: acao,
		EntityType: "financial_period", EntityID: idDaEntidade,
	})
}

func interpretarID(id string) (uuid.UUID, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, httpx.ErroDeValidacao("Identificador inválido",
			httpx.ErroDeCampo{Campo: "id", Mensagem: "identificador inválido"})
	}
	return uid, nil
}

func traduzir(err error) error {
	switch {
	case errors.Is(err, ErrNaoEncontrado):
		return httpx.ErroNaoEncontrado("Registro não encontrado")
	case errors.Is(err, ErrPeriodoFechado):
		return httpx.ErroConflito(ErrPeriodoFechado.Error())
	case errors.Is(err, storage.ErrNaoEncontrado):
		return httpx.ErroNaoEncontrado("Documento não encontrado")
	default:
		return httpx.ErroInterno("Falha ao processar transparência", err)
	}
}

// _ garante que a referência ao pacote permaneça estável.
var _ = documentotransparencia.FieldCreatedAt
