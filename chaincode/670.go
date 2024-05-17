package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type NftChaincode struct {
	contractapi.Contract
}


//////////////////////////////////	NFT Holder //////////////////////////////////
func _readNFT(ctx contractapi.TransactionContextInterface, tokenId string) (*Nft, error) {
	nftKey, err := ctx.GetStub().CreateCompositeKey(nftPrefix, []string{tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey %s: %v", tokenId, err)
	}

	nftBytes, err := ctx.GetStub().GetState(nftKey)
	if err != nil {
		return nil, fmt.Errorf("failed to GetState %s: %v", tokenId, err)
	}

	nft := new(Nft)
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

// ============== ERC721 metadata extension ===============

// Name returns a descriptive name for a collection of non-fungible tokens in this contract
// returns {String} Returns the name of the token

func (c *NftChaincode) Name(ctx contractapi.TransactionContextInterface) (string, error) {
	bytes, err := ctx.GetStub().GetState(nameKey)
	if err != nil {
		return "", fmt.Errorf("failed to get Name bytes: %s", err)
	}

	return string(bytes), nil
}

// Symbol returns an abbreviated name for non-fungible tokens in this contract.
// returns {String} Returns the symbol of the token

func (c *NftChaincode) Symbol(ctx contractapi.TransactionContextInterface) (string, error) {
	bytes, err := ctx.GetStub().GetState(symbolKey)
	if err != nil {
		return "", fmt.Errorf("failed to get Symbol: %v", err)
	}

	return string(bytes), nil
}

func (c *NftChaincode) TokenURI(ctx contractapi.TransactionContextInterface, tokenId string) (string, error) {
	nft, err := _readNFT(ctx, tokenId)
	if err != nil {
		return "", fmt.Errorf("failed to get TokenURI: %v", err)
	}
	return nft.TokenURI, nil
}


func (c *NftChaincode) MintWithTokenURI(ctx contractapi.TransactionContextInterface, tokenId string, tokenURI string) (*Nft, error) {

	// Check minter authorization - this sample assumes Org1 is the issuer with privilege to mint a new token
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, fmt.Errorf("failed to get clientMSPID: %v", err)
	}

	if clientMSPID != "Org1MSP" {
		return nil, fmt.Errorf("client is not authorized to set the name and symbol of the token")
	}

	// Get ID of submitting client identity
	minter64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return nil, fmt.Errorf("failed to get minter id: %v", err)
	}

	minterBytes, err := base64.StdEncoding.DecodeString(minter64)
	if err != nil {
		return nil, fmt.Errorf("failed to DecodeString minter64: %v", err)
	}
	minter := string(minterBytes)

	// Check if the token to be minted does not exist
	exists := _nftExists(ctx, tokenId)
	if exists {
		return nil, fmt.Errorf("the token %s is already minted.: %v", tokenId, err)
	}

	// Add a non-fungible token
	nft := new(Nft)
	nft.TokenId = tokenId
	nft.Owner = minter
	nft.TokenURI = tokenURI

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

	// A composite key would be balancePrefix.owner.tokenId, which enables partial
	// composite key query to find and count all records matching balance.owner.*
	// An empty value would represent a delete, so we simply insert the null character.

	balanceKey, err := ctx.GetStub().CreateCompositeKey(balancePrefix, []string{minter, tokenId})
	if err != nil {
		return nil, fmt.Errorf("failed to CreateCompositeKey to balanceKey: %v", err)
	}

	err = ctx.GetStub().PutState(balanceKey, []byte{'\u0000'})
	if err != nil {
		return nil, fmt.Errorf("failed to PutState balanceKey %s: %v", nftBytes, err)
	}

	// Emit the Transfer event
	transferEvent := new(Transfer)
	transferEvent.From = "0x0"
	transferEvent.To = minter
	transferEvent.TokenId = tokenId

	transferEventBytes, err := json.Marshal(transferEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal transferEventBytes: %v", err)
	}

	err = ctx.GetStub().SetEvent("Transfer", transferEventBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to SetEvent transferEventBytes %s: %v", transferEventBytes, err)
	}

	return nft, nil
}

//  Owner string `json:"owner"`
//	Count float64 `json:"count"`
//	AiInfos []AiInfo `json:"aiinfos"`
////////////////////////////////// WalletInfo //////////////////////////////////
func (c *NftChaincode) Wallet(ctx contractapi.TransactionContextInterface, owner string)  error {
	var user = WalletInfo{Owner: owner, Count: 0}
	userAsBytes, _ := json.Marshal(user)

	return ctx.GetStub().PutState(owner, userAsBytes)
}


//  AiTitle string `json:"aititle"`
// 	TokenId string `json:"tokenId"`
// 	AiSum float64 `json:"aisum"`
// 	LearningData float64 `json:"learningdata"`
////////////////////////////////// AiInfo //////////////////////////////////
func (c *NftChaincode) AI(ctx contractapi.TransactionContextInterface, Owner string, aititle string, tokenId string, aisum float64, learningdata string) error {

	// getState User
	userAsBytes, err := ctx.GetStub().GetState(Owner)

	if err != nil {
		return err
	} else if userAsBytes == nil { // no State! error
		return fmt.Errorf("\"Error\":\"User does not exist: " + Owner + "\"")
	}
	// state ok
	owner := WalletInfo{}
	err = json.Unmarshal(userAsBytes, &owner)
	if err != nil {
		return err
	}
	// create rate structure
	newlearningdata, _ := strconv.ParseFloat(learningdata, 64)
	var Info = AiInfo{AiTitle: aititle, TokenId: tokenId, AiSum: aisum, LearningData: newlearningdata}

	owner.AiInfos = append(owner.AiInfos, Info)
	// update to User World state
	userAsBytes, err = json.Marshal(owner)
	if err != nil {
		return fmt.Errorf("failed to Marshaling: %v", err)
	}

	err = ctx.GetStub().PutState(Owner, userAsBytes)
	if err != nil {
		return fmt.Errorf("failed to AddScore: %v", err)
	}
	return nil
}

func (c *NftChaincode) ReadAIinfo(ctx contractapi.TransactionContextInterface, Owner string) (string, error) {
	lookup, err := ctx.GetStub().GetState(Owner)
	if err != nil {
		return "", fmt.Errorf("\"Error\":\"User does not exist: " + Owner + "\"")
	}
	if lookup == nil {
		return "", fmt.Errorf("%s does not exist", Owner)
	}

	return string(lookup[:]), nil

}





////////////////////////////////// User(라벨러) //////////////////////////////////
func (c *NftChaincode) AddUser(ctx contractapi.TransactionContextInterface, username string) error {

	var user = UserInfo{User: username, Sum: 0}
	userAsBytes, _ := json.Marshal(user)

	return ctx.GetStub().PutState(username, userAsBytes)
}

// Add Score (수정 예정)
func (c *NftChaincode) AddScore(ctx contractapi.TransactionContextInterface, username string, project_name string, activity_score string) error {

	// getState User
	userAsBytes, err := ctx.GetStub().GetState(username)

	if err != nil {
		return err
	} else if userAsBytes == nil { // no State! error
		return fmt.Errorf("\"Error\":\"User does not exist: " + username + "\"")
	}
	// state ok
	user := UserInfo{}
	err = json.Unmarshal(userAsBytes, &user)
	if err != nil {
		return err
	}
	// create rate structure
	newScore, _ := strconv.ParseFloat(activity_score, 64)
	var Info = Info{ProjectTitle: project_name, ActivityScore: newScore}

	user.Infos = append(user.Infos, Info)

	// update to User World state
	userAsBytes, err = json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to Marshaling: %v", err)
	}

	err = ctx.GetStub().PutState(username, userAsBytes)
	if err != nil {
		return fmt.Errorf("failed to AddScore: %v", err)
	}
	return nil
}


func (c *NftChaincode) ReadScore(ctx contractapi.TransactionContextInterface, username string) (string, error) {
	lookup, err := ctx.GetStub().GetState(username)
	if err != nil {
		return "", fmt.Errorf("\"Error\":\"User does not exist: " + username + "\"")
	}
	if lookup == nil {
		return "", fmt.Errorf("%s does not exist", username)
	}

	return string(lookup[:]), nil

}
