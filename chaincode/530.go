package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/skip2/go-qrcode"
	"reflect"
	//"reflect"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	AppraisedValue int               `json:"AppraisedValue"`
	Color          string            `json:"Color"`
	ID             string            `json:"ID"`
	Owner          string            `json:"Owner"`
	Size           int               `json:"Size"`
	Time           string            `json:"Time"`
	QR             string            `json:"QR"`
	Additional     map[string]string `json:"Additional"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int, category string, value string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	asset = add(asset, category, value)
	asset = hash(asset)
	asset.Time = time.Now().Format("15:04:05 02-01-2006")
	code(asset)
	qr := asset.QR
	if err != nil {
		return err
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(qr, assetJSON)
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwriting original asset with new asset
	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	asset = hash(asset)
	asset.Time = time.Now().Format("15:04:05 02-01-2006")
	code(asset)
	qr, err := s.GetQR(ctx, id)
	fmt.Println(qr)
	if err != nil {
		return err
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(qr, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	qrcode, err := s.GetQR(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(qrcode)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	qrcode, err := s.GetQR(ctx, id)
	fmt.Println(qrcode)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	assetJSON, err := ctx.GetStub().GetState(qrcode)
	fmt.Println(assetJSON)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface, parameter string, value string) ([]*Asset, error) {
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

		t := reflect.TypeOf(asset)
		v := reflect.ValueOf(asset)

		for i := 0; i < t.NumField(); i++ {
			if parameter == t.Field(i).Name {
				if fmt.Sprint(v.Field(i).Interface()) == value {
					assets = append(assets, &asset)
				}
			}
		}

	}

	return assets, nil
}

func (s *SmartContract) GetQR(ctx contractapi.TransactionContextInterface, value string) (string, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	var assetQR string
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return assetQR, err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return assetQR, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return assetQR, err
		}

		t := reflect.TypeOf(asset)
		v := reflect.ValueOf(asset)

		for i := 0; i < t.NumField(); i++ {
			fmt.Println(t.Field(i).Name)
			if t.Field(i).Name == "ID" {
				fmt.Println("mphke")
				if fmt.Sprint(v.Field(i).Interface()) == value {
					fmt.Sprint(v.Field(i).Interface())
					QR := v.FieldByName("QR").Interface().(string)
					assetQR = QR
					fmt.Println(assetQR)
				}
			}
		}

	}

	return assetQR, nil
}

func hash(newasset Asset) Asset {
	hash := sha256.New()
	structString := fmt.Sprintf("%+v", newasset)
	_, err := hash.Write([]byte(structString))
	if err != nil {
		fmt.Println("Error creating SHA256 hash:", err)
	}
	hashValue := hash.Sum(nil)
	hexString := hex.EncodeToString(hashValue)
	newasset.QR = hexString
	return newasset

}

func code(newasset Asset) {

	err := qrcode.WriteFile(newasset.QR, qrcode.High, 256, fmt.Sprintf("/tmp/%s.png", newasset.ID))
	if err != nil {
		fmt.Println("Error generating QR code:", err)
	}

}

func add(newasset Asset, category string, value string) Asset {
	// Create a map and add some key-value pairs to it
	data := make(map[string]string)
	data[category] = value
	// Create a new struct and add the map to it
	newasset.Additional = data
	// Return the struct
	return newasset
}
