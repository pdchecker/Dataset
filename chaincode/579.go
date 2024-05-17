package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hyperledger_erc721/chaincode/model"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

/*
`TransferFrom` is invoke fnc that moves token
from is the owner's address, to is reciepient's address
*/
func (c *TokenERC721Contract) TransferFrom(ctx contractapi.TransactionContextInterface, from, to, tokenId string) (bool, error) {

	initialized, err := checkInitialized(ctx)

	if err != nil {
		return false, err
	}

	if !initialized {
		return false, fmt.Errorf("initialized first")
	}

	sender64, err := ctx.GetClientIdentity().GetID()

	if err != nil {
		return false, fmt.Errorf("failed to GetClientIdentity : %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)

	if err != nil {
		return false, fmt.Errorf("failed to DecodeString sender ID : %v", err)
	}

	sender := string(senderBytes)

	nft, err := _readNFT(ctx, tokenId)

	if err != nil {
		return false, fmt.Errorf("failed to _readNFT : %v", err)
	}

	owner := nft.Owner
	operator := nft.Approved
	operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)

	if err != nil {
		return false, fmt.Errorf("failed to get IsApprovedForAll : %v", err)
	}

	if owner != sender && operator != sender && !operatorApproval {
		return false, fmt.Errorf("the sender is not the current owner nor an authorized operator")
	}

	// Check if `from` is the current owner
	if owner != from {
		return false, fmt.Errorf("the from is not the current owner")
	}

	// Clear the approved client for this non-fungible token
	nft.Approved = ""

	// Overwrite a non-fungible token to assign a new owner.
	nft.Owner = to
	nftKey, err := ctx.GetStub().CreateCompositeKey(nftPrefix, []string{tokenId})

	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}

	nftBytes, err := json.Marshal(nft)
	if err != nil {
		return false, fmt.Errorf("failed to marshal approval: %v", err)
	}

	err = ctx.GetStub().PutState(nftKey, nftBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState nftBytes %s: %v", nftBytes, err)
	}

	// Remove a composite key from the balance of the current owner
	balanceKeyFrom, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{from, tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey from: %v", err)
	}

	err = ctx.GetStub().DelState(balanceKeyFrom)
	if err != nil {
		return false, fmt.Errorf("failed to DelState balanceKeyFrom %s: %v", nftBytes, err)
	}

	// Save a composite key to count the balance of a new owner
	balanceKeyTo, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{to, tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey to: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKeyTo, []byte{0})
	if err != nil {
		return false, fmt.Errorf("failed to PutState balanceKeyTo %s: %v", balanceKeyTo, err)
	}

	// Emit the Transfer event
	transferEvent := model.NewTransferMetadata(from, to, tokenId)

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return false, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
	}

	err = ctx.GetStub().SetEvent(TransferEventKey, transferEventBytes)
	if err != nil {
		return false, fmt.Errorf("failed to SetEvent transferEventBytes %s: %v", transferEventBytes, err)
	}

	return true, nil
}

/*
`MintWithTokenURI`is invoke fnc that mint a new non-fungible token
*/
func (c *TokenERC721Contract) MintWithTokenURI(ctx contractapi.TransactionContextInterface, tokenId string, tokenURI string) (*model.NFT, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return nil, fmt.Errorf("first initialized")
	}

	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, fmt.Errorf("failed to get clientMSPID: %v", err)
	}

	if clientMSPID != "Org1MSP" {
		return nil, fmt.Errorf("client is not authorized to set the name and symbol of the token")
	}

	minter64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get minter id: %v", err)
	}

	minterBytes, err := base64.StdEncoding.DecodeString(minter64)
	if err != nil {
		return nil, fmt.Errorf("failed to DecodeString minter64: %v", err)
	}

	minter := string(minterBytes)

	exists := _nftExists(ctx, tokenId)

	if exists {
		return nil, fmt.Errorf("the token %s is already minted", tokenId)
	}

	// Add a non-fungible token
	nft := model.NewNFT(tokenId, minter, tokenURI, "")

	nftKey, err := ctx.GetStub().CreateCompositeKey(nftPrefix, []string{tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey to nftKey: %v", err)
	}

	nftBytes, err := json.Marshal(nft)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal nft: %v", err)
	}

	err = ctx.GetStub().PutState(nftKey, nftBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to PutState nftBytes %s: %v", nftBytes, err)
	}

	// increase balance
	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{minter, tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey to balanceKey: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKey, []byte{'\u0000'})
	if err != nil {
		return nil, fmt.Errorf("failed to PutState balanceKey %s: %v", nftBytes, err)
	}

	// Emit the Transfer event
	transferEvent := model.NewTransferMetadata("0x0", minter, tokenId)

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
	}

	err = ctx.GetStub().SetEvent(TransferEventKey, transferEventBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to SetEvent transferEventBytes %s: %v", transferEventBytes, err)
	}

	return nft, nil
}

/*
`Approve` is invoke fnc that changes or reaffirms the approved client for a non-fungible token
*/

func (c *TokenERC721Contract) Approve(ctx contractapi.TransactionContextInterface, operator string, tokenId string) (bool, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return false, fmt.Errorf("first initialize")
	}

	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to DecodeString senderBytes: %v", err)
	}
	sender := string(senderBytes)

	nft, err := _readNFT(ctx, tokenId)
	if err != nil {
		return false, fmt.Errorf("failed to _readNFT: %v", err)
	}

	// Check if the sender is the current owner of the non-fungible token
	// or an authorized operator of the current owner
	owner := nft.Owner
	operatorApproval, err := c.IsApprovedForAll(ctx, owner, sender)
	if err != nil {
		return false, fmt.Errorf("failed to get IsApprovedForAll: %v", err)
	}
	if owner != sender && !operatorApproval {
		return false, fmt.Errorf("the sender is not the current owner nor an authorized operator")
	}

	// Update the approved operator of the non-fungible token
	nft.Approved = operator
	nftKey, err := ctx.GetStub().CreateCompositeKey(nftPrefix, []string{tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey %s: %v", nftKey, err)
	}

	nftBytes, err := json.Marshal(nft)
	if err != nil {
		return false, fmt.Errorf("failed to marshal nftBytes: %v", err)
	}

	err = ctx.GetStub().PutState(nftKey, nftBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState for nftKey: %v", err)
	}

	return true, nil
}

/*
`SetApprovalForAll` is invoke fnc that enables or disables approval for a third party ("operator")
this is manage all the message sender's assets
*/

func (c *TokenERC721Contract) SetApprovalForAll(ctx contractapi.TransactionContextInterface, operator string, approved bool) (bool, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return false, fmt.Errorf("please first initialize")
	}

	sender64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to GetClientIdentity: %v", err)
	}

	senderBytes, err := base64.StdEncoding.DecodeString(sender64)
	if err != nil {
		return false, fmt.Errorf("failed to DecodeString sender: %v", err)
	}
	sender := string(senderBytes)

	nftApproval := model.NewApproval(sender, operator, approved)

	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{sender, operator})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}

	approvalBytes, err := json.Marshal(nftApproval)
	if err != nil {
		return false, fmt.Errorf("failed to marshal approvalBytes: %v", err)
	}

	err = ctx.GetStub().PutState(approvalKey, approvalBytes)
	if err != nil {
		return false, fmt.Errorf("failed to PutState approvalBytes: %v", err)
	}

	err = ctx.GetStub().SetEvent("ApprovalForAll", approvalBytes)
	if err != nil {
		return false, fmt.Errorf("failed to SetEvent ApprovalForAll: %v", err)
	}

	return true, nil
}

/*
`Burn` is invoke fnc that burn a non-fungible token
*/
func (c *TokenERC721Contract) Burn(ctx contractapi.TransactionContextInterface, tokenId string) (bool, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return false, fmt.Errorf("please first initialize")
	}

	owner64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return false, fmt.Errorf("failed to GetClientIdentity owner64: %v", err)
	}

	ownerBytes, err := base64.StdEncoding.DecodeString(owner64)
	if err != nil {
		return false, fmt.Errorf("failed to DecodeString owner64: %v", err)
	}
	owner := string(ownerBytes)

	// Check if a caller is the owner of the non-fungible token
	nft, err := _readNFT(ctx, tokenId)
	if err != nil {
		return false, fmt.Errorf("failed to _readNFT nft : %v", err)
	}
	if nft.Owner != owner {
		return false, fmt.Errorf("non-fungible token %s is not owned by %s", tokenId, owner)
	}

	// Delete the token
	nftKey, err := ctx.GetStub().CreateCompositeKey(nftPrefix, []string{tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey tokenId: %v", err)
	}

	err = ctx.GetStub().DelState(nftKey)
	if err != nil {
		return false, fmt.Errorf("failed to DelState nftKey: %v", err)
	}

	// Remove a composite key from the balance of the owner
	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{owner, tokenId})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey balanceKey %s: %v", balanceKey, err)
	}

	err = ctx.GetStub().DelState(balanceKey)
	if err != nil {
		return false, fmt.Errorf("failed to DelState balanceKey %s: %v", balanceKey, err)
	}

	// Emit the Transfer event
	transferEvent := model.NewTransferMetadata(owner, "0x0", tokenId)

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return false, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return false, fmt.Errorf("failed to SetEvent transferEventBytes: %v", err)
	}

	return true, nil
}
