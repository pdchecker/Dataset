/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
        "encoding/json"
        "fmt"
        "strconv"

        "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a book
type SmartContract struct {
        contractapi.Contract
}

// Book describes basic details of what makes up a book


type Model struct {
	Name    string  `json:"name"`
	Grads   map[string][][]float64  `json:"grads"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
        Key    string `json:"Key"`
        Record *Model
}

// InitLedger adds a base set of books to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
        models := []Model{
        	Model{Name: "steam",
        	Grads: make(map[string][][]float64),
        	},
        }

        for i, model := range models {
                modelAsBytes, _ := json.Marshal(model)
                err := ctx.GetStub().PutState("MODEL"+strconv.Itoa(i), modelAsBytes)

                if err != nil {
                        return fmt.Errorf("Failed to put to world state. %s", err.Error())
                }
        }

        return nil
}

// CreateBook adds a new book to the world state with given details
func (s *SmartContract) CreateModel(ctx contractapi.TransactionContextInterface, modelNumber string, modelName string, frame map[string][]int) error {
        model := Model{
                Name:	modelName,
                Grads:	make(map[string][][]float64),
        }
        
        for k, v := range frame{
        	model.Grads[k] = make([][]float64, v[0])
		for i := range model.Grads[k]{
			model.Grads[k][i] = make([]float64, v[1])
		}
        }

        modelAsBytes, _ := json.Marshal(model)

        return ctx.GetStub().PutState(modelNumber, modelAsBytes)
}

func (s *SmartContract) GetGrad(ctx contractapi.TransactionContextInterface, modelNumber string, layerName string) ([][]float64, error) {
        modelAsBytes, err := ctx.GetStub().GetState(modelNumber)
        if err != nil {
                return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
        }

        if modelAsBytes == nil {
                return nil, fmt.Errorf("%s does not exist", modelNumber)
        }

        model := new(Model)
        _ = json.Unmarshal(modelAsBytes, model)

        return model.Grads[layerName], nil
}

func (s *SmartContract) PutGrad(ctx contractapi.TransactionContextInterface, modelNumber string, layerName string, position int, grad [][]float64) ([][]float64, error) {
	model, err := s.QueryModel(ctx, modelNumber)
	if err != nil {
                return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
        }
        if model == nil {
                return nil, fmt.Errorf("%s does not exist", modelNumber)
        }
        copy(model.Grads[layerName][position: position + len(grad)], grad)
	modelAsBytes, _ := json.Marshal(model)
        ctx.GetStub().PutState(modelNumber, modelAsBytes)
        return model.Grads[layerName], nil
}

func (s *SmartContract) QueryModel(ctx contractapi.TransactionContextInterface, modelNumber string) (*Model, error) {
        modelAsBytes, err := ctx.GetStub().GetState(modelNumber)
        if err != nil {
                return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
        }

        if modelAsBytes == nil {
                return nil, fmt.Errorf("%s does not exist", modelNumber)
        }

        model := new(Model)
        _ = json.Unmarshal(modelAsBytes, model)

        return model, nil
}

// QueryAllBooks returns all books found in world state
func (s *SmartContract) QueryAllModels(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
        startKey := ""
        endKey := ""

        resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

        if err != nil {
                return nil, err
        }
        defer resultsIterator.Close()

        results := []QueryResult{}

        for resultsIterator.HasNext() {
                queryResponse, err := resultsIterator.Next()

                if err != nil {
                        return nil, err
                }

                model := new(Model)
                _ = json.Unmarshal(queryResponse.Value, model)


                queryResult := QueryResult{Key: queryResponse.Key, Record: model}
                results = append(results, queryResult)
        }

        return results, nil
}


func main() {

        chaincode, err := contractapi.NewChaincode(new(SmartContract))

        if err != nil {
                fmt.Printf("Error create fabai chaincode: %s", err.Error())
                return
        }

        if err := chaincode.Start(); err != nil {
                fmt.Printf("Error starting fabai chaincode: %s", err.Error())
        }
}
