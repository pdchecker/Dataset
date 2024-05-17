package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const index = "state~name"

type SimpleChaincode struct {
	contractapi.Contract
}

type Asset struct {
	DocType  string `json:"docType"`
	ID       string `json:"ID"`
	REC      int    `json:"REC"`
	KRW      int    `json: "KRW"`
	Role     string `json: "role"`
}

type HistoryQueryResult struct {
	Record    *Asset    `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

type PaginatedQueryResult struct {
	Records             []*Asset `json:"records"`
	FetchedRecordsCount int32    `json:"fetchedRecordsCount"`
	Bookmark            string   `json:"bookmark"`
}

func (t *SimpleChaincode) CreateAsset(ctx contractapi.TransactionContextInterface, assetID string, REC int, KRW int, role string) error {
	exists, err := t.AssetExists(ctx, assetID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	}
	if exists {
		return fmt.Errorf("asset already exists: %s", assetID)
	}

	asset := &Asset{
		DocType:  "asset",  //아이디  - whwh
		ID:       assetID,  //키
		REC:      REC,      //보유 REC 개수
		KRW:      KRW,      //보유 원
		Role:     role,     //buyer , seller 저장
	}
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(assetID, assetBytes)
	if err != nil {
		return err
	}

	stateNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{asset.Role, asset.ID})
	if err != nil {
		return err
	}

	value := []byte{0x00}
	return ctx.GetStub().PutState(stateNameIndexKey, value)
}

// ReadAsset retrieves an asset from the ledger
func (t *SimpleChaincode) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", assetID, err)
	}
	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", assetID)
	}

	var asset Asset
	err = json.Unmarshal(assetBytes, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// DeleteAsset removes an asset key-value pair from the ledger
func (t *SimpleChaincode) DeleteAsset(ctx contractapi.TransactionContextInterface, assetID string) error {
	asset, err := t.ReadAsset(ctx, assetID)
	if err != nil {
		return err
	}

	err = ctx.GetStub().DelState(assetID)
	if err != nil {
		return fmt.Errorf("failed to delete asset %s: %v", assetID, err)
	}

	stateNameIndexKey, err := ctx.GetStub().CreateCompositeKey(index, []string{asset.Role, asset.ID})
	if err != nil {
		return err
	}

	// Delete index entry
	return ctx.GetStub().DelState(stateNameIndexKey)
}

func constructQueryResponseFromIterator(resultsIterator shim.StateQueryIteratorInterface) ([]*Asset, error) {
	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var asset Asset
		err = json.Unmarshal(queryResult.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (t *SimpleChaincode) GetAssetsByRange(ctx contractapi.TransactionContextInterface, startKey, endKey string) ([]*Asset, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	return constructQueryResponseFromIterator(resultsIterator)
}

// 모든 거래기록 반환
// 삭제된것 : isDeleted = true
// 모든 수정내역을 저장한다.
func (t *SimpleChaincode) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: ID %v", assetID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
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
			asset = Asset{
				ID: assetID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

func (t *SimpleChaincode) AssetExists(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return false, fmt.Errorf("failed to read asset %s from world state. %v", assetID, err)
	}

	return assetBytes != nil, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SimpleChaincode{})
	if err != nil {
		log.Panicf("Error creating asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}

func (t *SimpleChaincode) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (t *SimpleChaincode) UpdateAsset(ctx contractapi.TransactionContextInterface, assetID string, REC int, KRW int, role string) error {
	exists, err := t.AssetExists(ctx, assetID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", assetID)
	}

	asset := Asset{
		DocType:  "asset",
		ID:       assetID,
		REC:      REC,
		KRW:      KRW,
		Role:     role,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, assetJSON)
}
