package main

import (

	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"feedFactory/chaincode"

)

func main() {
	
	chaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error createing cryptoFungi chaincode: %v", err)
	}
	err = chaincode.Start()
	if err != nil {
		log.Panicf("Error starting cryptoFungi chaincode: %v", err)
	}
	
}