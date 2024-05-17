/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"contract/ijzcontract/chaincode"
)

func main() {
	ijzChaincode, err := contractapi.NewChaincode(&chaincode.IJZContract{})
	if err != nil {
		log.Panicf("Error creating ijzcontract: %v", err)
	}

	if err := ijzChaincode.Start(); err != nil {
		log.Panicf("Error starting ijzcontract: %v", err)
	}
}
