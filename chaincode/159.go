package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Package struct {
	ID                string   `json:"id"`
	DateOfManufacture string   `json:"dateOfManufacture"` // ISO 8601 date format
	PlaceOfOrigin     string   `json:"placeOfOrigin"`     // "City, Country"
	CurrentStatus     string   `json:"currentStatus"`
	CustomsDetails    []string `json:"customsDetails"`
	IsDamaged         bool     `json:"isDamaged"`
}

type NFT struct {
	ID       string `json:"id"`
	Owner    string `json:"owner"`
	Category string `json:"category"` // "Antibiotic", "Painkiller", etc.
}

// InitLedger - Initialize the ledger with default values
func (c *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// Initializations if needed
	return nil
}

// CreatePackage - Create a new package
func (c *SmartContract) CreatePackage(ctx contractapi.TransactionContextInterface, id, dateOfManufacture, placeOfOrigin string) error {
	// Validate dateOfManufacture and placeOfOrigin
	if _, err := time.Parse(time.RFC3339, dateOfManufacture); err != nil {
		return fmt.Errorf("invalid date format: %s", err.Error())
	}
	if placeOfOrigin == "" {
		return errors.New("place of origin must not be empty")
	}

	pkg := Package{
		ID:                id,
		DateOfManufacture: dateOfManufacture,
		PlaceOfOrigin:     placeOfOrigin,
		CurrentStatus:     "CREATED",
		CustomsDetails:    []string{},
		IsDamaged:         false,
	}
	packageBytes, _ := json.Marshal(pkg)
	return ctx.GetStub().PutState(id, packageBytes)
}

// UpdatePackageStatus - Update the status of a package
func (c *SmartContract) UpdatePackageStatus(ctx contractapi.TransactionContextInterface, id, newStatus string) error {
	// Retrieve the package from the ledger
	packageBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to get package: %s", err.Error())
	}
	if packageBytes == nil {
		return fmt.Errorf("package not found: %s", id)
	}

	var pkg Package
	err = json.Unmarshal(packageBytes, &pkg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal package: %s", err.Error())
	}

	// Update the package status
	pkg.CurrentStatus = newStatus
	packageBytes, _ = json.Marshal(pkg)
	return ctx.GetStub().PutState(id, packageBytes)
}

// MintNFT - Create a new NFT for the package
func (c *SmartContract) MintNFT(ctx contractapi.TransactionContextInterface, id, owner, category string) error {
	// Validate category if needed

	nft := NFT{
		ID:       id,
		Owner:    owner,
		Category: category,
	}
	nftBytes, _ := json.Marshal(nft)
	return ctx.GetStub().PutState(id, nftBytes)
}

// TransferNFT - Transfer an NFT to a new owner
func (c *SmartContract) TransferNFT(ctx contractapi.TransactionContextInterface, id, newOwner string) error {
	nftBytes, _ := ctx.GetStub().GetState(id)
	if nftBytes == nil {
		return fmt.Errorf("NFT not found")
	}
	var nft NFT
	_ = json.Unmarshal(nftBytes, &nft)
	nft.Owner = newOwner
	updatedNFTBytes, _ := json.Marshal(nft)
	return ctx.GetStub().PutState(id, updatedNFTBytes)
}

// GetNFT - Retrieve an NFT from the ledger
func (c *SmartContract) GetNFT(ctx contractapi.TransactionContextInterface, id string) (NFT, error) {
	nftBytes, _ := ctx.GetStub().GetState(id)
	if nftBytes == nil {
		return NFT{}, fmt.Errorf("NFT not found")
	}
	var nft NFT
	_ = json.Unmarshal(nftBytes, &nft)
	return nft, nil
}

// func main() {
// 	chaincode, err := contractapi.NewChaincode(new(SmartContract))
// 	if err != nil {
// 		fmt.Printf("Error create supply chain chaincode: %s", err.Error())
// 		return
// 	}

// 	if err := chaincode.Start(); err != nil {
// 		fmt.Printf("Error starting supply chain chaincode: %s", err.Error())
// 	}
// }
