package main

import (
	"encoding/json"
	"fmt"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type DataHash struct {
	Type string `json:"Type"`
	Hash string `json:"Hash"`
}

// Define the Smart Contract structure
type SmartContract struct {
	accountContract  *AccountContract
	dataContract     *DataContract
	transferContract *TransferContract
}

func NewSmartContract() *SmartContract {
	return &SmartContract{
		accountContract:  &AccountContract{},
		dataContract:     &DataContract{},
		transferContract: &TransferContract{},
	}
}

func (s *SmartContract) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// parse user cert test
	if hashByte := GetCreatorAddress(stub); hashByte != nil {
		fmt.Printf("creator address is %s \n", string(hashByte))
	}
	return shim.Success(nil)
}

func (s *SmartContract) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	switch function {
	// chaincode install support
	case "query":
		valBytes, _ := json.Marshal(100)
		return shim.Success(valBytes)
	case "invoke":
		return shim.Success(nil)
	// account manager
	case "createAccount":
		return s.accountContract.createAccount(stub, args)
	case "showAccount":
		return s.accountContract.showAccount(stub, args)
	case "frozenAccount":
		return s.accountContract.frozenAccount(stub, args)
	case "deleteAccount":
		return s.accountContract.deleteAccount(stub, args)
	case "mintToken":
		return s.accountContract.mintToken(stub, args)
	case "changeSecret":
		return s.accountContract.changeSecret(stub, args)
	// data manager
	case "setDataEvidence":
		return s.dataContract.setDataEvidence(stub, args)
	case "showDataEvidence":
		return s.dataContract.showDataEvidence(stub, args)
	case "setTitle":
		return s.dataContract.setTitle(stub, args)
	case "showTitles":
		return s.dataContract.showTitles(stub, args)
	case "showNameOfTitles":
		return s.dataContract.showNameOfTitles(stub, args)
	case "searchTitles":
		return s.dataContract.searchTitles(stub, args)
	// data transfer manager
	case "transferData":
		return s.transferContract.transferData(stub, args)
	case "showTransferRecord":
		return s.transferContract.showTransferRecord(stub, args)
	case "checkTransferred":
		return s.transferContract.checkTransferred(stub, args)
	default:
	}

	return shim.Error("Invalid Smart Contract function name.")
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(NewSmartContract())
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s \n", err)
	}
}
