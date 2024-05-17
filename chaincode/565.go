package main

import (
	"github.com/tokenERC1155/chaincode"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main()  {

	smartContract := new(chaincode.SmartContract)

	cc, err := contractapi.NewChaincode(smartContract)

	if err != nil {
		panic(err.Error())
	}
	if err = cc.Start(); err != nil {
		panic(err.Error())
	}
	
}


