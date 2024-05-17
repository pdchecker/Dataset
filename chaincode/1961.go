/*
 * Copyright IBM Corp All Rights Reserved
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	// Get the args from the transaction proposal
	args := stub.GetStringArgs()
	if len(args) != 2 {
		return shim.Error("Incorrect arguments. Expecting a key and a value")
	}

	// Set up any variables or assets here by calling stub.PutState()

	// We store the key and the value on the ledger
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to create asset: %s", args[0]))
	}
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "set" {
		result, err = set(stub, args)
	} else if fn == "getsaccname" {
		result, err = getsaccname(stub)
	} else if fn == "getsaccnamewritename" {
		result, err = getsaccnamewritename(stub)
	} else if fn == "setsaccname" {
		result, err = setsaccname(stub, args)
	} else if fn == "setsaccnamediffchannel" {
		result, err = setsaccnamediffchannel(stub, args)
	} else { // assume 'get' even if fn is nil
		result, err = get(stub, args)
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func set(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 2 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and a value")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: %s", args[0])
	}
	return args[1], nil
}

// Get returns the value of the specified asset key
func get(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return "", fmt.Errorf("Failed to get asset: %s with error: %s", args[0], err)
	}
	if value == nil {
		return "", fmt.Errorf("Asset not found: %s", args[0])
	}
	return string(value), nil
}

// Get "name" from chaincode sacc
func getsaccname(stub shim.ChaincodeStubInterface) (string, error) {

	params := []string{"get", "name"}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := stub.InvokeChaincode("sacc", queryArgs, "mychannel")
	if response.Status != shim.OK {
		return "", fmt.Errorf("Failed to query chaincode. Got error: %s", response.Payload)
	}
	return string(response.Payload), nil
}

// Get "name" from chaincode sacc and write to local chaincode ccctest
func getsaccnamewritename(stub shim.ChaincodeStubInterface) (string, error) {

	params := []string{"get", "name"}
	queryArgs := make([][]byte, len(params))
	for i, arg := range params {
		queryArgs[i] = []byte(arg)
	}

	response := stub.InvokeChaincode("sacc", queryArgs, "mychannel")
	if response.Status != shim.OK {
		return "", fmt.Errorf("Failed to query chaincode. Got error: %s", response.Payload)
	}

	err := stub.PutState("name", []byte(response.Payload))
	if err != nil {
		return "", fmt.Errorf("Failed to set asset: name")
	}

	return string(response.Payload), nil
}

// Set "name" in chaincode sacc from chaincode ccctest (same channel)
func setsaccname(stub shim.ChaincodeStubInterface, args []string) (string, error) {

	params := []string{"set", "name", args[0]}
	invokeArgs := make([][]byte, len(params))
	for i, arg := range params {
		invokeArgs[i] = []byte(arg)
	}

	response := stub.InvokeChaincode("sacc", invokeArgs, "mychannel")
	if response.Status != shim.OK {
		return "", fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
	}
	return args[0], nil
}

// Set "name" in chaincode sacc from chaincode ccctest (diff channels)
func setsaccnamediffchannel(stub shim.ChaincodeStubInterface, args []string) (string, error) {

	params := []string{"set", "name", args[0]}
	invokeArgs := make([][]byte, len(params))
	for i, arg := range params {
		invokeArgs[i] = []byte(arg)
	}

	response := stub.InvokeChaincode("sacc", invokeArgs, "newchannel")
	if response.Status != shim.OK {
		return "", fmt.Errorf("Failed to invoke chaincode. Got error: %s", response.Payload)
	}
	return args[0], nil
}
