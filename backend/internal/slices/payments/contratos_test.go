package payments

import "testing"

func TestParaDoadorAplicaRegraDeComunidade(t *testing.T) {
	// Regra de negócio: comunidade cadastrada OU nome livre, nunca os dois.
	// Havendo `community_id`, o nome livre é descartado.

	req := RequisicaoDeDoacao{
		NomeCompleto:   "  Maria de Nazaré  ",
		Email:          "maria@paroquia.org",
		Cidade:         " Francisco Morato ",
		Bairro:         " Centro ",
		IDComunidade:   "22222222-2222-2222-2222-222222222222",
		NomeComunidade: "Comunidade livre qualquer",
	}

	entrada := req.paraDoador()

	if entrada.CommunityID == nil {
		t.Fatal("community_id deveria ser convertido para UUID")
	}
	if entrada.CommunityName != "" {
		t.Fatalf("com comunidade cadastrada o nome livre deve ser descartado, veio %q", entrada.CommunityName)
	}
	if entrada.FullName != "Maria de Nazaré" {
		t.Fatalf("nome deveria ser normalizado, veio %q", entrada.FullName)
	}
	if entrada.City != "Francisco Morato" {
		t.Fatalf("cidade deveria ser normalizada, veio %q", entrada.City)
	}
}

func TestParaDoadorMantemNomeLivreSemComunidade(t *testing.T) {
	req := RequisicaoDeDoacao{
		NomeCompleto:   "José",
		Email:          "jose@paroquia.org",
		Cidade:         "Francisco Morato",
		NomeComunidade: "Capela São Bento",
	}

	entrada := req.paraDoador()

	if entrada.CommunityID != nil {
		t.Fatal("sem community_id não deveria haver vínculo com comunidade")
	}
	if entrada.CommunityName != "Capela São Bento" {
		t.Fatalf("nome livre deveria ser preservado, veio %q", entrada.CommunityName)
	}
}

func TestValidarComunidadeExigeUmaDasOpcoes(t *testing.T) {
	// Regra `requireCommunity` do sistema legado.
	vazio := RequisicaoDeDoacao{NomeComunidade: "   "}
	if vazio.ValidarComunidade() {
		t.Fatal("sem community_id nem nome, a doação deveria ser recusada")
	}

	comID := RequisicaoDeDoacao{IDComunidade: "22222222-2222-2222-2222-222222222222"}
	if !comID.ValidarComunidade() {
		t.Fatal("community_id deveria bastar")
	}

	comNome := RequisicaoDeDoacao{NomeComunidade: "Capela São Bento"}
	if !comNome.ValidarComunidade() {
		t.Fatal("nome livre deveria bastar")
	}
}
