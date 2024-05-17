package main

import (
	"log"

	chaincode "github.com/siddhantprateek/tractionx/contracts/chaincode"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	propertyContract := new(chaincode.PropertyContract)
	propertyChaincode, err := contractapi.NewChaincode(propertyContract)
	if err != nil {
		log.Panicf("Error creating property chaincode: %v", err)
	}

	if err := propertyChaincode.Start(); err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
