package main

/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */
import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
	"time"
)

// each object should contains its own contract means its own invoke and init ledger
// StockContract contract for handling writing and reading from the world state
type StockContract struct {
	contractapi.Contract
}

// Card is a struct for dividend info and owners
type Card struct {
	// TraderID added for more readable and query
	// make it string for compositekey
	TraderID string `json:"traderID"`
	// Count how many share owns this traderID
	Count       int    `json:"count"`
	StockSymbol string `json:"stockSymbol"`
	// Dividend toman/share
	Dividend int `json:"dividend"`
	// DividendPayments the plan of paying dividend
	DividendPayments []DividendPayment `json:"dividendPayment"`
}
type QueryCard struct {
	Cards []Card `json:"cards"`
}

// DividendPayment status of time plan of dividend pays
type DividendPayment struct {
	// Percentage of dividend count pays
	Percentage float32 `json:"percentage"`
	// PDate date of payment
	PDate time.Time `json:"pDate"`
	//// for changing status of NOT NEEDED I THINK IF  PAYING IS OUT OF BLOCKCHAIN JUST QUERY ISSUER
	//Paid bool `json:"paid"`
}

// InitLedger create all cards with TraderID and Issuer and other attr nil
func (sc *StockContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// add a reserved card for query and adding issuer and trader
	card := Card{
		TraderID:         "afrouz",
		Count:            0,
		StockSymbol:      "afrouz",
		Dividend:         0,
		DividendPayments: nil,
	}
	err := sc.AddCard(ctx, card.TraderID, string(card.Count), card.StockSymbol, string(card.Dividend))
	if err != nil {
		return fmt.Errorf("failed to init Cards %s", err.Error())
	}
	return nil

}

//todo: new trader register for adding cards, new issuer register, for dividend update it?

// AddCard calls putState of chaincode to add card maybe create a string to push in worldstate
func (sc *StockContract) AddCard(ctx contractapi.TransactionContextInterface, traderID string, countString string, stocksymbol string, dividendString string) error {
	count, _ := strconv.Atoi(countString)
	dividend, _ := strconv.Atoi(dividendString)
	indexName := "trader~stocksymbol"
	// NOTE: watch out for dividentpaymnet is empty
	card := Card{TraderID: traderID, Count: count, StockSymbol: stocksymbol, Dividend: dividend, DividendPayments: make([]DividendPayment, 0)}
	cardAsByte, _ := json.Marshal(card)
	cardKey, _ := ctx.GetStub().CreateCompositeKey(indexName, []string{card.TraderID, card.StockSymbol})
	err := ctx.GetStub().PutState(cardKey, cardAsByte)
	if err != nil {
		return fmt.Errorf("failed to put Card to world state %s", err.Error())
	}
	return nil
}

// QueryByTrader get all cards assigned to traderID input
func (sc *StockContract) QueryByTrader(ctx contractapi.TransactionContextInterface, traderID string) (QueryCard, error) {
	traderIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("trader~stocksymbol", []string{traderID})
	if err != nil {
		return QueryCard{}, fmt.Errorf("failed to get Cards of trader %s", err.Error())
	}
	defer traderIterator.Close()
	cards := QueryCard{}
	for traderIterator.HasNext() {
		response, err := traderIterator.Next()
		if err != nil {
			return QueryCard{}, err
		}
		card := Card{}
		_ = json.Unmarshal(response.Value, &card)
		cards.Cards = append(cards.Cards, card)

	}
	return cards, nil

}

// QueryByStockSymbol get all cards assigned to stock symbol input
func (sc *StockContract) QueryByStockSymbol(ctx contractapi.TransactionContextInterface, stockSymbol string) (QueryCard, error) {
	queryString := fmt.Sprintf("{\"selector\":{\"stockSymbol\":\"%s\"}}", stockSymbol)
	ssymbolIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return QueryCard{}, fmt.Errorf("failed to get Cards of stock symbol %s", err.Error())
	}
	defer ssymbolIterator.Close()
	cards := QueryCard{}
	for ssymbolIterator.HasNext() {
		response, err := ssymbolIterator.Next()
		if err != nil {
			return QueryCard{}, err
		}
		card := Card{}
		_ = json.Unmarshal(response.Value, &card)
		cards.Cards = append(cards.Cards, card)

	}
	return cards, nil

}

//todo: how dividendPayment added to buyer or remove from seller

// Trade exchange traded count of a stock symbol from seller to a buyer
func (sc *StockContract) Trade(ctx contractapi.TransactionContextInterface, seller string, buyer string, countString string, stockSymbol string) error {
	indexName := "trader~stocksymbol"
	count, _ := strconv.Atoi(countString)
	sellerCardKey, _ := ctx.GetStub().CreateCompositeKey(indexName, []string{seller, stockSymbol})
	sellerResponse, err := ctx.GetStub().GetState(sellerCardKey)
	if err != nil {
		return fmt.Errorf("failed to get Card from world state %s", err.Error())
	}
	if sellerResponse == nil {
		return fmt.Errorf("not such a card in worldstate")
	}
	fmt.Printf("seller card\n %s \n", sellerResponse)
	sellerResponseCard := Card{}
	_ = json.Unmarshal(sellerResponse, &sellerResponseCard)
	// add it here cause its refrence and maybe deleted in calling update count
	//dividendPaymentCard := sellerResponseCard.DividendPayments
	sellerResponseCard.Count -= count
	if sellerResponseCard.Count <= 0 {
		return fmt.Errorf("can't update count the count will be negative ")
	}
	// if not deleted buying another card maybe cause problem
	if sellerResponseCard.Count == 0 {
		// todo: delete dividend
	} else {
		fmt.Printf("Can Trader from seller \n")
		cardAsByte, _ := json.Marshal(sellerResponseCard)
		err = ctx.GetStub().PutState(sellerCardKey, cardAsByte)
		if err != nil {
			return fmt.Errorf("failed to put Card to world state %s", err.Error())
		}
	}
	buyerCardKey, _ := ctx.GetStub().CreateCompositeKey(indexName, []string{buyer, stockSymbol})
	buyerResponse, err := ctx.GetStub().GetState(buyerCardKey)
	fmt.Printf("buyer card\n %s \n", buyerResponse)

	if err != nil {
		return fmt.Errorf("failed to get Card from world state %s", err.Error())
	}
	if buyerResponse == nil {
		fmt.Printf("buyer in nil card %s \n", buyerResponse)

	} else {
		fmt.Print("buyer can buy the Card \n")
		buyerResponseCard := Card{}
		_ = json.Unmarshal(buyerResponse, &buyerResponseCard)
		buyerResponseCard.Count += count
		cardAsByte, _ := json.Marshal(buyerResponseCard)
		err = ctx.GetStub().PutState(buyerCardKey, cardAsByte)
		if err != nil {
			return fmt.Errorf("failed to put Card to world state %s", err.Error())
		}
	}
	return nil
}

// update count dividend  and dividendPayment of selected card(traderID,stockSymbol)
func (sc *StockContract) UpdateFields(ctx contractapi.TransactionContextInterface, traderID string, stockSymbol string, countString string, dividendString string, payment string) error {
	indexName := "trader~stocksymbol"
	cardKey, _ := ctx.GetStub().CreateCompositeKey(indexName, []string{traderID, stockSymbol})
	response, err := ctx.GetStub().GetState(cardKey)
	if err != nil {
		return fmt.Errorf("failed to get Card from world state %s", err.Error())
	}
	if response == nil {
		return fmt.Errorf("not such a card in worldstate")
	}
	responseCard := Card{}
	_ = json.Unmarshal(response, &responseCard)
	dividendPayment := DividendPayment{}
	_ = json.Unmarshal([]byte(payment), &dividendPayment)
	responseCard.Count, _ = strconv.Atoi(countString)
	responseCard.Dividend, _ = strconv.Atoi(dividendString)
	responseCard.DividendPayments = append(responseCard.DividendPayments, dividendPayment)
	cardAsByte, _ := json.Marshal(responseCard)
	err = ctx.GetStub().PutState(cardKey, cardAsByte)
	if err != nil {
		return fmt.Errorf("failed to put Card to world state %s", err.Error())
	}
	return nil
}

// UpdateDividend Update given trader and stocksymbol update dividend  and dividendPayment field
func (sc *StockContract) UpdateDividend(ctx contractapi.TransactionContextInterface, traderID string, stockSymbol string, dividendString string, payment string) error {
	indexName := "trader~stocksymbol"
	cardKey, _ := ctx.GetStub().CreateCompositeKey(indexName, []string{traderID, stockSymbol})
	response, err := ctx.GetStub().GetState(cardKey)
	if err != nil {
		return fmt.Errorf("failed to get Card from world state %s", err.Error())
	}
	if response == nil {
		return fmt.Errorf("not such a card in worldstate")
	}
	responseCard := Card{}
	_ = json.Unmarshal(response, &responseCard)
	dividendPayment := DividendPayment{}
	_ = json.Unmarshal([]byte(payment), &dividendPayment)
	responseCard.Dividend, _ = strconv.Atoi(dividendString)
	responseCard.DividendPayments = append(responseCard.DividendPayments, dividendPayment)
	cardAsByte, _ := json.Marshal(responseCard)
	err = ctx.GetStub().PutState(cardKey, cardAsByte)
	if err != nil {
		return fmt.Errorf("failed to put Card to world state %s", err.Error())
	}
	return nil
}

// AddDividendPayment func will append given slice of payments to the card
func (sc *StockContract) AddDividendPayment(ctx contractapi.TransactionContextInterface, traderID string, stockSymbol string, payment string) error {
	indexName := "trader~stocksymbol"
	cardKey, _ := ctx.GetStub().CreateCompositeKey(indexName, []string{traderID, stockSymbol})
	response, err := ctx.GetStub().GetState(cardKey)
	if err != nil {
		return fmt.Errorf("failed to get Card from world state %s", err.Error())
	}
	if response == nil {
		return fmt.Errorf("not such a card in worldstate")
	}
	responseCard := Card{}
	_ = json.Unmarshal(response, &responseCard)
	dividendPayment := DividendPayment{}
	_ = json.Unmarshal([]byte(payment), &dividendPayment)
	responseCard.DividendPayments = append(responseCard.DividendPayments, dividendPayment)
	cardAsByte, _ := json.Marshal(responseCard)
	err = ctx.GetStub().PutState(cardKey, cardAsByte)
	if err != nil {
		return fmt.Errorf("failed to put Card to world state %s", err.Error())
	}
	return nil
}

func (sc *StockContract) DeleteDividendPayment(ctx contractapi.TransactionContextInterface, traderID string, stockSymbol string) error {
	indexName := "trader~stocksymbol"
	cardKey, _ := ctx.GetStub().CreateCompositeKey(indexName, []string{traderID, stockSymbol})
	response, err := ctx.GetStub().GetState(cardKey)
	if err != nil {
		return fmt.Errorf("failed to get Card from world state %s", err.Error())
	}
	if response == nil {
		return fmt.Errorf("not such a card in worldstate")
	}
	responseCard := Card{}
	_ = json.Unmarshal(response, &responseCard)
	responseCard.DividendPayments = make([]DividendPayment, 0)
	cardAsByte, _ := json.Marshal(responseCard)
	err = ctx.GetStub().PutState(cardKey, cardAsByte)
	if err != nil {
		return fmt.Errorf("failed to put Card to world state %s", err.Error())
	}
	return nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(StockContract))

	if err != nil {
		fmt.Printf("Error create stock chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting stock chaincode: %s", err.Error())
	}
}
