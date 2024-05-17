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

//TransactionW writes state to blockchain
func TransactionW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 11 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 11")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Transaction ID ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	txnID, txnIDErr := strconv.Atoi(args[0])
	if txnIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODENOTALLWD, ERR: txnIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Transaction Date can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Buyer ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	buyerID, buyerIDErr := strconv.Atoi(args[2])
	if buyerIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODENOTALLWD, ERR: buyerIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("User ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	userID, userIDErr := strconv.Atoi(args[3])
	if userIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODENOTALLWD, ERR: userIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[4]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Repayment can not be empty")}
		return shim.Error(cErr.Error())
	}
	repayment, repaymentErr := strconv.ParseFloat(args[4], 64)
	if repaymentErr != nil {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODENOTALLWD, ERR: repaymentErr}
		return shim.Error(cErr.Error())
	}

	if len(args[5]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Amount can not be empty")}
		return shim.Error(cErr.Error())
	}
	amount, amountErr := strconv.ParseFloat(args[5], 64)
	if amountErr != nil {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODENOTALLWD, ERR: amountErr}
		return shim.Error(cErr.Error())
	}

	if len(args[6]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("InterestRate can not be empty")}
		return shim.Error(cErr.Error())
	}
	interestrate, interestrateErr := strconv.ParseFloat(args[6], 64)
	if interestrateErr != nil {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODENOTALLWD, ERR: interestrateErr}
		return shim.Error(cErr.Error())
	}

	if len(args[7]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Outstanding can not be empty")}
		return shim.Error(cErr.Error())
	}
	outstanding, outstandingErr := strconv.ParseFloat(args[7], 64)
	if outstandingErr != nil {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODENOTALLWD, ERR: outstandingErr}
		return shim.Error(cErr.Error())
	}

	if len(args[8]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("DueDate can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[9]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Bank can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[10]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("LoanStatus can not be empty")}
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

	objData := models.TransactionBase{
		TxnID:        txnID, //args[0],
		TxnDate:      args[1],
		BuyerID:      buyerID,      //args[2],
		UserID:       userID,       //args[3],
		Repayment:    repayment,    //args[4],
		Amount:       amount,       //args[5],
		InterestRate: interestrate, //args[6],
		Outstanding:  outstanding,  //args[7],
		DueDate:      args[8],
		Bank:         args[9],
		LoanStatus:   args[10],
		Created:      created,
		Createdby:    callerID,
	}

	obj := &models.Transactions{ObjectType: utils.TRNSAC, TxnID: args[0], Data: objData}
	return obj.PutState(stub)
}

//TransactionR reads state from blockchain
func TransactionR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.TransactionR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Transaction ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.TransactionR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Transaction ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.Transactions{TxnID: args[0]}
	return obj.GetState(stub)
}
