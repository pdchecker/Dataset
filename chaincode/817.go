/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TransactionContextInterface an interface to
// describe the minimum required functions for
// a transaction context in the commercial
// paper
type TransactionContextInterface interface {
	contractapi.TransactionContextInterface
	GetRecordList() ListInterface
}

// TransactionContext implementation of
// TransactionContextInterface for use with
// commercial paper contract
type TransactionContext struct {
	contractapi.TransactionContext
	recordList *list
}

// GetRecordList return paper list
func (tc *TransactionContext) GetRecordList() ListInterface {
	if tc.recordList == nil {
		tc.recordList = newList(tc)
	}

	return tc.recordList
}
