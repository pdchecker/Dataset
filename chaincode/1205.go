package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//"github.com/hyperledger/fabric/core/chaincode/shim"
//pb "github.com/hyperledger/fabric/protos/peer"

type KeyValueHandler struct {
}

func main() {

	/*var file, errOpen = os.OpenFile("kv-fabric.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errOpen != nil {
		log.Println(errOpen)
	}
	defer file.Close()

	log.SetOutput(file)*/

	fmt.Println("Started key_value")

	var errStart = shim.Start(new(KeyValueHandler))
	if errStart != nil {
		fmt.Printf("Error starting chaincode: %v \n", errStart)
	}

}

func (self *KeyValueHandler) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (self *KeyValueHandler) Get(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Call to Get must have 2 parameters")
	}

	var _, errGet = stub.GetState(args[0])
	if errGet != nil {
		return shim.Error("Error while GetState call")
	}

	var signature = args[1]
	var errSet = stub.SetEvent("keyValue/get", []byte(signature))
	if errSet != nil {
		return shim.Error("Error while setting event in GetState call")
	}

	//return shim.Success(value)
	return shim.Success(nil)
}

func (self *KeyValueHandler) Set(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error("Call to Set must have 3 parameters")
	}

	var key = args[0]
	var value = []byte(args[1])

	var errSet = stub.PutState(key, value)
	if errSet != nil {
		return shim.Error("Error while setting state")
	}

	var signature = args[2]
	var errEvent = stub.SetEvent("keyValue/set", []byte(signature))
	if errEvent != nil {
		return shim.Error("Error while setting event in SetState call")
	}

	return shim.Success(nil)
}

func (self *KeyValueHandler) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var function, args = stub.GetFunctionAndParameters()

	if function == "Get" {
		return self.Get(stub, args)
	}
	if function == "Set" {
		return self.Set(stub, args)
	}

	return shim.Error("Not yet implemented function called")
}
