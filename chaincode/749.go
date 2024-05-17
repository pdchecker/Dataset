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

// Package contract provides a sample smart contract.
// it uses util package to interact with the state database
// it also uses mock package to quickly test the solution
// by connecting to a remote couchDB instance.
package main

import (
	"github.com/Akachain/akc-go-sdk-v2/util"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SampleContract struct {
	contractapi.Contract
}

// Prefix for key in stateDB
const DocPrefix = "SAMPLE"

// Data - struct
type SampleData struct {
	Key1       string `json:"Key1"`
	Attribute1 string `json:"Attribute1"`
}

// Create something
func (s *SampleContract) CreateSampleObject(ctx contractapi.TransactionContextInterface, key string, val string) error {
	return util.CreateData(ctx.GetStub(), DocPrefix, []string{key}, &SampleData{Key1: key, Attribute1: val})
}
