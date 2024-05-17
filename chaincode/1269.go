package main

import (
	"strconv"
	"unsafe"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	peer "github.com/hyperledger/fabric-protos-go/peer"
)

type BadChaincode struct {
}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t *BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	var x int = 10
	ptr := &x // ptr 是一个指向 x 的指针
	// 将指针转换为整数类型，这个操作将地址作为数值变量来处理
	x = int(uintptr(unsafe.Pointer(ptr)))
	// 对数值进行计算
	y := x + 10
	// 将计算结果转换回指针类型
	stub.PutState(strconv.Itoa(y), []byte("pointer"))
	return shim.Success([]byte("success"))
}