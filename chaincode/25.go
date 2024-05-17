package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

type TokenContract struct {
	contractapi.Contract
}

type Stat struct {
	Supply   uint64 	`json:"supply"`
	MaxSupply  uint64 `json:"max_supply"`
	Issuer string 		`json:"issuer"`
	Symbol  string 		`json:"symbol"`
}

type Account struct {
	Balance uint64 	`json:"balance"`
	Symbol string 	`json:"symbol"`
}

// create token
func (t *TokenContract) Create(ctx contractapi.TransactionContextInterface, issuer string, max_supply uint64, symbol string) error {
	token := Stat{
		Supply:   0,
		MaxSupply:  max_supply,
		Issuer: issuer,
		Symbol:  symbol,
	}
	
	key := []string{t.Contract.GetName(), "stat"}
	ledgerKey, _ := ctx.GetStub().CreateCompositeKey(symbol , key)
	existing, _ := ctx.GetStub().GetState(ledgerKey)
	if existing != nil {
		return fmt.Errorf("token with symbol already exists")
	}
	tokenAsBytes, _ := json.Marshal(token)

	return ctx.GetStub().PutState(ledgerKey, tokenAsBytes)
}

// issue token
func (t *TokenContract) Issue(ctx contractapi.TransactionContextInterface, issuer string, symbol string, to string, quantity uint64, memo string) error {
	token, _ := t.QueryStat(ctx, symbol)
	if token == nil {
		return fmt.Errorf("token with symbol does not exist, create token before issue")
	}
	id, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return fmt.Errorf("get MSPID fail. %s", err.Error())
	}
	if id != issuer {
		return fmt.Errorf("only %s can issue token", token.Issuer)
	}
	token.Supply += quantity
	if token.Supply > token.MaxSupply {
		return fmt.Errorf("quantity exceeds available supply")
	}
	err = t.addBalance(ctx, to, quantity, symbol) 
	if err != nil {
		return fmt.Errorf("add balance fail, %s", err)
	}
	statKey, _ := ctx.GetStub().CreateCompositeKey(symbol, []string{t.Contract.GetName(), "stat"})
	tokenAsBytes, _ := json.Marshal(token)
	return ctx.GetStub().PutState(statKey, tokenAsBytes)
}

// transfer token
func (t *TokenContract) Transfer(ctx contractapi.TransactionContextInterface, from string, to string, quantity uint64, symbol string, memo string) error {
	if from == to {
		return fmt.Errorf("cannot transfer to self");
	}
	id, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return fmt.Errorf("get MSPID fail. %s", err.Error())
	}
	if id != from {
		return fmt.Errorf("miss authority")
	}
	err = t.subBalance(ctx, from, quantity, symbol)
	if err != nil {
		return fmt.Errorf("subtract balance fail, %s", err)
	}
	return t.addBalance(ctx, to, quantity, symbol)
}

// retire 
func (t* TokenContract) Retire(ctx contractapi.TransactionContextInterface , symbol string, quantity uint64, memo string) error {
	token, _ := t.QueryStat(ctx, symbol)
	if token == nil {
		return fmt.Errorf("token with symbol does not exist.")
	}
	id, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return fmt.Errorf("get MSPID fail. %s", err.Error())
	}
	if id != token.Issuer {
		return fmt.Errorf("invalid authority")
	}

	token.Supply -= quantity;
	err = t.subBalance(ctx, id, quantity, symbol)
	if err != nil {
		return fmt.Errorf("subtract balance fail, %s", err)
	}
	statKey, _ := ctx.GetStub().CreateCompositeKey(symbol, []string{t.Contract.GetName(), "stat"})
	tokenAsBytes, _ := json.Marshal(token)
	return ctx.GetStub().PutState(statKey, tokenAsBytes)
}

// query token status
func (t *TokenContract) QueryStat(ctx contractapi.TransactionContextInterface, symbol string) (*Stat, error) {
	key := []string{t.Contract.GetName(), "stat"}
	ledgerKey, _ := ctx.GetStub().CreateCompositeKey(symbol, key)
	tokenAsBytes, err := ctx.GetStub().GetState(ledgerKey)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if tokenAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", symbol)
	}

	stat := new(Stat)
	_ = json.Unmarshal(tokenAsBytes, stat)

	return stat, nil
}

// query balance
func (t *TokenContract) QueryAccount(ctx contractapi.TransactionContextInterface, symbol string, account string) (*Account, error) {
	accountKey, _ := ctx.GetStub().CreateCompositeKey(account, []string{t.Contract.GetName(), "account", symbol})
	accountAsBytes, err := ctx.GetStub().GetState(accountKey)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if accountAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", symbol)
	}

	acc := new(Account)
	_ = json.Unmarshal(accountAsBytes, acc)

	return acc, nil
}

func (t *TokenContract) addBalance(ctx contractapi.TransactionContextInterface, owner string, value uint64, symbol string) error {
	accountKey, _ := ctx.GetStub().CreateCompositeKey(owner, []string{t.Contract.GetName(), "account", symbol})
	account, _ := t.QueryAccount(ctx, symbol, owner)
	if account == nil {
		account = &Account{
			Balance: value,
			Symbol: symbol,
		}
	} else {
		account.Balance += value
	}
	accountAsBytes, _ := json.Marshal(account)
	return ctx.GetStub().PutState(accountKey, accountAsBytes)
}

func (t *TokenContract) subBalance(ctx contractapi.TransactionContextInterface, owner string, value uint64, symbol string) error {
	accountKey, _ := ctx.GetStub().CreateCompositeKey(owner, []string{t.Contract.GetName(), "account", symbol})

	account, _ := t.QueryAccount(ctx, symbol, owner)
	if account == nil {
		return fmt.Errorf("no balance object found")
	} else {
		if account.Balance < value {
			return fmt.Errorf("overdrawn balance")
		}
		account.Balance -= value
	}
	accountAsBytes, _ := json.Marshal(account)
	return ctx.GetStub().PutState(accountKey, accountAsBytes)
}

func main() {

	contract := new(TokenContract)

	contract.Name = "token"
	contract.Info.Version = "0.0.1"

	chaincode, err := contractapi.NewChaincode(contract)

	if err != nil {
		panic(fmt.Sprintf("Error creating chaincode. %s", err.Error()))
	}

	chaincode.Info.Title = "TokenChaincode"
	chaincode.Info.Version = "0.0.1"

	err = chaincode.Start()

	if err != nil {
		panic(fmt.Sprintf("Error starting chaincode. %s", err.Error()))
	}
}
