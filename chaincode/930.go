package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"hyperlibrary/common"
	"log"
)

func GetQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([][]byte, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([][]byte, error) {
	var assets [][]byte
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		assets = append(assets, queryResult.Value)
	}

	return assets, nil
}

func (t *SmartContract) GetHistory(ctx contractapi.TransactionContextInterface, id string) ([]common.History, error) {
	hi, err := ctx.GetStub().GetHistoryForKey(id)

	if err != nil {
		return []common.History{}, err
	}

	var entries []common.History

	for hi.HasNext() {
		h, err := hi.Next()

		if err != nil {
			return []common.History{}, err
		}

		date := common.GetApproxTime(h.Timestamp)
		desc := h.GetValue()
		log.Println(date, desc)

		var data map[string]interface{}
		err = json.Unmarshal(desc, &data)

		if err != nil {
			return []common.History{}, err
		}

		entries = append(entries, common.History{date, data, h.IsDelete})
	}

	return entries, nil
}
