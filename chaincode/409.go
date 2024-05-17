package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

//Contract for handling writing and reading from the world state
type Contract struct {
	contractapi.Contract
}

func main() {

	contract := new(Contract)

	cc, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
