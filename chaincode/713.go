package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//TransactionObj ...
type TransactionObj struct {
	TxnID        int       `json:"txnID"`
	TxnDate      string    `json:"txndate"`
	BuyerID      int       `json:"buyerID"`
	UserID       int       `json:"userID"`
	Repayment    float64   `json:"repayment"`
	Amount       float64   `json:"amount"`
	InterestRate float64   `json:"interestrate"`
	Outstanding  float64   `json:"outstanding"`
	DueDate      string    `json:"duedate"`
	Bank         string    `json:"bank"`
	LoanStatus   string    `json:"loanstatus"`
	Created       time.Time `json:"created"`
	Createdby     string    `json:"createdby"`
}

//TransactionHistory ...
type TransactionHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	Transaction *TransactionObj  `json:"transaction"`
}

//TransactionContract for handling writing and reading from the world state
type TransactionContract struct {
	contractapi.Contract
}


//Put adds a new key with value to the world state
func (sc *TransactionContract) Put(ctx contractapi.TransactionContextInterface, txnID int, txndate string, buyerID int, userID int, repayment float64, amount float64, interestrate float64, outstanding float64, duedate string, bank string, loanstatus string)	(err error) {

	if txnID == 0 {
		err = errors.New("Transaction ID can not be empty")
		return
	}

	obj := new(TransactionObj)
	obj.TxnID = txnID
	obj.TxnDate = txndate
	obj.BuyerID = buyerID
	obj.UserID = userID
	obj.Repayment = repayment
	obj.Amount = amount
	obj.InterestRate = interestrate
	obj.Outstanding = outstanding
	obj.DueDate = duedate
	obj.Bank = bank
	obj.LoanStatus = loanstatus

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(txnID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *TransactionContract) Get(ctx contractapi.TransactionContextInterface, key string) (*TransactionObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	transactionObj := new(TransactionObj)
	if err := json.Unmarshal(existingObj, transactionObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type TransactionObj", key)
	}
    return transactionObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *TransactionContract) History(ctx contractapi.TransactionContextInterface, key string) ([]TransactionHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []TransactionHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(TransactionObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := TransactionHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			Transaction:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
