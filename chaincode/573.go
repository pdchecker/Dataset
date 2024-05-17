package main

import (
	"github.com/xiebei1108/hyperledger-fabric-samples/financial-tracebility/chaincode-go/chaincode"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	contract := new(chaincode.Contract)
	contract.TransactionContextHandler = new(chaincode.TransactionContext)
	contract.Name = "org.example.finan-trace"
	contract.Info.Version = "0.0.1"

	contractChaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(fmt.Sprintf("Error creating contractChaincode. %s", err.Error()))
	}

	contractChaincode.Info.Title = "FinancialTraceabilityPaperChaincode"
	contractChaincode.Info.Version = "0.0.1"

	err = contractChaincode.Start()

	if err != nil {
		panic(fmt.Sprintf("Error starting contractChaincode. %s", err.Error()))
	}
}