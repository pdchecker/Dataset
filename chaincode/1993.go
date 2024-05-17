package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//"github.com/hyperledger/fabric/core/chaincode/shim"
//pb "github.com/hyperledger/fabric/protos/peer"

type IOHandler struct {
}

func main() {

	/*var file, errOpen = os.OpenFile("io-fabric.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errOpen != nil {
		log.Println(errOpen)
	}
	defer file.Close()

	log.SetOutput(file)*/

	fmt.Println("Started io")

	var errStart = shim.Start(new(IOHandler))
	if errStart != nil {
		fmt.Printf("Error starting chaincode: %v \n", errStart)
	}

}

const chars string = "!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~"

func (self *IOHandler) Init(stub shim.ChaincodeStubInterface) pb.Response {
	return shim.Success(nil)
}

func (self *IOHandler) Write(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		return shim.Error("Call to Write must have 4 parameters")
	}

	var size, errSize = strconv.Atoi(args[0])
	if errSize != nil {
		return shim.Error("Failed convert size")
	}
	var startKey, errKey = strconv.Atoi(args[1])
	if errKey != nil {
		return shim.Error("Failed convert key")
	}
	var retLen, errLen = strconv.Atoi(args[2])
	if errLen != nil {
		return shim.Error("Failed convert length")
	}

	var stateArr []byte
	for i := 0; i < size; i++ {
		var sK = strconv.Itoa(startKey + i)

		var val = getVal(startKey+i, retLen)
		var errSet = stub.PutState(sK, val)
		if errSet != nil {
			return shim.Error("Cannot put state")
		}
		stateArr = append(stateArr, val...)
	}
	var signature = args[3]
	var errEvent = stub.SetEvent("storage/write", []byte(signature))
	if errEvent != nil {
		return shim.Error("Cannot set event")
	}
	return shim.Success(nil)
}

func getVal(k int, retLen int) []byte {
	var ret = make([]byte, retLen)
	for i := 0; i < retLen; i++ {
		ret[i] = chars[(k+i)%len(chars)]
	}
	return ret
}

func (self *IOHandler) Scan(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 3 {
		return shim.Error("Call to Scan must have 3 parameters")
	}

	var size, errSize = strconv.Atoi(args[0])
	if errSize != nil {
		return shim.Error("Failed convert size: ")
	}
	var startKey, errSk = strconv.Atoi(args[1])
	if errSk != nil {
		return shim.Error("Failed to convert startKey: ")
	}
	var stateArr []byte
	for i := 0; i < size; i++ {
		var sK = strconv.Itoa(startKey + i)
		var state, errGet = stub.GetState(sK)
		if errGet != nil {
			return shim.Error("Error getting the state")
		}
		stateArr = append(stateArr, state...)
	}
	var signature = args[2]
	var errEvent = stub.SetEvent("storage/scan", []byte(signature))
	if errEvent != nil {
		return shim.Error("Error setting event")
	}
	return shim.Success(nil)
}

func (self *IOHandler) RevertScan(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 3 {
		return shim.Error("Call to RevertScan must have 3 parameters")
	}

	var size, errSize = strconv.Atoi(args[0])
	if errSize != nil {
		return shim.Error("Failed convert size: ")
	}
	var startKey, errSk = strconv.Atoi(args[1])
	if errSk != nil {
		return shim.Error("Failed to convert startKey: ")
	}
	var stateArr []byte
	for i := 0; i < size; i++ {
		var sK = strconv.Itoa(startKey + size - i - 1)
		var state, errGet = stub.GetState(sK)
		if errGet != nil {
			return shim.Error("Error getting the state")
		}
		stateArr = append(stateArr, state...)
	}
	var signature = args[2]
	var errEvent = stub.SetEvent("storage/revertScan", []byte(signature))
	if errEvent != nil {
		return shim.Error("Error setting event")
	}
	return shim.Success(nil)
}

func (self *IOHandler) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	var function, args = stub.GetFunctionAndParameters()

	if function == "Write" {
		return self.Write(stub, args)
	}
	if function == "Scan" {
		return self.Scan(stub, args)
	}
	if function == "RevertScan" {
		return self.RevertScan(stub, args)
	}

	return shim.Error("Not yet implemented function called")
}
