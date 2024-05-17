/*
*
*		Copyright CONUN KOREA. ALL Rights Reserved
*
 */

package main

import (
	"log"

	"github.com/conun/wrapchain/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	wrapchain, err := contractapi.NewChaincode(&chaincode.SmartContract{})

	if err != nil {
		log.Panicf("Error creating wrapchain chaincode: %v", err)
	}

	if err := wrapchain.Start(); err != nil {
		log.Panicf("Error starting wrapchain chaincode %v", err)
	}
}
