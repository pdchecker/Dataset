package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ============================================================================================================================
// Logger
// ============================================================================================================================

var logger = flogging.MustGetLogger("MKContract")


// ============================================================================================================================
// Contract Definitions
// ============================================================================================================================

type MKContract struct {
	contractapi.Contract
}


// ============================================================================================================================
// Asset Definitions - The ledger will store Mata Kuliah (MK) data
// ============================================================================================================================

type MataKuliah struct {
	ID      			string 	`json:"id"`
	IdSP				string 	`json:"idSp"`
	IdSMS				string 	`json:"idSms"`
	NamaMK				string 	`json:"namaMk"`
	KodeMK				string 	`json:"kodeMk"`
	SKS					int 	`json:"sks"`
	JenjangPendidikan	string 	`json:"jenjangPendidikan"`
}


// ============================================================================================================================
// Error Messages
// ============================================================================================================================

const (
	ER11 string = "ER11-Incorrect number of arguments. Required %d arguments, but you have %d arguments."
	ER12        = "ER12-MataKuliah with id '%s' already exists."
	ER13        = "ER13-MataKuliah with id '%s' doesn't exist."
	ER31        = "ER31-Failed to change to world state: %v."
	ER32        = "ER32-Failed to read from world state: %v."
	ER33        = "ER33-Failed to get result from iterator: %v."
	ER34        = "ER34-Failed unmarshaling JSON: %v."
	ER35        = "ER35-Failed parsing string to integer: %v."
	ER41        = "ER41-Access is not permitted with MSDPID '%s'."
	ER42        = "ER42-Unknown MSPID: '%s'."
)


// ============================================================================================================================
// CreateMk - Issues a new Mata Kuliah (MK) to the world state with given details.
// Arguments - ID, Id SP, Id SMS, Nama MK, Kode MK, SKS, Jenjang Pendidikan
// ============================================================================================================================

func (s *MKContract) CreateMk(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run CreateMk function with args: %+q.", args)

	if len(args) != 7 {
		logger.Errorf(ER11, 7, len(args))
		return fmt.Errorf(ER11, 7, len(args))
	}

	id:= args[0]
	idSp:= args[1]
	idSms:= args[2]
	namaMk:= args[3]
	kodeMk:= args[4]
	sksStr:= args[5]
	jenjangPendidikan:= args[6]

	exists, err := isMkExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		logger.Errorf(ER12, id)
		return fmt.Errorf(ER12, id)
	}

	sks, err := strconv.Atoi(sksStr)
	if err != nil {
		logger.Errorf(ER35, id)
		return fmt.Errorf(ER35, id)
	}

	mk := MataKuliah{
		ID:      			id,
		IdSMS:				idSms,
		IdSP:				idSp,
		NamaMK:				namaMk,
		KodeMK:				kodeMk,
		SKS:				sks,
		JenjangPendidikan:	jenjangPendidikan,
	}

	mkJSON, err := json.Marshal(mk)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, mkJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// UpdateMk - Updates an existing Mata Kuliah (MK) in the world state with provided parameters.
// Arguments - ID, Id SP, Id SMS, Nama MK, Kode MK, SKS, Jenjang Pendidikan
// ============================================================================================================================

func (s *MKContract) UpdateMk(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run UpdateMk function with args: %+q.", args)

	if len(args) != 7 {
		logger.Errorf(ER11, 7, len(args))
		return fmt.Errorf(ER11, 7, len(args))
	}

	id:= args[0]
	idSp:= args[1]
	idSms:= args[2]
	namaMk:= args[3]
	kodeMk:= args[4]
	sksStr:= args[5]
	jenjangPendidikan:= args[6]

	exists, err := isMkExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf(ER13, id)
	}

	sks, err := strconv.Atoi(sksStr)
	if err != nil {
		logger.Errorf(ER35, id)
		return fmt.Errorf(ER35, id)
	}

	mk := MataKuliah{
		ID:      			id,
		IdSP:				idSp,
		IdSMS:				idSms,
		NamaMK:				namaMk,
		KodeMK:				kodeMk,
		SKS:				sks,
		JenjangPendidikan:	jenjangPendidikan,
	}

	mkJSON, err := json.Marshal(mk)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, mkJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// DeleteMk - Deletes an given Mata Kuliah (MK) from the world state.
// Arguments - ID
// ============================================================================================================================

func (s *MKContract) DeleteMk(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run DeleteMk function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	exists, err := isMkExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf(ER13, id)
	}

	err = ctx.GetStub().DelState(id)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// GetAllMk - Returns all Mata Kuliah (MK) found in world state.
// No Arguments
// ============================================================================================================================

func (s *MKContract) GetAllMk(ctx contractapi.TransactionContextInterface) ([]*MataKuliah, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetAllMk function with args: %+q.", args)

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
// GetMkById - Get the Mata Kuliah (MK) stored in the world state with given id.
// Arguments - ID
// ============================================================================================================================

func (s *MKContract) GetMkById(ctx contractapi.TransactionContextInterface) (*MataKuliah, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetMkById function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	mkJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	if mkJSON == nil {
		return nil, fmt.Errorf(ER13, id)
	}

	var mk MataKuliah
	err = json.Unmarshal(mkJSON, &mk)
	if err != nil {
		return nil, fmt.Errorf(ER34, err)
	}

	return &mk, nil
}


// ============================================================================================================================
// GetMkByIdSp - Get the Mata Kuliah (MK) stored in the world state with given IdSp.
// Arguments - idSp
// ============================================================================================================================

func (t *MKContract) GetMkByIdSp(ctx contractapi.TransactionContextInterface) ([]*MataKuliah, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetMkByIdSp function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	idSp:= args[0]

	queryString := fmt.Sprintf(`{"selector":{"idSp":"%s"}}`, idSp)
	return getQueryResultForQueryString(ctx, queryString)
}


// ============================================================================================================================
// isMkExists - Returns true when Mata Kuliah (MK) with given ID exists in world state.
// ============================================================================================================================

func isMkExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	logger.Infof("Run isMkExists function with id: '%s'.", id)

	mkJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		logger.Errorf(ER32, err)
		return false, fmt.Errorf(ER32, err)
	}

	return mkJSON != nil, nil
}


// ============================================================================================================================
// constructQueryResponseFromIterator - Constructs a slice of assets from the resultsIterator.
// ============================================================================================================================

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*MataKuliah, error) {
	logger.Infof("Run constructQueryResponseFromIterator function.")

	var mkList []*MataKuliah

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(ER33, err)
		}

		var mk MataKuliah
		err = json.Unmarshal(queryResult.Value, &mk)
		if err != nil {
			return nil, fmt.Errorf(ER34, err)
		}
		mkList = append(mkList, &mk)
	}

	return mkList, nil
}


// ============================================================================================================================
// getQueryResultForQueryString - Get a query result from query string
// ============================================================================================================================

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*MataKuliah, error) {
	logger.Infof("Run getQueryResultForQueryString function with queryString: '%s'.", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}
