package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type registerDID struct {
}

func (t *registerDID) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *registerDID) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	if fn == "put" {
		return t.put(stub, args)
	} else if fn == "get" {
		return t.get(stub, args)
	} else if fn == "ping"{
		return t.ping(stub, args)
	}

	return shim.Error("Invalid invoke function name. Expecting \"put\" \"get\"")
}

func (t *registerDID) put(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to set asset: %s", args[0]))
	}

	err = stub.SetEvent("registerEvent", []byte("This is a test event!"))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to emit event: %s", err))
	}

	fmt.Println("Successfully emitted registerEvent")
	
	return shim.Success(nil)
}

func (t *registerDID) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	value, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to get asset: %s with error: %s", args[0], err))
	}
	if value == nil {
		return shim.Error(fmt.Sprintf("Asset not found: %s", args[0]))
	}
	return shim.Success(value)
}

func (t *registerDID) ping(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	return shim.Success([]byte("@@@@@@@@@@@@@@@@@@@@@@ Ping successful!"))
}

func main() {
	err := shim.Start(new(registerDID))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

