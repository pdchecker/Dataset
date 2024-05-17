package main

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ============================================================================================================================
// Logger
// ============================================================================================================================

var logger = flogging.MustGetLogger("NPDContract")


// ============================================================================================================================
// Contract Definitions
// ============================================================================================================================

type NPDContract struct {
	contractapi.Contract
}


// ============================================================================================================================
// Asset Definitions - The ledger will store Nilai Peserta Didik (NPD) data
// ============================================================================================================================

type NilaiPesertaDidik struct {
	ID      			string 			`json:"id"`
	IdKLS				string 			`json:"idKls"`
	IdPTK				string 			`json:"idPtk"`
	IdPD				string 			`json:"idPd"`
	NilaiAngka			float64			`json:"nilaiAngka"`
	NilaiHuruf			string 			`json:"nilaiHuruf"`
	NilaiIndex			float64 		`json:"nilaiIndex"`
}


// ============================================================================================================================
// Struct Definitions - PesertaDidik (PD)
// ============================================================================================================================

type PesertaDidik struct {
	ID      			string 	`json:"id"`
	TotalMutu			float64	`json:"totalMutu"`
	TotalSKS			int 	`json:"totalSks"`
	IPK					float64	`json:"ipk"`
}


// ============================================================================================================================
// Struct Definitions - KelasKuliah (KLS)
// ============================================================================================================================

type KelasKuliah struct {
	ID      			string 		`json:"id"`
	SKS					int 		`json:"sks"`
}


// ============================================================================================================================
// Error Messages
// ============================================================================================================================

const (
	ER11 string = "ER11-Incorrect number of arguments. Required %d arguments, but you have %d arguments."
	ER12        = "ER12-NilaiPesertaDidik with id '%s' already exists."
	ER13        = "ER13-NilaiPesertaDidik with id '%s' doesn't exist."
	ER31        = "ER31-Failed to change to world state: %v."
	ER32        = "ER32-Failed to read from world state: %v."
	ER33        = "ER33-Failed to get result from iterator: %v."
	ER34        = "ER34-Failed unmarshaling JSON: %v."
	ER35        = "ER35-Failed parsing string to integer: %v."
	ER36        = "ER36-Failed parsing string to float: %v."
	ER37        = "ER37-Failed to query another chaincode (%s): %v."
	ER41        = "ER41-Access is not permitted with MSDPID '%s'."
	ER42        = "ER42-Unknown MSPID: '%s'."
)


// ============================================================================================================================
// Channel Name & Contract Name In The Channel
// ============================================================================================================================

const (
	AcademicChannel	string = "academicchannel"
	PDContract 		string = "pdcontract"
	KLSContract 	string = "klscontract"
)


// ============================================================================================================================
// CreateNpd - Issues a new Nilai Peserta Didik (NPD) to the world state with given details.
// Arguments - ID, Id KLS, Id PTK, Id PD, Nilai Angka, Nilai Huruf, Nilai Index
// ============================================================================================================================

func (s *NPDContract) CreateNpd(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run CreateNpd function with args: %+q.", args)

	if len(args) != 7 {
		logger.Errorf(ER11, 7, len(args))
		return fmt.Errorf(ER11, 7, len(args))
	}

	id:= args[0]
	idKls:= args[1]
	idPtk:= args[2]
	idPd:= args[3]
	nilaiAngkaStr:= args[4]
	nilaiHuruf:= args[5]
	nilaiIndexStr:= args[6]

	exists, err := isNpdExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		logger.Errorf(ER12, id)
		return fmt.Errorf(ER12, id)
	}

	nilaiAngka, err := strconv.ParseFloat(nilaiAngkaStr, 64)
	if err != nil {
		logger.Errorf(ER36, id)
		return fmt.Errorf(ER36, id)
	}

	nilaiIndex, err := strconv.ParseFloat(nilaiIndexStr, 64)
	if err != nil {
		logger.Errorf(ER36, id)
		return fmt.Errorf(ER36, id)
	}

	pd, err := getPdById(ctx, idPd)
	if err != nil {
		return err
	}

	kls, err := getKlsById(ctx, idKls)
	if err != nil {
		return err
	}

	totalMutu := pd.TotalMutu + (float64(kls.SKS) * nilaiIndex)
	totalSks := pd.TotalSKS + kls.SKS
	ipk := math.Round((totalMutu / float64(totalSks)) * 100) / 100

	err = updatePdRecord(ctx, idPd, fmt.Sprintf("%v", totalMutu), fmt.Sprintf("%v", totalSks), fmt.Sprintf("%v", ipk))
	if err != nil {
		return err
	}

	npd := NilaiPesertaDidik{
		ID:      			id,
		IdKLS:				idKls,
		IdPTK:				idPtk,
		IdPD:				idPd,
		NilaiAngka:			nilaiAngka,
		NilaiHuruf:			nilaiHuruf,
		NilaiIndex:			nilaiIndex,
	}

	npdJSON, err := json.Marshal(npd)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, npdJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// UpdateNpd - Updates an existing Nilai Peserta Didik (NPD) in the world state with provided parameters.
// Arguments - ID, Id KLS, Id PTK, Id PD, Nilai Angka, Nilai Huruf, Nilai Index
// ============================================================================================================================

func (s *NPDContract) UpdateNpd(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run UpdateNpd function with args: %+q.", args)

	if len(args) != 7 {
		logger.Errorf(ER11, 7, len(args))
		return fmt.Errorf(ER11, 7, len(args))
	}

	id:= args[0]
	idKls:= args[1]
	idPtk:= args[2]
	idPd:= args[3]
	nilaiAngkaStr:= args[4]
	nilaiHuruf:= args[5]
	nilaiIndexStr:= args[6]

	npd, err := getNpdStateById(ctx, id)
	if err != nil {
		return err
	}

	nilaiAngka, err := strconv.ParseFloat(nilaiAngkaStr, 64)
	if err != nil {
		logger.Errorf(ER36, id)
		return fmt.Errorf(ER36, id)
	}

	nilaiIndex, err := strconv.ParseFloat(nilaiIndexStr, 64)
	if err != nil {
		logger.Errorf(ER36, id)
		return fmt.Errorf(ER36, id)
	}

	pd, err := getPdById(ctx, idPd)
	if err != nil {
		return err
	}

	kls, err := getKlsById(ctx, idKls)
	if err != nil {
		return err
	}

	totalMutu := pd.TotalMutu + (float64(kls.SKS) * nilaiIndex) - (float64(kls.SKS) * npd.NilaiIndex)
	totalSks := pd.TotalSKS
	ipk := math.Round((totalMutu / float64(totalSks)) * 100) / 100

	err = updatePdRecord(ctx, idPd, fmt.Sprintf("%v", totalMutu), fmt.Sprintf("%v", totalSks), fmt.Sprintf("%v", ipk))
	if err != nil {
		return err
	}

	npd.IdKLS = idKls
	npd.IdPTK = idPtk
	npd.IdPD = idPd
	npd.NilaiAngka = nilaiAngka
	npd.NilaiHuruf = nilaiHuruf
	npd.NilaiIndex = nilaiIndex

	npdJSON, err := json.Marshal(npd)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, npdJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}

// ============================================================================================================================
// DeleteNpd - Deletes an given Nilai Peserta Didik (NPD) from the world state.
// Arguments - ID
// ============================================================================================================================

func (s *NPDContract) DeleteNpd(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run DeleteNpd function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	npd, err := getNpdStateById(ctx, id)
	if err != nil {
		return err
	}

	pd, err := getPdById(ctx, npd.IdPD)
	if err != nil {
		return err
	}

	kls, err := getKlsById(ctx, npd.IdKLS)
	if err != nil {
		return err
	}

	totalMutu := pd.TotalMutu - (float64(kls.SKS) * npd.NilaiIndex)
	totalSks := pd.TotalSKS - kls.SKS
	ipk := math.Round((totalMutu / float64(totalSks)) * 100) / 100

	err = updatePdRecord(ctx, npd.IdPD, fmt.Sprintf("%v", totalMutu), fmt.Sprintf("%v", totalSks), fmt.Sprintf("%v", ipk))
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(id)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// GetAllNpd - Returns all Nilai Peserta Didik (NPD) found in world state.
// No Arguments
// ============================================================================================================================

func (s *NPDContract) GetAllNpd(ctx contractapi.TransactionContextInterface) ([]*NilaiPesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetAllNpd function with args: %+q.", args)

	if len(args) != 0 {
		logger.Errorf(ER11, 0, len(args))
		return nil, fmt.Errorf(ER11, 0, len(args))
	}

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}


// ============================================================================================================================
// GetNpdById - Get the Nilai Peserta Didik (NPD) stored in the world state with given id.
// Arguments - ID
// ============================================================================================================================

func (s *NPDContract) GetNpdById(ctx contractapi.TransactionContextInterface) (*NilaiPesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetNpdById function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	npd, err := getNpdStateById(ctx, id)
	if err != nil {
		return nil, err
	}

	return npd, nil
}


// ============================================================================================================================
// GetNpdByIdKls - Get the Nilai Peserta Didik (NPD) stored in the world state with given IdKls.
// Arguments - idKls
// ============================================================================================================================

func (t *NPDContract) GetNpdByIdKls(ctx contractapi.TransactionContextInterface) ([]*NilaiPesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetNpdByIdKls function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	idKls:= args[0]

	queryString := fmt.Sprintf(`{"selector":{"idKls":"%s"}}`, idKls)
	return getQueryResultForQueryString(ctx, queryString)
}


// ============================================================================================================================
// GetNpdByIdPd - Get the Nilai Peserta Didik (NPD) stored in the world state with given IdPd.
// Arguments - idPd
// ============================================================================================================================

func (t *NPDContract) GetNpdByIdPd(ctx contractapi.TransactionContextInterface) ([]*NilaiPesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetNpdByIdPd function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	idPd:= args[0]

	queryString := fmt.Sprintf(`{"selector":{"idPd":"%s"}}`, idPd)
	return getQueryResultForQueryString(ctx, queryString)
}


// ============================================================================================================================
// GetNpdLastTxIdById - Get Last Tx Id of Nilai Peserta Didik (NPD) with given Id.
// Arguments - ID
// ============================================================================================================================

func (t *NPDContract) GetNpdLastTxIdById(ctx contractapi.TransactionContextInterface) (string, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetNpdLastTxIdById function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return "", fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	response, err := resultsIterator.Next()
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	return response.TxId, nil
}


// ============================================================================================================================
// isNpdExists - Returns true when Nilai Peserta Didik (NPD) with given ID exists in world state.
// ============================================================================================================================

func isNpdExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	logger.Infof("Run isNpdExists function with id: '%s'.", id)

	npdJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		logger.Errorf(ER32, err)
		return false, fmt.Errorf(ER32, err)
	}

	return npdJSON != nil, nil
}


// ============================================================================================================================
// getNpdStateById - Get KLS state with given id.
// ============================================================================================================================

func getNpdStateById(ctx contractapi.TransactionContextInterface, id string) (*NilaiPesertaDidik, error) {
	logger.Infof("Run getNpdStateById function with id: '%s'.", id)

	npdJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	if npdJSON == nil {
		return nil, fmt.Errorf(ER13, id)
	}

	var npd NilaiPesertaDidik
	err = json.Unmarshal(npdJSON, &npd)
	if err != nil {
		return nil, fmt.Errorf(ER34, err)
	}

	return &npd, nil
}


// ============================================================================================================================
// getKlsById - Get KLS with given idKls.
// ============================================================================================================================

func getKlsById(ctx contractapi.TransactionContextInterface, idKls string) (*KelasKuliah, error) {
	logger.Infof("Run getKlsById function with idKls: '%s'.", idKls)

	params := []string{"GetKlsById", idKls}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := ctx.GetStub().InvokeChaincode(KLSContract, queryArgs, AcademicChannel)
	if response.Status != shim.OK {
		return nil, fmt.Errorf(ER37, KLSContract, response.Payload)
	}

	var kls KelasKuliah
	err := json.Unmarshal([]byte(response.Payload), &kls)
	if err != nil {
		return nil, fmt.Errorf(ER34, err)
	}

	return &kls, nil
}


// ============================================================================================================================
// getPdById - Get PD with given idPd.
// ============================================================================================================================

func getPdById(ctx contractapi.TransactionContextInterface, idPd string) (*PesertaDidik, error) {
	logger.Infof("Run getPdById function with idPd: '%s'.", idPd)

	params := []string{"GetPdById", idPd}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := ctx.GetStub().InvokeChaincode(PDContract, queryArgs, AcademicChannel)
	if response.Status != shim.OK {
		return nil, fmt.Errorf(ER37, PDContract, response.Payload)
	}

	var pd PesertaDidik
	err := json.Unmarshal([]byte(response.Payload), &pd)
	if err != nil {
		return nil, fmt.Errorf(ER34, err)
	}

	return &pd, nil
}


// ============================================================================================================================
// updatePdRecord - Update PD Record with given.
// ============================================================================================================================

func updatePdRecord(ctx contractapi.TransactionContextInterface, idPd string, totalMutu string, totalSks string, ipk string) (error) {
	logger.Infof("Run updatePdRecord function with idPd: '%s'.", idPd)

	params := []string{"UpdatePdRecord", idPd, totalMutu, totalSks, ipk}
	invokeArgs := make([][]byte, len(params))
	for i, arg := range params {
		invokeArgs[i] = []byte(arg)
	}

	response := ctx.GetStub().InvokeChaincode(PDContract, invokeArgs, AcademicChannel)
	if response.Status != shim.OK {
		return fmt.Errorf(ER37, PDContract, response.Payload)
	}

	return nil
}


// ============================================================================================================================
// constructQueryResponseFromIterator - Constructs a slice of assets from the resultsIterator.
// ============================================================================================================================

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*NilaiPesertaDidik, error) {
	logger.Infof("Run constructQueryResponseFromIterator function.")

	var npdList []*NilaiPesertaDidik

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(ER33, err)
		}

		var npd NilaiPesertaDidik
		err = json.Unmarshal(queryResult.Value, &npd)
		if err != nil {
			return nil, fmt.Errorf(ER34, err)
		}
		npdList = append(npdList, &npd)
	}

	return npdList, nil
}


// ============================================================================================================================
// getQueryResultForQueryString - Get a query result from query string
// ============================================================================================================================

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*NilaiPesertaDidik, error) {
	logger.Infof("Run getQueryResultForQueryString function with queryString: '%s'.", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}
