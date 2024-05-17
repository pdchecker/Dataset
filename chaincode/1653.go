package main

import (
	"encoding/json"
	"io"
	"math/rand"
	"net"
	"os"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type BadChaincode struct{}

func (t *BadChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("success"))
}

func (t BadChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	_, err := net.Dial("tcp", "google.com:80")
	f, err := os.Open("a.txt")
	defer f.Close()
	buf := make([]byte, 12) // 实例化一个长度为4的[]byte

	for {
		n, err2 := f.Read(buf) // 将内容读至buf
		if n == 0 || err2 == io.EOF {
			return shim.Success([]byte("success"))
		}
	}

	tByte, err := json.Marshal(time.Now())
	rand.Seed(time.Now().Unix())
	rdata := rand.Int63n(100)
	err = stub.PutState("key", tByte)
	err = stub.PutState("key", rdata)
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success([]byte("success"))
}