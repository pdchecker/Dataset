package main

import (
	"fmt"

	"github.com/bridge/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {

	bridgeContract, err := contractapi.NewChaincode(&chaincode.SmartContract{})
	if err != nil {
		fmt.Println(fmt.Sprintf("Error init smart contract %s", err))
	}
	if err := bridgeContract.Start(); err != nil {
		fmt.Println(fmt.Sprintf("Error while starting SmartContract Bridge %s", err))
	}

}

// mint
// withdraw
//
