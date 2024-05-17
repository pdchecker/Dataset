package main

import (
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"
	"github.com/tokenERC20/util"
)

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	contractapi.Contract
}

// AddToken
func AddToken(ctx contractapi.TransactionContextInterface, amount, wallet string) error {
	var err error
	var toCoin, addAmount decimal.Decimal

	if addAmount, err = util.ParsePositive(amount); err != nil {
		return fmt.Errorf("%s is not postive integer", amount)
	}

	toCoin = decimal.Zero

	currentBalanceBytes, err := ctx.GetStub().GetState(wallet)
	if err != nil {
		return fmt.Errorf("Failed to read wallet account %s from world state ", wallet)
	}
	var currentBalance string
	if currentBalanceBytes == nil {
		currentBalance = "0"
	} else {
		currentBalance = string(currentBalanceBytes)
	}

	toCoin, _ = decimal.NewFromString(currentBalance)
	toCoin = toCoin.Add(addAmount).Truncate(0)

	if err = ctx.GetStub().PutState(wallet, []byte(toCoin.String())); err != nil {
		return fmt.Errorf("error occured while saving to world state %s", err)
	}
	return nil
}

// sub Tokens
func SubstractToken(ctx contractapi.TransactionContextInterface, amount, wallet string) error {
	var err error
	var substractAmount, fromCoin decimal.Decimal

	if substractAmount, err = util.ParsePositive(amount); err != nil {
		return fmt.Errorf("Amount must be an integer string")
	}

	currentBalanceBytes, err := ctx.GetStub().GetState(wallet)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s", wallet)
	}
	if currentBalanceBytes == nil {
		return fmt.Errorf("The balance does not exists")
	}

	fromCoin, _ = decimal.NewFromString(string(currentBalanceBytes))

	if fromCoin.Cmp(substractAmount) < 0 {
		return fmt.Errorf("Amount is bigger than current balance")
	} else {
		if err = ctx.GetStub().PutState(wallet, []byte(fromCoin.Sub(substractAmount).String())); err != nil {
			return fmt.Errorf("error occured while saving to world state %s", err)
		}
	}
	return nil
}

// transfer tokens helper
func MoveToken(ctx contractapi.TransactionContextInterface, fromwallet, towallet, amount string) error {
	var err error
	var substractAmount, fromCoin, toCoin, addAmount decimal.Decimal

	if substractAmount, err = util.ParsePositive(amount); err != nil {
		return fmt.Errorf("Error occured while Parsing Anmount mustbe integer %s", substractAmount)
	}

	addAmount = substractAmount

	fromCurrentBalanceBytes, err := ctx.GetStub().GetState(fromwallet)
	if err != nil {
		return err
	}
	fromCoin, _ = decimal.NewFromString(string(fromCurrentBalanceBytes))
	if fromCoin.Cmp(substractAmount) < 0 {
		return fmt.Errorf("Amount is bigger than expected")
	}
	if err = ctx.GetStub().PutState(fromwallet, []byte(fromCoin.Sub(substractAmount).String())); err != nil {
		return fmt.Errorf("Error occured while saving to DB")
	}

	toCoin = decimal.Zero

	toCurrentBalanceBytes, err := ctx.GetStub().GetState(towallet)
	if err != nil {
		return err
	}

	toCoin, _ = decimal.NewFromString(string(toCurrentBalanceBytes))
	toCoin = toCoin.Add(addAmount).Truncate(0)
	if err = ctx.GetStub().PutState(towallet, []byte(toCoin.String())); err != nil {
		return err
	}
	return nil
}
