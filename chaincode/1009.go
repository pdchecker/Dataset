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

// Repository Package follow the repository pattern
// https://www.c-sharpcorner.com/UploadFile/b1df45/getting-started-with-repository-pattern-using-C-Sharp/
// It provides a generic interface to interact with our entities
package main

import (
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// We currently don't support delete function and getAll document
// There will not be any deletion. GetAll is quite dangerous as we never know
// what it can break.
type Repo interface {
	Create(ctx contractapi.TransactionContextInterface, entity interface{}, docPrefix string, keys []string) error
	Update(ctx contractapi.TransactionContextInterface, entity interface{}, docPrefix string, keys []string) error
	Get(ctx contractapi.TransactionContextInterface, docPrefix string, keys []string) (interface{}, error)
	GetQueryStringWithPagination(ctx contractapi.TransactionContextInterface, queryString string) (shim.StateQueryIteratorInterface, error)
	IsExist(ctx contractapi.TransactionContextInterface, docPrefix string, keys []string) (bool, error)
	GetQueryString(ctx contractapi.TransactionContextInterface, queryString string) (shim.StateQueryIteratorInterface, error)
	GetAndCheckExist(ctx contractapi.TransactionContextInterface, docPrefix string, keys []string) (bool, interface{}, error)
}
