package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	// "github.com/tuhoag/promark/chaincodes/proof/chaincode"
)

func main() {
	proofChaincode, err := contractapi.NewChaincode(&ProofSmartContract{})
	if err != nil {
		log.Panicf("Error creating proof chaincode: %v", err)
	}

	if err := proofChaincode.Start(); err != nil {
		log.Panicf("Error starting proof chaincode: %v", err)
	}
}
