package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

func Bridge(ctx contractapi.TransactionContextInterface, fcName, toWallet, amount, msg, signature, swapId string) (bool, error) {

	params := []string{fcName, toWallet, amount, msg, signature, swapId}
	queryArgs := make([][]byte, len(params))
	for i, args := range params {
		queryArgs[i] = []byte(args)
	}

	res := ctx.GetStub().InvokeChaincode("conx", queryArgs, "mychannel")
	if res.Status != shim.OK {
		return false, fmt.Errorf("error occured while invoke %s", res.Message)
	}

	return true, nil
}
