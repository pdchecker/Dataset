package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

var myMap = map[int]int{
	1: 1,
	2: 5,
	3: 10,
	4: 50,
}

var mySlice = []int{100, 200, 300}

type BadChaincode struct{}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	returnValue := 0
	for imkey, imvalue := range myMap {
		returnValue = returnValue*imkey - imvalue
	}
	for index, elem := range mySlice {
		elem = elem + index
	}
	return shim.Success([]byte("value: " + string(returnValue)))
}