/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// ReadAsset reads the information from collection
func (s *SmartContract) ReadAsset1(ctx contractapi.TransactionContextInterface, assetID string) (*Asset2, error) {

	log.Printf("ReadAsset: collection %v, ID %v", assetCollection, assetID)
	assetJSON, err := ctx.GetStub().GetPrivateData(assetCollection, assetID) //get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}

	//No Asset found, return empty response
	if assetJSON == nil {
		log.Printf("%v does not exist in collection %v", assetID, assetCollection)
		return nil, nil
	}

	var asset *Asset2
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return asset, nil

}

func (s *SmartContract) ReadLog(ctx contractapi.TransactionContextInterface, assetID string) (*LogAsset, error) {

	log.Printf("ReadAsset: collection %v, ID %v", assetCollection, assetID)
	assetJSON, err := ctx.GetStub().GetPrivateData(assetCollection, assetID) //get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset: %v", err)
	}

	//No Asset found, return empty response
	if assetJSON == nil {
		log.Printf("%v does not exist in collection %v", assetID, assetCollection)
		return nil, nil
	}

	var asset *LogAsset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return asset, nil

}



// GetAssetByRange performs a range query based on the start and end keys provided. Range
// queries can be used to read data from private data collections, but can not be used in
// a transaction that also writes to private data.
func (s *SmartContract) GetAssetByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*Asset2, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(assetCollection, startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []*Asset2{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset *Asset2
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}

		results = append(results, asset)
	}

	return results, nil

}

func (s *SmartContract) GetLogAssetByRange(ctx contractapi.TransactionContextInterface, startKey string, endKey string) ([]*LogAsset, error) {

	resultsIterator, err := ctx.GetStub().GetPrivateDataByRange(assetCollection, startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []*LogAsset{}

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset *LogAsset
		err = json.Unmarshal(response.Value, &asset)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}

		results = append(results, asset)
	}

	return results, nil

}

