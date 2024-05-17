package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	studentChaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error creating student chaincode: %v", err)
	}

	if err := studentChaincode.Start(); err != nil {
		log.Panicf("Error starting student chaincode: %v", err)
	}
}
