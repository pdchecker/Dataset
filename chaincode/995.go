package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-samples/asset-transfer-basic/chaincode-go/chaincode"
)

func main() {
	chaincode, err := contractapi.NewChaincode(&chaincode.SmartContract{})

	if err != nil {
		fmt.Printf("Error create D_Grow chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting D_Grow chaincode: %s", err.Error())
	}

}
