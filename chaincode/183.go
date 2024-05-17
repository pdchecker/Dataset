package main

import (
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	. "github.com/newHouseSale/ChainCode"
)

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create houseSale chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting houseSale chaincode: %s", err.Error())
	}
}
