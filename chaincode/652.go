package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const recordObyType = "record"

// add more attributes, use couchDB indexing

type singleRecord struct {
	Tags        []string  // {general-check-up, dockter anak}
	Date        time.Time `json:"date"`
	Description string    `json:"description"`
	Doctor      string    `json:"doctor"`
	Hospital    string    `json:"hospital"`
	Record      []byte    `json:"record"` // save pdf files using reader/writer interface
}

// PatientEHR struct
type PatientEHR struct {
	PatientID    string         `json:"id"`
	Name         string         `json:"name"`
	SingleRecord []singleRecord `json:"singleRecord"`
}

// ToLedgerValue covnerts a []byte to JSON for ledger
func (EHR *PatientEHR) ToLedgerValue() ([]byte, error) {
	return json.Marshal(EHR)
}

// SaveState saves the lot struct to the ledger
func (EHR *PatientEHR) SaveState(ctx contractapi.TransactionContextInterface) error {
	compositeKey, err := ctx.GetStub().CreateCompositeKey(recordObyType, []string{EHR.PatientID, EHR.Name})
	if err != nil {
		message := fmt.Sprintf("unable to create a composite key: %s", err.Error())
		return errors.New(message)
	}

	ledgerValue, err := EHR.ToLedgerValue()
	if err != nil {
		message := fmt.Sprintf("unable to compose a ledger value: %s", err.Error())
		return errors.New(message)
	}

	return ctx.GetStub().PutState(compositeKey, ledgerValue)
}

// LoadState loads the data from the ledger into the Lot object if data is found.
// Returns false if a Lot object wasn't found in the ledger; otherwise returns true
func (EHR *PatientEHR) LoadState(ctx contractapi.TransactionContextInterface) (bool, error) {
	compositeKey, err := ctx.GetStub().CreateCompositeKey(recordObyType, []string{EHR.PatientID, EHR.Name})
	if err != nil {
		message := fmt.Sprintf("unable to create a composite key: %s", err.Error())
		return false, errors.New(message)
	}

	ledgerValue, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		message := fmt.Sprintf("unable to read the ledger value: %s", err.Error())
		return false, errors.New(message)
	}

	if ledgerValue == nil {
		return false, nil
	}

	return true, json.Unmarshal(ledgerValue, &EHR)
}
