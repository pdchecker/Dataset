package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const bilateralChaincodeName string = "bilateralchannel_cc"

func main() {

	chaincode, err := contractapi.NewChaincode(new(SimpleChaincode))
	if err != nil {

		fmt.Errorf("Error while creating new chaincode")
		return
	}

	if err := chaincode.Start(); err != nil {

		fmt.Errorf("Error while starting chaincode")
	}
}

func (t *SimpleChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

func (t *SimpleChaincode) Invoke(ctx contractapi.TransactionContextInterface) ([]byte, error) {
	function, args := ctx.GetStub().GetFunctionAndParameters()
	fmt.Println("invoke MultilateralChannel is running " + function)

	if function == "pingChaincode" {
		return t.pingChaincode(ctx)
	} else if function == "pingChaincodeQuery" {
		return t.pingChaincodeQuery(ctx)

		// Transient fund functions
	} else if function == "createTransientFund" {
		return t.createTransientFund(ctx, args)

		// Other Functions
	} else if function == "getState" {
		return t.getStateAsBytes(ctx, args)
	} else if function == "resetChannel" {
		return t.resetChannel(ctx)
	}

	fmt.Println("MultilateralChannel invoke did not find func: " + function) //error
	return nil, fmt.Errorf("Received unknown function")
}

func (t *SimpleChaincode) pingChaincode(ctx contractapi.TransactionContextInterface) ([]byte, error) {
	pingChaincodeAsBytes, err := ctx.GetStub().GetState("pingchaincode")
	if err != nil {
		jsonResp := "Error: Failed to get state for pingchaincode"
		return nil, fmt.Errorf(jsonResp)
	} else if pingChaincodeAsBytes == nil {
		pingChaincode := PingChaincode{"pingchaincode", 1}
		pingChaincodeAsBytes, err = json.Marshal(pingChaincode)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

		err = ctx.GetStub().PutState("pingchaincode", pingChaincodeAsBytes)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
	} else {
		pingChaincode := &PingChaincode{}
		err = json.Unmarshal([]byte(pingChaincodeAsBytes), pingChaincode)
		pingChaincode.Number++
		pingChaincodeAsBytes, err = json.Marshal(pingChaincode)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		err = ctx.GetStub().PutState("pingchaincode", pingChaincodeAsBytes)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
	}
	return pingChaincodeAsBytes, nil

}

func (t *SimpleChaincode) pingChaincodeQuery(
	ctx contractapi.TransactionContextInterface) ([]byte, error) {

	pingChaincodeAsBytes, err := ctx.GetStub().GetState("pingchaincode")
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return pingChaincodeAsBytes, nil
}

func (t *SimpleChaincode) getStateAsBytes(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	err := checkArgArrayLength(args, 1)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	key := args[0]
	valAsbytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	} else if valAsbytes == nil {
		errMsg := fmt.Sprintf("Error: Key does not exist (%s)", key)
		return nil, fmt.Errorf(errMsg)
	}

	return valAsbytes, nil
}

func (t *SimpleChaincode) resetChannel(ctx contractapi.TransactionContextInterface) ([]byte, error) {
	err := resetAllTransientFund(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return nil, nil
}
