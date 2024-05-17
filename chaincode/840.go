package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type ECAsset struct {
	ID               string `json:"ID"`
	EndorsementCount int    `json:"EndorsementCount"`
}

type StatusAsset struct {
	ID     string `json:"ID"`
	Status int    `json:"Status"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

func (s *SmartContract) ReadEC(ctx contractapi.TransactionContextInterface, id string) (*ECAsset, error) {
	resultIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("id~ec~txID", []string{id})
	if err != nil {
		return nil, fmt.Errorf("Could not retrieve value for %s: %s", id, err.Error())
	}
	defer resultIterator.Close()

	// Check the variable existed
	if !resultIterator.HasNext() {
		return nil, fmt.Errorf("No variable by the id %s exists", id)
	}

	// Iterate through result set and compute final ec
	ec := 0
	for i := 0; resultIterator.HasNext(); i++ {
		// Get the next row
		responseRange, err := resultIterator.Next()
		if err != nil {
			return nil, err
		}

		// Split the composite key into its component parts
		_, keyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.GetKey())
		if err != nil {
			return nil, err
		}

		// Retrieve the delta value and operation
		curEcStr := keyParts[1]

		// Convert the value string and perform the operation
		curEc, err := strconv.Atoi(curEcStr)
		if err != nil {
			return nil, err
		}
		ec += curEc
	}

	asset := ECAsset{
		ID:               id,
		EndorsementCount: ec,
	}

	return &asset, nil
}

func (s *SmartContract) EndorsementInc(ctx contractapi.TransactionContextInterface, id string) error {
	// Retrieve info needed for the update procedure
	txid := ctx.GetStub().GetTxID()
	compositeIndexName := "id~ec~txID"

	// Create the composite key that will allow us to query for all deltas on a particular variable
	comopositeKey, err := ctx.GetStub().CreateCompositeKey(compositeIndexName, []string{id, "1", txid})
	if err != nil {
		return fmt.Errorf("Could not create a composite key for %s: %s", id, err.Error())
	}

	// Save the composite key index
	err = ctx.GetStub().PutState(comopositeKey, []byte{0x00})
	if err != nil {
		return fmt.Errorf("Could not put operation for %s in the ledger: %s", id, err.Error())
	}

	return nil
}

//func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
//	exists, err := s.AssetExists(ctx, id)
//	if err != nil {
//		return err
//	}
//	if !exists {
//		return fmt.Errorf("the asset %s does not exist", id)
//	}
//
//	return ctx.GetStub().DelState(id)
//}

func (s *SmartContract) StatusAssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) SetStatus(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.StatusAssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("The StatusAsset %s existed", id)
	}

	asset := StatusAsset{
		ID:     id,
		Status: 1,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]string, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		results = append(results, string(queryResponse.Value))
	}

	return results, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}
