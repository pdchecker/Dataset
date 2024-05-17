/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const ownerRecIndex = "owner~id"
const assetRecIndex = "assetId~doctype" //Id+Doctype
const tokenRecIndex = "tokenId~doctype"
const ASSET = "ASSET"
const TOKEN = "TOKEN"
const OWNER = "OWNER"

// SmartContract provides functions for managing a Asset and Token
type SmartContract struct {
	contractapi.Contract
}

type Asset struct {
	ID       string   `json:"id"`
	DocType  string   `json:"doctype"`
	Desc     string   `json:"desc"`
	Name     string   `json:"name"`
	Address  string   `json:"address"`
	Owner    []string `json:"owner"`
	ISActive bool     `json:"isActive"`
}

type Token struct {
	ID             string   `json:"id"`
	Name           string   `json:"name"`
	Symbol         string   `json:"symbol"`
	DocType        string   `json:"doctype"`
	AssetID        string   `json:"assetid"`
	TotalToken     int      `json:"totatCount"`
	AvailableToken int      `json:"avaCount"`
	ReserveToken   int      `json:"resCount"`
	Owner          []string `json:"owner"`
	PricePerToken  float32  `json:"pricePerToken"`
}

type Owner struct {
	Id            string `json:"id"`
	DocType       string `json:"doctype"`
	ParentId      string `json:"parentId"`
	ParentDocType string `json:"parentDocType"`
	Balance       int    `json:"balance"`
}

type Transaction struct {
	Id       string `json:"id"`
	TokenId  string `json:"tokenId"`
	Sender   string `json:"sender"`
	Receiver string `json:"receiver"`
	Amount   int    `json:"amount"`
}

type Record struct {
	AssetRec []Asset
	TokenRec []Token
	OwnerRec Owner
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, assetInputString string) error {
	var assetInput Asset
	err := json.Unmarshal([]byte(assetInputString), &assetInput)
	if err != nil {
		return fmt.Errorf("Error while doing unmarshal of input string : %v", err.Error())
	}
	fmt.Println("Input String :", assetInput)

	//Validate input Parameters
	if len(strings.TrimSpace(assetInput.ID)) == 0 {
		return fmt.Errorf("Asset Id should not be empty")
	}
	if assetInput.DocType != "ASSET" {
		return fmt.Errorf(`Doc Type for Asset should be "ASSET"`)
	}
	for index, owner := range assetInput.Owner {
		if len(strings.TrimSpace(owner)) == 0 {
			return fmt.Errorf("Owner %v is null", index+1)
		}
	}

	//Check if asset ID is present or not
	assetDetailer, err := s.GetDetails(ctx, assetInput.ID, ASSET, "")
	if err != nil {
		return err
	}
	if assetDetailer != nil {
		return fmt.Errorf("Asset already exist with ID : %v", assetInput.ID)
	}

	assetAsBytes, err := json.Marshal(assetInput)
	if err != nil {
		return fmt.Errorf("failed to marshal of Asset records : %v", err.Error())
	}

	//Inserting Asset record
	assetCompositeKey, err := ctx.GetStub().CreateCompositeKey(assetRecIndex, []string{assetInput.ID, assetInput.DocType})
	if err != nil {
		return fmt.Errorf("Error while creating composite key for asset %v and err is :%v", assetInput.ID, err.Error())
	}
	err = ctx.GetStub().PutState(assetCompositeKey, assetAsBytes)
	if err != nil {
		return fmt.Errorf("Error while inserting data to couchDB : %v", err.Error())
	}

	//Creating CompositeKey for owner's record and inserting to ledger
	for _, ownerName := range assetInput.Owner {
		compositeKey, err := ctx.GetStub().CreateCompositeKey(ownerRecIndex, []string{ownerName, assetInput.ID})
		if err != nil {
			return fmt.Errorf("failed to create composite key for owner %v and err is :%v", ownerName, err.Error())
		}

		ownerRec := Owner{
			Id:            ownerName,
			DocType:       "OWNER",
			ParentId:      assetInput.ID,
			ParentDocType: assetInput.DocType,
		}

		ownerBytes, err := json.Marshal(ownerRec)
		if err != nil {
			return fmt.Errorf("Error while doing Marshal : %v", err.Error())
		}
		err = ctx.GetStub().PutState(compositeKey, ownerBytes)
		if err != nil {
			return fmt.Errorf("Error while inserting record for owner %v and error is : ", ownerName, err.Error())
		}
	}
	return nil
}

func (s *SmartContract) GetDetails(ctx contractapi.TransactionContextInterface, id string, docType string, owner string) (interface{}, error) {
	var recIndex string
	var recIndexParam []string
	if docType == "ASSET" {
		recIndex = assetRecIndex
		recIndexParam = []string{id, "ASSET"}
	} else if docType == "TOKEN" {
		recIndex = tokenRecIndex
		recIndexParam = []string{id, "TOKEN"}

	} else if docType == "OWNER" {
		recIndex = ownerRecIndex
		recIndexParam = []string{owner, id}
	}
	fmt.Println("RecIndex : ", recIndex)
	fmt.Println("RecIndex Param : ", recIndexParam)

	compositeKey, err := ctx.GetStub().CreateCompositeKey(recIndex, recIndexParam)
	if err != nil {
		return nil, fmt.Errorf("failed to create composite key for id %v and err is :%v", id, err.Error())
	}
	fmt.Println("Composite Key : ", compositeKey)
	objectBytes, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		return nil, fmt.Errorf("Failed to read data from world state %s", err.Error())
	}
	if objectBytes == nil {
		return nil, nil
	}

	fmt.Println("objectBytes : ", string(objectBytes))

	switch docType {
	case "ASSET":
		var asset Asset
		err = json.Unmarshal(objectBytes, &asset)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal asset data: %s", err.Error())
		}

		fmt.Println("Asset Rec : ", asset)
		return asset, nil
	case "TOKEN":
		var token Token
		err = json.Unmarshal(objectBytes, &token)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal token data: %s", err.Error())
		}
		fmt.Println("token Rec : ", token)
		return token, nil
	case "OWNER":
		var owner Owner
		err = json.Unmarshal(objectBytes, &owner)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal owner data: %s", err.Error())
		}
		fmt.Println("owner Rec : ", owner)
		return owner, nil
	default:
		return nil, fmt.Errorf("Unsupported docType: %s", docType)
	}

}

/*Total Asset and Token holding by Owner*/
func (s *SmartContract) TotalAssetPerOwnerWithoutQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]Owner, error) {
	//queryString := fmt.Sprintf(`{"selector":{"id":"%s"}}`, owner)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("Error found during GetQueryResult :%s", err.Error())
	}
	defer resultsIterator.Close()

	var ownerTotalAsset []Owner
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Error found query Result :%s", err.Error())
		}
		fmt.Println("queryResult : ", queryResult)
		ownerBytes, err := ctx.GetStub().GetState(queryResult.Key)
		if err != nil {
			return nil, fmt.Errorf("Error found During get state :%s", err.Error())
		}
		fmt.Println("ownerBytes : ", ownerBytes)

		var ownerRecord Owner
		err = json.Unmarshal(ownerBytes, &ownerRecord)
		if err != nil {
			return nil, fmt.Errorf("Error found During unmarshal :%s", err.Error())
		}
		fmt.Println("ownerRecord : ", ownerRecord)
		ownerTotalAsset = append(ownerTotalAsset, ownerRecord)
		fmt.Println("ownerTotalAsset : ", ownerTotalAsset)
		fmt.Println("********************************")
	}
	return ownerTotalAsset, nil
}

/**************************************************/

func (s *SmartContract) MintToken(ctx contractapi.TransactionContextInterface, tokenInputString string) error {
	var tokenInput Token
	err := json.Unmarshal([]byte(tokenInputString), &tokenInput)
	if err != nil {
		return fmt.Errorf("Error while doing unmarshal of input string : %v", err.Error())
	}
	fmt.Println("Input String :", tokenInput)

	//Validate input Parameters
	if len(strings.TrimSpace(tokenInput.ID)) == 0 {
		return fmt.Errorf("Token Id should not be empty")
	}
	if len(strings.TrimSpace(tokenInput.Name)) == 0 {
		return fmt.Errorf("Token name should not be empty")
	}
	if len(strings.TrimSpace(tokenInput.Symbol)) == 0 {
		return fmt.Errorf("Token Symbol should not be empty")
	}
	if tokenInput.DocType != "TOKEN" {
		return fmt.Errorf(`Doc Type for Asset should be "TOKEN"`)
	}
	if tokenInput.TotalToken <= 0 {
		return fmt.Errorf("Total Token should be +ve")
	}
	if tokenInput.PricePerToken <= 0 {
		return fmt.Errorf("Price per token should be +ve")
	}
	if float32(tokenInput.ReserveToken) > (float32(tokenInput.TotalToken) * 0.75) {
		fmt.Errorf("Reserved token is greater than %d%% of total token", 75)
	}

	//Check if token ID is present or not
	tokenDetailer, err := s.GetDetails(ctx, tokenInput.ID, TOKEN, "")
	if err != nil {
		return err
	}
	if tokenDetailer != nil {
		return fmt.Errorf("Token already exist with ID : %v", tokenInput.ID)
	}

	//Check if asset ID is present or not
	assetDetailer, err := s.GetDetails(ctx, tokenInput.AssetID, ASSET, "")
	if err != nil {
		return err
	}
	if assetDetailer == nil {
		return fmt.Errorf("Asset does not exist with ID : %v", tokenInput.AssetID)
	}
	asset, ok := assetDetailer.(Asset)
	if !ok {
		return fmt.Errorf("Failed to convert Detailer to Asset type")
	}

	tokenInput.Owner = asset.Owner

	//Calculating reserve token
	if tokenInput.ReserveToken == 0 {
		tokenInput.ReserveToken = int(float32(tokenInput.TotalToken) * 0.75)
	}

	//calculating avaliable token
	tokenInput.AvailableToken = tokenInput.TotalToken - tokenInput.ReserveToken

	tokenAsBytes, err := json.Marshal(tokenInput)
	if err != nil {
		return fmt.Errorf("failed to Marshal of Token records : %v", err.Error())
	}

	//Inserting Token record
	tokenCompositeKey, err := ctx.GetStub().CreateCompositeKey(tokenRecIndex, []string{tokenInput.ID, tokenInput.DocType})
	if err != nil {
		return fmt.Errorf("failed to create composite key for token %v and err is :%v", tokenInput.ID, err.Error())
	}
	err = ctx.GetStub().PutState(tokenCompositeKey, tokenAsBytes)
	if err != nil {
		return fmt.Errorf("failed to insert data to couchDB : %v", err.Error())
	}

	//Creating CompositeKey for owner's record and inserting to ledger
	tokenPerUser := tokenInput.AvailableToken / len(tokenInput.Owner)
	for _, ownerName := range tokenInput.Owner {
		compositeKey, err := ctx.GetStub().CreateCompositeKey(ownerRecIndex, []string{ownerName, tokenInput.ID})
		if err != nil {
			return fmt.Errorf("failed to create composite key for owner %v and err is :%v", ownerName, err.Error())
		}

		if tokenInput.AvailableToken-tokenPerUser >= tokenPerUser {
			tokenInput.AvailableToken = tokenInput.AvailableToken - tokenPerUser
		} else {
			tokenPerUser = tokenInput.AvailableToken
		}

		ownerRec := Owner{
			Id:            ownerName,
			DocType:       "OWNER",
			ParentId:      tokenInput.ID,
			ParentDocType: tokenInput.DocType,
			Balance:       tokenPerUser,
		}

		ownerBytes, err := json.Marshal(ownerRec)
		if err != nil {
			return fmt.Errorf("failed to marshal ownerRec : %v", err.Error())
		}
		err = ctx.GetStub().PutState(compositeKey, ownerBytes)
		if err != nil {
			return fmt.Errorf("failed to insert record for owner %v and error is : ", ownerName, err.Error())
		}
	}
	return nil
}

func (s *SmartContract) BalanceOf(ctx contractapi.TransactionContextInterface, ownerToken string) (int, error) {
	ownerInput := struct {
		Owner   string `json:"owner"`
		TokenId string `json:"tokenid"`
	}{}

	err := json.Unmarshal([]byte(ownerToken), &ownerInput)
	if err != nil {
		return 0, fmt.Errorf("Error while doing unmarshal of input string : %v", err.Error())
	}
	fmt.Println("Input String :", ownerInput)

	ownerDetailer, err := s.GetDetails(ctx, ownerInput.TokenId, OWNER, ownerInput.Owner)
	if err != nil {
		return 0, err
	}
	if ownerDetailer == nil {
		return 0, fmt.Errorf("failed to get owner details for tokenID : %v", ownerInput.TokenId)
	}

	owner, ok := ownerDetailer.(Owner)
	if !ok {
		return 0, fmt.Errorf("Failed to convert Detailer to Owner type")
	}

	return owner.Balance, nil
}

func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, transferTxnInput string) error {
	transferTxn := struct {
		TokenId          string `json:"tokenid"`
		Sender           string `json:"sender"`
		Receiver         string `json:"receiver"`
		AmountToTransfer int    `json:"amountToTransfer"`
	}{}

	err := json.Unmarshal([]byte(transferTxnInput), &transferTxn)
	if err != nil {
		return fmt.Errorf("failed to unmarshal of input string : %v", err.Error())
	}
	fmt.Println("Input String :", transferTxn)

	//Validate tokenid
	tokenDetailer, err := s.GetDetails(ctx, transferTxn.TokenId, TOKEN, "")
	if err != nil {
		return err
	}
	if tokenDetailer == nil {
		return fmt.Errorf("Token does not exist with ID : %v", transferTxn.TokenId)
	}
	fmt.Println("Token Details :", tokenDetailer)

	//Get Sender details
	senderDetailer, err := s.GetDetails(ctx, transferTxn.TokenId, OWNER, transferTxn.Sender)
	if err != nil {
		return err
	}
	if senderDetailer == nil {
		return fmt.Errorf("failed to get sender details for tokenID : %v", transferTxn.TokenId)
	}
	senderDetail, ok := senderDetailer.(Owner)
	if !ok {
		return fmt.Errorf("Failed to convert Detailer to Sender type")
	}
	fmt.Println("Sender Details :", senderDetail)

	//Validate sender balance
	if senderDetail.Balance < transferTxn.AmountToTransfer {
		return fmt.Errorf("Insufficient Balance to transfer")
	}

	//Getting Receiver
	var isReceiverExist bool = false
	var receiverDetail Owner
	receiverDetailer, err := s.GetDetails(ctx, transferTxn.TokenId, OWNER, transferTxn.Receiver)
	if err != nil {
		return err
	}
	if receiverDetailer != nil {
		isReceiverExist = true
		receiverDetail, ok = receiverDetailer.(Owner)
		if !ok {
			return fmt.Errorf("Failed to convert Detailer to Receiver type")
		}
	}
	fmt.Println("Receiver Details :", receiverDetail)
	fmt.Println("Receiver exist :", isReceiverExist)

	senderCompositeKey, err := ctx.GetStub().CreateCompositeKey(ownerRecIndex, []string{transferTxn.Sender, transferTxn.TokenId})
	if err != nil {
		return fmt.Errorf("failed to create composite key for sender %v and err is :%v", transferTxn.Sender, err.Error())
	}

	receiverCompositeKey, err := ctx.GetStub().CreateCompositeKey(ownerRecIndex, []string{transferTxn.Receiver, transferTxn.TokenId})
	if err != nil {
		return fmt.Errorf("failed to create composite key for receiver %v and err is :%v", transferTxn.Receiver, err.Error())
	}

	//create Transaction record
	txID := ctx.GetStub().GetTxID()
	txn := Transaction{
		Id:       txID,
		TokenId:  transferTxn.TokenId,
		Sender:   transferTxn.Sender,
		Receiver: transferTxn.Receiver,
		Amount:   transferTxn.AmountToTransfer,
	}

	txnBytes, err := json.Marshal(txn)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(txID, txnBytes)
	if err != nil {
		return err
	}

	if isReceiverExist {
		//Update balance for Sender
		updatedSenderBalance, err := sub(senderDetail.Balance, transferTxn.AmountToTransfer)
		if err != nil {
			return err
		}
		senderDetail.Balance = updatedSenderBalance
		fmt.Println("Sender Updated balance :", senderDetail.Balance)

		updatedReceiverBalance, err := add(receiverDetail.Balance, transferTxn.AmountToTransfer)
		if err != nil {
			return err
		}
		receiverDetail.Balance = updatedReceiverBalance
		fmt.Println("Receiver Updated balance :", receiverDetail.Balance)

		senderBytes, err := json.Marshal(senderDetail)
		if err != nil {
			return err
		}
		receiverBytes, err := json.Marshal(receiverDetail)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(senderCompositeKey, senderBytes)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(receiverCompositeKey, receiverBytes)
		if err != nil {
			return err
		}

	} else {
		//Update balance for Sender
		updatedSenderBalance, err := sub(senderDetail.Balance, transferTxn.AmountToTransfer)
		if err != nil {
			return err
		}
		senderDetail.Balance = updatedSenderBalance
		fmt.Println("Sender Updated balance :", senderDetail.Balance)

		receiverDetail = Owner{
			Id:            transferTxn.Receiver,
			DocType:       "OWNER",
			ParentId:      transferTxn.TokenId,
			ParentDocType: "TOKEN",
			Balance:       transferTxn.AmountToTransfer,
		}

		senderBytes, err := json.Marshal(senderDetail)
		if err != nil {
			return err
		}
		receiverBytes, err := json.Marshal(receiverDetail)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(senderCompositeKey, senderBytes)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(receiverCompositeKey, receiverBytes)
		if err != nil {
			return err
		}
	}
	fmt.Println("*****************************")
	return nil
}

// sub two number checking for overflow
func sub(b int, q int) (int, error) {

	// Check overflow
	var diff int
	diff = b - q

	if (diff > b) == (b >= 0 && q >= 0) {
		return 0, fmt.Errorf("Math: Subtraction overflow occurred  %d - %d", b, q)
	}

	return diff, nil
}

// add two number checking for overflow
func add(b int, q int) (int, error) {

	// Check overflow
	var sum int
	sum = q + b

	if (sum < q) == (b >= 0 && q >= 0) {
		return 0, fmt.Errorf("Math: addition overflow occurred %d + %d", b, q)
	}

	return sum, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
