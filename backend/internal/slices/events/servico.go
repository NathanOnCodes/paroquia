package events

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"mime/multipart"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/NathanOnCodes/paroquia/backend/internal/platform/audit"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/authx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/entx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/httpx"
	"github.com/NathanOnCodes/paroquia/backend/internal/platform/storage"
)

// Tipos de imagem aceitos, como no sistema legado.
var mimeAceitos = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

// Servico é a superfície de negócio dos eventos.
type Servico interface {
	ListarPublicas(ctx context.Context) ([]Resposta, error)
	ObterPublico(ctx context.Context, slug string) (Resposta, error)

	ListarAdmin(ctx context.Context) ([]Resposta, error)
	Obter(ctx context.Context, id string) (Resposta, error)
	Criar(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDeEscrita, imagem *multipart.FileHeader) (Resposta, error)
	Atualizar(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita, imagem *multipart.FileHeader) (Resposta, error)
	DefinirStatus(ctx context.Context, ator authx.UsuarioAutenticado, id string, status string) (Resposta, error)
}

// Dependencias agrupa as injeções do serviço.
type Dependencias struct {
	Repositorio   Repositorio
	Armazenamento storage.ArmazenamentoDeArquivos
	Auditoria     audit.RegistradorDeAuditoria
	Log           *slog.Logger
	TamanhoMaximo int64
	// URLDaImagem resolve a URL pública de um caminho armazenado.
	URLDaImagem func(caminho string) string
	Relogio     func() time.Time
}

type servico struct{ Dependencias }

// NovoServico monta o serviço de eventos.
func NovoServico(d Dependencias) Servico {
	if d.Log == nil {
		d.Log = slog.Default()
	}
	if d.Relogio == nil {
		d.Relogio = time.Now
	}
	if d.URLDaImagem == nil {
		d.URLDaImagem = func(string) string { return "" }
	}
	return &servico{Dependencias: d}
}

func (s *servico) ListarPublicas(ctx context.Context) ([]Resposta, error) {
	// A janela de 24h para trás mantém eventos recém-encerrados visíveis,
	// como no comportamento legado.
	aPartirDe := s.Relogio().Add(-24 * time.Hour).UTC()
	linhas, err := s.Repositorio.ListarPublicas(ctx, aPartirDe, 50)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar eventos", err)
	}
	return paraRespostas(linhas, s.URLDaImagem), nil
}

func (s *servico) ObterPublico(ctx context.Context, slug string) (Resposta, error) {
	linha, err := s.Repositorio.ObterPublicoPorSlug(ctx, slug)
	if err != nil {
		return Resposta{}, traduzir(err)
	}
	return paraResposta(linha, s.URLDaImagem), nil
}

func (s *servico) ListarAdmin(ctx context.Context) ([]Resposta, error) {
	linhas, err := s.Repositorio.ListarAdmin(ctx, 200)
	if err != nil {
		return nil, httpx.ErroInterno("Falha ao listar eventos", err)
	}
	return paraRespostas(linhas, s.URLDaImagem), nil
}

func (s *servico) Obter(ctx context.Context, id string) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}
	linha, err := s.Repositorio.Obter(ctx, uid)
	if err != nil {
		return Resposta{}, traduzir(err)
	}
	return paraResposta(linha, s.URLDaImagem), nil
}

func (s *servico) Criar(ctx context.Context, ator authx.UsuarioAutenticado, req RequisicaoDeEscrita, imagem *multipart.FileHeader) (Resposta, error) {
	inicio, fim, err := req.interpretarDatas()
	if err != nil {
		return Resposta{}, httpx.ErroDeValidacao("Data inválida",
			httpx.ErroDeCampo{Campo: "starts_at", Mensagem: "data inválida"})
	}
	if fim != nil && fim.Before(inicio) {
		return Resposta{}, httpx.ErroDeValidacao("A data final deve ser posterior à inicial",
			httpx.ErroDeCampo{Campo: "ends_at", Mensagem: "deve ser posterior a starts_at"})
	}

	slug, err := s.slugDisponivel(ctx, req, uuid.Nil)
	if err != nil {
		return Resposta{}, err
	}

	// A imagem sobe primeiro: se o registro falhar, apagamos o arquivo
	// (mesmo rollback do sistema legado).
	var caminhoDaImagem *string
	if imagem != nil {
		caminho, cerr := s.enviarImagem(ctx, imagem)
		if cerr != nil {
			return Resposta{}, cerr
		}
		caminhoDaImagem = &caminho
	}

	var publicadoEm *time.Time
	if req.Status == "published" {
		agora := s.Relogio().UTC()
		publicadoEm = &agora
	}

	autorID, _ := uuid.Parse(ator.ID)
	linha, err := s.Repositorio.Criar(ctx, EntradaDeEscrita{
		Titulo:          strings.TrimSpace(req.Titulo),
		Slug:            slug,
		Descricao:       req.Descricao,
		CaminhoDaImagem: caminhoDaImagem,
		IniciaEm:        inicio,
		TerminaEm:       fim,
		Local:           req.Local,
		Status:          req.Status,
		IDAutor:         entx.UUIDOpcional(autorID),
		PublicadoEm:     publicadoEm,
	})
	if err != nil {
		if caminhoDaImagem != nil {
			s.removerImagem(ctx, *caminhoDaImagem)
		}
		return Resposta{}, traduzir(err)
	}

	s.auditar(ctx, "create_event", linha.ID.String(), ator.ID)
	return paraResposta(linha, s.URLDaImagem), nil
}

func (s *servico) Atualizar(ctx context.Context, ator authx.UsuarioAutenticado, id string, req RequisicaoDeEscrita, imagem *multipart.FileHeader) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}
	inicio, fim, err := req.interpretarDatas()
	if err != nil {
		return Resposta{}, httpx.ErroDeValidacao("Data inválida",
			httpx.ErroDeCampo{Campo: "starts_at", Mensagem: "data inválida"})
	}

	atual, err := s.Repositorio.Obter(ctx, uid)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	slug, err := s.slugDisponivel(ctx, req, uid)
	if err != nil {
		return Resposta{}, err
	}

	// Nova imagem substitui a antiga, que só é apagada depois do sucesso.
	var caminhoDaImagem *string
	imagemAntiga := atual.ImagePath
	if imagem != nil {
		caminho, cerr := s.enviarImagem(ctx, imagem)
		if cerr != nil {
			return Resposta{}, cerr
		}
		caminhoDaImagem = &caminho
	} else {
		caminhoDaImagem = atual.ImagePath
	}

	var publicadoEm *time.Time
	switch {
	case req.Status == "published" && atual.Status != "published":
		agora := s.Relogio().UTC()
		publicadoEm = &agora
	case req.Status != "published":
		publicadoEm = nil
	default:
		publicadoEm = atual.PublishedAt
	}

	autorID, _ := uuid.Parse(ator.ID)
	linha, err := s.Repositorio.Atualizar(ctx, uid, EntradaDeEscrita{
		Titulo:          strings.TrimSpace(req.Titulo),
		Slug:            slug,
		Descricao:       req.Descricao,
		CaminhoDaImagem: caminhoDaImagem,
		IniciaEm:        inicio,
		TerminaEm:       fim,
		Local:           req.Local,
		Status:          req.Status,
		IDAutor:         entx.UUIDOpcional(autorID),
		PublicadoEm:     publicadoEm,
	})
	if err != nil {
		if imagem != nil && caminhoDaImagem != nil {
			s.removerImagem(ctx, *caminhoDaImagem)
		}
		return Resposta{}, traduzir(err)
	}

	if imagem != nil && imagemAntiga != nil && *imagemAntiga != *caminhoDaImagem {
		s.removerImagem(ctx, *imagemAntiga)
	}

	s.auditar(ctx, "update_event", linha.ID.String(), ator.ID)
	return paraResposta(linha, s.URLDaImagem), nil
}

func (s *servico) DefinirStatus(ctx context.Context, ator authx.UsuarioAutenticado, id string, status string) (Resposta, error) {
	uid, err := interpretarID(id)
	if err != nil {
		return Resposta{}, err
	}

	var publicadoEm *time.Time
	if status == "published" {
		agora := s.Relogio().UTC()
		publicadoEm = &agora
	}

	linha, err := s.Repositorio.DefinirStatus(ctx, uid, status, publicadoEm)
	if err != nil {
		return Resposta{}, traduzir(err)
	}

	s.auditar(ctx, "update_event_status", linha.ID.String(), ator.ID)
	return paraResposta(linha, s.URLDaImagem), nil
}

// slugDisponivel usa o slug informado ou deriva do título e resolve colisões
// com o sufixo -2, -3…, como no `uniqueEventSlug` legado.
func (s *servico) slugDisponivel(ctx context.Context, req RequisicaoDeEscrita, ignorar uuid.UUID) (string, error) {
	base := strings.TrimSpace(req.Slug)
	if base == "" {
		base = Slugificar(req.Titulo)
	}
	if base == "" {
		return "", httpx.ErroDeValidacao("Não foi possível derivar o endereço do evento",
			httpx.ErroDeCampo{Campo: "slug", Mensagem: "informe um endereço válido"})
	}

	candidato := base
	for sufixo := 2; ; sufixo++ {
		existe, err := s.Repositorio.SlugExiste(ctx, candidato, ignorar)
		if err != nil {
			return "", httpx.ErroInterno("Falha ao verificar endereço do evento", err)
		}
		if !existe {
			return candidato, nil
		}
		candidato = fmt.Sprintf("%s-%d", base, sufixo)
		if sufixo > 500 {
			return "", httpx.ErroConflito("Não foi possível gerar um endereço único")
		}
	}
}

func (s *servico) enviarImagem(ctx context.Context, arquivo *multipart.FileHeader) (string, error) {
	tipo := strings.ToLower(strings.TrimSpace(arquivo.Header.Get("Content-Type")))
	extensao, ok := mimeAceitos[tipo]
	if !ok {
		return "", httpx.ErroDeValidacao("Tipo de imagem não permitido",
			httpx.ErroDeCampo{Campo: "image", Mensagem: "use JPEG, PNG ou WebP"})
	}
	if s.TamanhoMaximo > 0 && arquivo.Size > s.TamanhoMaximo {
		return "", httpx.ErroDeValidacao("Imagem acima do tamanho permitido",
			httpx.ErroDeCampo{Campo: "image", Mensagem: "arquivo grande demais"})
	}

	origem, err := arquivo.Open()
	if err != nil {
		return "", httpx.ErroInterno("Falha ao ler imagem", err)
	}
	defer func() { _ = origem.Close() }()

	caminho := fmt.Sprintf("events/%s%s", uuid.NewString(), extensao)
	if _, err := s.Armazenamento.Salvar(ctx, caminho, origem, arquivo.Size, tipo); err != nil {
		if errors.Is(err, storage.ErrArquivoGrandeDemais) {
			return "", httpx.ErroDeValidacao("Imagem acima do tamanho permitido",
				httpx.ErroDeCampo{Campo: "image", Mensagem: "arquivo grande demais"})
		}
		return "", httpx.ErroInterno("Falha ao armazenar imagem", err)
	}
	return caminho, nil
}

// removerImagem nunca derruba a operação principal.
func (s *servico) removerImagem(ctx context.Context, caminho string) {
	if err := s.Armazenamento.Remover(ctx, caminho); err != nil {
		s.Log.WarnContext(ctx, "falha ao remover imagem",
			slog.String("caminho", caminho), slog.String("erro", err.Error()))
	}
}

func (s *servico) auditar(ctx context.Context, acao, idDaEntidade, idDoAtor string) {
	if s.Auditoria == nil {
		return
	}
	s.Auditoria.Registrar(ctx, audit.RegistroDeAuditoria{
		ActorID: idDoAtor, Action: acao,
		EntityType: "event", EntityID: idDaEntidade,
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
		return httpx.ErroNaoEncontrado("Evento não encontrado")
	case errors.Is(err, ErrSlugEmUso):
		return httpx.ErroConflito("Já existe um evento com este endereço")
	default:
		return httpx.ErroInterno("Falha ao processar evento", err)
	}
}

// ExtensaoDe resume a extensão esperada de um tipo MIME, para os logs.
func ExtensaoDe(mimeType string) string {
	if ext, ok := mimeAceitos[strings.ToLower(mimeType)]; ok {
		return ext
	}
	return strings.TrimPrefix(filepath.Ext(mimeType), ".")
}
