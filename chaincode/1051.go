package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type BasicRepository struct {
	Stub shim.ChaincodeStubInterface
}

func (repo *BasicRepository) Find(key string) []byte {
	bytes, _ := repo.Stub.GetState(key)
	return bytes
}

func (repo *BasicRepository) FindAndUnmarshal(key string, dest interface{}) error {
	bytes := repo.Find(key)
	if bytes == nil || len(bytes) == 0 {
		return nil
	}

	err := json.Unmarshal([]byte(bytes), dest)
	if err != nil {
		return fmt.Errorf("Failed to unmarshall obj", key)
	}
	return nil
}

func (repo *BasicRepository) Exists(key string) (bool, error) {
	return repo.Find(key) != nil, nil
}
