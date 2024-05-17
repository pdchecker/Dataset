package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (t *SimpleChaincode) initAccount(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	// AccountID, Currency, Amount, Status
	err := checkArgArrayLength(args, 4)
	if err != nil {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 4")
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("AccountID must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, fmt.Errorf("Currency must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, fmt.Errorf("Amount must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, fmt.Errorf("Status must be a non-empty string")
	}

	accountId := args[0]
	currency := strings.ToUpper(args[1])
	amount, err := strconv.ParseFloat(args[2], 64)
	status := strings.ToUpper(args[3])
	if err != nil {
		return nil, fmt.Errorf("Amount must be a numeric string")
	}

	accountAsBytes, err := ctx.GetStub().GetState(accountId)
	if err != nil {
		return nil, fmt.Errorf("Failed to get account: " + err.Error())
	} else if accountAsBytes != nil {
		errMsg := fmt.Sprintf(
			"Error: This account already exists (%s)",
			accountId)
		return nil, fmt.Errorf(errMsg)
	}

	account := Account{}
	account.ObjectType = accountObjectType
	account.AccountID = accountId
	account.Currency = currency
	account.Amount = amount
	account.Status = status

	accountAsBytes, err = json.Marshal(account)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(accountId, accountAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return accountAsBytes, nil
}

func (t *SimpleChaincode) getChannelLiquidity(
	ctx contractapi.TransactionContextInterface) ([]byte, error) {

	queryString := fmt.Sprintf(
		`{"selector":{"docType":"%s"}}`,
		accountObjectType)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var totalLiquidity float64
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		jsonByteObj := queryResponse.Value
		account := Account{}
		json.Unmarshal(jsonByteObj, &account)
		totalLiquidity += account.Amount
	}
	totalLiquidityString := strconv.FormatFloat(totalLiquidity, 'f', -1, 64)
	return []byte(totalLiquidityString), nil
}

func (t *SimpleChaincode) updateAccount(
	ctx contractapi.TransactionContextInterface,
	accountID string,
	currency string,
	amount float64,
	status string) ([]byte, error) {

	account, err := getAccountStructFromID(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	} else if account.Currency != currency {
		return nil, fmt.Errorf("Currency provided does not match with currency set by account")
	}

	account.Amount = amount
	account.Status = status
	accountAsBytes, err := json.Marshal(account)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(accountID, accountAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return accountAsBytes, nil
}

func getListOfAccounts(
	ctx contractapi.TransactionContextInterface) ([]string, error) {

	var accountList []string

	queryString := fmt.Sprintf(
		`{"selector":{"docType":"%s"}}`,
		accountObjectType)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return accountList, err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return accountList, err
		}
		jsonByteObj := queryResponse.Value
		account := Account{}
		json.Unmarshal(jsonByteObj, &account)
		accountList = append(accountList, account.AccountID)
	}
	return accountList, nil
}

func updateAccountBalance(
	ctx contractapi.TransactionContextInterface,
	accountID string,
	currency string,
	amount float64,
	isMinus bool) error {

	var err error

	if len(accountID) <= 0 {
		return fmt.Errorf("AccountID must be a non-empty string")
	}
	if len(currency) <= 0 {
		return fmt.Errorf("Currency must be a non-empty string")
	}
	if amount < 0 {
		return fmt.Errorf("Amount must be a positive value")
	}

	account, err := getAccountStructFromID(ctx, accountID)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	if account.Status == "PAUSED" {
		return errors.New("Account Status is : " + account.Status)
	} else if account.Currency != currency {
		errStr := fmt.Sprintf(
			"Currency set for account [%s] does not match currency provided [%s]",
			account.Currency,
			currency)
		return errors.New(errStr)
	}

	if isMinus {
		if amount > account.Amount {
			return fmt.Errorf("Amount to be deducted from account cannot exceed account balance")
		} else {
			account.Amount -= amount
		}
	} else {
		account.Amount += amount
	}

	UpdatedAcctAsBytes, err := json.Marshal(account)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(accountID, UpdatedAcctAsBytes)
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	return nil
}

func (t *SimpleChaincode) deleteAccount(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	err := checkArgArrayLength(args, 1)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	accountId := args[0]

	// Access Control
	err = verifyIdentity(ctx, regulatorName)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	valAsbytes, err := ctx.GetStub().GetState(accountId)
	if err != nil {
		errMsg := fmt.Sprintf(
			"Error: Failed to get state for account (%s)",
			accountId)
		return nil, fmt.Errorf(errMsg)
	} else if valAsbytes == nil {
		errMsg := fmt.Sprintf(
			"Error: Account does not exist (%s)",
			accountId)
		return nil, fmt.Errorf(errMsg)
	}

	err = ctx.GetStub().DelState(accountId)
	if err != nil {
		return nil, fmt.Errorf("Failed to delete state")
	}
	return valAsbytes, nil
}

func (t *SimpleChaincode) updateAccountStatus(
	ctx contractapi.TransactionContextInterface,
	args []string,
	status string) ([]byte, error) {

	err := checkArgArrayLength(args, 1)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("AccountID must be a non-empty string")
	}

	accountID := args[0]
	account, err := getAccountStructFromID(ctx, accountID)
	account.Status = status

	accountAsBytes, err := json.Marshal(account)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	err = ctx.GetStub().PutState(accountID, accountAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	return accountAsBytes, nil
}
