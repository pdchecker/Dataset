package main

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"strconv"

	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ============================================================================================================================
// Logger
// ============================================================================================================================

var logger = flogging.MustGetLogger("PDContract")


// ============================================================================================================================
// Contract Definitions
// ============================================================================================================================

type PDContract struct {
	contractapi.Contract
}


// ============================================================================================================================
// Asset Definitions - The ledger will store Peserta Didik (PD) data
// ============================================================================================================================

type PesertaDidik struct {
	ID      			string 	`json:"id"`
	IdSP				string 	`json:"idSp"`
	IdSMS				string 	`json:"idSms"`
	NamaPD				string 	`json:"namaPd"`
	NIPD				string 	`json:"nipd"`
	Username			string 	`json:"username"`
	TotalMutu			float64	`json:"totalMutu"`
	TotalSKS			int 	`json:"totalSks"`
	IPK					float64	`json:"ipk"`
	Status				int		`json:"status"`
}


// ============================================================================================================================
// Error Messages
// ============================================================================================================================

const (
	ER11 string = "ER11-Incorrect number of arguments. Required %d arguments, but you have %d arguments."
	ER12        = "ER12-PesertaDidik with id '%s' already exists."
	ER13        = "ER13-PesertaDidik with id '%s' doesn't exist."
	ER31        = "ER31-Failed to change to world state: %v."
	ER32        = "ER32-Failed to read from world state: %v."
	ER33        = "ER33-Failed to get result from iterator: %v."
	ER34        = "ER34-Failed unmarshaling JSON: %v."
	ER35        = "ER35-Failed parsing string to integer: %v."
	ER36        = "ER36-Failed parsing string to float: %v."
	ER41        = "ER41-Access is not permitted with MSDPID '%s'."
	ER42        = "ER42-Unknown MSPID: '%s'."
)


// ============================================================================================================================
// PD Status
// ============================================================================================================================

const (
	BELUMLULUS int	= 0
	LULUS        	= 1
)


// ============================================================================================================================
// CreatePd - Issues a new Peserta Didik (PD) to the world state with given details.
// Arguments - ID, Id SP, Id SMS, Nama PD, NIPD, Username
// ============================================================================================================================

func (s *PDContract) CreatePd(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run CreatePd function with args: %+q.", args)

	if len(args) != 6 {
		logger.Errorf(ER11, 6, len(args))
		return fmt.Errorf(ER11, 6, len(args))
	}

	id:= args[0]
	idSp:= args[1]
	idSms:= args[2]
	namaPd:= args[3]
	nipd:= args[4]
	username:= args[5]

	exists, err := isPdExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		logger.Errorf(ER12, id)
		return fmt.Errorf(ER12, id)
	}

	pd := PesertaDidik{
		ID:      			id,
		IdSP:				idSp,
		IdSMS:				idSms,
		NamaPD:				namaPd,
		NIPD:				nipd,
		Username:			username,
		TotalMutu:			0.00,
		TotalSKS:			0,
		IPK:				0.00,
		Status:				BELUMLULUS,
	}

	pdJSON, err := json.Marshal(pd)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, pdJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// UpdatePd - Updates an existing Peserta Didik (PD) in the world state with provided parameters.
// Arguments - ID, Id SP, Id SMS, Nama PD, NIPD
// ============================================================================================================================

func (s *PDContract) UpdatePd(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run UpdatePd function with args: %+q.", args)

	if len(args) != 5 {
		logger.Errorf(ER11, 5, len(args))
		return fmt.Errorf(ER11, 5, len(args))
	}

	id:= args[0]
	idSp:= args[1]
	idSms:= args[2]
	namaPd:= args[3]
	nipd:= args[4]

	pd, err := getPdStateById(ctx, id)
	if err != nil {
		return err
	}

	pd.IdSP = idSp
	pd.IdSMS = idSms
	pd.NamaPD = namaPd
	pd.NIPD = nipd

	pdJSON, err := json.Marshal(pd)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, pdJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// UpdatePdRecord - Update TotalMutu, TotalSKS, and IPK of an existing Peserta Didik (PD) in the world state.
// Arguments - ID, TotalMutu, TotalSKS, and IPK
// ============================================================================================================================

func (s *PDContract) UpdatePdRecord(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run UpdatePdRecord function with args: %+q.", args)

	if len(args) != 4 {
		logger.Errorf(ER11, 4, len(args))
		return fmt.Errorf(ER11, 4, len(args))
	}

	id:= args[0]
	totalMutuStr:= args[1]
	totalSksStr:= args[2]
	ipkStr:= args[3]

	pd, err := getPdStateById(ctx, id)
	if err != nil {
		return err
	}

	totalMutu, err := strconv.ParseFloat(totalMutuStr, 64)
	if err != nil {
		logger.Errorf(ER35, id)
		return fmt.Errorf(ER35, id)
	}

	totalSks, err := strconv.Atoi(totalSksStr)
	if err != nil {
		logger.Errorf(ER35, id)
		return fmt.Errorf(ER35, id)
	}

	ipk, err := strconv.ParseFloat(ipkStr, 64)
	if err != nil {
		logger.Errorf(ER36, id)
		return fmt.Errorf(ER36, id)
	}

	pd.TotalMutu = totalMutu
	pd.TotalSKS = totalSks
	pd.IPK = ipk

	pdJSON, err := json.Marshal(pd)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, pdJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// SetPdGraduated - Set status of an existing Peserta Didik (PD) in the world state to 'LULUS'.
// Arguments - ID
// ============================================================================================================================

func (s *PDContract) SetPdGraduated(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run SetPdGraduated function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	pd, err := getPdStateById(ctx, id)
	if err != nil {
		return err
	}

	pd.Status = LULUS

	pdJSON, err := json.Marshal(pd)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, pdJSON)
	if err != nil {
		logger.Errorf(ER31, err)
	}

	return err
}


// ============================================================================================================================
// SetPdGraduatedBatch - Set status of an existing Peserta Didik (PD) in the world state to 'LULUS' in batch.
// Arguments - List ID PD
// ============================================================================================================================

func (s *PDContract) SetPdGraduatedBatch(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run SetPdGraduatedBatch function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return fmt.Errorf(ER11, 1, len(args))
	}

	listPdStr:= args[0]

	listPdStr = strings.Replace(listPdStr, "[", "", -1)
	listPdStr = strings.Replace(listPdStr, "]", "", -1)
	splitter := regexp.MustCompile(` *, *`)
	listPd :=  splitter.Split(listPdStr, -1)

	for _, id := range listPd {
		pd, err := getPdStateById(ctx, id)
		if err != nil {
			return err
		}

		pd.Status = LULUS

		pdJSON, err := json.Marshal(pd)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(id, pdJSON)
		if err != nil {
			logger.Errorf(ER31, err)
			return err
		}
	}

	return nil
}


// ============================================================================================================================
// DeletePd - Deletes an given Peserta Didik (PD) from the world state.
// Arguments - ID
// ============================================================================================================================

func (s *PDContract) DeletePd(ctx contractapi.TransactionContextInterface) error {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run DeletePd function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	exists, err := isPdExists(ctx, id)
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
// GetAllPd - Returns all Peserta Didik (PD) found in world state.
// No Arguments
// ============================================================================================================================

func (s *PDContract) GetAllPd(ctx contractapi.TransactionContextInterface) ([]*PesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetAllPd function with args: %+q.", args)

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
// GetPdById - Get the Peserta Didik (PD) stored in the world state with given id.
// Arguments - ID
// ============================================================================================================================

func (s *PDContract) GetPdById(ctx contractapi.TransactionContextInterface) (*PesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetPdById function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	id:= args[0]

	pd, err := getPdStateById(ctx, id)
	if err != nil {
		return nil, err
	}

	return pd, nil
}


// ============================================================================================================================
// GetPdByIdSp - Get the Peserta Didik (PD) stored in the world state with given IdSp.
// Arguments - idSp
// ============================================================================================================================

func (t *PDContract) GetPdByIdSp(ctx contractapi.TransactionContextInterface) ([]*PesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetPdByIdSp function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	idSp:= args[0]

	queryString := fmt.Sprintf(`{"selector":{"idSp":"%s"}}`, idSp)
	return getQueryResultForQueryString(ctx, queryString)
}


// ============================================================================================================================
// GetPdByIdSms - Get the Peserta Didik (PD) stored in the world state with given IdSms.
// Arguments - idSms
// ============================================================================================================================

func (t *PDContract) GetPdByIdSms(ctx contractapi.TransactionContextInterface) ([]*PesertaDidik, error) {
	args := ctx.GetStub().GetStringArgs()[1:]

	logger.Infof("Run GetPdByIdSms function with args: %+q.", args)

	if len(args) != 1 {
		logger.Errorf(ER11, 1, len(args))
		return nil, fmt.Errorf(ER11, 1, len(args))
	}

	idSms:= args[0]

	queryString := fmt.Sprintf(`{"selector":{"idSms":"%s"}}`, idSms)
	return getQueryResultForQueryString(ctx, queryString)
}


// ============================================================================================================================
// isPdExists - Returns true when Peserta Didik (PD) with given ID exists in world state.
// ============================================================================================================================

func isPdExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	logger.Infof("Run isPdExists function with id: '%s'.", id)

	pdJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		logger.Errorf(ER32, err)
		return false, fmt.Errorf(ER32, err)
	}

	return pdJSON != nil, nil
}


// ============================================================================================================================
// getPdStateById - Get PD state with given id.
// ============================================================================================================================

func getPdStateById(ctx contractapi.TransactionContextInterface, id string) (*PesertaDidik, error) {
	logger.Infof("Run getPdStateById function with id: '%s'.", id)

	pdJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	if pdJSON == nil {
		return nil, fmt.Errorf(ER13, id)
	}

	var pd PesertaDidik
	err = json.Unmarshal(pdJSON, &pd)
	if err != nil {
		return nil, fmt.Errorf(ER34, err)
	}

	return &pd, nil
}


// ============================================================================================================================
// constructQueryResponseFromIterator - Constructs a slice of assets from the resultsIterator.
// ============================================================================================================================

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*PesertaDidik, error) {
	logger.Infof("Run constructQueryResponseFromIterator function.")

	var pdList []*PesertaDidik

	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(ER33, err)
		}

		var pd PesertaDidik
		err = json.Unmarshal(queryResult.Value, &pd)
		if err != nil {
			return nil, fmt.Errorf(ER34, err)
		}
		pdList = append(pdList, &pd)
	}

	return pdList, nil
}


// ============================================================================================================================
// getQueryResultForQueryString - Get a query result from query string
// ============================================================================================================================

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*PesertaDidik, error) {
	logger.Infof("Run getQueryResultForQueryString function with queryString: '%s'.", queryString)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf(ER32, err)
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}
