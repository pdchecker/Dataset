package main

import (
	"simple-payment-application-chaincode/contracts"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	simplePaymentApp := new(contracts.SimplePaymentApp)

	cc, err := contractapi.NewChaincode(simplePaymentApp)

	if err != nil {
		panic(err.Error())
	}

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
