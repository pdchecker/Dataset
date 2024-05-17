package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bridge/base"
	"github.com/bridge/bridge"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Details struct {
	Id        string `json:"Id"`
	Key       string `json:"Key"`
	User      string `json:"User"`
	Amount    string `json:"Value"`
	Message   string `json:"Message"`
	Signature string `json:"Signature"`
}

type TxDetails struct {
	From   string `json:"From"`
	To     string `json:"To"`
	Action string `json:"Action"`
	Value  string `json:"Value"`
}

type Event struct {
	Id    string `json:"Id"`
	User  string `json:"User"`
	Value string `json:"Value"`
}

const DepositPrefix = "depostix~prefix"
const WithdrawPrefix = "withdraw~prefix"
const TokenContract = "token"

// deposit
func (s *SmartContract) MintAndTransfer(ctx contractapi.TransactionContextInterface, data string) (interface{}, error) {

	var dataJson Details

	err := json.Unmarshal([]byte(data), &dataJson)
	if err != nil {
		return nil, err
	}
	if id, err := ctx.GetStub().GetState(dataJson.Id); err != nil || id != nil {
		return nil, fmt.Errorf(fmt.Sprintf("key Id is already exists %s,\n %s", err, string(id)))
	}

	l, err := hex.DecodeString(dataJson.Key)
	if err != nil {
		return nil, err
	}
	hash := sha256.New()
	hash.Write(l)
	md := hash.Sum(nil)
	mdStr := hex.EncodeToString(md)

	if c := strings.Compare(dataJson.Id, mdStr); c != 0 {
		return nil, fmt.Errorf("keys are not matching")
	}

	_, err = bridge.Bridge(ctx, "MintAndTransfer", dataJson.User, dataJson.Amount, dataJson.Message, dataJson.Signature, dataJson.Id)
	if err != nil {
		return nil, err
	}

	if err = ctx.GetStub().PutState(dataJson.Id, []byte("True")); err != nil {
		return nil, err
	}

	response := &TxDetails{
		From:   "Bridge",
		To:     dataJson.User,
		Action: "Mint",
		Value:  dataJson.Amount,
	}

	// set event
	mintEevent := &Event{Id: dataJson.Id, User: dataJson.User, Value: dataJson.Amount}
	mintEeventJSON, err := json.Marshal(mintEevent)
	if err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	err = ctx.GetStub().SetEvent("MintAndTransfer", mintEeventJSON)
	if err != nil {
		return nil, fmt.Errorf(base.EventError)
	}

	resp, _ := json.Marshal(response)
	_ = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), resp)

	return string(resp), nil
}

func (s *SmartContract) BurnFrom(ctx contractapi.TransactionContextInterface, data string) (interface{}, error) {
	var dataJson Details

	err := json.Unmarshal([]byte(data), &dataJson)
	if err != nil {
		return nil, err
	}

	if id, err := ctx.GetStub().GetState(dataJson.Id); err != nil || id != nil {
		return nil, fmt.Errorf("key Id is already exists %s, %s", err, string(id))
	}

	_, err = bridge.Bridge(ctx, "BurnFrom", dataJson.User, dataJson.Amount, dataJson.Message, dataJson.Signature, dataJson.Id)
	if err != nil {
		return nil, err
	}

	if err = ctx.GetStub().PutState(dataJson.Id, []byte("True")); err != nil {
		return nil, err
	}

	response := &TxDetails{
		From:   dataJson.User,
		To:     "0x0",
		Action: "BurnFrom",
		Value:  dataJson.Amount,
	}

	// set event
	burnEevent := &Event{Id: dataJson.Id, User: dataJson.User, Value: dataJson.Amount}
	burnEeventtJSON, err := json.Marshal(burnEevent)
	if err != nil {
		return nil, fmt.Errorf(base.JSONParseError)
	}
	err = ctx.GetStub().SetEvent("BurnFrom", burnEeventtJSON)
	if err != nil {
		return nil, fmt.Errorf(base.EventError)
	}

	resp, _ := json.Marshal(response)
	_ = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), resp)

	return string(resp), nil
}
