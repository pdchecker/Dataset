/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/srj15/Hyperledger-Fabric-Go/chaincode-go/chaincode"
)

func main() {
	tranChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating transaction chaincode: %v", err)
	}

	if err := tranChaincode.Start(); err != nil {
		log.Panicf("Error starting transaction chaincode: %v", err)
	}
}
