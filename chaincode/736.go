package main

import (
	"log"
	"chaincode/chaincode"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	deliveryOrderChaincode, err := contractapi.NewChaincode(chaincode.NewDeliveryOrderContract())

	if err != nil {
		log.Panicf("Error creating DeliveryOrderChaincode: %v", err)
	}

	if err := deliveryOrderChaincode.Start(); err != nil {
		log.Panicf("Error starting DeliveryOrderChaincode: %v", err)
	}
}