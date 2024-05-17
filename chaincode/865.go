package main

import (
	"asset"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	cc, err := contractapi.NewChaincode(&asset.SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := cc.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
