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

func  TestUserCategoryContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(UserCategoryContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test UserCategoryContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("UserCategoryContract:Put"),
		[]byte("1"),
		[]byte("1"),
		[]byte("1.1"),
		[]byte("Rating Desc"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test UserCategoryContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("UserCategoryContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	userCategoryObj := new(UserCategoryObj)
	err = json.Unmarshal(getResp.Payload, userCategoryObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(userCategoryObj, "userCategoryObj should not be nil")

	retrievedID := strconv.Itoa(userCategoryObj.CatID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}