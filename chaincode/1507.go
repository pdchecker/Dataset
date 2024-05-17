package main

import (
	"os"
	"testing"
	"encoding/json"
	"strconv"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func  TestLoanRatingContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(LoanRatingContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test LoanRatingContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("LoanRatingContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("1.1"),
		[]byte("Rating Desc"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test LoanRatingContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("LoanRatingContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	loanRatingObj := new(LoanRatingObj)
	err = json.Unmarshal(getResp.Payload, loanRatingObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(loanRatingObj, "loanRatingObj should not be nil")

	retrievedID := strconv.Itoa(loanRatingObj.RatingID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}