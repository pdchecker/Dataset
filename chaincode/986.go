package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset in the SLA
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of org1 and org2 for SLA.
//Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	Availibility int    `json:"Availibility"`
	ID           string `json:"ID"`
	Wallet       int    `json:"Wallet"`
	Latency      int    `json:"Latency"`
}

// InitContract adds a the agreed parameter of the SLA to the ledger
func (s *SmartContract) InitContract(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Availibility: 99, Wallet: 100, Latency: 99},
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

// AddPracticalAsset issues a new asset2 to the world state with given details.
func (s *SmartContract) AddPracticalAsset(ctx contractapi.TransactionContextInterface, id string, availibility int, wallet int, latency int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:           id,
		Availibility: availibility,
		Wallet:       wallet,
		Latency:      latency,
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
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, availibility int, wallet int, latency int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:           id,
		Availibility: availibility,
		Wallet:       wallet,
		Latency:      latency,
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

// ComputeParametes calculate the eligibility for the refund if any.

func (s *SmartContract) ComputeParameters(ctx contractapi.TransactionContextInterface) (string, error) {
	var id1 string = "asset1"
	asset, err := s.ReadAsset(ctx, id1)
	if err != nil {
		return "", err
	}
	refAvailibility := asset.Availibility

	//praAvaility

	var id2 string = "asset2"
	_, _ = s.ReadAsset(ctx, id2)
	if err != nil {
		return "", err
	}
	praAvailibility := asset.Availibility

	if praAvailibility < refAvailibility {
		fmt.Printf("TransferRefund")
	}
	return "", nil
}

// TransferRefund updates the wallet field of asset2 with given id in world state, and returns the oldWallet.
func (s *SmartContract) TransferRefund(ctx contractapi.TransactionContextInterface, id string, refund int) (int, error) {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return 0, err
	}

	oldWallet := asset.Wallet
	newWallet := oldWallet + refund
	asset.Wallet = newWallet

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return 0, err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return 0, err
	}

	return oldWallet, nil
}
