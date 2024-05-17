/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// import "github.com/hyperledger/fabric-chaincode-go/shim"

func main() {
	err := shim.Start(NewChaincode())
	if err != nil {
		panic(err)
	}

}
