/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	carContract := new(CarContract)
	orderContract := new(OrderContract)

	chaincode, err := contractapi.NewChaincode(carContract, orderContract)

	if err != nil {
		panic("Could not create chaincode from CarContract." + err.Error())
	}

	err = chaincode.Start()

	if err != nil {
		panic("Failed to start chaincode. " + err.Error())
	}
}
