package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//LoanMarketShareObj ...
type LoanMarketShareObj struct {
	ShareID     int       `json:"shareID"`
	TitleHolder string    `json:"titleholder"`
	Amount      float64   `json:"amount"`
	Repayments  float64   `json:"repayments"`
	Statutes    string    `json:"statutes"`
	Rating      float64   `json:"rating"`
	Status      string    `json:"status"`
	Created     time.Time `json:"created"`
	Createdby   string    `json:"createdby"`
}

//LoanMarketShareHistory ...
type LoanMarketShareHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	LoanMarketShare *LoanMarketShareObj  `json:"loanmarketshare"`
}

//LoanMarketShareContract for handling writing and reading from the world state
type LoanMarketShareContract struct {
	contractapi.Contract
}


//Put adds a new key with value to the world state
func (sc *LoanMarketShareContract) Put(ctx contractapi.TransactionContextInterface, shareID int, titleholder string, amount float64, repayments float64, statutes string, rating float64, status string)	(err error) {

	if shareID == 0 {
		err = errors.New("Loan Rating ID can not be empty")
		return
	}
	
	obj := new(LoanMarketShareObj)
	obj.ShareID = shareID
	obj.TitleHolder = titleholder
	obj.Amount = amount
	obj.Repayments = repayments
	obj.Statutes = statutes
	obj.Rating = rating
	obj.Status = status
	

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(shareID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *LoanMarketShareContract) Get(ctx contractapi.TransactionContextInterface, key string) (*LoanMarketShareObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	loanMarketShareObj := new(LoanMarketShareObj)
	if err := json.Unmarshal(existingObj, loanMarketShareObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type LoanMarketShareObj", key)
	}
    return loanMarketShareObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *LoanMarketShareContract) History(ctx contractapi.TransactionContextInterface, key string) ([]LoanMarketShareHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []LoanMarketShareHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(LoanMarketShareObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := LoanMarketShareHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			LoanMarketShare:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
