/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"contract/klscontract/chaincode"
)

func main() {
	klsChaincode, err := contractapi.NewChaincode(&chaincode.KLSContract{})
	if err != nil {
		log.Panicf("Error creating klscontract: %v", err)
	}

	if err := klsChaincode.Start(); err != nil {
		log.Panicf("Error starting klscontract: %v", err)
	}
}
