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

//PropertyW writes state to blockchain
func PropertyW(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 14 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting 14")}
		return shim.Error(cErr.Error())
	}
	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Property ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	proID, proIDErr := strconv.Atoi(args[0])
	if proIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODENOTALLWD, ERR: proIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[1]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Seller ID can not be empty")}
		return shim.Error(cErr.Error())
	}
	sellerID, sellerIDErr := strconv.Atoi(args[1])
	if sellerIDErr != nil {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODENOTALLWD, ERR: sellerIDErr}
		return shim.Error(cErr.Error())
	}

	if len(args[2]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Property Type can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[3]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Property Name can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[4]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Description can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[5]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Address can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[6]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[7]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location Lat can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[8]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Location Long can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[9]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Views can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[10]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Viewers Stats can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[11]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Entry Date can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[12]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Expiry Date can not be empty")}
		return shim.Error(cErr.Error())
	}
	if len(args[13]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyW, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Status can not be empty")}
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

	objData := models.PropertyBase{
		ProID:        proID,    //args[0]
		SellerID:     sellerID, //args[1]
		ProType:      args[2],
		ProName:      args[3],
		Desc:         args[4],
		Address:      args[5],
		Location:     args[6],
		LocationLat:  args[7],
		LocationLong: args[8],
		Views:        args[9],
		ViewerStats:  args[10],
		EntryDate:    args[11],
		ExpiryDate:   args[12],
		Status:       args[13],
		Created:      created,
		Createdby:    callerID,
	}

	obj := &models.Property{ObjectType: utils.PRPRTY, ProID: args[0], Data: objData}
	return obj.PutState(stub)
}

//PropertyR reads state from blockchain
func PropertyR(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		cErr := &utils.ChainError{FCN: utils.PropertyR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Incorrect no of input args, got " + fmt.Sprintf("%v", len(args)) + " expecting Property ID only")}
		return shim.Error(cErr.Error())
	}

	if len(args[0]) == 0 {
		cErr := &utils.ChainError{FCN: utils.PropertyR, KEY: "", CODE: utils.CODEUNPROCESSABLEENTITY, ERR: errors.New("Property ID can not be empty")}
		return shim.Error(cErr.Error())
	}

	obj := &models.Property{ProID: args[0]}
	return obj.GetState(stub)
}
