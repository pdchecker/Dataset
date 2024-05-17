// Copyright (c) 2021 akachain
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package mock provides mock methods to simulate how a peer node interact with CouchDB
// and simulate a transaction proposal. This does not requires a working Fabric Peer like
// FabricSDK but require connection to a remote CouchDB instance.
//
// The workflow is as follow
//
// 1) Create a core.yaml follow similar template in test/contract/core.yaml, remember to change the configuration of the couchDB as you need
//
// 2) Create a NewMockStubExtend object that point to the core.yaml and the smart contract object
//
// 3) Process Indexes (if need)
//
// 4) Perform MockInvokeTransaction
//
// For more details, please find test example in the README file
package main

import (
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/satori/go.uuid"
	"testing"
)

// MockInvokeTransaction creates a mock invoke transaction using MockStubExtend
func MockInvokeTransaction(t *testing.T, stub *MockStubExtend, args [][]byte) string {
	txId := genTxID()
	res := stub.MockInvoke(txId, args)
	if res.Status != shim.OK {
		return string(res.Message)
	}
	// fmt.Println(res.Payload)
	return string(res.Payload)
}

// MockQueryTransaction creates a mock query transaction using MockStubExtend
func MockQueryTransaction(t *testing.T, stub *MockStubExtend, args [][]byte) string {
	txId := genTxID()
	res := stub.MockInvoke(txId, args)
	if res.Status != shim.OK {
		t.FailNow()
		return string(res.Message)
	}
	return string(res.Payload)
}

// MockIInit creates a mock invoke transaction using MockStubExtend
func MockInitTransaction(t *testing.T, stub *MockStubExtend, args [][]byte) string {
	txId := genTxID()
	res := stub.MockInit(txId, args)
	if res.Status != shim.OK {
		return string(res.Message)
	}
	return string(res.Payload)
}

// Generate random transaction ID
func genTxID() string {
	// or error handling
	uid := uuid.NewV4()
	txId := fmt.Sprintf("%s", uid)
	return txId
}
