package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"hyperledger_erc721/chaincode/model"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func _readNFT(ctx contractapi.TransactionContextInterface, tokenId string) (*model.NFT, error) {
	nftKey, err := ctx.GetStub().CreateCompositeKey(nftPrefix, []string{tokenId})

	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey %s: %v", tokenId, err)
	}

	nftBytes, err := ctx.GetStub().GetState(nftKey)

	if err != nil {
		return nil, fmt.Errorf("failed to GetState %s: %v", tokenId, err)
	}

	nft := model.NewNFT("", "", "", "")
	err = json.Unmarshal(nftBytes, nft)
	if err != nil {
		return nil, fmt.Errorf("failed to Unmarshal nftBytes: %v", err)
	}

	return nft, nil
}

func _nftExists(ctx contractapi.TransactionContextInterface, tokenId string) bool {
	nftKey, err := ctx.GetStub().CreateCompositeKey(nftPrefix, []string{tokenId})
	if err != nil {
		panic("error creating CreateCompositeKey:" + err.Error())
	}

	nftBytes, err := ctx.GetStub().GetState(nftKey)
	if err != nil {
		panic("error GetState nftBytes:" + err.Error())
	}

	return len(nftBytes) > 0
}

func (c *TokenERC721Contract) BalanceOf(ctx contractapi.TransactionContextInterface, owner string) int {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		panic(err.Error())
	}
	if !initialized {
		panic("first initialized")
	}

	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(balancePrefix, []string{owner})
	if err != nil {
		panic("Error creating asset chaincode:" + err.Error())
	}

	balance := 0
	for iterator.HasNext() {
		_, err := iterator.Next()
		if err != nil {
			return 0
		}
		balance++

	}
	return balance
}

func (c *TokenERC721Contract) OwnerOf(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return "", fmt.Errorf("first initialize")
	}

	nft, err := _readNFT(ctx, tokenId)
	if err != nil {
		return "", fmt.Errorf("could not process OwnerOf for tokenId: %w", err)
	}

	return nft.Owner, nil
}

func (c *TokenERC721Contract) IsApprovedForAll(ctx contractapi.TransactionContextInterface, owner string, operator string) (bool, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return false, fmt.Errorf("first initialize")
	}

	approvalKey, err := ctx.GetStub().CreateCompositeKey(approvalPrefix, []string{owner, operator})
	if err != nil {
		return false, fmt.Errorf("failed to CreateCompositeKey: %v", err)
	}
	approvalBytes, err := ctx.GetStub().GetState(approvalKey)
	if err != nil {
		return false, fmt.Errorf("failed to GetState approvalBytes %s: %v", approvalBytes, err)
	}

	if len(approvalBytes) < 1 {
		return false, nil
	}

	approval := model.NewApproval("", "", false)
	err = json.Unmarshal(approvalBytes, approval)
	if err != nil {
		return false, fmt.Errorf("failed to Unmarshal: %v, string %s", err, string(approvalBytes))
	}

	return approval.Approved, nil

}

/*
`GetApproved` is query fnc that returns the approved client for a single non-fungible token
*/
func (c *TokenERC721Contract) GetApproved(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return "false", fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return "false", fmt.Errorf("please first initialize")
	}

	nft, err := _readNFT(ctx, tokenId)
	if err != nil {
		return "false", fmt.Errorf("failed GetApproved for tokenId : %v", err)
	}

	return *nft.GetApproved(), nil
}

/*
`TokenURI` is query fnc that TokenURI returns a distinct Uniform Resource Identifier (URI) for a given token.
*/
func (c *TokenERC721Contract) TokenURI(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return "", fmt.Errorf("please first initialize")
	}

	nft, err := _readNFT(ctx, tokenId)
	if err != nil {
		return "", fmt.Errorf("failed to get TokenURI: %v", err)
	}
	return nft.TokenURI, nil
}

/*
`TotalSupply` is query fnc that counts non-fungible tokens tracked by this contract.
*/
func (c *TokenERC721Contract) TotalSupply(ctx contractapi.TransactionContextInterface) int {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		panic("failed to check if contract ia already initialized:" + err.Error())
	}
	if !initialized {
		panic("please first initialize")
	}

	iterator, err := ctx.GetStub().GetStateByPartialCompositeKey(nftPrefix, []string{})
	if err != nil {
		panic("Error creating GetStateByPartialCompositeKey:" + err.Error())
	}

	totalSupply := 0
	for iterator.HasNext() {
		_, err := iterator.Next()
		if err != nil {
			return 0
		}
		totalSupply++

	}
	return totalSupply

}

/*
`ClientAccountBalance` is query fnc that returns the balance of the requesting client's account.
*/
func (c *TokenERC721Contract) ClientAccountBalance(ctx contractapi.TransactionContextInterface) (int, error) {

	initialized, err := checkInitialized(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to check if contract ia already initialized: %v", err)
	}
	if !initialized {
		return 0, fmt.Errorf("please first initialize")
	}

	clientAccountID64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return 0, fmt.Errorf("failed to GetClientIdentity minter: %v", err)
	}

	clientAccountIDBytes, err := base64.StdEncoding.DecodeString(clientAccountID64)
	if err != nil {
		return 0, fmt.Errorf("failed to DecodeString sender: %v", err)
	}

	clientAccountID := string(clientAccountIDBytes)

	return c.BalanceOf(ctx, clientAccountID), nil
}
