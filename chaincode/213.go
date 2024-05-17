package main

import (
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"nft-ticket/chaincode"
)

func main() {
	assetChaincode, err := contractapi.NewChaincode(&chaincode.NFTTicketChaincode{})
	if err != nil {
		fmt.Printf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		fmt.Printf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
