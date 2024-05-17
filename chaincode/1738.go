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

//SellerW writes state to blockchain
func SellerW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 5 {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 5")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Seller ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	sellerID, sellerIDErr := strconv.Atoi(args[0])
	if sellerIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODENOTALLWD, ERR: sellerIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("User ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	userID, userIDErr := strconv.Atoi(args[1])
	if userIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODENOTALLWD, ERR: userIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Seller Type can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Details can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[4]) == 0 {
		cErr := &utils.ChainError{FCN: utils.SellerW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Registration Date can not be empty")}
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

	objData := models.SellerBase{
		SellerID:   sellerID, //args[0],
		UserID:     userID,   //args[1],
		SellerType: args[2],
		Details:    args[3],
		RegDate:    args[4],
		Created:    created,
		Createdby:  callerID,
	}

	obj := &models.Seller{ObjectType: utils.SELLER, SellerID: args[0], Data: objData}
	return obj.PutState(stub)
}

//SellerR reads state from blockchain
func SellerR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.SellerR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Seller ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.SellerR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Seller ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.Seller{SellerID: args[0]}
	return obj.GetState(stub)
}
