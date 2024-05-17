/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"contract/npdcontract/chaincode"
)

func main() {
	npdChaincode, err := contractapi.NewChaincode(&chaincode.NPDContract{})
	if err != nil {
		log.Panicf("Error creating npdcontract: %v", err)
	}

	if err := npdChaincode.Start(); err != nil {
		log.Panicf("Error starting npdcontract: %v", err)
	}
}
