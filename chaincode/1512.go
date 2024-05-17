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

func  TestBankBranchContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(BankBranchContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test BankBranchContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("BankBranchContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("Test Branch"),
		[]byte("1"),
		[]byte("1"),
		[]byte("Test Location 1"),
		[]byte("Test LocationLat 1"),
		[]byte("Test LocationLong 1"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test BankBranchContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("BankBranchContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	bankBranchObj := new(BankBranchObj)
	err = json.Unmarshal(getResp.Payload, bankBranchObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(bankBranchObj, "bankBranchObj should not be nil")

	retrievedID := strconv.Itoa(bankBranchObj.BranchID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}