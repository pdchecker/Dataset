package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (doc *DeliveryOrderContract) RequestDO(ctx contractapi.TransactionContextInterface, requestDetailsJSON string) error {
    var requestDetails RequestDetails
    err := json.Unmarshal([]byte(requestDetailsJSON), &requestDetails)
    if err != nil {
        return fmt.Errorf("Failed to parse the JSON string: %v", err)
    }

    // Check if a request with the given ID already exists
    existingRequestJSON, err := ctx.GetStub().GetState(requestDetails.ID)
    if err != nil {
        return fmt.Errorf("Failed to read from world state: %v", err)
    }
    if existingRequestJSON != nil {
        return fmt.Errorf("A request with ID %s already exists", requestDetails.ID)
    }

    requestDetails.Status = "request"
    requestDetailsAsBytes, _ := json.Marshal(requestDetails)
    return ctx.GetStub().PutState(requestDetails.ID, requestDetailsAsBytes)
}


func (doc *DeliveryOrderContract) UpdateDO(ctx contractapi.TransactionContextInterface, requestDetailsJSON string) error {
    var requestDetails RequestDetails
    err := json.Unmarshal([]byte(requestDetailsJSON), &requestDetails)
    if err != nil {
        return fmt.Errorf("Failed to parse the JSON string: %v", err)
    }

    // Check if a request with the given ID exists
    existingRequestJSON, err := ctx.GetStub().GetState(requestDetails.ID)
    if err != nil {
        return fmt.Errorf("Failed to read from world state: %v", err)
    }
    if existingRequestJSON == nil {
        return fmt.Errorf("No request with ID %s exists", requestDetails.ID)
    }

    // Update the request details
    requestDetailsAsBytes, _ := json.Marshal(requestDetails)
    return ctx.GetStub().PutState(requestDetails.ID, requestDetailsAsBytes)
}

