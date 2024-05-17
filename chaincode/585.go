/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"contract/tskcontract/chaincode"
)

func main() {
	tskChaincode, err := contractapi.NewChaincode(&chaincode.TSKContract{})
	if err != nil {
		log.Panicf("Error creating tskcontract: %v", err)
	}

	if err := tskChaincode.Start(); err != nil {
		log.Panicf("Error starting tskcontract: %v", err)
	}
}
