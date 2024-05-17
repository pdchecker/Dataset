package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
//Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
// The following is the original definition of Asset ... by Hanool (04.14, 13:31)
// type Asset struct {
// 	AppraisedValue int    `json:"AppraisedValue"`
// 	Color          string `json:"Color"`
// 	ID             string `json:"ID"`
// 	Owner          string `json:"Owner"`
// 	Size           int    `json:"Size"`
// }

// The follwing is the new definition of Asset ... by Hanool (04.14, 13:31)
type Asset struct {
	Capacity  float64 `json:"Capacity"`
	DateClose string  `json:"DateClose"`
	DateProd  string  `json:"DateProd"`
	ID        string  `json:"ID"`
	Kind      int32   `json:"Kind"`
	Owner     string  `json:"Owner"`
	Producer  string  `json:"Producer"`
	Status    string  `json:"Status"`
}

type BatteryAsset struct {
	Asset
	Spec       string `json:"Spec"`
	DateStop   string `json:"DateStop"`
	DateSalv   string `json:"DateSalv"`
	ReasonSalv string `json:"ReasonSalv"`
}

type SecondAsset struct {
	Asset
	RefAssetID string `json:"RefAssetID"`
	Spec       string `json:"Spec"`
}

type RecycleAsset struct {
	Asset
	Amount     float64 `json:"Amount"`
	Measure    string  `json:"Measure"`
	RefAssetID string  `json:"RefAssetID"`
}

// InitLedger adds a base set of assets to the ledger
// InitLedger incorporates a new definition of asset, batteryasset (by Hanool (04.14 13:34))
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// assets := []Asset{
	// 	{ID: "asset1", Color: "blue", Size: 5, Owner: "Tomoko", AppraisedValue: 300},
	// 	{ID: "asset2", Color: "red", Size: 5, Owner: "Brad", AppraisedValue: 400},
	// 	{ID: "asset3", Color: "green", Size: 10, Owner: "Jin Soo", AppraisedValue: 500},
	// 	{ID: "asset4", Color: "yellow", Size: 10, Owner: "Max", AppraisedValue: 600},
	// 	{ID: "asset5", Color: "black", Size: 15, Owner: "Adriana", AppraisedValue: 700},
	// 	{ID: "asset6", Color: "white", Size: 15, Owner: "Michel", AppraisedValue: 800},
	// }

	batteryassets := []BatteryAsset{
		{Asset: Asset{ID: "asset1", Capacity: 50.0, DateClose: "None", DateProd: "18-10-31", Kind: 1, Owner: "DaeguCity", Producer: "Samsung", Status: "New"},
			Spec:       "2|3",
			DateStop:   "20-10-30",
			DateSalv:   "20-11-30",
			ReasonSalv: "Repair",
		},
		{Asset: Asset{ID: "asset2", Capacity: 250.0, DateClose: "None", DateProd: "19-10-30", Kind: 1, Owner: "JejuCity", Producer: "SK", Status: "New"},
			Spec:       "3|3",
			DateStop:   "22-10-30",
			DateSalv:   "22-12-30",
			ReasonSalv: "Accident",
		},
	}

	for _, asset := range batteryassets {
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
// New definition of asset applies here as well by Hanool (04.14, 13:58)
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string,
	capacity float64, dtclose string, dtprod string, dtstop string, dtsalv string,
	kind int32, owner string, producer string, spec string, status string,
	reason string) error {

	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the battery asset %s already exists", id)
	}

	batteryasset := BatteryAsset{
		Asset: Asset{
			ID:        id,
			Capacity:  capacity,
			DateClose: dtclose,
			DateProd:  dtprod,
			Kind:      kind,
			Owner:     owner,
			Producer:  producer,
			Status:    status,
		},
		Spec:       spec,
		DateStop:   dtstop,
		DateSalv:   dtsalv,
		ReasonSalv: reason,
	}
	assetJSON, err := json.Marshal(batteryasset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
// New definition of asset applies here as well by Hanool (04.14, 14:18)
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*BatteryAsset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var batteryasset BatteryAsset
	err = json.Unmarshal(assetJSON, &batteryasset)
	if err != nil {
		return nil, err
	}

	return &batteryasset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
// New definition of asset applies here as well by Hanool (04.14, 14:19)
// Update Status of battery asset...
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string,
	status string) error {

	// Delete by Hanool (04.14, 14:54)
	// exists, err := s.AssetExists(ctx, id)
	// if err != nil {
	// 	return err
	// }
	// if !exists {
	// 	return fmt.Errorf("the asset %s does not exist", id)
	// }
	batteryasset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return err
	}

	// overwriting original asset with new asset
	newbatteryasset := BatteryAsset{
		Asset: Asset{
			ID:        id,
			Capacity:  batteryasset.Capacity,
			DateClose: batteryasset.DateClose,
			DateProd:  batteryasset.DateProd,
			Kind:      batteryasset.Kind,
			Owner:     batteryasset.Owner,
			Producer:  batteryasset.Producer,
			Status:    status,
		},
		Spec:       batteryasset.Spec,
		DateStop:   batteryasset.DateStop,
		DateSalv:   batteryasset.DateSalv,
		ReasonSalv: batteryasset.ReasonSalv,
	}

	assetJSON, err := json.Marshal(newbatteryasset)
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

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, id string, newOwner string) (string, error) {
	batteryasset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldOwner := batteryasset.Owner
	batteryasset.Owner = newOwner

	assetJSON, err := json.Marshal(batteryasset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, assetJSON)
	if err != nil {
		return "", err
	}

	return oldOwner, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*BatteryAsset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*BatteryAsset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset BatteryAsset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
