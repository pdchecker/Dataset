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

func  TestRoleContract(t *testing.T) {
	os.Setenv("MODE","TEST")
	
	assert := assert.New(t)
	uid := uuid.New().String()

	cc, err := contractapi.NewChaincode(new(RoleContract))
	assert.Nil(err, "error should be nil")

	stub := shimtest.NewMockStub("TestStub", cc)
	assert.NotNil(stub, "Stub is nil, TestStub creation failed")

	// - - - test RoleContract:Put function - - - 
	putResp := stub.MockInvoke(uid,[][]byte{
		[]byte("RoleContract:Put"),
		[]byte("1"),
		[]byte("user category"),
		[]byte("user type"),
		[]byte("role name"),
		[]byte("role desc"),
	})
	assert.EqualValues(OK, putResp.GetStatus(), putResp.GetMessage())
	

	// - - - test RoleContract:Get function - - - 
	testID := "1"
	getResp := stub.MockInvoke(uid, [][]byte{
		[]byte("RoleContract:Get"),
		[]byte(testID),
	})
	assert.EqualValues(OK, getResp.GetStatus(), getResp.GetMessage())
	assert.NotNil(getResp.Payload, "getResp.Payload should not be nil")
	
	roleObj := new(RoleObj)
	err = json.Unmarshal(getResp.Payload, roleObj)
	assert.Nil(err, "json.Unmarshal error should be nil")
	assert.NotNil(roleObj, "roleObj should not be nil")

	retrievedID := strconv.Itoa(roleObj.RoleID)
	assert.EqualValues(testID, retrievedID, "testID and retrievedID mismatch")
}