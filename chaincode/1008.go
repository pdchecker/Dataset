package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"testing"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
	"github.com/hyperledger/fabric-protos-go/ledger/queryresult"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// Changed to Transaction id being generated uniquely for even test cases
var TxIdSeed int = 0

type CouchDBMockStub struct {
	*shimtest.MockStub
	ObjectType string
}

func NewCouchDBMockStateRangeQueryIterator(queryresults []queryresult.KV) *CouchDBMockStateRangeQueryIterator {
	iter := new(CouchDBMockStateRangeQueryIterator)
	if !iter.HasNext() {
		return iter
	}
	iter.QueryResults = queryresults
	iter.CurrentIndex = 0
	iter.Closed = false
	return iter
}

type CouchDBMockStateRangeQueryIterator struct {
	QueryResults []queryresult.KV
	CurrentIndex int
	Closed       bool
}

func NewCouchDBMockStub(name string, cc shim.Chaincode) *CouchDBMockStub {
	mock := shimtest.NewMockStub(name, cc)
	cmock := CouchDBMockStub{mock, ""}
	return &cmock
}

type CouchDBQuery struct {
	Selector map[string]string `json:"selector"`
}

func MockInit(stub *CouchDBMockStub, function string, args []byte) pb.Response {
	mockInvokeArgs := [][]byte{[]byte(function), args}
	txId := generateTransactionId()
	res := stub.MockStub.MockInit(txId, mockInvokeArgs)
	return res
}

func checkInvoke(t *testing.T, stub *CouchDBMockStub, function string, args []byte) pb.Response {
	fmt.Println("inside checkInvoke")
	mockInvokeArgs := [][]byte{[]byte(function), args}
	fmt.Println("mockinvokeArgs ", mockInvokeArgs)
	txId := generateTransactionId()
	fmt.Println("txId ", txId)
	res := stub.MockInvoke(txId, mockInvokeArgs)
	fmt.Println("res ", res)
	if res.Status != shim.OK {
		t.FailNow()
	}
	fmt.Println("res ", res)
	return res
}
func checkInvokeForResetLedger(t *testing.T, stub *CouchDBMockStub, function string) pb.Response {
	mockInvokeArgs := [][]byte{[]byte(function)}
	txId := generateTransactionId()
	res := stub.MockInvoke(txId, mockInvokeArgs)
	if res.Status != shim.OK {
		t.FailNow()
	}
	return res
}

// TODO: What is this function supposed to do?
// It is not asserting anything at the moment... anything missing here?
func checkInvoke_forError(t *testing.T, stub *CouchDBMockStub, function string, args []byte) pb.Response {
	mockInvokeArgs := [][]byte{[]byte(function), args}
	txId := generateTransactionId()
	res := stub.MockInvoke(txId, mockInvokeArgs)
	if res.Status != shim.OK {
	}
	return res
}

func (stub *CouchDBMockStub) GetQueryResult(query string) (shim.StateQueryIteratorInterface, error) {
	// Not implemented since the mock engine does not have a query engine.
	// However, a very simple query engine that supports string matching
	// could be implemented to test that the framework supports queries

	fmt.Printf("%+v\n", query)
	cdbquery := CouchDBQuery{}
	json.Unmarshal([]byte(query), &cdbquery)
	fmt.Printf("%+v\n", cdbquery)
	fmt.Printf("object type %s\n", stub.ObjectType)
	iter, err := stub.GetStateByPartialCompositeKey(stub.ObjectType, []string{})

	defer iter.Close()
	if err != nil {
		return nil, err
	}
	//	fmt.Printf("%+v\n", iter)
	//	fmt.Printf("%+v\n", stub.State)
	if iter.HasNext() {
		return iter, nil
	}
	ret := []queryresult.KV{}
	for iter.HasNext() {
		kv, err := iter.Next()
		if err != nil {
			return nil, err
		}
		fmt.Printf("%+v\n", kv)
		// TODO: choose items which matches query condition.
		ret = append(ret, *kv)
	}
	retiter := NewCouchDBMockStateRangeQueryIterator(ret)
	fmt.Printf("%+v\n", retiter)
	return retiter, nil
	//	return nil, errors.New("Not Implemented")
}

// HasNext returns true if the range query iterator contains additional keys
// and values.
func (iter *CouchDBMockStateRangeQueryIterator) HasNext() bool {
	if iter.Closed {
		// previously called Close()
		return false
	}
	if iter.CurrentIndex >= len(iter.QueryResults) {
		return false
	}
	return true
}

// Next returns the next key and value in the range query iterator.
func (iter *CouchDBMockStateRangeQueryIterator) Next() (*queryresult.KV, error) {
	if iter.Closed == true {
		return nil, errors.New("MockStateRangeQueryIterator.Next() called after Close()")
	}

	if iter.HasNext() == false {
		return nil, errors.New("MockStateRangeQueryIterator.Next() called when it does not HaveNext()")
	}
	ret := &iter.QueryResults[iter.CurrentIndex]
	iter.CurrentIndex++
	return ret, nil
	//	return nil, errors.New("MockStateRangeQueryIterator.Next() went past end of range")
}

// Close closes the range query iterator. This should be called when done
// reading from the iterator to free up resources.
func (iter *CouchDBMockStateRangeQueryIterator) Close() error {
	iter.Closed = true
	return nil
}

func generateTransactionId() string {
	TxIdSeed++
	s := strconv.Itoa(TxIdSeed)
	return s
}
