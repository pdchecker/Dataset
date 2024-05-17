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

//BankBranchW writes state to blockchain
func BankBranchW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 8 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 8")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Branch ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	branchID, branchIDErr := strconv.Atoi(args[0])
	if branchIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODENOTALLWD, ERR: branchIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	bankID, bankIDErr := strconv.Atoi(args[1])
	if bankIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODENOTALLWD, ERR: bankIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Branch Name can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Branch Manager User ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	branchmanageruserID, branchmanageruserIDErr := strconv.Atoi(args[3])
	if branchmanageruserIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODENOTALLWD, ERR: branchmanageruserIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[4]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Branch Manager Role ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	branchmanagerroleID, branchmanagerroleIDErr := strconv.Atoi(args[4])
	if branchmanagerroleIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODENOTALLWD, ERR: branchmanagerroleIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[5]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[6]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location Lat can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[7]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location Long can not be empty")}
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

	objData := models.BankBranchBase{
		BranchID:            branchID, //args[0],
		BankID:              bankID,   //args[1],
		BranchName:          args[2],
		BranchManagerUserID: branchmanageruserID, //args[3],
		BranchManagerRoleID: branchmanagerroleID, //args[4],
		Location:            args[5],
		LocationLat:         args[6],
		LocationLong:        args[7],
		Created:             created,
		Createdby:           callerID,
	}

	obj := &models.BankBranch{ObjectType: utils.BNKBRA, BranchID: args[0], Data: objData}
	return obj.PutState(stub)
}

//BankBranchR reads state from blockchain
func BankBranchR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.BankBranchR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Branch ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BankBranchR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Branch ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.BankBranch{BranchID: args[0]}
	return obj.GetState(stub)
}
