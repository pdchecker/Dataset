package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	TId   string
	CId   string
	TTs   float64
	ID    string  `json:"id"`
	Ts    int64   `json:"ts"`
	Sym   string  `json:"sym"`
	Size  float64 `json:"size"`
	Side  string  `json:"side"`
	Price float64 `json:"price"`
	TP    float64 `json:"tp"`
	SL    float64 `json:"sl"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	tts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	ttso, err := strconv.ParseFloat(fmt.Sprintf("%d.%d", tts.Seconds, tts.Nanos), 64)

	assets := []Asset{
		{TId: ctx.GetStub().GetTxID(), CId: ctx.GetStub().GetChannelID(), TTs: ttso, ID: "1", Ts: 1621543917, Sym: "XAU_USD", Size: 1.0, Side: "LONG", Price: 1900.1, TP: 1930.1, SL: 1890.1},
		{TId: ctx.GetStub().GetTxID(), CId: ctx.GetStub().GetChannelID(), TTs: ttso, ID: "2", Ts: 1621543918, Sym: "XAU_USD", Size: 1.0, Side: "SHORT", Price: 1900.2, TP: 1930.2, SL: 1890.2},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, ts int64, sym string, size float64, side string, price float64, tp float64, sl float64) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	tts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	ttso, err := strconv.ParseFloat(fmt.Sprintf("%d.%d", tts.Seconds, tts.Nanos), 64)

	asset := Asset{
		TId:   ctx.GetStub().GetTxID(),
		CId:   ctx.GetStub().GetChannelID(),
		TTs:   ttso,
		ID:    id,
		Ts:    ts,
		Sym:   sym,
 		Size:  size,
		Side:  side,
		Price: price,
		TP:    tp,
		SL:    sl,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, ts int64, sym string, size float64, side string, price float64, tp float64, sl float64) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	tts, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return err
	}
	ttso, err := strconv.ParseFloat(fmt.Sprintf("%d.%d", tts.Seconds, tts.Nanos), 64)

	// overwriting original asset with new asset
	asset := Asset{
		TId:   ctx.GetStub().GetTxID(),
		CId:   ctx.GetStub().GetChannelID(),
		TTs:   ttso,
		ID:    id,
		Ts:    ts,
		Sym:   sym,
		Size:  size,
		Side:  side,
		Price: price,
		TP:    tp,
		SL:    sl,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
//func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newSize float64) error {
//	asset, err := s.ReadAsset(ctx, id)
//	if err != nil {
//		return err
//	}
//
//	asset.Size = newSize
//	assetJSON, err := json.Marshal(asset)
//	if err != nil {
//		return err
//	}
//
//	return ctx.GetStub().PutState(id, assetJSON)
//}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
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