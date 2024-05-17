package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"testing"
)


func testRegisterOrder(stub *shimtest.MockStub) {
	args := [][]byte{[]byte("RegisterOrder"), []byte("123order"), []byte("32028219123123"), []byte("zjj"), []byte("1772132131")}
	stub.MockInvoke("1", args)
	res := stub.MockInvoke("1", args)
	fmt.Println(string(res.Payload))
}

func testQueryOrder(stub *shimtest.MockStub) {
	args := [][]byte{[]byte("QueryOrder"), []byte("123order"), []byte("32028219123123"), []byte("zjj"), []byte("1772132131")}
	res := stub.MockInvoke("1", args)
	if res.Status != 200 || res.Payload == nil {
		fmt.Println(string(res.Message))
	} else {
		fmt.Println(string(res.Payload))
	}
}

func testConfirmOrder(stub *shimtest.MockStub) {
	args := [][]byte{[]byte("ConfirmOrder"), []byte("123order"), []byte("32028219123123"), []byte("zjj"), []byte("1772132131")}
	res := stub.MockInvoke("1", args)
	if res.Status != 200 || res.Payload == nil {
		fmt.Println(string(res.Message))
	} else {
		fmt.Println(string(res.Payload))
	}
}

func testDownPayment(stub *shimtest.MockStub) {
	args := [][]byte{[]byte("DownPayment"), []byte("123order"), []byte("32028219123123"), []byte("zjj"), []byte("1772132131"), []byte("万科"), []byte("天宇府"), []byte("89"), []byte("802")}
	res := stub.MockInvoke("1", args)
	if res.Status != 200 || res.Payload == nil {
		fmt.Println(string(res.Message))
	} else {
		fmt.Println(string(res.Payload))
	}}

func testConfirmDownPayment(stub *shimtest.MockStub) {
	args := [][]byte{[]byte("ConfirmDownPayment"), []byte("123order"), []byte("32028219123123"), []byte("zjj"), []byte("1772132131")}
	res := stub.MockInvoke("1", args)
	if res.Status != 200 || res.Payload == nil {
		fmt.Println(string(res.Message))
	} else {
		fmt.Println(string(res.Payload))
	}}

func testFullPayment(stub *shimtest.MockStub) {
	args := [][]byte{[]byte("FullPayment"), []byte("123order"), []byte("32028219123123"), []byte("zjj"), []byte("1772132131")}
	res := stub.MockInvoke("1", args)
	if res.Status != 200 || res.Payload == nil {
		fmt.Println(string(res.Message))
	} else {
		fmt.Println(string(res.Payload))
	}
}

func testConfirmFullPayment(stub *shimtest.MockStub) {
	args := [][]byte{[]byte("ConfirmFullPayment"), []byte("123order"), []byte("32028219123123"), []byte("zjj"), []byte("1772132131")}
	res := stub.MockInvoke("1", args)
	if res.Status != 200 || res.Payload == nil {
		fmt.Println(string(res.Message))
	} else {
		fmt.Println(string(res.Payload))
	}
}

func getChaincode() (scc *contractapi.ContractChaincode) {
	cc := new(SmartContract)
	cc.Name = "SmartContract"
	//cc.BeforeTransaction
	scc, err := contractapi.NewChaincode(cc)
	if  err != nil {
		fmt.Println(err)
	}
	return
}

func getMockStub(cc *contractapi.ContractChaincode) (stub *shimtest.MockStub) {
	stub = shimtest.NewMockStub("salesTest", cc)
	return
}

func Test_houseSale(t *testing.T) {
	cc := getChaincode()
	stub := getMockStub(cc)
	testRegisterOrder(stub)
	testQueryOrder(stub)
	testConfirmOrder(stub)
	testQueryOrder(stub)
	testDownPayment(stub)
	testQueryOrder(stub)
	testConfirmDownPayment(stub)
	testQueryOrder(stub)
	testFullPayment(stub)
	testQueryOrder(stub)
	testConfirmFullPayment(stub)
	testQueryOrder(stub)
}

func Test_MockWrongQuery(t *testing.T) {
	cc := getChaincode()
	stub := getMockStub(cc)
	testQueryOrder(stub)
}

func Test_MockWrongOrderOperation(t *testing.T) {
	cc := getChaincode()
	stub := getMockStub(cc)

	testRegisterOrder(stub)
	testQueryOrder(stub)
	testDownPayment(stub)
	testQueryOrder(stub)
}