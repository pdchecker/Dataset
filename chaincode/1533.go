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

func  TestSellerContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(SellerContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test SellerContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("SellerContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("seller type"),
		[]byte("details"),
		[]byte("regdate"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test SellerContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("SellerContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	sellerObj := new(SellerObj)
	err = json.Unmarshal(getResp.Payload, sellerObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(sellerObj, "sellerObj should not be nil")

	retrievedID := strconv.Itoa(sellerObj.SellerID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}