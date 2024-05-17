package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// SimpleAsset implements a simple chaincode to manage an asset
type SimpleAsset struct {
}
type outputEvent struct {
	EventName string
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *SimpleAsset) Init(stub shim.ChaincodeStubInterface) peer.Response {
	fmt.Printf("init...")
	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *SimpleAsset) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()
	switch fn {
	case "set":
		return t.set(stub, args)
	case "get":
		return t.get(stub, args)
	case "rangeQuery":
		return t.rangeQuery(stub, args)
	}
	// Return the result as success payload
	return shim.Error("指定的函数名称错误")
}

// Set stores the asset (both key and value) on the ledger. If the key exists,
// it will override the value with the new one
func (t *SimpleAsset) set(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("给定的参数个数不符合要求")
	}
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error("Failed to set asset: " + string(args[0]) + " " + string(args[1]))
	}
	event := outputEvent{
		EventName: "set",
	}
	payload, err := json.Marshal(event)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = stub.SetEvent("chaincode-event", payload)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("信息添加成功"))
}

// Get returns the value of the specified asset key
func (t *SimpleAsset) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// 检查传入数据长度
	if len(args) != 1 {
		return shim.Error("给定的参数个数不符合要求")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("根据userid查询信息时发生错误")
	}
	if value == nil {
		return shim.Error("根据指定的userid没有查询到相关的信息")
	}

	return shim.Success(value)
}

// range query with start and end key
func (t *SimpleAsset) rangeQuery(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("给定的参数个数不符合要求")
	}
	startKey := args[0]
	endKey := args[1]
	resultsIterator, err := stub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error("范围查询信息时发生错误")
	}
	defer resultsIterator.Close()
	var buffer string
	buffer = "["
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("范围查询信息时发生错误")
		}
		buffer = buffer + string(queryResponse.Value) + ","
	}
	buffer = buffer + "]"
	return shim.Success([]byte(buffer))
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SimpleAsset)); err != nil {
		fmt.Printf("Error starting SimpleAsset chaincode: %s", err)
	}
}
