package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
)

// Technical state errors
// TODO: make errors constant
var (
	ErrStateNotFound = errors.New("State was not found")
)

// PutState takes care of marshalling the state value before storing it in the World State with the specified key
func PutState(ctx contractapi.TransactionContextInterface, key string, v interface{}) (err error) {
	value, err := json.Marshal(v)
	if err != nil {
		return
	}

	err = ctx.GetStub().PutState(key, value)
	return
}

// GetState is a function to get data from the World State that will return an error if the state does not exist and that will unmarshal to a specified empty interface if found
func GetState(ctx contractapi.TransactionContextInterface, key string, v interface{}) (err error) {
	bytes, err := ctx.GetStub().GetState(key)
	if err != nil {
		return
	}
	if bytes == nil {
		err = ErrStateNotFound
		return
	}

	err = json.Unmarshal(bytes, v)
	return
}

// GetStateHistory returns all different states a specific key has stored in the history of the blockchain
func GetStateHistory(ctx contractapi.TransactionContextInterface, key string) (*bytes.Buffer, error) {
	historyIterator, err := ctx.GetStub().GetHistoryForKey(key)
	if err != nil {
		return nil, err
	}
	defer historyIterator.Close()

	// buffer is a JSON array containing historic values for the iou
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for historyIterator.HasNext() {
		var response *queryresult.KeyModification
		response, err = historyIterator.Next()
		if err != nil {
			return nil, err
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		// If it was a delete operation on a given key, then there is no value on the key anymore.
		// So only write the response.Value as-is when it was not a delete operation.
		if !response.IsDelete {
			buffer.WriteString(", \"Value\":")
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	return &buffer, nil
}
