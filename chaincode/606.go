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

func  TestBuyerContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(BuyerContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test BuyerContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("BuyerContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("buyer type"),
		[]byte("buyer details"),
		[]byte("reg date"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test BuyerContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("BuyerContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	buyerObj := new(BuyerObj)
	err = json.Unmarshal(getResp.Payload, buyerObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(buyerObj, "buyerObj should not be nil")

	retrievedID := strconv.Itoa(buyerObj.BuyerID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}