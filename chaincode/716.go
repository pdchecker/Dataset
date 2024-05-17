package main

import (
	"log"

	"vehiclecontract/chaincode"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	//Aqui invocamos nossas funções para o contrato inteligente
	assetChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
