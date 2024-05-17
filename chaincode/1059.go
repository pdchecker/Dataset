/**
 * 
 * Copyright (c) 2022, Oracle and/or its affiliates. All rights reserved.
 * 
 */
package main

import (
	"fmt"
	"sample.com/LoyaltyToken/lib/chaincode"
	"sample.com/LoyaltyToken/lib/util"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func main() {
	util.ChaincodeName = "LoyaltyToken"
	err := shim.Start(new(chaincode.ChainCode))
	if err != nil {
		fmt.Printf("Error starting chaincode: %s", err)
	}
}
