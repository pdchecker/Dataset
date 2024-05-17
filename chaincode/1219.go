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

//BankW writes state to blockchain
func BankW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 8 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 8")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	bankID, bankIDErr := strconv.Atoi(args[0])
	if bankIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODENOTALLWD, ERR: bankIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank Name can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank HQ Address can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank Category can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[4]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank Admin UserID can not be empty")}
		return shim.Error(cErr.Error())
	}
	bankAdminUserID, bankAdminUserIDErr := strconv.Atoi(args[4])
	if bankAdminUserIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODENOTALLWD, ERR: bankAdminUserIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[5]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[6]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location Lat can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[7]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location Long can not be empty")}
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

	objData := models.BankBase{
		BankID:          bankID, //args[0],
		BankName:        args[1],
		HqAddress:       args[2],
		BankCategory:    args[3],
		BankAdminUserID: bankAdminUserID, //args[4],
		Location:        args[5],
		LocationLat:     args[6],
		LocationLong:    args[7],
		Created:         created,
		Createdby:       callerID,
	}

	obj := &models.Bank{ObjectType: utils.BANKOO, BankID: args[0], Data: objData}
	return obj.PutState(stub)
}

//BankR reads state from blockchain
func BankR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.BankR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Bank ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.Bank{BankID: args[0]}
	return obj.GetState(stub)
}
