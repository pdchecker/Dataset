package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Product struct {
	ProductId   string     `json:"ProductId"`
	ProductName string     `json:"ProductName"`
	Raws        []*Raw     `json:"Raws"`
	Price       float64    `json:"Price"`
	Status      string     `json:"Status"`
	Description string     `json:"Description"`
	TimeStamps  TimeStamps `json:"TimeStamps"`
	Actors      Actors     `json:"Actors"`
	HashCode    string     `json:"HashCode"`
}

type TimeStamps struct {
	Created     string `json:"Created"`
	Ordered     string `json:"Ordered"`
	Distributed string `json:"Distributed"`
	Received    string `json:"Received"`
	Sold        string `json:"Sold"`
}

type Actors struct {
	ManufacturerId string `json:"ManufacturerId"`
	DistributorId  string `json:"DistributorId"`
	RetailerId     string `json:"RetailerId"`
}

type ProductHistory struct {
	Record    *Product  `json:"Record"`
	TxId      string    `json:"TxId"`
	Timestamp time.Time `json:"Timestamp"`
	IsDelete  bool      `json:"IsDelete"`
}

type Raw struct {
	RawId          string `json:"RawId"`
	RawName        string `json:"RawName"`
	CreatedDate    string `json:"CreateDate"`
	SuppliedDate   string `json:"SuppliedDate"`
	SupplierId     string `json:"SupplierId"`
	ManufacturerId string `json:"ManufacturerId"`
	Status         string `json:"Status"`
	HashCode       string `json:"HashCode"`
}

type RawHistory struct {
	Record    *Raw      `json:"Record"`
	TxId      string    `json:"TxId"`
	Timestamp time.Time `json:"Timestamp"`
	IsDelete  bool      `json:"IsDelete"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	return nil
}

/* 	RAW FUNCTIONS
CREATE RAW
UPDATE RAW
ORDER RAW
SUPPLY RAW

GET RAW
GET RAW HISTORY
*/

func (s *SmartContract) CreateRaw(ctx contractapi.TransactionContextInterface, supplierId string, rawId string, rawName string, createDate string, status string, hashCode string) error {

	var raw = Raw{
		RawId:          rawId,
		RawName:        rawName,
		CreatedDate:    createDate,
		SuppliedDate:   "",
		SupplierId:     supplierId,
		ManufacturerId: "",
		Status:         status,
		HashCode:       hashCode,
	}

	rawAsBytes, _ := json.Marshal(raw)
	return ctx.GetStub().PutState(raw.RawId, rawAsBytes)
}

func (s *SmartContract) UpdateRaw(ctx contractapi.TransactionContextInterface, rawId string, rawName string, status string, hashCode string) error {

	rawAsBytes, _ := ctx.GetStub().GetState(rawId)
	if rawAsBytes == nil {
		return fmt.Errorf("cannot find this raw")
	}

	raw := new(Raw)
	_ = json.Unmarshal(rawAsBytes, raw)
	raw.RawName = rawName
	raw.Status = status
	raw.HashCode = hashCode
	rawAsBytes, _ = json.Marshal(raw)

	return ctx.GetStub().PutState(raw.RawId, rawAsBytes)
}

func (s *SmartContract) OrderRaw(ctx contractapi.TransactionContextInterface, rawId string, manufacturerId string, status string, hashCode string) error {

	rawAsBytes, _ := ctx.GetStub().GetState(rawId)
	if rawAsBytes == nil {
		return fmt.Errorf("cannot find this raw")
	}

	raw := new(Raw)
	_ = json.Unmarshal(rawAsBytes, raw)
	raw.ManufacturerId = manufacturerId
	raw.Status = status
	raw.HashCode = hashCode
	updatedRawAsBytes, _ := json.Marshal(raw)

	return ctx.GetStub().PutState(raw.RawId, updatedRawAsBytes)
}

func (s *SmartContract) SupplyRaw(ctx contractapi.TransactionContextInterface, rawId string, suppliedDate string, status string, hashCode string) error {

	rawAsBytes, _ := ctx.GetStub().GetState(rawId)
	if rawAsBytes == nil {
		return fmt.Errorf("cannot find this raw")
	}

	raw := new(Raw)
	_ = json.Unmarshal(rawAsBytes, raw)
	raw.SuppliedDate = suppliedDate
	raw.Status = status
	raw.HashCode = hashCode
	rawAsBytes, _ = json.Marshal(raw)

	return ctx.GetStub().PutState(raw.RawId, rawAsBytes)
}

func (s *SmartContract) GetRaw(ctx contractapi.TransactionContextInterface, rawId string) (*Raw, error) {

	rawAsBytes, err := ctx.GetStub().GetState(rawId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if rawAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", rawId)
	}

	raw := new(Raw)
	_ = json.Unmarshal(rawAsBytes, &raw)

	return raw, nil
}

func (s *SmartContract) GetRawHistories(ctx contractapi.TransactionContextInterface, rawId string) ([]RawHistory, error) {

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(rawId)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	var histories []RawHistory

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var raw Raw
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &raw)
			if err != nil {
				return nil, err
			}
		} else {
			raw = Raw{
				RawId: rawId,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		rawHistory := RawHistory{
			Record:    &raw,
			TxId:      response.TxId,
			Timestamp: timestamp,
			IsDelete:  response.IsDelete,
		}
		histories = append(histories, rawHistory)
	}

	return histories, nil
}

/* 	PRODUCT FUNCTIONS
CREATE PRODUCT
UPDATE PRODUCT
ORDER PRODUCT
PROVIDE PRODUCT
DELEVER PRODUCT
RECEIVE PRODUCT
SELL PRODUCT
*/

func (s *SmartContract) CreateProduct(ctx contractapi.TransactionContextInterface, manufacturerId string, productId string, productName string, price float64, rawIds string, status string, description string, createdDate string, hashCode string) error {

	timestamps := TimeStamps{}
	timestamps.Created = createdDate
	timestamps.Ordered = ""
	timestamps.Distributed = ""
	timestamps.Received = ""
	timestamps.Sold = ""

	actors := Actors{}
	actors.ManufacturerId = manufacturerId
	actors.DistributorId = ""
	actors.RetailerId = ""

	var raws []*Raw
	rawIdsSplit := strings.Split(rawIds, ",")
	for _, rawId := range rawIdsSplit {
		raw, _ := s.GetRaw(ctx, rawId)
		raws = append(raws, raw)
	}

	var product = Product{
		ProductId:   productId,
		ProductName: productName,
		Price:       price,
		Status:      status,
		Raws:        raws,
		Description: description,
		TimeStamps:  timestamps,
		Actors:      actors,
		HashCode:    hashCode,
	}

	productAsBytes, _ := json.Marshal(product)
	return ctx.GetStub().PutState(product.ProductId, productAsBytes)
}

func (s *SmartContract) UpdateProduct(ctx contractapi.TransactionContextInterface, productId string, productName string, price float64, rawIds string, status string, description string, hashCode string) error {

	productAsBytes, _ := ctx.GetStub().GetState(productId)
	if productAsBytes == nil {
		return fmt.Errorf("cannot find this product")
	}

	product := new(Product)
	_ = json.Unmarshal(productAsBytes, product)

	var raws []*Raw
	rawIdsSplit := strings.Split(rawIds, ",")
	for _, rawId := range rawIdsSplit {
		raw, _ := s.GetRaw(ctx, rawId)
		raws = append(raws, raw)
	}

	product.ProductName = productName
	product.Price = price
	product.Raws = raws
	product.Description = description
	product.Status = status
	product.HashCode = hashCode

	updatedProductAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(product.ProductId, updatedProductAsBytes)
}

func (s *SmartContract) OrderProduct(ctx contractapi.TransactionContextInterface, retailerId string, productId string, orderedDate string, status string, hashCode string) error {

	productAsBytes, _ := ctx.GetStub().GetState(productId)
	if productAsBytes == nil {
		return fmt.Errorf("cannot find this product")
	}

	product := new(Product)
	_ = json.Unmarshal(productAsBytes, product)

	product.Actors.RetailerId = retailerId
	product.TimeStamps.Ordered = orderedDate
	product.Status = status
	product.HashCode = hashCode

	updatedProductAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(product.ProductId, updatedProductAsBytes)
}

func (s *SmartContract) ProvideProduct(ctx contractapi.TransactionContextInterface, manufacturerId string, productId string, distributorId string, status string, hashCode string) error {

	productAsBytes, _ := ctx.GetStub().GetState(productId)
	if productAsBytes == nil {
		return fmt.Errorf("cannot find this product")
	}

	product := new(Product)
	_ = json.Unmarshal(productAsBytes, product)

	product.Actors.DistributorId = distributorId
	product.Status = status
	product.HashCode = hashCode

	updatedProductAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(product.ProductId, updatedProductAsBytes)
}

func (s *SmartContract) DeliveryProduct(ctx contractapi.TransactionContextInterface, distributorId string, productId string, deliveryDate string, status string, hashCode string) error {

	productAsBytes, _ := ctx.GetStub().GetState(productId)
	if productAsBytes == nil {
		return fmt.Errorf("cannot find this product")
	}

	product := new(Product)
	_ = json.Unmarshal(productAsBytes, product)

	product.TimeStamps.Distributed = deliveryDate
	product.Status = status
	product.HashCode = hashCode

	updatedProductAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(product.ProductId, updatedProductAsBytes)
}

func (s *SmartContract) ReceiveProduct(ctx contractapi.TransactionContextInterface, productId string, receivedDate string, status string, hashCode string) error {

	productAsBytes, _ := ctx.GetStub().GetState(productId)
	if productAsBytes == nil {
		return fmt.Errorf("cannot find this product")
	}

	product := new(Product)
	_ = json.Unmarshal(productAsBytes, product)

	product.TimeStamps.Received = receivedDate
	product.Status = status
	product.HashCode = hashCode

	updatedProductAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(product.ProductId, updatedProductAsBytes)
}

func (s *SmartContract) SellProduct(ctx contractapi.TransactionContextInterface, productId string, soldDate string, status string, hashCode string) error {

	productAsBytes, _ := ctx.GetStub().GetState(productId)
	if productAsBytes == nil {
		return fmt.Errorf("cannot find this product")
	}

	product := new(Product)
	_ = json.Unmarshal(productAsBytes, product)

	product.TimeStamps.Sold = soldDate
	product.Status = status
	product.HashCode = hashCode

	updatedProductAsBytes, _ := json.Marshal(product)

	return ctx.GetStub().PutState(product.ProductId, updatedProductAsBytes)
}

func (s *SmartContract) GetProduct(ctx contractapi.TransactionContextInterface, productId string) (*Product, error) {

	productAsBytes, err := ctx.GetStub().GetState(productId)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if productAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", productId)
	}

	product := new(Product)
	_ = json.Unmarshal(productAsBytes, &product)

	return product, nil
}

func (s *SmartContract) GetProductHistories(ctx contractapi.TransactionContextInterface, productId string) ([]ProductHistory, error) {

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(productId)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	defer resultsIterator.Close()

	var histories []ProductHistory

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var product Product
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &product)
			if err != nil {
				return nil, err
			}
		} else {
			product = Product{
				ProductId: productId,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		productHistory := ProductHistory{
			Record:    &product,
			TxId:      response.TxId,
			Timestamp: timestamp,
			IsDelete:  response.IsDelete,
		}
		histories = append(histories, productHistory)
	}

	return histories, nil
}

// MAIN FUNCTION

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fsc chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fsc chaincode: %s", err.Error())
	}
}
