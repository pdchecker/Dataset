package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"strconv"
)

type OrderContract struct {
	contractapi.Contract
}

type Order struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	OrderType   string        `json:"orderType"`
	Country     string        `json:"country"`
	Description string        `json:"description"`
	Owner       string        `json:"owner"`
	Goods       []*GoodsAsset `json:"goods"`
}

type GoodsAsset struct {
	GoodId   string `json:"goodId"`
	Quantity string `json:"quantity"`
	Unit     string `json:"unit"`
}

func (c *OrderContract) CreateOrder(ctx contractapi.TransactionContextInterface, id, orderType, country, name, description string, goods []*GoodsAsset) error {
	log.Printf("CreateOrder -> INFO: creating order id: %s, orderType: %s, country: %s, name: %s, description: %s...\n", id, orderType, country, name, description)
	log.Println("CreateOrder -> INFO: checking if order already exists...")
	exists, err := c.OrderExists(ctx, id)

	if err != nil {
		return err
	}
	if exists {
		log.Println("CreateOrder -> ERROR: order already exists")
		return fmt.Errorf("order with id %s exists", id)
	}

	log.Println("CreateOrder -> INFO: validating requested goods...")
	for _, good := range goods {
		log.Printf("CreateOrder -> INFO: retreiving good id: %s", good.GoodId)
		goodJson, err := ctx.GetStub().GetState(good.GoodId)
		if err != nil {
			log.Printf("CreateOrder -> ERROR: error retreiving good id: %s\n", good.GoodId)
			return err
		}
		if goodJson == nil {
			log.Println("CreateOrder -> ERROR: good do not exist")
			return fmt.Errorf("specified good do not exist")
		}
		log.Println("CreateOrder -> INFO: unmarshalling good")
		var goodRule *Good
		err = json.Unmarshal(goodJson, &goodRule)
		if err != nil {
			log.Println("CreateOrder -> ERROR: failed to unmarshall data")
			return err
		}

		log.Println("CreateOrder -> INFO: validating good quantity...")

		quantity, err := strconv.ParseFloat(good.Quantity, 64)
		if err != nil {
			log.Println("CreateOrder -> ERROR: invalid quantity format")
			return nil
		}
		err = nil
		var limit float64
		if orderType == "import" {
			limit, err = strconv.ParseFloat(goodRule.ImportLimit, 64)
		} else if orderType == "export" {
			limit, err = strconv.ParseFloat(goodRule.ExportLimit, 64)
		} else {
			log.Println("CreateOrder -> ERROR: invalid orderType")
			return fmt.Errorf("invalid order type")
		}

		if err != nil {
			log.Println("CreateOrder -> ERROR: failed to parse limit")
			return err
		}

		log.Println("CreateOrder -> INFO: validating quantity limit")

		if quantity > limit {
			log.Println("CreateOrder -> ERROR: limit exceeded")
			return fmt.Errorf("export/import limit exceeded")
		}
	}
	log.Println("CreateOrder -> INFO: retrieving client identity")
	clientId, err := getSubmittingClientIdentity(ctx)
	if err != nil {
		log.Println("CreateOrder -> ERROR: failed to retrieve client identity")
		return err
	}

	order := Order{
		ID:          id,
		Name:        name,
		Country:     country,
		OrderType:   orderType,
		Description: description,
		Owner:       clientId,
		Goods:       goods,
	}

	log.Println("CreateOrder -> INFO: creating json data...")

	orderJson, err := json.Marshal(order)
	if err != nil {
		log.Println("CreateOrder -> ERROR: failed to marshal data")
		return err
	}

	log.Println("CreateOrder -> INFO: putting data to world state")
	err = ctx.GetStub().PutState(id, orderJson)
	if err != nil {
		log.Println("CreateOrder -> ERROR: failed to put data to world state")
	}
	return err
}

func (c *OrderContract) OrderExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	log.Printf("CreateOrder -> INFO: checking order id: %s existence", id)
	orderJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		log.Println("CreateOrder -> ERROR: failed to read from world state")
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return orderJSON != nil, nil
}

func (c *OrderContract) ReadOrder(ctx contractapi.TransactionContextInterface, id string) (*Order, error) {
	log.Printf("ReadOrder -> INFO: reading order id: %s\n", id)
	orderJson, err := ctx.GetStub().GetState(id)
	if err != nil {
		log.Println("ReadOrder -> ERROR: failed to read data from world state")
		return nil, err
	}
	if orderJson == nil {
		log.Printf("ReadOrder -> INFO: order id: %s not found\n", id)
		return nil, fmt.Errorf("order id: %s not found", id)
	}
	var order *Order
	log.Println("ReadOrder -> INFO: unmarshalling data")
	err = json.Unmarshal(orderJson, &order)
	if err != nil {
		log.Println("ReadOrder -> ERROR: failed to unmarshal data")
		return nil, err
	}
	log.Printf("ReadOrder -> INFO: order %s received", string(orderJson))
	return order, nil
}

// getSubmittingClientIdentity utils to get client id from credentials.
func getSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	log.Println("getSubmittingClientIdentity -> INFO: retrieving client id")
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		log.Println("getSubmittingClientIdentity -> ERROR: failed to read client id")
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	log.Println("getSubmittingClientIdentity -> INFO: decoding client id")
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		log.Println("getSubmittingClientIdentity -> ERROR: failed to decode id")
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	log.Println("getSubmittingClientIdentity -> INFO: retrieved client id: " + string(decodeID))
	return string(decodeID), nil
}

// ClearData delete data for simulations. Not part of features.
func (c *OrderContract) ClearData(ctx contractapi.TransactionContextInterface, startKey, endKey string) error {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return err
	}
	defer resultsIterator.Close()
	var keys []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var order Order
		err = json.Unmarshal(queryResponse.Value, &order)
		if err != nil {
			return err
		}
		keys = append(keys, order.ID)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(keys); i++ {
		err = ctx.GetStub().DelState(keys[i])
		if err != nil {
			return err
		}
	}
	return nil
}
