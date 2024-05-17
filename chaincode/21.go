package main

import (
  "encoding/binary"
  "encoding/hex"
  "errors"
  "fmt"
  "crypto/sha256"
  "time"

  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// StoreContract contract for handling writing and reading from the world state
type StoreContract struct {
  contractapi.Contract
}


// Store adds a new key with value to the world state
func (sc *StoreContract) Store(ctx contractapi.TransactionContextInterface, value string) (string, error) {

  // Compute a key for the value, the key depends on the content of the value
  // and the current timestamp
  t := uint64(time.Now().Unix())
  byteT := make([]byte, 8)
  binary.LittleEndian.PutUint64(byteT, t)
  keyPreimage := append([]byte(value),  byteT...)
  keyBytes := sha256.Sum256(keyPreimage)

  key := hex.EncodeToString(keyBytes[:])

  err := ctx.GetStub().PutState(key[:], []byte(value))

  if err != nil {
    return "", errors.New("Unable to interact with world state")
  }

  return key, nil
}

// Retrieve returns the value at key in the world state
func (sc *StoreContract) Retrieve(ctx contractapi.TransactionContextInterface, key string) (string, error) {
    existing, err := ctx.GetStub().GetState(key)

    if err != nil {
        return "", errors.New("Unable to interact with world state")
    }

    if existing == nil {
        return "", fmt.Errorf("Cannot read world state pair with key %s. Does not exist", key)
    }

    return string(existing), nil
}
