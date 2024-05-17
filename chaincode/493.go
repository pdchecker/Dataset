package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ... [Your Other Struct Definitions Here] ...

func (pc *PalmOilContract) QueryCommodityByID(ctx contractapi.TransactionContextInterface, commodityID string) (*Commodity, error) {
	commodityJSON, err := ctx.GetStub().GetState(commodityID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if commodityJSON == nil {
		return nil, fmt.Errorf("the commodity with ID %s does not exist", commodityID)
	}

	var commodity Commodity
	err = json.Unmarshal(commodityJSON, &commodity)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal commodity JSON: %v", err)
	}

	return &commodity, nil
}

// QueryAllCommodities retrieves all commodities from the ledger
func (pc *PalmOilContract) QueryAllCommodities(ctx contractapi.TransactionContextInterface) ([]*Commodity, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("COM_", "COM_zzzzzzzzzz")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var commodities []*Commodity
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var commodity Commodity
		err = json.Unmarshal(queryResponse.Value, &commodity)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal commodity JSON: %v", err)
		}

		commodities = append(commodities, &commodity)
	}

	return commodities, nil
}

// QueryAllProcessedCommodities retrieves all processed commodities from the ledger
func (pc *PalmOilContract) QueryAllProcessedCommodities(ctx contractapi.TransactionContextInterface) ([]*ProcessedCommodity, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("PCD_", "PCD_zzzzzzzzzz")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var processedCommodities []*ProcessedCommodity
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var processedCommodity ProcessedCommodity
		json.Unmarshal(queryResponse.Value, &processedCommodity)
		processedCommodities = append(processedCommodities, &processedCommodity)
	}

	return processedCommodities, nil
}