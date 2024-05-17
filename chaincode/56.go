package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

func (sc *SmartContract) Transfer(
	ctx contractapi.TransactionContextInterface,
	sender, receiver string, amount int64,
) error {
	sbal, err := sc.GetBalance(ctx, sender)
	if err != nil {
		return err
	}
	rbal, err := sc.GetBalance(ctx, receiver)
	if err != nil {
		return err
	}
	if sbal < amount {
		return fmt.Errorf("not enough tokens in sender")
	}
	err = sc.SetBalance(ctx, sender, sbal-amount)
	if err != nil {
		return err
	}
	return sc.SetBalance(ctx, receiver, rbal+amount)
}

func (sc *SmartContract) GetBalance(
	ctx contractapi.TransactionContextInterface, account string,
) (int64, error) {
	b, err := ctx.GetStub().GetState("balance" + account)
	if err != nil {
		return 0, err
	}
	if b == nil {
		return 0, nil
	}
	var value int64
	err = json.Unmarshal(b, &value)
	return value, err
}

func (sc *SmartContract) SetBalance(
	ctx contractapi.TransactionContextInterface, account string, value int64,
) error {
	b, _ := json.Marshal(value)
	return ctx.GetStub().PutState("balance"+account, b)
}
