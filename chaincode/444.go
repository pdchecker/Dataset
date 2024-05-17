package main

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

// TransactionContextInterface an interface to
// describe the minimum required functions for
// a transaction context in the nalanet
type TransactionContextInterface interface {
	contractapi.TransactionContextInterface
	GetLandList() ListInterface
}

// TransactionContext implementation of
// TransactionContextInterface for use with
// land contract
type TransactionContext struct {
	contractapi.TransactionContext
	landList *list
}

// GetLandList returns land list
func (tc *TransactionContext) GetLandList() ListInterface {
	if tc.landList == nil {
		tc.landList = newList(tc)
	}

	return tc.landList
}
