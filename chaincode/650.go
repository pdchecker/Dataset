/*
*
* 	Copyright CONUN KOREA. All Rights Reserved
*
 */
package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/tokenERC20/chaincode"
)

func main() {
	tokenChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating token-erc-20 chaincode: %v", err)
	}

	if err := tokenChaincode.Start(); err != nil {
		log.Panicf("Error starting token-erc-20 chaincode: %v", err)
	}
}
