package main

import (
	"encoding/json"
	"fmt"

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

func (c *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// Initializations if needed
	return nil
}

func (c *SmartContract) CreatePackage(ctx contractapi.TransactionContextInterface, id string, dateOfManufacture string, placeOfOrigin string) error {
	// Add validations for dateOfManufacture and placeOfOrigin

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

// Add functions for the supply chain package management...
// ... such as updating customs details, marking as damaged, etc.

// NFT functions

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

func (c *SmartContract) GetNFT(ctx contractapi.TransactionContextInterface, id string) (NFT, error) {
	nftBytes, _ := ctx.GetStub().GetState(id)
	if nftBytes == nil {
		return NFT{}, fmt.Errorf("NFT not found")
	}
	var nft NFT
	_ = json.Unmarshal(nftBytes, &nft)
	return nft, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create supply chain chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting supply chain chaincode: %s", err.Error())
	}
}
