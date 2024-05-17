package main

// package name이 main이 아니면 인스턴스화가 되지 않음

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

type SmartContract struct {
	contractapi.Contract
}

type Goods struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Price    int    `json:"price"`
	WalletID string `json:"walletId"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"key"`
	Record *Goods `json:"record"`
}

type GoodsKey struct {
	Key string
	Idx int
}

type Wallet struct {
	Name  string `json:"name"`
	ID    string `json:"id"`
	Token int    `json:"token"`
}

func (s *SmartContract) InitWallet(ctx contractapi.TransactionContextInterface) error {
	customer := Wallet{Name: "Hyper", ID: "1Q2W3E4R", Token: 100}
	seller := Wallet{Name: "Ledger", ID: "5T6Y7U8I", Token: 200}

	customerBytes, _ := json.Marshal(customer)

	if err := ctx.GetStub().PutState(customer.ID, customerBytes); err != nil {
		return err
	}

	sellerBytes, _ := json.Marshal(seller)

	if err := ctx.GetStub().PutState(seller.ID, sellerBytes); err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) GetWallet(ctx contractapi.TransactionContextInterface, walletId string) (*Wallet, error) {
	walletBytes, err := ctx.GetStub().GetState(walletId)
	if err != nil {
		return nil, err
	}

	var wallet Wallet

	_ = json.Unmarshal(walletBytes, &wallet)

	return &wallet, nil
}

func (s *SmartContract) generateKey(ctx contractapi.TransactionContextInterface) (*GoodsKey, error) {
	isFirst := false

	goodsKeyBytes, err := ctx.GetStub().GetState("latestKey")
	if err != nil {
		return nil, err
	}

	goodsKey := GoodsKey{}

	fmt.Printf("goodsKey: %v\n", string(goodsKeyBytes))

	_ = json.Unmarshal(goodsKeyBytes, &goodsKey)

	if len(goodsKey.Key) == 0 || goodsKey.Key == "" {
		isFirst = true
		goodsKey.Key = "GS"
	}

	if !isFirst {
		goodsKey.Idx = goodsKey.Idx + 1
	}

	return &goodsKey, nil
}

func (s *SmartContract) SetGoods(ctx contractapi.TransactionContextInterface, name, category, price, walletId string) error {
	goodsKey, err := s.generateKey(ctx)
	if err != nil {
		return err
	}

	keyIdx := goodsKey.Idx

	fmt.Printf("key : %v, idx : %v\n", goodsKey.Key, keyIdx)

	goodsPrice, _ := strconv.Atoi(price)

	goods := Goods{Name: name, Category: category, Price: goodsPrice, WalletID: walletId}

	goodsBytes, _ := json.Marshal(goods)
	keyString := goodsKey.Key + fmt.Sprint(keyIdx)

	fmt.Printf("goodsKey: %v\n", keyString)

	if err := ctx.GetStub().PutState(keyString, goodsBytes); err != nil {
		return err
	}

	goodsKeyBytes, _ := json.Marshal(goodsKey)

	if err := ctx.GetStub().PutState("latestKey", goodsKeyBytes); err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) GetAllGoods(ctx contractapi.TransactionContextInterface) ([]*QueryResult, error) {
	goodsKeyBytes, _ := ctx.GetStub().GetState("latestKey")

	goodsKey := GoodsKey{}

	_ = json.Unmarshal(goodsKeyBytes, &goodsKey)

	idxStr := fmt.Sprint(goodsKey.Idx + 1)

	startKey := "GS0"
	endKey := goodsKey.Key + idxStr

	resultsIter, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}

	defer resultsIter.Close()

	results := []*QueryResult{}

	for resultsIter.HasNext() {
		queryResp, err := resultsIter.Next()
		if err != nil {
			return nil, err
		}

		goods := Goods{}
		_ = json.Unmarshal(queryResp.Value, &goods)

		queryResult := QueryResult{Key: queryResp.Key, Record: &goods}
		results = append(results, &queryResult)
	}

	return results, nil
}

func (s *SmartContract) PurchaseGoods(ctx contractapi.TransactionContextInterface, customerId, goodsKey string) error {
	goodsBytes, err := ctx.GetStub().GetState(goodsKey)
	if err != nil {
		return err
	}

	if goodsBytes == nil {
		return errors.New("goods entity not found")
	}

	goods := Goods{}
	_ = json.Unmarshal(goodsBytes, &goods)

	sellerId := goods.WalletID
	goodsPrice := goods.Price

	customerBytes, err := ctx.GetStub().GetState(customerId)
	if err != nil {
		return err
	}

	if customerBytes == nil {
		return errors.New("customer entity not found")
	}

	customer := Wallet{}
	_ = json.Unmarshal(customerBytes, &customer)

	sellerBytes, err := ctx.GetStub().GetState(sellerId)
	if err != nil {
		return err
	}

	if sellerBytes == nil {
		return errors.New("seller entity not found")
	}

	seller := Wallet{}
	_ = json.Unmarshal(sellerBytes, &seller)

	customer.Token = customer.Token - goodsPrice
	seller.Token = seller.Token + goodsPrice
	goods.WalletID = customerId

	updatedCustomerBytes, _ := json.Marshal(customer)
	updatedSellerBytes, _ := json.Marshal(seller)
	updateGoodsBytes, _ := json.Marshal(goods)

	if err := ctx.GetStub().PutState(customerId, updatedCustomerBytes); err != nil {
		return err
	}
	if err := ctx.GetStub().PutState(sellerId, updatedSellerBytes); err != nil {
		return err
	}
	if err := ctx.GetStub().PutState(goodsKey, updateGoodsBytes); err != nil {
		return err
	}

	fmt.Printf("customer Token: %v, seller Token: %v\n", customer.Token, seller.Token)

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Printf("Error create chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting chaincode: %s", err.Error())
	}
}
