package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"jarvispowered.com/color_rpg/chain/chaincode"
)

func main() {
	crpgChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating crpg chaincode: %v", err)
	}

	if err := crpgChaincode.Start(); err != nil {
		log.Panicf("Error starting crpg chaincode: %v", err)
	}
}
