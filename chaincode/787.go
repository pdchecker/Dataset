package main

import (
	land "contract/nalanet"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	contract := new(land.Contract)
	contract.TransactionContextHandler = new(land.TransactionContext)
	contract.Name = "org..nalanet"
	contract.Info.Version = "0.0.1"

	chaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode. %s", err.Error()))
	}

	chaincode.Info.Title = "LandChaincode"
	chaincode.Info.Version = "0.0.1"

	err = chaincode.Start()

	if err != nil {
		panic(fmt.Sprintf("Error starting chaincode. %s", err.Error()))
	}
}
