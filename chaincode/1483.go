package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	"log"
	"os"
)

type HelloWorld struct{}

func (h *HelloWorld) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (h *HelloWorld) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	log.Printf("Invoking %s with args: %v", function, args)

	if function == "save" {
		return h.invoke(stub, args)
	} else if function == "get" {
		return h.query(stub, args)
	}

	return shim.Error("Invalid function call")
}

func (h *HelloWorld) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of args")
	}

	key := args[0]
	value, err := json.Marshal(args[1])
	if err != nil {
		return shim.Error(fmt.Sprintf("marshalling value: %s", err.Error()))
	}

	if err := stub.PutState(key, value); err != nil {
		return shim.Error(err.Error())
	}

	if err := stub.SetEvent("notification", []byte(fmt.Sprintf("key %s successfully saved", key))); err != nil {
		return shim.Error("error happened emitting event: " + err.Error())
	}

	return shim.Success(nil)
}

func (h *HelloWorld) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of args")
	}

	key := args[0]

	value, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Error getting state: " + err.Error())
	}

	if value == nil {
		return shim.Error("no such key")
	}

	return shim.Success(value)
}

func main() {
	ccid := os.Getenv("CHAINCODE_ID")
	address := os.Getenv("CHAINCODE_ADDRESS")

	server := &shim.ChaincodeServer{
		CCID:    ccid,
		Address: address,
		CC:      &HelloWorld{},
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	log.Printf("Started ccid %s on %s", ccid, address)

	if err := server.Start(); err != nil {
		log.Fatalf("Error starting HelloWorld chaincode: %s", err)
	}
}
