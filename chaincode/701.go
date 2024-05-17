/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"contract/mkcontract/chaincode"
)

func main() {
	mkChaincode, err := contractapi.NewChaincode(&chaincode.MKContract{})
	if err != nil {
		log.Panicf("Error creating mkcontract: %v", err)
	}

	if err := mkChaincode.Start(); err != nil {
		log.Panicf("Error starting mkcontract: %v", err)
	}
}
