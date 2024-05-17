package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const GENESIS_MINT_AMOUNT = 100000000

// TradeChaincode ...
type TradeChaincode struct {
	contractapi.Contract
}

// MeowType ...
type MeowType struct {
	Type   string `json:"type"`
	Amount uint32 `json:"amount"`
}

// TransferType ...
type TransferType struct {
	Type      string `json:"type"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    uint32 `json:"amount"`
	Timestamp string `json:"timestamp"`
}

// RewardType ...
type RewardType struct {
	Type   string `json:"type"`
	To     string `json:"to"`
	Amount uint32 `json:"amount"`
}

//BuyAIModelType ...
type BuyAIModelType struct {
	Timestamp string       `json:"timestamp"`
	History   []RewardType `json:"history"`
}

// AIModelType ...
type AIModelType struct {
	Type             string   `json:"type"`
	Name             string   `json:"name"`
	Language         string   `json:"language"`
	Price            uint32   `json:"price"`
	Owner            string   `json:"owner"`
	Score            uint32   `json:"score"`
	Downloaded       uint32   `json:"downloaded"`
	Description      string   `json:"description"`
	VerificationOrgs []string `json:"verification_orgs"`
	Contents         string   `json:"contents`
	Timestamp        string   `json:"timestamp"`
}

// InitLedger ...
func (t *TradeChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	isInitBytes, err := ctx.GetStub().GetState("isInit")
	if err != nil {
		return fmt.Errorf("failed GetState('isInit')")
	} else if isInitBytes == nil {
		initMeow := MeowType{
			Type:   "GenesisMint",
			Amount: GENESIS_MINT_AMOUNT,
		}

		initMeowAsBytes, err := json.Marshal(initMeow)
		if err != nil {
			return fmt.Errorf("failed to json.Marshal(). %v", err)
		}

		ctx.GetStub().PutState(makeMeowKey("bank"), initMeowAsBytes)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}

		ctx.GetStub().PutState("isInit", []byte{0x1})
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}

		return nil
	} else {
		return fmt.Errorf("already initialized")
	}
}

// GetCurrentMeow ...
func (t *TradeChaincode) GetCurrentMeow(ctx contractapi.TransactionContextInterface, uid string) (*MeowType, error) {
	currentMeow := &MeowType{}
	currentMeowAsBytes, err := ctx.GetStub().GetState(makeMeowKey(uid))
	if err != nil {
		return nil, err
	} else if currentMeowAsBytes == nil {
		currentMeow.Type = "CurrentMeowAmount"
		currentMeow.Amount = 0
	} else {
		err = json.Unmarshal(currentMeowAsBytes, currentMeow)
		if err != nil {
			return nil, err
		}
	}

	return currentMeow, nil
}

// Transfer ...
func (t *TradeChaincode) Transfer(ctx contractapi.TransactionContextInterface, from string, to string, amount uint32, timestamp string, meowType string) error {
	// INSERT Transfer history
	transferMeow := TransferType{
		Type:      "transfer",
		From:      from,
		To:        to,
		Amount:    amount,
		Timestamp: timestamp,
	}

	transferMeowAsBytes, err := json.Marshal(transferMeow)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal(). %v", err)
	}
	transferMeowKey := makeFromToMeowKey(from, to, timestamp)
	ctx.GetStub().PutState(transferMeowKey, transferMeowAsBytes)

	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	// UPDATE Current From Meow
	currentFromMeow, err := t.GetCurrentMeow(ctx, from)
	if err != nil {
		return fmt.Errorf("failed to get current meow. %v", err)
	}

	if currentFromMeow.Amount < amount {
		return fmt.Errorf("meow is lacking.. %v", err)
	}

	currentFromMeow.Amount -= amount

	currentFromMeowAsBytes, err := json.Marshal(currentFromMeow)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal(). %v", err)
	}
	ctx.GetStub().PutState(makeMeowKey(from), currentFromMeowAsBytes)

	// UPDATE Current To Meow
	currentToMeow, err := t.GetCurrentMeow(ctx, to)
	if err != nil {
		return fmt.Errorf("failed to get current meow. %v", err)
	}

	currentToMeow.Amount += amount

	currentToMeowAsBytes, err := json.Marshal(currentToMeow)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal(). %v", err)
	}
	ctx.GetStub().PutState(makeMeowKey(to), currentToMeowAsBytes)

	// TODO
	// Transfer
	// Before amount (from, to)
	// After amount (from, to)
	return nil
}

// GetModel ...
func (t *TradeChaincode) GetModel(ctx contractapi.TransactionContextInterface, modelKey string) (*AIModelType, error) {
	funNameAsBytes := []byte("GetAIModelInfoWithKey")
	argAsBytes := []byte(modelKey)
	args := [][]byte{funNameAsBytes, argAsBytes}
	result := ctx.GetStub().InvokeChaincode("ai-model", args, "ai-model")
	aiModelInfo := &AIModelType{}
	json.Unmarshal(result.Payload, aiModelInfo)

	return aiModelInfo, nil
}

// BuyModel ...
func (t *TradeChaincode) BuyModel(ctx contractapi.TransactionContextInterface, uid string, modelKey string, price uint32, timestamp string) error {
	checkBuyAIModelAsBytes, err := ctx.GetStub().GetState(makeBuyAIModelKey(uid, modelKey))
	if err != nil {
		return fmt.Errorf("failed to get BuyAiModel. %v", err)
	} else if checkBuyAIModelAsBytes != nil {
		return fmt.Errorf("already buy model ...")
	}

	aiModel := &AIModelType{}
	aiModel, err = t.GetModel(ctx, modelKey)
	fmt.Println(aiModel)

	if price != aiModel.Price {
		return fmt.Errorf("the price mismatch in blockchain ..")
	}

	currentMeow, err := t.GetCurrentMeow(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to get current meow. %v", err)
	}

	if currentMeow.Amount < price {
		return fmt.Errorf("meow is lacking.. %v", err)
	}

	// NOTE
	// modelKey -> AI_uid_modelName_version(unique)
	seller := strings.Split(modelKey, "_")[1]
	verificationOrgs := aiModel.VerificationOrgs
	fmt.Println(verificationOrgs)

	if price%10 != 0 {
		return fmt.Errorf("only available in units of 10 meow")
	}

	income := price * 8 / 10
	verifyReward := price * 1 / 10
	manageReward := price * 1 / 10

	t.Transfer(ctx, uid, seller, income, timestamp, "income")
	t.Transfer(ctx, uid, "admin", manageReward, timestamp, "reward")

	buyAIModel := BuyAIModelType{
		Timestamp: timestamp,
		History: []RewardType{
			{
				Type:   "income",
				To:     seller,
				Amount: income,
			},
			{
				Type:   "manageReward",
				To:     "admin",
				Amount: manageReward,
			},
		},
	}

	for _, org := range verificationOrgs {
		// TODO
		// divide verifyReward
		t.Transfer(ctx, uid, org, verifyReward, timestamp, "reward")
		verifyRewardHistory := RewardType{
			Type:   "verifyReward",
			To:     org,
			Amount: verifyReward,
		}
		buyAIModel.History = append(buyAIModel.History, verifyRewardHistory)
	}

	currentMeow.Amount -= price

	currentMeowAsBytes, err := json.Marshal(currentMeow)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal(). %v", err)
	}
	ctx.GetStub().PutState(makeMeowKey(uid), currentMeowAsBytes)

	buyAIModelAsBytes, err := json.Marshal(buyAIModel)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal(). %v", err)
	}
	ctx.GetStub().PutState(makeBuyAIModelKey(uid, modelKey), buyAIModelAsBytes)

	return nil
}

// TODO
// func getIsBuyModel
// func getAsset

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*TransferType, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var transferHistorys []*TransferType
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var transferHistory TransferType
		err = json.Unmarshal(queryResult.Value, &transferHistory)
		if err != nil {
			return nil, err
		}
		transferHistorys = append(transferHistorys, &transferHistory)
	}

	return transferHistorys, nil
}

func (t *TradeChaincode) getAllHistory(uid string) error {
	return nil
}

func (t *TradeChaincode) GetQueryHistory(ctx contractapi.TransactionContextInterface, uid string) ([]*TransferType, error) {
	queryString := fmt.Sprintf(`{"selector":{"type":"transfer","$or":[{"from":"%s"},{"to":"%s"}]}}`, uid, uid)
	return getQueryResultForQueryString(ctx, queryString)
}

func (t *TradeChaincode) GetQueryFromHistory(ctx contractapi.TransactionContextInterface, uid string) ([]*TransferType, error) {
	queryString := fmt.Sprintf(`{"selector":{"type":"transfer","from":"%s"}}`, uid)
	return getQueryResultForQueryString(ctx, queryString)
}

func (t *TradeChaincode) GetQueryToHistory(ctx contractapi.TransactionContextInterface, uid string) ([]*TransferType, error) {
	queryString := fmt.Sprintf(`{"selector":{"type":"transfer","to":"%s"}}`, uid)
	return getQueryResultForQueryString(ctx, queryString)
}

func makeBuyAIModelKey(uid string, model string) string {
	var sb strings.Builder

	sb.WriteString("B_")
	sb.WriteString(uid)
	sb.WriteString("_")
	sb.WriteString(model)

	return sb.String()
}

func makeMeowKey(uid string) string {
	var sb strings.Builder

	sb.WriteString("M_")
	sb.WriteString(uid)

	return sb.String()
}

func makeFromToMeowKey(from string, to string, timestamp string) string {
	var sb strings.Builder

	sb.WriteString("F_")
	sb.WriteString(from)
	sb.WriteString("_T_")
	sb.WriteString(to)
	sb.WriteString("_")
	sb.WriteString(timestamp)

	return sb.String()
}
