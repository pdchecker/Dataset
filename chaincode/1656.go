package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ============================================================================================================================
// Logger
// ============================================================================================================================

var logger = flogging.MustGetLogger("SPContract")


// ============================================================================================================================
// Contract Definitions
// ============================================================================================================================

type SPContract struct {
	contractapi.Contract
}


// ============================================================================================================================
// Asset Definitions - The ledger will store Satuan Pendidikan (SP) data
// ============================================================================================================================

type SatuanPendidikan struct {
	ID      		string `json:"id"`
	IdMSP			string `json:"idMsp"`
	NamaSP			string `json:"namaSp"`
	UsernameAdmin	string `json:"usernameAdmin"`
}


// ============================================================================================================================
// Error Messages
// ============================================================================================================================

const (
	ER11 string = "ER11-Incorrect number of arguments. Required %d arguments, but you have %d arguments."
	ER12        = "ER12-SatuanPendidikan with id '%s' already exists."
	ER13        = "ER13-SatuanPendidikan with id '%s' doesn't exist."
	ER31        = "ER31-Failed to change to world state: %v."
	ER32        = "ER32-Failed to read from world state: %v."
	ER33        = "ER33-Failed to get result from iterator: %v."
	ER34        = "ER34-Failed unmarshaling JSON: %v."
	ER41        = "ER41-Access is not permitted with MSDPID '%s'."
	ER42        = "ER42-Unknown MSPID: '%s'."
)


// ============================================================================================================================
// CreateSp - Issues a new Satuan Pendidikan (SP) to the world state with given details.
// Arguments - ID, Id MSP, Nama SP, Username Admin SP
// ============================================================================================================================

func (s *SPContract) CreateSp(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run CreateSp function with args: %+q.", args)

	if len(args) != 4 {
		logger.Errorf(ER11, 4, len(args))
		return fmt.Errorf(ER11, 4, len(args))
	}

	id:= args[0]
	idMsp:= args[1]
	namaSp:= args[2]
	usernameAdmin:= args[3]

	exists, err := isSpExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		logger.Errorf(ER12, id)
		return fmt.Errorf(ER12, id)
	}

	sp := SatuanPendidikan{
		ID:      		id,
		IdMSP:			idMsp,
		NamaSP:			namaSp,
		UsernameAdmin:	usernameAdmin,
	}

	spJSON, err := json.Marshal(sp)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, spJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// UpdateSp - Updates an existing Satuan Pendidikan (SP) in the world state with provided parameters.
// Arguments - ID, Id MSP, Nama SP
// ============================================================================================================================

func (s *SPContract) UpdateSp(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run UpdateSp function with args: %+q.", args)

	if len(args) != 3 {
		logger.Errorf(ER11, 3, len(args))
		return fmt.Errorf(ER11, 3, len(args))
	}

	id:= args[0]
	idMsp:= args[1]
	namaSp:= args[2]

	sp, err := getSpStateById(ctx, id)
	if err != nil {
		return err
	}

	sp.IdMSP = idMsp
	sp.NamaSP = namaSp

	spJSON, err := json.Marshal(sp)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, spJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// DeleteSp - Deletes an given Satuan Pendidikan (SP) from the world state.
// Arguments - ID
// ============================================================================================================================

func (s *SPContract) DeleteSp(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run DeleteSp function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	exists, err := isSpExists(ctx, id)
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
// GetAllSp - Returns all Satuan Pendidikan (SP) found in world state.
// No Arguments
// ============================================================================================================================

func (s *SPContract) GetAllSp(ctx contractapi.TransactionContextInterface) ([]*SatuanPendidikan, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetAllSp function with args: %+q.", args)

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
// GetSpById - Get the Satuan Pendidikan (SP) stored in the world state with given id.
// Arguments - ID
// ============================================================================================================================

func (s *SPContract) GetSpById(ctx contractapi.TransactionContextInterface) (*SatuanPendidikan, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetSpById function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	sp, err := getSpStateById(ctx, id)
	if err != nil {
		return nil, err
	}

	return sp, nil
}


// ============================================================================================================================
// isSpExists - Returns true when Satuan Pendidikan (SP) with given ID exists in world state.
// ============================================================================================================================

func isSpExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	logger.Infof("Run isSpExists function with id: '%s'.", id)

	spJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		logger.Errorf(ER32, err)
		return false, fmt.Errorf(ER32, err)
	}

	return spJSON != nil, nil
}


// ============================================================================================================================
// getSpStateById - Get SP state with given id.
// ============================================================================================================================

func getSpStateById(ctx contractapi.TransactionContextInterface, id string) (*SatuanPendidikan, error) {
	logger.Infof("Run getSpStateById function with id: '%s'.", id)

	spJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	if spJSON == nil {
		return nil, fmt.Errorf(ER13, id)
	}

	var sp SatuanPendidikan
	err = json.Unmarshal(spJSON, &sp)
	if err != nil {
		return nil, fmt.Errorf(ER34, err)
	}

	return &sp, nil
}


// ============================================================================================================================
// constructQueryResponseFromIterator - Constructs a slice of assets from the resultsIterator.
// ============================================================================================================================

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*SatuanPendidikan, error) {
	logger.Infof("Run constructQueryResponseFromIterator function.")

	var spList []*SatuanPendidikan

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(ER33, err)
		}

		var sp SatuanPendidikan
		err = json.Unmarshal(queryResult.Value, &sp)
		if err != nil {
			return nil, fmt.Errorf(ER34, err)
		}
		spList = append(spList, &sp)
	}

	return spList, nil
}


// ============================================================================================================================
// getQueryResultForQueryString - Get a query result from query string
// ============================================================================================================================

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*SatuanPendidikan, error) {
	logger.Infof("Run getQueryResultForQueryString function with queryString: '%s'.", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}
