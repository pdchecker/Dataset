package main

import (
	"errors"
	"fmt"
	"time"
	"strconv"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//LoanRatingObj ...
type LoanRatingObj struct {
	RatingID   int       `json:"ratingID"`
	LoanID     int       `json:"loanID"`
	Rating     float64   `json:"rating"`
	RatingDesc string    `json:"ratingdesc"`
	Created   time.Time `json:"created"`
	Createdby string    `json:"createdby"`
}

//LoanRatingHistory ...
type LoanRatingHistory struct {
	TxID string `json:"txID"`
	Timestamp time.Time  `json:"timestamp"`
	LoanRating *LoanRatingObj  `json:"loanrating"`
}

//LoanRatingContract for handling writing and reading from the world state
type LoanRatingContract struct {
	contractapi.Contract
}

//Put adds a new key with value to the world state
func (sc *LoanRatingContract) Put(ctx contractapi.TransactionContextInterface, ratingID int, loanID int, rating float64, ratingdesc string)	(err error) {

	if ratingID == 0 {
		err = errors.New("Loan Rating ID can not be empty")
		return
	}
	
	obj := new(LoanRatingObj)
	obj.RatingID = ratingID
	obj.LoanID = loanID
	obj.Rating = rating
	obj.RatingDesc = ratingdesc
	

	if obj.Created, err = GetTimestamp(ctx); err != nil {
		return
	}

	if obj.Createdby, err = GetCallerID(ctx); err != nil {
		return
	}

	key := strconv.Itoa(ratingID)
	objBytes, _ := json.Marshal(obj)	
	err = ctx.GetStub().PutState(key, []byte(objBytes))
    return 
}

//Get retrieves the value linked to a key from the world state
func (sc *LoanRatingContract) Get(ctx contractapi.TransactionContextInterface, key string) (*LoanRatingObj, error) {
	
    existingObj, err := ctx.GetStub().GetState(key)
    if err != nil {
        return nil, err
    }

    if existingObj == nil {
        return nil, fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

	loanRatingObj := new(LoanRatingObj)
	if err := json.Unmarshal(existingObj, loanRatingObj); err != nil {
		return nil, fmt.Errorf("Data retrieved from world state for key %s was not of type LoanRatingObj", key)
	}
    return loanRatingObj, nil
}

//History retrieves the history linked to a key from the world state
func (sc *LoanRatingContract) History(ctx contractapi.TransactionContextInterface, key string) ([]LoanRatingHistory, error) {

	iter, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
        return nil, err
	}
	defer func() { _ = iter.Close() }()

	var results []LoanRatingHistory
	for iter.HasNext() {
		state, err := iter.Next()
		if err != nil {
			return nil, err
		}

		entryObj := new(LoanRatingObj)
		if errNew := json.Unmarshal(state.Value, entryObj); errNew != nil {
			return nil, errNew
		}

		entry := LoanRatingHistory{
			TxID:		state.GetTxId(),
			Timestamp:	time.Unix(state.GetTimestamp().GetSeconds(), 0),
			LoanRating:	entryObj,
		}

		results = append(results, entry)
	}
	return results, nil
}
