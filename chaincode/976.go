package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type DecapolisContract struct {
	// contains filtered or unexported fields
	contractapi.Contract
}

type Product struct {
	id    string `json:"id"`
	name  string `json:"name"`
	lot   string `json:"lot"`
	block string `json:"block"`
}

type ErrorType struct {
}

func (err *ErrorType) Error() string {
	return "Custom Error"
}

func (s *DecapolisContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	products := []Product{
		{id: "1",
			name:  "1",
			lot:   "1",
			block: "1",
		},
	}

	for _, product := range products {
		productJSON, err := json.Marshal(product)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(product.id, productJSON)
		if err != nil {
			return fmt.Errorf("Failed to put to world state. %v", err)
		}
	}
	fmt.Println("Instantiated Successfully")
	return nil
}
