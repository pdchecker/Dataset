/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}
type ProductState struct {
	SellerID        string `json:"SellerID"`
	ProductID       string `json:"ProductID"`
	State           string `json:"State"`
	BuyerID         string `json:"BuyerID"`
	BuyDate         uint64 `json:"BuyDate"`
	Deliver_address string `json:"Deliver_address"`
	DeliverID       string `json:"DeliverID"`
	Reject_info     string `json:"Reject_info"`
	Update_date     uint64 `json:"Update_date"`
}
type sellerInfo struct {
	SellerID string          `json:"SellerID"`
	Score    string          `json:"Score"`
	Historys []sellerHistory `json:"Historys"`
}
type sellerHistory struct {
	ProductID   string `json:"ProductID"`
	BuyerID     string `json:"SellerID"`
	State       string `json:"State"`
	Update_date uint64 `json:"Update_date"`
}

type buyerInfo struct {
	BuyerID  string         `json:"BuyerID"`
	Score    string         `json:"Score"`
	Historys []buyerHistory `json:"Historys"`
}
type buyerHistory struct {
	ProductID   string `json:"ProductID"`
	SellerID    string `json:"SellerID"`
	State       string `json:"State"`
	Update_date uint64 `json:"Update_date"`
}

//product
func (s *SmartContract) InitProduct(ctx contractapi.TransactionContextInterface, ProductID string, SellerID string, State string,
	BuyerID string, BuyDate uint64, Deliver_address string,
	Reject_info string, Update_date uint64) error {

	var product = ProductState{ProductID: ProductID, SellerID: SellerID, State: State, BuyerID: BuyerID, BuyDate: BuyDate, Deliver_address: Deliver_address, Reject_info: Reject_info, Update_date: Update_date}
	userAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(ProductID, userAsBytes)
}

//state
func (s *SmartContract) UpdateProductState(ctx contractapi.TransactionContextInterface, ProductID string, State string) error {

	userAsBytes, err := ctx.GetStub().GetState(ProductID)

	if err != nil {
		return err
	} else if userAsBytes == nil {
		return fmt.Errorf("User does not exist " + ProductID + "/")
	}

	productState := ProductState{}
	err = json.Unmarshal(userAsBytes, &productState)
	if err != nil {
		return err
	}

	newState := State

	productState.State = newState
	productState.ProductID = ProductID

	userAsBytes, err = json.Marshal(productState)

	if err != nil {
		return fmt.Errorf("failed to Marshaling:%v", err)

	}

	err = ctx.GetStub().PutState(ProductID, userAsBytes)

	if err != nil {
		return fmt.Errorf("failed to AddRating %v", err)

	}
	return nil
}

func (s *SmartContract) UpdateDeliverState(ctx contractapi.TransactionContextInterface, ProductID string, BuyerID string, DeliverID string, State string) error {

	productAsBytes, err := ctx.GetStub().GetState(ProductID)

	if err != nil {
		return err
	} else if productAsBytes == nil {
		return fmt.Errorf("User does not exist " + ProductID + "/")
	}

	productState := ProductState{}
	err = json.Unmarshal(productAsBytes, &productState)
	if err != nil {
		return err
	}

	productState.State = State
	productState.DeliverID = DeliverID

	productAsBytes, err = json.Marshal(productState)

	if err != nil {
		return fmt.Errorf("failed to Marshaling:%v", err)

	}

	err = ctx.GetStub().PutState(ProductID, productAsBytes)

	if err != nil {
		return fmt.Errorf("failed to AddRating %v", err)

	}
	return nil

}

func (s *SmartContract) GetProductState(ctx contractapi.TransactionContextInterface, ProductID string) (string, error) {

	StateAsBytes, err := ctx.GetStub().GetState(ProductID)

	if err != nil {
		return "", fmt.Errorf("failed to read from world State,%s", err.Error())
	}

	if StateAsBytes == nil {
		return "", fmt.Errorf("%s  does not exist", ProductID)

	}

	return string(StateAsBytes[:]), nil

}

//seller function
func (s *SmartContract) InitSellerInfo(ctx contractapi.TransactionContextInterface, SellerID string, Score string) error {

	var seller = sellerInfo{SellerID: SellerID, Score: Score}
	userAsBytes, _ := json.Marshal(seller)

	return ctx.GetStub().PutState(SellerID, userAsBytes)
}

func (s *SmartContract) UpdateSellerHistory(ctx contractapi.TransactionContextInterface, SellerID string, ProductID string, BuyerID string, State string, Update_date uint64) error {

	sellerAsBytes, err := ctx.GetStub().GetState(SellerID)

	if err != nil {
		return err
	} else if sellerAsBytes == nil {
		return fmt.Errorf("User does not exist" + SellerID + "/")
	}

	seller := sellerInfo{}
	err = json.Unmarshal(sellerAsBytes, &seller)

	if err != nil {
		return err
	}

	sellerHistory := sellerHistory{ProductID: ProductID, BuyerID: BuyerID, State: State, Update_date: Update_date}

	seller.Historys = append(seller.Historys, sellerHistory)

	sellerAsBytes, err = json.Marshal(seller)

	if err != nil {
		return fmt.Errorf("failed to Marshaling:%v", err)

	}

	err = ctx.GetStub().PutState(SellerID, sellerAsBytes)

	if err != nil {
		return fmt.Errorf("failed to AddContract %v", err)

	}

	return nil
}

func (s *SmartContract) GetSellerHistory(ctx contractapi.TransactionContextInterface, SellerID string) (string, error) {

	userAsBytes, err := ctx.GetStub().GetState(SellerID)

	if err != nil {
		return "", fmt.Errorf("failed to read from world State,%s", err.Error())
	}

	if userAsBytes == nil {
		return "", fmt.Errorf("%s  does not exist", SellerID)

	}

	return string(userAsBytes[:]), nil

}

//deliver

func (s *SmartContract) InitBuyerInfo(ctx contractapi.TransactionContextInterface, BuyerID string, Score string) error {

	var Buyer = buyerInfo{BuyerID: BuyerID, Score: Score}
	BuyerAsBytes, _ := json.Marshal(Buyer)

	return ctx.GetStub().PutState(BuyerID, BuyerAsBytes)
}

func (s *SmartContract) UpdateBuyerHistory(ctx contractapi.TransactionContextInterface, BuyerID string, ProductID string, SellerID string, State string, Update_date uint64) error {

	buyerAsBytes, err := ctx.GetStub().GetState(BuyerID)

	if err != nil {
		return err
	} else if buyerAsBytes == nil {
		return fmt.Errorf("User does not exist" + BuyerID + "/")
	}

	buyer := buyerInfo{}
	err = json.Unmarshal(buyerAsBytes, &buyer)

	if err != nil {
		return err
	}

	buyerHistory := buyerHistory{ProductID: ProductID, SellerID: SellerID, State: State, Update_date: Update_date}

	buyer.Historys = append(buyer.Historys, buyerHistory)

	buyerAsBytes, err = json.Marshal(buyer)

	if err != nil {
		return fmt.Errorf("failed to Marshaling:%v", err)

	}

	err = ctx.GetStub().PutState(BuyerID, buyerAsBytes)

	if err != nil {
		return fmt.Errorf("failed to AddContract %v", err)

	}

	return nil
}

func (s *SmartContract) GetBuyerHistory(ctx contractapi.TransactionContextInterface, BuyerID string) (string, error) {

	buyerAsBytes, err := ctx.GetStub().GetState(BuyerID)

	if err != nil {
		return "", fmt.Errorf("failed to read from world State,%s", err.Error())
	}

	if buyerAsBytes == nil {
		return "", fmt.Errorf("%s  does not exist", BuyerID)

	}

	return string(buyerAsBytes[:]), nil

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
