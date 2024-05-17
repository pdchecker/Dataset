package main

import (
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type randomKEY struct {
}

func (t *randomKEY) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *randomKEY) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	fn, args := stub.GetFunctionAndParameters()

	if fn == "put" {
		return t.put(stub, args)
	}
	return shim.Error("Invalid invoke function name. Expecting \"put\" \"get\"")
}

func (t *randomKEY) put(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to set asset: %s", args[0]))
	}
	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(randomKEY))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

