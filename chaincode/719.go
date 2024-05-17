package main

import (
	"log"

	"github.com/chunsik-is-meow/blockchain/src/asset/chaincodes/ai-model/contract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	aiChaincode, err := contractapi.NewChaincode(&contract.AIChaincode{})
	if err != nil {
		log.Panicf("Error creating aiChaincode: %v", err)
	}

	if err := aiChaincode.Start(); err != nil {
		log.Panicf("Error starting aiChaincode: %v", err)
	}
}
