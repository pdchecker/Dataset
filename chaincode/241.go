package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//LoanObj ...
type LoanObj struct {
	LoanID     int       `json:"loanID"`
	PropertyID int       `json:"propertyID"`
	UserID     int       `json:"userID"`
	BuyerID    int       `json:"buyerID"`
	Repayment  float64   `json:"repayment"`
	LoanStatus string    `json:"loanstatus"`
	PerfRating float64   `json:"perfrating"`
	Created    time.Time `json:"created"`
	Createdby  string    `json:"createdby"`
}

//LoanHistory ...
type LoanHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	Loan *LoanObj  `json:"loan"`
}

//LoanContract for handling writing and reading from the world state
type LoanContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *LoanContract) Put(ctx contractapi.TransactionContextInterface, loanID int, propertyID int, userID int, buyerID int, repayment float64, loanstatus string, perfrating float64)	(err error) {

	if loanID == 0 {
		err = errors.New("Loan ID can not be empty")
		return
	}
	
	obj := new(LoanObj)
	obj.LoanID = loanID
	obj.PropertyID = propertyID
	obj.UserID = userID
	obj.BuyerID = buyerID
	obj.Repayment = repayment
	obj.LoanStatus = loanstatus
	obj.PerfRating = perfrating	

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(loanID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *LoanContract) Get(ctx contractapi.TransactionContextInterface, key string) (*LoanObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	loanObj := new(LoanObj)
	if err := json.Unmarshal(existingObj, loanObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type LoanObj", key)
	}
    return loanObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *LoanContract) History(ctx contractapi.TransactionContextInterface, key string) ([]LoanHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []LoanHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(LoanObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := LoanHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			Loan:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
