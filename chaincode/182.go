/*
SPDX-License-Identifier: MIT
*/

package main

import (
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/vidalpaul/golang-chaincodes/tokens/erc20/chaincode"
)

func main() {
	tokenChaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		log.Panicf("Error creating ERC20 token chaincode: %v", err)
	}

	if err := tokenChaincode.Start(); err != nil {
		log.Panicf("Error starting ERC20 token chaincode: %v", err)
	}
}
