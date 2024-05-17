// SPDX-License-Identifier: MIT
package main

import (
	"strings"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
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

func (c *StepContract) Init(ctx contractapi.TransactionContextInterface, chainID string, chainStep string, companyName string) error {


    // Set the initialized values in the chaincode state
    c.IChainID = chainID
    c.IChainStep = chainStep
    c.ICompanyName = companyName

    // Set the owner of the chaincode (the organization that instantiated it)
		clientID, err := ctx.GetClientIdentity().GetID()

		if err != nil {
			return fmt.Errorf("failed to get client ID: %v", err)
		}

    c.IOwner = extractCommonName(clientID)

    return nil
}


func (c *StepContract) OnlyOwner(ctx contractapi.TransactionContextInterface) (bool, error) {
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to get client ID: %v", err)
	}

	// Extract common name (CN) from the client ID (assuming it's a certificate)
	clientCN := extractCommonName(clientID)

	return clientCN == c.IOwner, nil
}

// Function to extract the Common Name (CN) from a certificate
func extractCommonName(cert string) string {
	parts := strings.Split(cert, ",")

	for _, part := range parts {
		if strings.Contains(part, "CN=") {
			return strings.TrimSpace(strings.TrimPrefix(part, "CN="))
		}
	}

	return ""
}

func (c *StepContract) PublishBatch(ctx contractapi.TransactionContextInterface, category string, productID string, parent []ParentInfo, productName string, uom string, quantity uint64) error {
	isOwner, err := c.OnlyOwner(ctx)
	if err != nil {
		return fmt.Errorf("failed to check ownership: %v", err)
	}
	if !isOwner {
		return fmt.Errorf("not the owner")
	}

	event := BatchEvent{
		Category:    category,
		ProductID:   productID,
		Parent:      parent,
		ProductName: productName,
		UOM:         uom,
		Quantity:    quantity,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	err = ctx.GetStub().SetEvent("emitBatch", eventBytes)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	return nil
}

func (c *StepContract) StoreHash(ctx contractapi.TransactionContextInterface, category string, productID string, multi []Multihash) error {
	isOwner, err := c.OnlyOwner(ctx)
	if err != nil {
		return fmt.Errorf("failed to check ownership: %v", err)
	}
	if !isOwner {
		return fmt.Errorf("not the owner")
	}

	event := HashesEvent{
		Category:  category,
		ProductID: productID,
		Multi:     multi,
	}

	eventBytes, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %v", err)
	}

	err = ctx.GetStub().SetEvent("emitHashes", eventBytes)
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
