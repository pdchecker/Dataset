package main

import (
	"encoding/json"
	"fmt"
	"time"
	"log"
	"github.com/golang/protobuf/ptypes"
	// "strconv"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract 
}

type Asset struct {
	Key string `json:"key"`
	Value float64 `json:"value"`
}

type RangedQueryResult struct {
	FromKey string `json:"fromKey"`
	ToKey string `json:"toKey"`
	Record *Asset
}

type HistoryQueryResult struct {
	Record *Asset `json:"record"`
	TxId string `json:"txId`
	Timestamp time.Time `json:"timestamp"`
	IsDelete bool `json:"isDelete`
}

func (s *SmartContract) Get(ctx contractapi.TransactionContextInterface, key string) (*Asset, error) {
	assetAsBytes, err := ctx.GetStub().GetState(key)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from SimpleAsset world state. %s", err.Error())
	}

	if assetAsBytes == nil {
		return nil, fmt.Errorf("Asset Key %s does not exist", key)
	}

	asset :=new(Asset)
	_ = json.Unmarshal(assetAsBytes, asset)

	return asset, nil
}

func (s *SmartContract) Set(ctx contractapi.TransactionContextInterface, key string, value float64) error {

	asset := Asset {
		Key: key,
		Value: value,
	}
	assetAsBytes, _ := json.Marshal(asset)

	return ctx.GetStub().PutState(key, assetAsBytes)
}
/*
	1. History method
	2. Transfer method
	3. Main function
*/

func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, from string, to string, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("Incorrect Transfering amount. Must be more than ZERO")
	}
	fromAsset, err := s.Get(ctx, from)
	if err != nil {
		return fmt.Errorf("Failed to get senders's state. %s", err.Error())
	}

	if fromAsset.Value < amount {
		return fmt.Errorf("insufficient money from sender")
	}

	toAsset, err := s.Get(ctx, to)
	if err != nil {
		return fmt.Errorf("Failed to get reciever's state. %s", err.Error())
	}

	fromAsset.Value -= amount
	toAsset.Value += amount
	fromAsBytes, _ := json.Marshal(fromAsset)
	toAsBytes, _ := json.Marshal(toAsset)

	ctx.GetStub().PutState(from, fromAsBytes)
	ctx.GetStub().PutState(to, toAsBytes)

	return nil
}

func (s* SmartContract) History(ctx contractapi.TransactionContextInterface, key string) ([]HistoryQueryResult, error) {
	log.Printf("Getting History For %s", key)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(key)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Asset{}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult {
			TxId: response.TxId,
			Timestamp: timestamp,
			Record: &asset,
			IsDelete: response.IsDelete,
		}

		records = append(records, record)
	}
	return records, nil
}

func (s* SmartContract)GetKeyRange(ctx contractapi. TransactionContextInterface) ([]Asset, error) {
	var startKey string = ""
	var endKey string= ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err !=  nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []Asset{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}
//TODO : get query results
		asset := new(Asset)
		_ = json.Unmarshal(queryResponse.Value, asset)
		asset.Key = queryResponse.Key

		// queriedAsset := Asset {
		// 	Key: queryResponse.Key,
		// 	Value: strconv.ParseFloat(queryResponse.Value, 64),
		// }

		results = append(results, *asset)
	}
	return results, nil
}


func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create SimpleAsset chaincode: %s", err.Error())
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err.Error())
	}
}