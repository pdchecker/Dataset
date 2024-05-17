package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	//"github.com/skip2/go-qrcode"
	//"reflect"
	//"reflect"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
	"strings"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type LNGDocument struct {
    DocumentID   string		`json:"DocumentID"`
    DocumentCID  string		`json:"DocumentCID"`
    Owner        string		`json:"Owner"`
    AccessList   []string	`json:"AccessList"`
	QR 			 string		`json:"QR"`
	Timestamp 		 string		`json:"Timestamp"`
}

type AssetInfo struct {
	DocumentID string `json:"DocumentID"`
	Owner      string `json:"Owner"`
	DocumentCID  string		`json:"DocumentCID"`
	Timestamp 		 string		`json:"Timestamp"`
}

type AssetInfoOwner struct {
	DocumentID string `json:"DocumentID"`
	Owner      string `json:"Owner"`
	DocumentCID  string		`json:"DocumentCID"`
	AccessList   []string	`json:"AccessList"`
	Timestamp 		 string		`json:"Timestamp"`
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateDocument(ctx contractapi.TransactionContextInterface, DocumentID string, DocumentCID string, Owner string, access string) error{
	exists, err := s.AssetExists(ctx, DocumentID, Owner)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", DocumentID)
	}

	asset := LNGDocument{
		DocumentID:		DocumentID,
		DocumentCID:    DocumentCID,
		Owner:          Owner,
		Timestamp:		time.Now().Format("15:04:05 02-01-2006"),
	}

	asset = accesslist(asset, access)
	asset = hash(asset)
	qr := asset.QR

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(qr, assetJSON)

}

func (s *SmartContract) UpdateAccess(ctx contractapi.TransactionContextInterface, access string, DocumentID string, Owner string) error {
	asset, err := s.ReturnAsset(ctx, DocumentID, Owner)
	if err != nil {
		return err
	}

	// Update only the access list field of the asset
	updatedAsset := accesslist(asset, access)

	assetJSON, err := json.Marshal(updatedAsset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(asset.QR, assetJSON)
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, DocumentID string, Owner string) error {
	qrcode, err := s.GetQR(ctx, DocumentID, Owner)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	exists, err := s.AssetExists(ctx, DocumentID, Owner)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", DocumentID)
	}

	return ctx.GetStub().DelState(qrcode)
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, DocumentID string, Owner string) (bool, error) {
	qrcode, err := s.GetQR(ctx, DocumentID, Owner)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	assetJSON, err := ctx.GetStub().GetState(qrcode)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	
	return assetJSON != nil, nil
}

func (s *SmartContract) ReturnAsset(ctx contractapi.TransactionContextInterface, DocumentID string, Owner string) (LNGDocument, error) {
	qrcode, err := s.GetQR(ctx, DocumentID, Owner)
	if err != nil {
		return LNGDocument{}, fmt.Errorf("failed to read from world state: %v", err)
	}
	assetJSON, err := ctx.GetStub().GetState(qrcode)
	if err != nil {
		return LNGDocument{}, fmt.Errorf("failed to read from world state: %v", err)
	}
	asset, err := DecodeAsset(assetJSON)
	if err != nil {
		return LNGDocument{}, fmt.Errorf("failed to decode asset: %v", err)
	}
	return asset, nil
}

func (s *SmartContract) GetCIDByAccessName(ctx contractapi.TransactionContextInterface, accessName string) ([]AssetInfo, error) {
	var filteredAssets []AssetInfo

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return filteredAssets, err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return filteredAssets, err
		}

		var asset LNGDocument
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return filteredAssets, err
		}

		for _, name := range asset.AccessList {
			if name == accessName {
				assetInfo := AssetInfo{
					DocumentID: asset.DocumentID,
					DocumentCID: asset.DocumentCID,
					Owner:      asset.Owner,
					Timestamp: 		asset.Timestamp,
					
				}
				filteredAssets = append(filteredAssets, assetInfo)
				break
			}
		}
	}

	return filteredAssets, nil
}

func (s *SmartContract) GetCIDByOwnerName(ctx contractapi.TransactionContextInterface, ownerName string) ([]AssetInfoOwner, error) {
	var filteredAssets []AssetInfoOwner

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return filteredAssets, err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return filteredAssets, err
		}

		var asset LNGDocument
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return filteredAssets, err
		}
		if ownerName == asset.Owner {
			assetInfo := AssetInfoOwner{
				DocumentID: asset.DocumentID,
				DocumentCID: asset.DocumentCID,
				Owner:      asset.Owner,
				AccessList: asset.AccessList,
				Timestamp: 		asset.Timestamp,
			}
			filteredAssets = append(filteredAssets, assetInfo)
		}
	}
	return filteredAssets, nil
}


func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*LNGDocument, error) {
    resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
    if err != nil {
        return nil, fmt.Errorf("failed to get assets: %v", err)
    }
    defer resultsIterator.Close()

    var assets []*LNGDocument
    for resultsIterator.HasNext() {
        queryResponse, err := resultsIterator.Next()
        if err != nil {
            return nil, fmt.Errorf("failed to iterate over assets: %v", err)
        }

        var asset LNGDocument
        err = json.Unmarshal(queryResponse.Value, &asset)
        if err != nil {
            return nil, fmt.Errorf("failed to unmarshal asset: %v", err)
        }

        assets = append(assets, &asset)
    }

    return assets, nil
}


func (s *SmartContract) GetQR(ctx contractapi.TransactionContextInterface, value1 string, value2 string) (string, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	var asserDoc string
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return asserDoc, err
	}
	defer resultsIterator.Close()

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return asserDoc, err
		}

		var asset LNGDocument
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return asserDoc, err
		}

		if asset.DocumentID == value1 && asset.Owner == value2 {
			return asset.QR, nil
		}
	}

	return "", nil

}


func hash(newasset LNGDocument) LNGDocument {
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

func accesslist(oldAsset LNGDocument, access string) LNGDocument {
	oldAsset.AccessList = make([]string, 0)

	accessList := strings.Split(access, ", ")

	for _, a := range accessList {
		oldAsset.AccessList = append(oldAsset.AccessList, a)
	}

	return oldAsset
}



func DecodeAsset(data []byte) (LNGDocument, error) {
    var asset LNGDocument
    err := json.Unmarshal(data, &asset)
    if err != nil {
        return LNGDocument{}, err
    }
    return asset, nil
}