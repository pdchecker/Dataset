package main

/* Imports
 * Bibliotecas necessarias para manipulacao de bites, arquivos JSON, strings e formatacao
 * Adicionalmente, bibliotecas para smart contracts
 */
 
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

// Estrutura SmartContract
type SmartContract struct{}

// Definição de classes
// Entidade (Proprietario, Concessionaria, Montadora)
type Entidade struct {
	Nome                          string `json:"nome"`
	ECnpjCpf                      int    `json:"eCnpjCpf"`
	Senha                         string `json:"senha"`
	Tipo                          string `json:"tipo"`
	EDocumentoResponsavelRegistro int    `json:"eDocumentoResponsavelRegistro"`
	Id                            string `json:"id"`
}

// Veiculo
type Veiculo struct {
	VChassis                      string `json:"vChassis"`
	Renavam                       int    `json:"renavam"`
	VCnpjCpf                      int    `json:"vCnpjCpf"`
	Especificacao                 string `json:"especificacao"`
	DataFabricacao                string `json:"dataFabricacao"`
	DataUtilizacao                string `json:"dataUtilizacao"`
	DataRevenda                   string `json:"dataRevenda"`
	VDocumentoResponsavelRegistro int    `json:"vDocumentoResponsavelRegistro"`
	Id                            string `json:"id"`
}

// Registros de Manutencao
type RegistroManutencao struct {
	MChassis                      string  `json:"mChassis"`
	PnFalha                       int     `json:"pnFalha"`
	TipoFalha                     string  `json:"tipoFalha"`
	SintomaFalha                  string  `json:"sintomaFalha"`
	InvestigacaoFalha             string  `json:"investigacaoFalha"`
	CodigoFalha                   int     `json:"codigoFalha"`
	ManutencaoRealizada           string  `json:"manutencaoRealizada"`
	PnTrocado                     int     `json:"pnTrocado"`
	NSerieTrocado                 int     `json:"nSerieTrocado"`
	NotaFiscal                    string  `json:"notaFiscal"`
	CustoManutencao               float64 `json:"custoManutencao"`
	DataReparo                    string  `json:"dataReparo"`
	KmAtual                       int     `json:"kmAtual"`
	ConsumoCombustivel            float64 `json:"consumoCombustivel"`
	CargaCarregada                int     `json:"cargaCarregada"`
	MDocumentoResponsavelRegistro int     `json:"mDocumentoResponsavelRegistro"`
	ResponsavelRegistro           string  `json:"responsavelRegistro"`
	Id                            string  `json:"id"`
}

// Arquivo JSON a ser retornado
type JsonReturnedEntidade struct {
	Key    string   `json:"Key"`
	Record Entidade `json:"Record"`
}

type JsonReturnedVeiculo struct {
	Key    string  `json:"Key"`
	Record Veiculo `json:"Record"`
}

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}

/*
 * Funcao Init é chamado pelo fabric na inicializacao da blockchain.
 * Uma melhor pratica documentada é de manter essa funcao em branco,
 * uma vez que roda sempre que o chaincode é iniciado, e manter uma funcao separada (iniciarLedger())
 * com um código que pode ser rodado quando apropriado.
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * A funcao Invoke é o recurso chamado para quaisquer requisicoes para o chaincode.
 * Cada função deve ser definida aqui, para que possam ser chamadas como parametros da funcao invoke.
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	
	function, args := APIstub.GetFunctionAndParameters()

	if function == "buscaVeiculoPorEntidade" { 
		return s.buscaVeiculoPorEntidade(APIstub, args)
	} else if function == "buscaVeiculoPorChassis" { 
		return s.buscaVeiculoPorChassis(APIstub, args)
	} else if function == "buscaManutencoes" { 
		return s.buscaManutencoes(APIstub, args)
	} else if function == "buscaEntidade" { 
		return s.buscaEntidade(APIstub, args)
	} else if function == "iniciarLedger" { 
		return s.iniciarLedger(APIstub)
	} else if function == "novaEntidade" { 
		return s.novaEntidade(APIstub, args)
	} else if function == "novoVeiculo" { 
		return s.novoVeiculo(APIstub, args)
	} else if function == "novaManutencao" { 
		return s.novaManutencao(APIstub, args)
	} else if function == "mudancaProprietario" { 
		return s.mudancaProprietario(APIstub, args)
	} else if function == "mudancaSenha" { 
		return s.mudancaSenha(APIstub, args)
	}

	return shim.Error(`Funcao Invalida.\nFuncoes Disponiveis:
	\nbuscaVeiculoPorEntidade(Documento)
	\nbuscaManutencoes(Chassis)
	\nbuscaEntidade(Documento)
	\niniciarLedger()
	\nnovaEntidade(nome,documento,senha,tipo)
	\nnovoVeiculo(chassis,renavam,documento_proprietario,especificacao,data_fabricacao,data_inicio_utilizacao,data_venda)
	\nnovaManutencao(pn_falha,tipo_falha,sintoma_falha,investigacao_falha,codigo_falha,manutencao_realizada,pn_Trocado,n_Serie_PN_Trocado,nota_fiscal,custo_manutencao,data_reparo,km_atual,consumo_combsutivel,carga_carregada,responsavel_registro)
	\nmudancaProprietario(chassis,documento_comprador,senha_comprador,senha_vendedor,documento_concessionaria,senha_concessionaria)
	\nmudancaSenha(documento,senha_antiga,senha_nova)
	`)
}

//BUSCAS

// Busca veiculo cadastrado sob certa entidade
func (s *SmartContract) buscaVeiculoPorEntidade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Verifica que apenas um argumento foi passado
	if len(args) != 1 {
		return shim.Error("Numero incorreto de argumentos. Inclua apenas um documento (CPF ou CNPJ, sem pontos)")
	}

	// Recebe o argumento como byte[], transforma em string e envia a query pelo documento do proprietario
	queryIntString := "{\"selector\":{\"vCnpjCpf\":" + args[0] + "}}"
	queryString := fmt.Sprintf(queryIntString)
	queryResults, err := getQueryResultForQueryString(APIstub, queryString)

	// Analisa o resultado e retorna se busca sucedida.
	fmt.Println(BytesToString(queryResults))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte(queryResults))
}

// Busca veiculo por chassis
func (s *SmartContract) buscaVeiculoPorChassis(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Verifica que apenas um argumento foi passado
	if len(args) != 1 {
		return shim.Error("Numero incorreto de argumentos. Inclua apenas um chassis.")
	}

	// Recebe o argumento como byte[], transforma em string e envia a query pelo documento do proprietario
	queryIntString := "{\"selector\":{\"vChassis\":\"" + args[0] + "\"}}"
	queryString := fmt.Sprintf(queryIntString)
	queryResults, err := getQueryResultForQueryString(APIstub, queryString)

	// Analisa o resultado e retorna se busca sucedida.
	fmt.Println(BytesToString(queryResults))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// Busca Manutencoes
func (s *SmartContract) buscaManutencoes(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Verifica que apenas um argumento foi passado
	if len(args) != 1 {
		return shim.Error("Numero incorreto de argumentos. Inclua apenas um chassis")
	}

	// Recebe o argumento como byte[], transforma em string e envia a query pelo documento do proprietario
	queryIntString := "{\"selector\":{\"mChassis\":\"" + args[0] + "\"}}"
	queryString := fmt.Sprintf(queryIntString)
	queryResults, err := getQueryResultForQueryString(APIstub, queryString)

	// Analisa o resultado e retorna se busca sucedida.
	fmt.Println(BytesToString(queryResults))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// Busca entidade
func (s *SmartContract) buscaEntidade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// Verifica que apenas um argumento foi passado
	if len(args) != 1 {
		return shim.Error("Numero incorreto de argumentos. Inclua apenas um documento (CPF ou CNPJ, sem pontos)")
	}

	// Recebe o argumento como byte[], transforma em string e envia a query pelo documento do proprietario
	queryIntString := "{\"selector\":{\"eCnpjCpf\":" + args[0] + "}}"
	queryString := fmt.Sprintf(queryIntString)
	queryResults, err := getQueryResultForQueryString(APIstub, queryString)

	if string(queryResults) == "" || string(queryResults) == "[]" {
		fmt.Println("Usuario nao existe")
	}

	// Analisa o resultado e retorna se busca sucedida.
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(queryResults)
}

// CRIAR
func (s *SmartContract) novaEntidade(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 6 {
		return shim.Error("Numero incorreto de argumentos (6 sao esperados):\nNome, Documento (CPF/CNPJ), Senha, Tipo (Montadora, Concessionaria, Proprietario), documento do responsavel pelo registro, senha do responsavel pelo registro.")
	}

	//Transformando documentos em numeros
	NECnpjCpf, _ := strconv.Atoi(args[1])
	NEDoc, _ := strconv.Atoi(args[4])

	//Verifica se o usuario responsavel pelo registro existe, se a senha confere e se é autorizado
	argsB := make([]string, 1)
	argsB[0] = args[4]

	respEntidadeByte := s.buscaEntidade(APIstub, argsB)
	if string(respEntidadeByte.Payload) == "[]" {
		return shim.Error("Usuário responsável não existe.")
	}
	respEntidadeFix := []byte(strings.Trim(string(respEntidadeByte.Payload), "[]"))
	var respEntidade = JsonReturnedEntidade{}
	json.Unmarshal(respEntidadeFix, &respEntidade)
	fmt.Println(respEntidade.Record.Senha)

	if respEntidade.Record.Senha != args[5] {
		return shim.Error("Senha do responsavel invalida.")
	}
	if respEntidade.Record.Tipo == "proprietario" {
		return shim.Error("Usuário não autorizado para realizar essa operação.")
	}

	argsB[0] = args[1]
	novaEntidadeByte := s.buscaEntidade(APIstub, argsB)
	if string(novaEntidadeByte.Payload) != "[]" {
		return shim.Error("Usuário já existe.")
	}

	//Encontra o ultimo ID da blockchain para ser o ID+1 (nkey)
	resultsIterator, err := APIstub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	i := 0
	for resultsIterator.HasNext() {
		_, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		i = i + 1
	}

	nkey := strconv.Itoa(i + 1)

	//Populando a variavel entidade com os dados do registro
	var entidade = Entidade{Nome: args[0], ECnpjCpf: NECnpjCpf, Senha: args[2], Tipo: args[3], EDocumentoResponsavelRegistro: NEDoc, Id: nkey}

	//Transformando a entidade em um registro JSON
	entidadeAsBytes, _ := json.Marshal(entidade)

	//Envia o registro da entidade em JSON para a blockchain sob o ID nkey.
	APIstub.PutState(nkey, entidadeAsBytes)

	return shim.Success(entidadeAsBytes)
}

func (s *SmartContract) novoVeiculo(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 9 {
		return shim.Error("Numero incorreto de argumentos (9 sao esperados):\nChassis, Renavam, Documento (CPF/CNPJ), Especificacao, data de fabricacao, data de inicio de utilizacao e data de revenda, documento do responsavel pelo registro e sua senha")
	}

	nVRenavam, _ := strconv.Atoi(args[1])
	nVCnpjCpf, _ := strconv.Atoi(args[2])
	nVDoc, _ := strconv.Atoi(args[7])
	senhaIntermediador := args[8]

	//Encontra o ultimo ID da blockchain para ser o ID+1 (nkey)
	resultsIterator, err := APIstub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	i := 0
	for resultsIterator.HasNext() {
		_, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		i = i + 1
	}

	//Verifica se veiculo ja nao existe
	exVeiculo := s.checaExistencia(APIstub, "vChassis", args[0])
	if exVeiculo == true {
		return shim.Error("Veiculo já cadastrado")
	}

	//Verifica novo proprietario existe.
	exNovoProprietario := s.checaExistencia(APIstub, "eCnpjCpf", args[2])
	if exNovoProprietario == false {
		return shim.Error("Proprietario inexistente")
	}

	//Verifica se o intermediador existe e se é concessionaria ou montadora
	exIntermediador := s.checaExistencia(APIstub, "eCnpjCpf", args[7])
	tipoIntermediadorCons := s.checaEntidade(APIstub, args[7], "concessionaria", "tipo")
	tipoIntermediadorMont := s.checaEntidade(APIstub, args[7], "montadora", "tipo")
	if (exIntermediador == false) || ((tipoIntermediadorCons == false) && (tipoIntermediadorMont == false)) {
		return shim.Error("Intermediador invalido")
	}

	//Verifica se senha esta correta
	checkSenhaIntermediario := s.checaEntidade(APIstub, args[7], senhaIntermediador, "senha")
	if checkSenhaIntermediario == false {
		return shim.Error("Senha invalida")
	}

	nkey := strconv.Itoa(i + 1)

	var veiculo = Veiculo{VChassis: args[0], Renavam: nVRenavam, VCnpjCpf: nVCnpjCpf, Especificacao: args[3], DataFabricacao: args[4], DataUtilizacao: args[5], DataRevenda: args[6], VDocumentoResponsavelRegistro: nVDoc, Id: nkey}

	veiculoAsBytes, _ := json.Marshal(veiculo)

	APIstub.PutState(nkey, veiculoAsBytes)

	return shim.Success(veiculoAsBytes)
}

func (s *SmartContract) novaManutencao(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 17 {
		return shim.Error("Numero incorreto de argumentos (17 sao esperados):\nChassis, Numero da Peça falhada, tipo de falha, sintoma de falha, investigacao da falha, codigo de fala, manutencao realizada, numero da peca trocada, numero de serie da peca trocada, nota fiscal, custo de manutencao, data de reparo, km atual, consumo de combustivel atual, carga carrega, documento do responsavel pelo registro, senha do responsavel pelo registro.")
	}

	nMPnFalha, _ := strconv.Atoi(args[1])
	nMCodFalha, _ := strconv.Atoi(args[5])
	nMPnTrocado, _ := strconv.Atoi(args[7])
	nMNSerieTrocado, _ := strconv.Atoi(args[8])
	nMCustoManutencao, _ := strconv.ParseFloat(args[10], 64)
	nMKmAtual, _ := strconv.Atoi(args[12])
	nMConsumo, _ := strconv.ParseFloat(args[13], 64)
	nMCarga, _ := strconv.Atoi(args[14])
	nMDoc, _ := strconv.Atoi(args[15])

	//Validar Responsavel
	argsB := make([]string, 1)
	argsB[0] = args[15]

	respEntidadeByte := s.buscaEntidade(APIstub, argsB)
	if string(respEntidadeByte.Payload) == "[]" {
		return shim.Error("Usuário responsável não existe.")
	}
	respEntidadeFix := []byte(strings.Trim(string(respEntidadeByte.Payload), "[]"))
	var respEntidade = JsonReturnedEntidade{}
	json.Unmarshal(respEntidadeFix, &respEntidade)
	if respEntidade.Record.Senha != args[16] {
		return shim.Error("Senha invalida para responsavel")
	}
	nomeResponsavelRegistro := respEntidade.Record.Nome

	//Encontra o ultimo ID da blockchain para ser o ID+1 (nkey)
	resultsIterator, err := APIstub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	i := 0
	for resultsIterator.HasNext() {
		_, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		i = i + 1
	}

	nkey := strconv.Itoa(i + 1)

	var registroManutencao = RegistroManutencao{MChassis: args[0], PnFalha: nMPnFalha, TipoFalha: args[2], SintomaFalha: args[3], InvestigacaoFalha: args[4], CodigoFalha: nMCodFalha, ManutencaoRealizada: args[6], PnTrocado: nMPnTrocado, NSerieTrocado: nMNSerieTrocado, NotaFiscal: args[9], CustoManutencao: nMCustoManutencao, DataReparo: args[11], KmAtual: nMKmAtual, ConsumoCombustivel: nMConsumo, CargaCarregada: nMCarga, MDocumentoResponsavelRegistro: nMDoc, ResponsavelRegistro: nomeResponsavelRegistro, Id: nkey}

	registroManutencaoAsBytes, _ := json.Marshal(registroManutencao)

	APIstub.PutState(nkey, registroManutencaoAsBytes)

	return shim.Success(registroManutencaoAsBytes)
}


func (s *SmartContract) mudancaProprietario(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	chassis := args[0]
	novoProprietario, _ := strconv.Atoi(args[1])
	dataTransacao := args[2]
	senhaNovoProprietario := args[3]
	senhaVelhoProprietario := args[4]
	intermediador, _ := strconv.Atoi(args[5])
	senhaIntermediador := args[6]

	//Verifica se chassis, novo proprietario e intermediador existem.
	exChassis := s.checaExistencia(APIstub, "vChassis", chassis)
	exNovoProprietario := s.checaExistencia(APIstub, "eCnpjCpf", args[1])
	exIntermediador := s.checaExistencia(APIstub, "eCnpjCpf", args[5])
	if exChassis == false {
		return shim.Error("Chassis Inexistente")
	}
	if exNovoProprietario == false {
		return shim.Error("Novo Proprietario Inexistente")
	}
	if exIntermediador == false {
		return shim.Error("Intermediador Inexistente")
	}

	//Verifica se o intermediador é concessionaria ou montadora
	tipoIntermediadorCons := s.checaEntidade(APIstub, args[5], "concessionaria", "tipo")
	tipoIntermediadorMont := s.checaEntidade(APIstub, args[5], "montadora", "tipo")
	if (tipoIntermediadorCons == false) && (tipoIntermediadorMont == false) {
		return shim.Error("Intermediador invalido")
	}

	//Recupera dados do veiculo
	argsB := make([]string, 1)
	argsB[0] = chassis

	oldVeiculoByte := s.buscaVeiculoPorChassis(APIstub, argsB)
	oldVeiculoFix := []byte(strings.Trim(string(oldVeiculoByte.Payload), "[]"))
	var veiculoJson = JsonReturnedVeiculo{}
	json.Unmarshal(oldVeiculoFix, &veiculoJson)

	//Verifica se senhas sao corretas
	checkSenhaNovo := s.checaEntidade(APIstub, args[1], senhaNovoProprietario, "senha")
	checkSenhaVelho := s.checaEntidade(APIstub, strconv.Itoa(veiculoJson.Record.VCnpjCpf), senhaVelhoProprietario, "senha")
	checkSenhaIntermediario := s.checaEntidade(APIstub, args[5], senhaIntermediador, "senha")
	if (checkSenhaNovo == false) || (checkSenhaVelho == false) || (checkSenhaIntermediario == false) {
		return shim.Error("Senha invalida")
	}

	// Grava veiculo sob o ID atual
	nkey := veiculoJson.Record.Id

	var veiculo = Veiculo{VChassis: veiculoJson.Record.VChassis, Renavam: veiculoJson.Record.Renavam, VCnpjCpf: novoProprietario, Especificacao: veiculoJson.Record.Especificacao, DataFabricacao: veiculoJson.Record.DataFabricacao, DataUtilizacao: veiculoJson.Record.DataUtilizacao, DataRevenda: dataTransacao, VDocumentoResponsavelRegistro: intermediador, Id: veiculoJson.Record.Id}
	veiculoAsBytes, _ := json.Marshal(veiculo)
	APIstub.PutState(nkey, veiculoAsBytes)
	return shim.Success(veiculoAsBytes)

}

func (s *SmartContract) mudancaSenha(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3: Documento, Senha antiga, Senha nova")
	}

	documento := args[0]
	senhaAntiga := args[1]
	senhaNova := args[2]

	exProprietario := s.checaExistencia(APIstub, "eCnpjCpf", documento)
	if exProprietario == false {
		return shim.Error("Proprietario Inexistente")
	}

	checkSenha := s.checaEntidade(APIstub, documento, senhaAntiga, "senha")
	if checkSenha == false {
		return shim.Error("Senha Incorreta")
	}

	//Buscar registro
	argsB := make([]string, 1)
	argsB[0] = documento

	entidadeByte := s.buscaEntidade(APIstub, argsB)
	entidadeFix := []byte(strings.Trim(string(entidadeByte.Payload), "[]"))
	var entidadeJson = JsonReturnedEntidade{}
	json.Unmarshal(entidadeFix, &entidadeJson)
	entidadeJson.Record.Senha = senhaNova
	newEntidadeAsBytes, _ := json.Marshal(entidadeJson.Record)
	nkey := entidadeJson.Record.Id
	APIstub.PutState(nkey, newEntidadeAsBytes)

	return shim.Success(newEntidadeAsBytes)
}

func (s *SmartContract) iniciarLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	entidades := []Entidade{
		Entidade{Nome: "Montadora de Caminhoes LTDA", ECnpjCpf: 9999999999, Senha: "69bdd58e17ab101986d8cf7a7f9279db", Tipo: "montadora", EDocumentoResponsavelRegistro: 9999999999, Id: "0"}, //senha: montadora, hashed MD5
		Entidade{Nome: "Concessionaria A", ECnpjCpf: 8888888888, Senha: "9aa4898a07dc811fddd4b9d8655fce8f", Tipo: "concessionaria", EDocumentoResponsavelRegistro: 9999999999, Id: "1"},       //senha: concessionariaa, hashed MD5
		Entidade{Nome: "Caio", ECnpjCpf: 7777777777, Senha: "af1dbc5648a563e9a5bd97d0eb68f41b", Tipo: "proprietario", EDocumentoResponsavelRegistro: 8888888888, Id: "2"},                     //senha: papas, hashed MD5
		Entidade{Nome: "Patricia", ECnpjCpf: 6666666666, Senha: "86241b5767f022a036a93a9b55af2e71", Tipo: "proprietario", EDocumentoResponsavelRegistro: 8888888888, Id: "3"},                 //senha: branco, hashed MD5
		Entidade{Nome: "Charlie", ECnpjCpf: 5555555555, Senha: "21591c1cb4eacdf98eab4454f9dbbd09", Tipo: "proprietario", EDocumentoResponsavelRegistro: 8888888888, Id: "4"},                  //senha: sarayu, hashed MD5
	}

	veiculos := []Veiculo{
		Veiculo{VChassis: "E111111", Renavam: 38461734, VCnpjCpf: 9999999999, Especificacao: "Pesado 13L 6x4 500cv", DataFabricacao: "01/11/2019", DataUtilizacao: "", DataRevenda: "", VDocumentoResponsavelRegistro: 9999999999, Id: "5"},
		Veiculo{VChassis: "E222222", Renavam: 49682922, VCnpjCpf: 9999999999, Especificacao: "Medio 7L 4x2 250cv", DataFabricacao: "05/11/2019", DataUtilizacao: "", DataRevenda: "", VDocumentoResponsavelRegistro: 9999999999, Id: "6"},
		Veiculo{VChassis: "E333333", Renavam: 59692833, VCnpjCpf: 8888888888, Especificacao: "Medio 7L 6x4 300cv", DataFabricacao: "10/09/2019", DataUtilizacao: "", DataRevenda: "", VDocumentoResponsavelRegistro: 9999999999, Id: "7"},
		Veiculo{VChassis: "E444444", Renavam: 66049299, VCnpjCpf: 8888888888, Especificacao: "Pesado 13L 6x2 400cv", DataFabricacao: "15/09/2019", DataUtilizacao: "", DataRevenda: "", VDocumentoResponsavelRegistro: 9999999999, Id: "8"},
		Veiculo{VChassis: "E555555", Renavam: 94827493, VCnpjCpf: 7777777777, Especificacao: "Pesado 13L 6x4 450cv", DataFabricacao: "13/01/2019", DataUtilizacao: "22/06/2019", DataRevenda: "22/06/2019", VDocumentoResponsavelRegistro: 9999999999, Id: "9"},
		Veiculo{VChassis: "E666666", Renavam: 20595753, VCnpjCpf: 7777777777, Especificacao: "Pesado 13L 6x4 500cv", DataFabricacao: "05/02/2019", DataUtilizacao: "11/05/2019", DataRevenda: "11/05/2019", VDocumentoResponsavelRegistro: 9999999999, Id: "10"},
		Veiculo{VChassis: "E777777", Renavam: 28102409, VCnpjCpf: 7777777777, Especificacao: "Pesado 13L 6x2 400cv", DataFabricacao: "23/02/2019", DataUtilizacao: "14/06/2019", DataRevenda: "14/06/2019", VDocumentoResponsavelRegistro: 9999999999, Id: "11"},
		Veiculo{VChassis: "E888888", Renavam: 21052708, VCnpjCpf: 6666666666, Especificacao: "Medio 7L 8x2 350cv", DataFabricacao: "01/03/2019", DataUtilizacao: "10/04/2019", DataRevenda: "10/04/2019", VDocumentoResponsavelRegistro: 9999999999, Id: "12"},
		Veiculo{VChassis: "E999999", Renavam: 21052711, VCnpjCpf: 6666666666, Especificacao: "Medio 7L 8x2 350cv", DataFabricacao: "01/03/2019", DataUtilizacao: "10/04/2019", DataRevenda: "10/04/2019", VDocumentoResponsavelRegistro: 9999999999, Id: "13"},
	}

	registroManutencao := []RegistroManutencao{
		//RegistroManutencao{PnFalha:"",TipoFalha:"",SintomaFalha:"",InvestigacaoFalha:"",CodigoFalha:"",ManutencaoRealizada:"",PnTrocado:"",NSerieTrocado:"",NotaFiscal:"",CustoManutencao:"",DataReparo:"",KmAtual:"",ConsumoCombustivel:"",CargaCarregada:"",ResponsavelRegistro :""},
		RegistroManutencao{MChassis: "E555555", PnFalha: 7769, TipoFalha: "Quebra", SintomaFalha: "Perda de potencia.", InvestigacaoFalha: "Codigos de falha escaneados", CodigoFalha: 238, ManutencaoRealizada: "Troca de peça quebrada.", PnTrocado: 5306, NSerieTrocado: 936861, NotaFiscal: "410520888888888888885500110000000011193203701", CustoManutencao: 2696, DataReparo: "01/03/2020", KmAtual: 35155, ConsumoCombustivel: 2.34, CargaCarregada: 54, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "14"},
		RegistroManutencao{MChassis: "E555555", PnFalha: 5727, TipoFalha: "Quebra", SintomaFalha: "Fumaça branca.", InvestigacaoFalha: "Vazamento de oleo identificado.", CodigoFalha: 222, ManutencaoRealizada: "Troca de peça quebrada.", PnTrocado: 8097, NSerieTrocado: 348430, NotaFiscal: "410520888888888888885500110000000021640017072", CustoManutencao: 2947, DataReparo: "11/03/2020", KmAtual: 35988, ConsumoCombustivel: 2.22, CargaCarregada: 54, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "15"},
		RegistroManutencao{MChassis: "E555555", PnFalha: 7148, TipoFalha: "Software", SintomaFalha: "Freios não funcionando.", InvestigacaoFalha: "Verificacao de codigos de falha.", CodigoFalha: 607, ManutencaoRealizada: "Download de novo software na unidade de controle.", PnTrocado: 0, NSerieTrocado: 0, NotaFiscal: "411020888888888888885500110000000031180774293", CustoManutencao: 121, DataReparo: "01/08/2020", KmAtual: 81230, ConsumoCombustivel: 2.61, CargaCarregada: 54, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "16"},
		RegistroManutencao{MChassis: "E666666", PnFalha: 1290, TipoFalha: "Quebra", SintomaFalha: "Perda de potência de frenagem.", InvestigacaoFalha: "Vazamento no ar comprimido.", CodigoFalha: 456, ManutencaoRealizada: "Troca de peça quebrada.", PnTrocado: 3485, NSerieTrocado: 577830, NotaFiscal: "410720888888888888885500110000000041754768734", CustoManutencao: 1949, DataReparo: "12/04/2020", KmAtual: 44033, ConsumoCombustivel: 2.8, CargaCarregada: 74, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "17"},
		RegistroManutencao{MChassis: "E666666", PnFalha: 1952, TipoFalha: "Quebra", SintomaFalha: "Ruido na suspensão", InvestigacaoFalha: "Desgaste na suspensao.", CodigoFalha: 515, ManutencaoRealizada: "Troca de peça quebrada.", PnTrocado: 4060, NSerieTrocado: 133920, NotaFiscal: "411020888888888888885500110000000051219054995", CustoManutencao: 4000, DataReparo: "01/08/2020", KmAtual: 56041, ConsumoCombustivel: 2.49, CargaCarregada: 74, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "18"},
		RegistroManutencao{MChassis: "E777777", PnFalha: 5885, TipoFalha: "Software", SintomaFalha: "Marcha não engatando.", InvestigacaoFalha: "Problema identificado no software da caixa.", CodigoFalha: 863, ManutencaoRealizada: "Download de novo software na unidade de controle.", PnTrocado: 0, NSerieTrocado: 0, NotaFiscal: "410320888888888888885500110000000061697193086", CustoManutencao: 145, DataReparo: "14/09/2020", KmAtual: 23231, ConsumoCombustivel: 2.87, CargaCarregada: 74, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "19"},
		RegistroManutencao{MChassis: "E777777", PnFalha: 1245, TipoFalha: "Quebra", SintomaFalha: "Ruido no motor.", InvestigacaoFalha: "Sujeira e desgaste nas correias", CodigoFalha: 795, ManutencaoRealizada: "Troca de peça quebrada.", PnTrocado: 5328, NSerieTrocado: 607783, NotaFiscal: "410720888888888888885500110000000071489081347", CustoManutencao: 2788, DataReparo: "22/11/2020", KmAtual: 44430, ConsumoCombustivel: 2.44, CargaCarregada: 74, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "20"},
		RegistroManutencao{MChassis: "E777777", PnFalha: 2079, TipoFalha: "Quebra", SintomaFalha: "Nivel de óleo baixo.", InvestigacaoFalha: "Consumo de oleo pela junta.", CodigoFalha: 985, ManutencaoRealizada: "Troca de peça quebrada.", PnTrocado: 6462, NSerieTrocado: 842840, NotaFiscal: "410820888888888888885500110000000081402701728", CustoManutencao: 3515, DataReparo: "14/12/2020", KmAtual: 48955, ConsumoCombustivel: 2.57, CargaCarregada: 74, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "21"},
		RegistroManutencao{MChassis: "E777777", PnFalha: 4512, TipoFalha: "Software", SintomaFalha: "Veículo não liga.", InvestigacaoFalha: "Falha no software do motor de arranque.", CodigoFalha: 648, ManutencaoRealizada: "Download de novo software na unidade de controle.", PnTrocado: 0, NSerieTrocado: 0, NotaFiscal: "411220888888888888885500110000000091746789739", CustoManutencao: 98, DataReparo: "01/02/2021", KmAtual: 75099, ConsumoCombustivel: 2.11, CargaCarregada: 74, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "22"},
		RegistroManutencao{MChassis: "E888888", PnFalha: 5841, TipoFalha: "Quebra", SintomaFalha: "Nível de água baixo.", InvestigacaoFalha: "Falha na valvula do tanque de expansao.", CodigoFalha: 627, ManutencaoRealizada: "Troca de peça quebrada.", PnTrocado: 1522, NSerieTrocado: 101289, NotaFiscal: "410220888888888888885500110000000101489270291", CustoManutencao: 2968, DataReparo: "01/04/2020", KmAtual: 26987, ConsumoCombustivel: 2.63, CargaCarregada: 22, MDocumentoResponsavelRegistro: 8888888888, ResponsavelRegistro: "Concessionaria A", Id: "23"},
	}

	i := 0
	for i < len(entidades) {
		fmt.Println("i is ", i)
		entidadesBytes, _ := json.Marshal(entidades[i])
		APIstub.PutState(strconv.Itoa(i), entidadesBytes)
		fmt.Println("Added", entidades[i])
		i = i + 1
	}
	j := 0
	for j < len(veiculos) {
		fmt.Println("j is ", j)
		veiculosBytes, _ := json.Marshal(veiculos[j])
		APIstub.PutState(strconv.Itoa(i+j), veiculosBytes)
		fmt.Println("Added", veiculos[j])
		j = j + 1
	}
	k := 0
	for k < len(registroManutencao) {
		fmt.Println("k is ", k)
		registroManutencaoBytes, _ := json.Marshal(registroManutencao[k])
		APIstub.PutState(strconv.Itoa(i+j+k), registroManutencaoBytes)
		fmt.Println("Added", registroManutencao[k])
		k = k + 1
	}

	return shim.Success(nil)
}

//SUPORTE

func (s *SmartContract) checaExistencia(APIstub shim.ChaincodeStubInterface, jsonvar string, parametro string) bool {
	var ResultBusca sc.Response
	argsB := make([]string, 1)
	argsB[0] = parametro

	if jsonvar == "eCnpjCpf" {
		ResultBusca = s.buscaEntidade(APIstub, argsB)
		fmt.Println("ecnpj")
	} else {
		fmt.Println("chassis")
		ResultBusca = s.buscaVeiculoPorChassis(APIstub, argsB)
	}

	fmt.Println(string(ResultBusca.Payload))
	if (string(ResultBusca.Payload) == "") || (string(ResultBusca.Payload) == "[]") {
		return false
	}

	return true
}

func (s *SmartContract) checaEntidade(APIstub shim.ChaincodeStubInterface, parametro string, variavel string, tipoVariavel string) bool {
	var ResultBusca sc.Response
	argsB := make([]string, 1)
	argsB[0] = parametro

	ResultBusca = s.buscaEntidade(APIstub, argsB)

	entidadeFix := []byte(strings.Trim(string(ResultBusca.Payload), "[]"))
	var entidadeJson = JsonReturnedEntidade{}
	json.Unmarshal(entidadeFix, &entidadeJson)

	if tipoVariavel == "tipo" {
		if variavel != entidadeJson.Record.Tipo {
			return false
		}
	}
	if tipoVariavel == "senha" {
		if variavel != entidadeJson.Record.Senha {
			return false
		}
	}
	return true
}

func getLatestKey(APIstub shim.ChaincodeStubInterface) int {

	startKey := ""
	endKey := ""

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return -1
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	i := 0
	for resultsIterator.HasNext() {
		_, err := resultsIterator.Next()
		if err != nil {
			return -1
		}
		i = i + 1
	}

	return i

}

// =========================================================================================
// getQueryResultForQueryString executes the passed in query string.
// Result set is built and returned as a byte array containing the JSON results.
// =========================================================================================
func getQueryResultForQueryString(stub shim.ChaincodeStubInterface, queryString string) ([]byte, error) {

	fmt.Printf("- getQueryResultForQueryString queryString:\n%s\n", queryString)

	resultsIterator, err := stub.GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	buffer, err := constructQueryResponseFromIterator(resultsIterator)
	if err != nil {
		return nil, err
	}

	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return buffer.Bytes(), nil
}

//Byte to string
func BytesToString(data []byte) string {
	return string(data[:])
}

// ===========================================================================================
// constructQueryResponseFromIterator constructs a JSON array containing query results from
// a given result iterator
// ===========================================================================================
func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) (*bytes.Buffer, error) {
	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}
