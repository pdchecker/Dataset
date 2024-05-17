package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (doc *DeliveryOrderContract) QueryRequestDetailByID(ctx contractapi.TransactionContextInterface, id string) (*RequestDetails, error) {
	requestDetailsAsBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, err
	}
	var requestDetails RequestDetails
	json.Unmarshal(requestDetailsAsBytes, &requestDetails)
	return &requestDetails, nil
}

func (doc *DeliveryOrderContract) QueryAllRequestDetails(ctx contractapi.TransactionContextInterface) ([]*RequestDetails, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("DO_", "DO_zzzzzzzzzz")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var requestDetailsList []*RequestDetails
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var requestDetails RequestDetails
		json.Unmarshal(queryResponse.Value, &requestDetails)
		requestDetailsList = append(requestDetailsList, &requestDetails)
	}

	return requestDetailsList, nil
}


