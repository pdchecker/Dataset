package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (s *SmartContract) ReadBill(ctx contractapi.TransactionContextInterface, billID string) (*Bill, error) {

	log.Printf("ReadBill: collection %v, ID %v", billCollection, billID)
	billJSON, err := ctx.GetStub().GetPrivateData(billCollection, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to read bill: %v", err)
	}

	if billJSON == nil {
		log.Printf("%v does not exist in collection %v", billID, billCollection)
		return nil, nil
	}

	var bill *Bill
	err = json.Unmarshal(billJSON, &bill)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return bill, nil

}

func (s *SmartContract) ReadBillPrivateDetails(ctx contractapi.TransactionContextInterface, collection string, billID string) (*BillPrivateDetails, error) {
	log.Printf("ReadBillPrivateDetails: collection %v, ID %v", collection, billID)
	billDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to read bill details: %v", err)
	}
	if billDetailsJSON == nil {
		log.Printf("BillPrivateDetails for %v does not exist in collection %v", billID, collection)
		return nil, nil
	}

	var billDetails *BillPrivateDetails
	err = json.Unmarshal(billDetailsJSON, &billDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return billDetails, nil
}

func (s *SmartContract) ReadOrg1BillPrivateDetails(ctx contractapi.TransactionContextInterface, billID string) (*BillPrivateDetails, error) {

	billDetailsJSON, err := ctx.GetStub().GetPrivateData("hosp1apolloMSPPrivateCollection", billID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from marble details %s", err.Error())
	}
	if billDetailsJSON == nil {
		return nil, fmt.Errorf("%s does not exist", billID)
	}

	var billDetails *BillPrivateDetails
	err = json.Unmarshal(billDetailsJSON, &billDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return billDetails, nil

}

func (s *SmartContract) ReadOrg2BillPrivateDetails(ctx contractapi.TransactionContextInterface, billID string) (*BillPrivateDetails, error) {

	billDetailsJSON, err := ctx.GetStub().GetPrivateData("Ins1starhealthMSPPrivateCollection", billID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from marble details %s", err.Error())
	}
	if billDetailsJSON == nil {
		return nil, fmt.Errorf("%s does not exist", billID)
	}

	var billDetails *BillPrivateDetails
	err = json.Unmarshal(billDetailsJSON, &billDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return billDetails, nil

}
