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

//BuyerW writes state to blockchain
func BuyerW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 5 {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 5")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Buyer ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	buyerID, buyerIDErr := strconv.Atoi(args[0])
	if buyerIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODENOTALLWD, ERR: buyerIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("User ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	userID, userIDErr := strconv.Atoi(args[1])
	if userIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODENOTALLWD, ERR: userIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Buyer Type can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Details can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[4]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BuyerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Reg Date can not be empty")}
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

	objData := models.BuyerBase{
		BuyerID:   buyerID, //args[0],
		UserID:    userID,  //args[1],
		BuyerType: args[2],
		Details:   args[3],
		RegDate:   args[4],
		Created:   created,
		Createdby: callerID,
	}

	obj := &models.Buyer{ObjectType: utils.BUYERO, BuyerID: args[0], Data: objData}
	return obj.PutState(stub)
}

//BuyerR reads state from blockchain
func BuyerR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.BuyerR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Buyer ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.BuyerR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Buyer ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.Buyer{BuyerID: args[0]}
	return obj.GetState(stub)
}
