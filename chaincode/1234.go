package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/abbeydabiri/hdlchaincode/models"
	"github.com/abbeydabiri/hdlchaincode/utils"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//PermissionsW writes state to blockchain
func PermissionsW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		cErr := &utils.ChainError{FCN: utils.PermissionsW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 4")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PermissionsW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Permission ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	permID, permIDErr := strconv.Atoi(args[0])
	if permIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.PermissionsW, KEY: "", CODE: utils.CODENOTALLWD, ERR: permIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PermissionsW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Permission Role ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	permroleID, permroleIDErr := strconv.Atoi(args[1])
	if permroleIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.PermissionsW, KEY: "", CODE: utils.CODENOTALLWD, ERR: permroleIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PermissionsW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Permission Name can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PermissionsW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Permission Module can not be empty")}
		return shim.Error(cErr.Error())
	}

	epochTime, _ := stub.GetTxTimestamp()
	created := time.Unix(epochTime.GetSeconds(), 0)
	// Bypass whilst running unit test
	var callerID string
	if os.Getenv("MODE") != "TEST" {
		var err *utils.ChainError
		callerID, err = utils.GetCallerID(stub)
		if err != nil {
			return shim.Error(err.Error())
		}
	} else {
		callerID = "Test Logic Caller"
	}

	objData := models.PermissionsBase{
		PermID:     permID,     //args[0],
		PermRoleID: permroleID, //args[1],
		PermName:   args[2],
		PermModule: args[3],
		Created:    created,
		Createdby:  callerID,
	}

	obj := &models.Permissions{ObjectType: utils.PERMSN, PermID: args[0], Data: objData}
	return obj.PutState(stub)
}

//PermissionsR reads state from blockchain
func PermissionsR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.PermissionsR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Permissions ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PermissionsR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Permissions ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.Permissions{PermID: args[0]}
	return obj.GetState(stub)
}
