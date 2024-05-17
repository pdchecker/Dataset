package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"tallysolutions.com/owner/chaincode-go/chaincode"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating owner chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting owner chaincode: %v", err)
	}
}
