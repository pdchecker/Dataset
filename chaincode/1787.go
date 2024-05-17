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

//UserCategoryW writes state to blockchain
func UserCategoryW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 3 {
		cErr := &utils.ChainError{FCN: utils.UserCategoryW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 3")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.UserCategoryW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Category ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	catID, catIDErr := strconv.Atoi(args[0])
	if catIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.UserCategoryW, KEY: "", CODE: utils.CODENOTALLWD, ERR: catIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.UserCategoryW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Category Name can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.UserCategoryW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Category Description can not be empty")}
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

	objData := models.UserCategoryBase{
		CatID:          catID, //args[0],
		CatName:        args[1],
		CatDescription: args[2],
		Created:        created,
		Createdby:      callerID,
	}

	obj := &models.UserCategory{ObjectType: utils.USECAT, CatID: args[0], Data: objData}
	return obj.PutState(stub)
}

//UserCategoryR reads state from blockchain
func UserCategoryR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.UserCategoryR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Category ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.UserCategoryR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Category ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.UserCategory{CatID: args[0]}
	return obj.GetState(stub)
}
