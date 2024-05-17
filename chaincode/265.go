/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"contract/pdcontract/chaincode"
)

func main() {
	pdChaincode, err := contractapi.NewChaincode(&chaincode.PDContract{})
	if err != nil {
		log.Panicf("Error creating pdcontract: %v", err)
	}

	if err := pdChaincode.Start(); err != nil {
		log.Panicf("Error starting pdcontract: %v", err)
	}
}
