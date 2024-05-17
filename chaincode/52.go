/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"contract/smscontract/chaincode"
)

func main() {
	smsChaincode, err := contractapi.NewChaincode(&chaincode.SMSContract{})
	if err != nil {
		log.Panicf("Error creating smscontract: %v", err)
	}

	if err := smsChaincode.Start(); err != nil {
		log.Panicf("Error starting smscontract: %v", err)
	}
}
