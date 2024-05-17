package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	// "github.com/tuhoag/promark/src/chaincodes/campaign/chaincode"
)

func main() {
	campaignChaincode, err := contractapi.NewChaincode(&CampaignSmartContract{})
	if err != nil {
		log.Panicf("Error creating campaign chaincode: %v", err)
	}

	if err := campaignChaincode.Start(); err != nil {
		log.Panicf("Error starting campaign chaincode: %v", err)
	}
}
