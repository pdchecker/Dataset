package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (t *SimpleChaincode) GetAllBankRequestStruct(
	ctx contractapi.TransactionContextInterface) ([]BankRequest, error) {

	queryString := fmt.Sprintf(
		`{"selector":{"docType":"%s"}}`,
		bankRequestObjectType)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	bankRequestArr := []BankRequest{}
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		bankRequestAsBytes := queryResponse.Value
		bankRequest := BankRequest{}
		json.Unmarshal(bankRequestAsBytes, &bankRequest)
		bankRequestArr = append(bankRequestArr, bankRequest)
	}
	return bankRequestArr, nil
}

func (t SimpleChaincode) createBankRequest(
	ctx contractapi.TransactionContextInterface,
	requestID string,
	bankID string,
	netValue float64,
	nettableList []string,
	nonNettableList []string) (*BankRequest, error) {

	bankRequest := &BankRequest{}
	bankRequest.ObjectType = bankRequestObjectType
	bankRequest.BankRequestID = requestID
	bankRequest.BankID = bankID
	bankRequest.NetValue = netValue
	bankRequest.NettableList = nettableList
	bankRequest.NonNettableList = nonNettableList

	bankRequestAsBytes, err := json.Marshal(bankRequest)
	if err != nil {
		return bankRequest, err
	}
	err = ctx.GetStub().PutState(requestID, bankRequestAsBytes)
	if err != nil {
		return bankRequest, err
	}

	return bankRequest, nil
}

func (t *SimpleChaincode) ResetBankRequests(ctx contractapi.TransactionContextInterface) error {
	bankRequestArr, err := t.GetAllBankRequestStruct(ctx)
	if err != nil {
		return err
	}
	for _, bankRequest := range bankRequestArr {
		err = ctx.GetStub().DelState(bankRequest.BankRequestID)
		if err != nil {
			return err
		}
	}

	return nil
}
