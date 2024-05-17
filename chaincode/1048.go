package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

// WriteLedger
func WriteLedger(obj interface{}, stub shim.ChaincodeStubInterface, objectType string, keys []string) error {
	//create primary composite key
	var key string
	if val, err := stub.CreateCompositeKey(objectType, keys); err != nil {
		return errors.New(fmt.Sprintf("%s-Error creating composite primary key %s", objectType, err))
	} else {
		key = val
	}
	//object serialization
	bytes, err := json.Marshal(obj)
	if err != nil {
		return errors.New(fmt.Sprintf("%s-Failed to serialize json data error: %s", objectType, err))
	}

	//push object to the Ledger
	if err := stub.PutState(key, bytes); err != nil {
		return errors.New(fmt.Sprintf("%s-Error writing to the blockchain ledger: %s", objectType, err))
	}
	return nil
}

// Delete asset from the ledger
func DelLedger(stub shim.ChaincodeStubInterface, objectType string, keys []string) error {
	//create primary composite key
	var key string
	if val, err := stub.CreateCompositeKey(objectType, keys); err != nil {
		return errors.New(fmt.Sprintf("%s-Error creating primary composite key %s", objectType, err))
	} else {
		key = val
	}
	//push object to the Ledger
	if err := stub.DelState(key); err != nil {
		return errors.New(fmt.Sprintf("%s-error while deleting from the ledger: %s", objectType, err))
	}
	return nil
}

// returns all assets using their composite key attributes ( can be used to retrieve all assets of a particular type using a simple key: for example "companyMat" as argument to retrieve all companies)
func GetStateByPartialCompositeKeys(stub shim.ChaincodeStubInterface, objectType string, keys []string) (results [][]byte, err error) {
	if len(keys) == 0 {

		resultIterator, err := stub.GetStateByPartialCompositeKey(objectType, keys)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("%s-Error while retrieving data: %s", objectType, err))
		}
		defer resultIterator.Close()

		for resultIterator.HasNext() {
			val, err := resultIterator.Next()
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%s-Error in returned data: %s", objectType, err))
			}

			results = append(results, val.GetValue())
		}
	} else {
		for _, v := range keys {

			key, err := stub.CreateCompositeKey(objectType, []string{v})
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%s-error while creating  composite key: %s", objectType, err))
			}

			bytes, err := stub.GetState(key)
			if err != nil {
				return nil, errors.New(fmt.Sprintf("%s-Error getting data : %s", objectType, err))
			}

			if bytes != nil {
				results = append(results, bytes)
			}
		}
	}

	return results, nil
}
