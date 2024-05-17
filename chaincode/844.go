package main

import (
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	//"github.com/hyperledger/fabric-chaincode-go/shim"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct{}

// InitLedger is called during chaincode instantiation to initialize any data
func (t *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	// Get the args from the TX proposal
	args := ctx.GetStub().GetStringArgs()
	if len(args) != 2 {
		return fmt.Errorf("incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := ctx.GetStub().PutState(args[0], []byte(args[1]))
	if err != nil {
		return fmt.Errorf("failed to create asset: %s", args[0])
	}

	return nil
}

// Invoke is called per transaction on the chaincode. Each TX is either a 'get' or a 'set' on the asset
// created by Init function. The 'Set' method may create a new asset by specifying a new key-value pair
//func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
//	// Extract the function and args from the transaction proposal
//	fn, args := stub.GetFunctionAndParameters()
//
//	var result string
//	var err error
//	if fn == "set" {
//		result, err = set(stub, args)
//	} else {
//		result, err = get(stub, args)
//	}
//
//	if err != nil {
//		return shim.Error(err.Error())
//	}
//
//	// Return the result as success payload
//	return shim.Success([]byte(result))
//}

// setAsset stores the asset (both key and value) on the ledger. If the key exists, it will override the value with the new one
func (t *SmartContract) setAsset(ctx contractapi.TransactionContextInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("incorrect arguments. Expecting a key and a value")
	}

	err := ctx.GetStub().PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("failed to set asset: %s", args[0])
	}

	return args[1], nil
}

// getAsset returns the value of the specified asset key
func (t *SmartContract) getAsset(ctx contractapi.TransactionContextInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a  key")
	}

	value, err := ctx.GetStub().GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("failed to get asset: %s with error: %s", args[0], err)
	}

	if value == nil {
		return "", fmt.Errorf("asset not found: %s", args[0])
	}

	return string(value), nil
}

// main function starts up the chaincode in the container during instantiate
//func main() {
//	if err := shim.Start(new(SimpleAsset)); err != nil {
//		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
//	}
//}
