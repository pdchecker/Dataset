/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/Subskribo-BV/dnn-fabric-chaincode/service"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	cc, err := contractapi.NewChaincode(&service.SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := cc.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
