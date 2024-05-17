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

//LoanRatingW writes state to blockchain
func LoanRatingW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 4")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Rating ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	ratingID, ratingIDErr := strconv.Atoi(args[0])
	if ratingIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODENOTALLWD, ERR: ratingIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Loan ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	loanID, loanIDErr := strconv.Atoi(args[1])
	if loanIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODENOTALLWD, ERR: loanIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Rating can not be empty")}
		return shim.Error(cErr.Error())
	}
	rating, ratingErr := strconv.ParseFloat(args[2], 64)
	if ratingErr != nil {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODENOTALLWD, ERR: ratingErr}
		return shim.Error(cErr.Error())
	}

	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.LoanRatingW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Rating Desc can not be empty")}
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

	objData := models.LoanRatingBase{
		RatingID:   ratingID, //args[0],
		LoanID:     loanID,   //args[1],
		Rating:     rating,   //args[2],
		RatingDesc: args[3],
		Created:    created,
		Createdby:  callerID,
	}

	obj := &models.LoanRating{ObjectType: utils.LONRAT, RatingID: args[0], Data: objData}
	return obj.PutState(stub)
}

//LoanRatingR reads state from blockchain
func LoanRatingR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.LoanRatingR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Rating ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.LoanRatingR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Rating ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.LoanRating{RatingID: args[0]}
	return obj.GetState(stub)
}
