package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) AddTwoNumbers(ctx contractapi.TransactionContextInterface, num1 int, num2 int) (int, error) {

	sum := num1 + num2
	strNum := strconv.Itoa(sum)
	ctx.GetStub().PutState("Addition_res", []byte(strNum))
	return sum, nil
}

func (s *SmartContract) getSummationResult(ctx contractapi.TransactionContextInterface) (int, error) {
	sum, err := ctx.GetStub().GetState("Addition_res")
	if err != nil {
		return 0, err
	}
	sumInt, err := strconv.Atoi(string(sum))
	if err != nil {
		return 0, err
	}
	return sumInt, nil
}
