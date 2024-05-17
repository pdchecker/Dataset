// SPDX-License-Identifier: MIT
package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"errors"
)

type StepContract struct {
	contractapi.Contract
	IOwner        string `json:"i_owner"`
	ICompanyName  string `json:"i_company_name"`
	IChainID      string `json:"i_chain_id"`
	IChainStep    string `json:"i_chain_step"`
}

type Multihash struct {
	FileName      string `json:"file_name"`
	HashFunction  string `json:"hash_function"`
	Size          string `json:"size"`
	Hash          string `json:"hash"`
}

type ParentInfo struct {
	ParentContract string `json:"parent_contract"`
	ProductID      string `json:"product_id"`
}

type BatchEvent struct {
	Category      string      `json:"category"`
	ProductID     string      `json:"product_id"`
	Parent        []ParentInfo `json:"parent"`
	ProductName   string      `json:"product_name"`
	UOM           string      `json:"uom"`
	Quantity      uint64      `json:"quantity"`
}

type HashesEvent struct {
	Category  string     `json:"category"`
	ProductID string     `json:"product_id"`
	Multi     []Multihash `json:"multi"`
}

// Errors
var ErrNotOwner = errors.New("not the owner")
var ErrContractClosed = errors.New("contract closed")

func (c *StepContract) Init(ctx contractapi.TransactionContextInterface, chainID string, chainStep string, companyName string) error {
	// Set the initialized values in the chaincode state
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return fmt.Errorf("failed to get creator: %v", err)
	}

	c.IOwner = string(creator) // Convert []byte to string if IOwner is a string type
	c.IChainID = chainID
	c.IChainStep = chainStep
	c.ICompanyName = companyName
	return nil
}

func (s *StepContract) OnlyOwner(ctx contractapi.TransactionContextInterface) {
	owner, err := ctx.GetStub().GetCreator()
	if err != nil || string(owner) != s.IOwner {
		panic(ErrNotOwner.Error())
	}
}

func (c *StepContract) PublishBatch(ctx contractapi.TransactionContextInterface, category string, productID string, parent []ParentInfo, productName string, uom string, quantity uint64) error {
	c.OnlyOwner(ctx)

	// Create a composite key using the specified attributes
	batchKey, err := ctx.GetStub().CreateCompositeKey("Batch", []string{category, productID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Check if the batch already exists
	batchExists, err := ctx.GetStub().GetState(batchKey)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if batchExists != nil {
		return fmt.Errorf("batch with category %s and product ID %s already exists", category, productID)
	}

	// Store the batch data in the world state
	batch := BatchEvent{
		Category:    category,
		ProductID:   productID,
		Parent:      parent,
		ProductName: productName,
		UOM:         uom,
		Quantity:    quantity,
	}
	batchBytes, err := json.Marshal(batch)
	if err != nil {
		return fmt.Errorf("failed to marshal batch: %v", err)
	}

	err = ctx.GetStub().PutState(batchKey, batchBytes)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Emit the batch event
	err = ctx.GetStub().SetEvent("emitBatch", batchBytes)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func (c *StepContract) StoreHash(ctx contractapi.TransactionContextInterface, category string, productID string, multi []Multihash) error {
	c.OnlyOwner(ctx)

	// Create a composite key using the specified attributes
	hashKey, err := ctx.GetStub().CreateCompositeKey("Hash", []string{category, productID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	// Check if the hash data already exists
	hashExists, err := ctx.GetStub().GetState(hashKey)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if hashExists != nil {
		return fmt.Errorf("hash with category %s and product ID %s already exists", category, productID)
	}

	// Store the hash data in the world state
	hash := HashesEvent{
		Category:  category,
		ProductID: productID,
		Multi:     multi,
	}
	hashBytes, err := json.Marshal(hash)
	if err != nil {
		return fmt.Errorf("failed to marshal hash: %v", err)
	}

	err = ctx.GetStub().PutState(hashKey, hashBytes)
	if err != nil {
		return fmt.Errorf("failed to put state: %v", err)
	}

	// Emit the hashes event
	err = ctx.GetStub().SetEvent("emitHashes", hashBytes)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&StepContract{})
	if err != nil {
		fmt.Printf("Error creating StepContract chaincode: %v", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting StepContract chaincode: %v", err)
	}
}
