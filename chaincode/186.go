/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"bytes"
	"math/big"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	//"github.com/hyperledger/fabric-chaincode-go/shim/ext/cid"
	//"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Product struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Stage       string      `json:"status"`
	Signatures  []Signature `json:"signatures"`
	CreatedBy   string      `json:"createdBy"`
	MaxStage    string      `json:"maxStage"`
	Completed   bool        `json:"completed"`
}

type Signature struct {
	Name  string `json:"name"`
	Stage string `json:"stage"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Product
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	var product = Product{Name: "Test Product", Description: "Just a test product to make sure chaincode is running", Stage: "0", CreatedBy: "admin", MaxStage: "3"}
	productAsBytes, _ := json.Marshal(product)
	err := ctx.GetStub().PutState("1", productAsBytes)
	if err != nil {
		return fmt.Errorf("Failed to put to world state. %s", err.Error())
	}
	startingId := new(big.Int).SetInt64(2)
	startingIdAsByte := startingId.Bytes()
	err = ctx.GetStub().PutState("MaxProductId", startingIdAsByte) //set starting id as 2
	if err != nil {
		return fmt.Errorf("Failed to put to world state. %s", err.Error())
	}
	return nil
}

// Task 2
func (s *SmartContract) CreateProduct(ctx contractapi.TransactionContextInterface, name string, description string, maxStage string) error {

	var APIstub = ctx.GetStub()
	var cid = ctx.GetClientIdentity()
	canCreate, found, err := cid.GetAttributeValue("canCreate")
	if err != nil {
		return fmt.Errorf("Error when getting user rights")
	}
	if !found {
		return fmt.Errorf("User does not have right to perform this action.")
	}
	if canCreate != "true" {
		return fmt.Errorf("User does not have right to perform this action.")
	}

	username, found, _ := cid.GetAttributeValue("username")
	var product = Product{Name: name, Description: description, Stage: "0", CreatedBy: username, MaxStage: maxStage, Completed: false}

	productIdAsBytes, _ := APIstub.GetState("MaxProductId")
	productId := new(big.Int).SetBytes(productIdAsBytes)
	productIdAsString := productId.String()

	productAsBytes, _ := json.Marshal(product)

	APIstub.PutState(productIdAsString, productAsBytes)

	increment := new(big.Int).SetInt64(1)
	newProductId := new(big.Int).Add(productId, increment)
	APIstub.PutState("MaxProductId", newProductId.Bytes())

	return nil

}

// Task 2
func (s *SmartContract) QueryAllProducts(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("0", "999999999")
	if err != nil {
		return nil, err
	}

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		product := new(Product)
		_ = json.Unmarshal(queryResponse.Value, product)

		if product.Signatures == nil {
		 		product.Signatures = []Signature{}
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: product}
		results = append(results, queryResult)
	}
	// Task 3
	product := new(Product)
	product.Signatures = []Signature{}
	queryResult := QueryResult{Key: "testing", Record: product}
	results = append(results, queryResult)
	return results, nil

}

func (s *SmartContract) QueryProduct(ctx contractapi.TransactionContextInterface, productId string) (*Product, error) {

		productAsBytes, err := ctx.GetStub().GetState(productId)

		if err != nil {
			return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
		}

		if productAsBytes == nil {
			return nil, fmt.Errorf("%s does not exist", productId)
		}

		product := new(Product)
		_ = json.Unmarshal(productAsBytes, product)

		if product.Signatures == nil {
		 		product.Signatures = []Signature{}
		}

		return product, nil
}

// Task 5
func (s *SmartContract) SignProduct(ctx contractapi.TransactionContextInterface, productId string) error {
	if productId == "" {
		return fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}

	var APIstub = ctx.GetStub()
	var cid = ctx.GetClientIdentity()

	productAsBytes, err := APIstub.GetState(productId)

	if err != nil {
		return fmt.Errorf("Error while retrieving product")
	}

	product := Product{}
	username, _, _ := cid.GetAttributeValue("username")
	json.Unmarshal(productAsBytes, &product)

	canSignStageAsBytes, found, _ := cid.GetAttributeValue("canSignProduct")
	if !found {
		return fmt.Errorf("User cannot sign product")
	}

	if product.Stage == product.MaxStage {
		return fmt.Errorf("Cannot be signed.")
	}
	canSignStage, _ := strconv.Atoi(canSignStageAsBytes)
	currentStage, _ := strconv.Atoi(product.Stage)

	if canSignStage <= currentStage {
		return fmt.Errorf("User does not have rights to sign.")
	}

	product.Stage = strconv.Itoa(currentStage + 1)
	if product.Signatures == nil {
		product.Signatures = []Signature{}
	}

	if product.Stage == product.MaxStage {
		product.Completed = true
	}
	var signature = Signature{Name: username, Stage: strconv.Itoa(currentStage)}
	product.Signatures = append(product.Signatures, signature)

	productAsBytes, _ = json.Marshal(product)
	APIstub.PutState(productId, productAsBytes)

	return nil
}

// Task 8
func (s *SmartContract) GetIncompleteProducts(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	var query string
	query = "{\"selector\":{\"completed\": false}}"
	resultsIterator, err := ctx.GetStub().GetQueryResult(query)
	if err != nil {
		return nil, err
	}

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		product := new(Product)
		_ = json.Unmarshal(queryResponse.Value, product)

		if product.Signatures == nil {
		 		product.Signatures = []Signature{}
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: product}
		results = append(results, queryResult)
	}

	return results, nil
}

// Task 9
func (s *SmartContract) SearchProducts(ctx contractapi.TransactionContextInterface, keyword string) ([]QueryResult, error) {
	var query string
	query = fmt.Sprintf("{\"selector\":{\"name\": {\"$regex\":\"(?i)^.*?%s.*?$\"}}}", keyword)
	resultsIterator, err := ctx.GetStub().GetQueryResult(query)
	if err != nil {
		return nil, err
	}

	results := []QueryResult{}

	for resultsIterator.HasNext() {
			queryResponse, err := resultsIterator.Next()

			if err != nil {
				return nil, err
			}

			product := new(Product)
			_ = json.Unmarshal(queryResponse.Value, product)

			if product.Signatures == nil {
					product.Signatures = []Signature{}
			}

		queryResult := QueryResult{Key: queryResponse.Key, Record: product}
		results = append(results, queryResult)
	}
	return results, nil
}

func buildJSON(resultsIterator shim.StateQueryIteratorInterface, buffer bytes.Buffer) bytes.Buffer {
	buffer.WriteString("[")
	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, _ := resultsIterator.Next()
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"ProductId\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}

	buffer.WriteString("]")
	return buffer
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
