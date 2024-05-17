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

func  TestLoanContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(LoanContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test LoanContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("LoanContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("1"),
		[]byte("1"),
		[]byte("1.1"),
		[]byte("active"),
		[]byte("2.2"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test LoanContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("LoanContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	loanObj := new(LoanObj)
	err = json.Unmarshal(getResp.Payload, loanObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(loanObj, "loanObj should not be nil")

	retrievedID := strconv.Itoa(loanObj.LoanID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}