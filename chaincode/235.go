package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract defines the structure of the chaincode
type SmartContract struct {
	contractapi.Contract
}

// NFT represents an art collectible with details
type NFT struct {
	ID    string `json:"id"`
	Owner string `json:"owner"`
	Price int    `json:"price"` // Price is in FabBits (1 FabCoin = 1000 FabBits)
}

// InitLedger initializes the ledger with some default values (if needed)
func (t *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	nfts := []NFT{
		{ID: "NFT1", Owner: "Alice", Price: 10000},
		{ID: "NFT2", Owner: "Bob", Price: 15000},
	}

	for _, nft := range nfts {
		nftBytes, err := json.Marshal(nft)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(nft.ID, nftBytes)
		if err != nil {
			return err
		}
	}
	return nil
}

// Mint creates a new NFT in the ledger
func (t *SmartContract) Mint(ctx contractapi.TransactionContextInterface, id string, priceInFabCoin float64) error {
	if priceInFabCoin <= 0 {
		return fmt.Errorf("price must be a positive value")
	}

	priceInFabBits := int(priceInFabCoin * 1000) // Convert FabCoin to FabBits

	ownerId, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client identity: %s", err)
	}

	nft := NFT{
		ID:    id,
		Owner: ownerId,
		Price: priceInFabBits,
	}
	nftBytes, err := json.Marshal(nft)
	if err != nil {
		return fmt.Errorf("failed to marshal NFT: %s", err)
	}

	return ctx.GetStub().PutState(id, nftBytes)
}

// Transfer changes the ownership of an NFT
func (t *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	nftBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to get NFT: %s", err)
	}
	if nftBytes == nil {
		return fmt.Errorf("NFT not found")
	}

	nft := NFT{}
	err = json.Unmarshal(nftBytes, &nft)
	if err != nil {
		return fmt.Errorf("failed to unmarshal NFT: %s", err)
	}

	nft.Owner = newOwner
	updatedNftBytes, err := json.Marshal(nft)
	if err != nil {
		return fmt.Errorf("failed to marshal updated NFT: %s", err)
	}

	return ctx.GetStub().PutState(id, updatedNftBytes)
}

// PurchaseNFT allows a user to buy an NFT from another user
func (t *SmartContract) PurchaseNFT(ctx contractapi.TransactionContextInterface, nftID string, paymentInFabCoin float64) error {
	paymentInFabBits := int(paymentInFabCoin * 1000)

	buyerID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get buyer identity: %s", err)
	}

	nftBytes, err := ctx.GetStub().GetState(nftID)
	if err != nil {
		return fmt.Errorf("failed to get NFT: %s", err)
	}
	if nftBytes == nil {
		return fmt.Errorf("NFT not found")
	}

	nft := NFT{}
	err = json.Unmarshal(nftBytes, &nft)
	if err != nil {
		return fmt.Errorf("failed to unmarshal NFT: %s", err)
	}

	if buyerID == nft.Owner {
		return fmt.Errorf("buyer cannot be the current owner")
	}

	if paymentInFabBits < nft.Price {
		return fmt.Errorf("insufficient payment, NFT price is %d FabBits", nft.Price)
	}

	// Logic to deduct FabBits from buyer's account and add to seller's account
	// [Assuming the existence of account management logic in the chaincode]

	nft.Owner = buyerID
	updatedNftBytes, err := json.Marshal(nft)
	if err != nil {
		return fmt.Errorf("failed to marshal updated NFT: %s", err)
	}

	return ctx.GetStub().PutState(nftID, updatedNftBytes)
}

// QueryNFT retrieves details of an NFT
func (t *SmartContract) QueryNFT(ctx contractapi.TransactionContextInterface, nftID string) (*NFT, error) {
	nftBytes, err := ctx.GetStub().GetState(nftID)
	if err != nil {
		return nil, fmt.Errorf("failed to get NFT: %s", err)
	}
	if nftBytes == nil {
		return nil, fmt.Errorf("NFT not found")
	}

	nft := new(NFT)
	err = json.Unmarshal(nftBytes, nft)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal NFT: %s", err)
	}
	return nft, nil
}

// SetAccountBalance sets the balance of a given account (for testing)
func (t *SmartContract) SetAccountBalance(ctx contractapi.TransactionContextInterface, accountID string, balanceInFabCoin float64) error {
	balanceInFabBits := int(balanceInFabCoin * 1000)
	return ctx.GetStub().PutState(accountID, []byte(strconv.Itoa(balanceInFabBits)))
}

// main function can be uncommented when deploying the chaincode
// func main() {
// 	chaincode, err := contractapi.NewChaincode(&SmartContract{})
// 	if err != nil {
// 		fmt.Printf("Error creating NFT chaincode: %s", err.Error())
// 		return
// 	}
// 	if err := chaincode.Start(); err != nil {
// 		fmt.Printf("Error starting NFT chaincode: %s", err.Error())
// 	}
// }
