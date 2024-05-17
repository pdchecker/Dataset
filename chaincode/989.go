package main

import (
	farm "farm/contract"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	contract := new(farm.FarmContract)
	contract.Name = "farm_contract"
	contract.Info.Version = "0.0.1"

	chaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode. %s", err.Error()))
	}

	chaincode.Info.Title = "FarmChaincode"
	chaincode.Info.Version = "0.0.1"

	err = chaincode.Start()

	if err != nil {
		panic(fmt.Sprintf("Error starting chaincode. %s", err.Error()))
	}
}
