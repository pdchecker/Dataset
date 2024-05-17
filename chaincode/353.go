package main

import (
	"fmt"
	"github.com/YGrylls/sourcePlatform/contract/process"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	contract := new(process.Contract)
	contract.TransactionContextHandler = new(process.TransactionContext)
	contract.Name = Name
	contract.Info.Version = Version

	chaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode. %s", err.Error()))
	}

	chaincode.Info.Title = "ProcessChaincode"
	chaincode.Info.Version = Version

	err = chaincode.Start()
	if err != nil {
		panic(fmt.Sprintf("Error starting chaincode. %s", err.Error()))
	}
}
