package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const billCollection = "billCollection"

type Bill struct {
	ID                string `json:"ID"`
	Name              string `json:"name"` //the fieldtags are needed to keep case from bouncing around
	Owner             string `json:"owner"`
	PublicDescription string `json:"publicDescription"`
}

type BillPrivateDetails struct {
	ID         string `json:"ID"`
	BillAmount string `json:"billamount"`
}

type SmartContract struct {
	contractapi.Contract
}

//Admin's create bill
func (s *SmartContract) CreateBill(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientBillJSON, ok := transientMap["bill_properties"]
	if !ok {
		return fmt.Errorf("bill not found in the transient map input")
	}

	type billTransientInput struct {
		ID                string `json:"ID"`
		Name              string `json:"name"`
		BillAmount        string `json:"billamount"`
		PublicDescription string `json:"publicDesc"`
	}

	var billInput billTransientInput
	err = json.Unmarshal(transientBillJSON, &billInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(billInput.ID) == 0 {
		return fmt.Errorf("aBillID field must be a non-empty string")
	}
	if len(billInput.Name) == 0 {
		return fmt.Errorf("name field must be a non-empty string")
	}
	if len(billInput.BillAmount) == 0 {
		return fmt.Errorf("Bill amount field must be a non-empty string")
	}
	if len(billInput.PublicDescription) == 0 {
		return fmt.Errorf("PublicDescription field must be a non-empty string")
	}

	billAsBytes, err := ctx.GetStub().GetPrivateData(billCollection, billInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if billAsBytes != nil {
		fmt.Println("Asset already exists: " + billInput.ID)
		return fmt.Errorf("this asset already exists: " + billInput.ID)
	}

	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateBill cannot be performed: Error %v", err)
	}

	bill := Bill{
		ID:                billInput.ID,
		Name:              billInput.Name,
		PublicDescription: billInput.PublicDescription,
		Owner:             clientID,
	}
	billJSONasBytes, err := json.Marshal(bill)
	if err != nil {
		return fmt.Errorf("failed to marshal bill into JSON: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(billCollection, billInput.ID, billJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put bill into private data collecton: %v", err)
	}

	billPrivateDetails := BillPrivateDetails{
		ID:         billInput.ID,
		BillAmount: billInput.BillAmount,
	}

	billPrivateDetailsAsBytes, err := json.Marshal(billPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(orgCollection, billInput.ID, billPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put bill private details: %v", err)
	}
	return nil
}

//Update bill Admn or Insurance

func (s *SmartContract) UpdateBill(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	transientBillJSON, ok := transientMap["bill_properties"]
	if !ok {
		return fmt.Errorf("bill not found in the transient map input")
	}

	type billTransientInput struct {
		ID                string `json:"ID"`
		Name              string `json:"name"`
		BillAmount        string `json:"billamount"`
		PublicDescription string `json:"publicDesc"`
	}

	var billInput billTransientInput
	err = json.Unmarshal(transientBillJSON, &billInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(billInput.ID) == 0 {
		return fmt.Errorf("BillID field must be a non-empty string")
	}
	if len(billInput.Name) == 0 {
		return fmt.Errorf("name field must be a non-empty string")
	}
	if len(billInput.BillAmount) == 0 {
		return fmt.Errorf("bill amount field must be a non-empty string")
	}
	if len(billInput.PublicDescription) == 0 {
		return fmt.Errorf("PublicDescription field must be a non-empty string")
	}

	billAsBytes, err := s.BillExists(ctx, billInput.ID)
	if err != nil {
		return err
	}
	if !billAsBytes {
		return fmt.Errorf("the bill %s does not exist", billInput.ID)
	}

	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateBill cannot be performed: Error %v", err)
	}

	bill := Bill{
		ID:                billInput.ID,
		Name:              billInput.Name,
		PublicDescription: billInput.PublicDescription,
		Owner:             clientID,
	}
	billJSONasBytes, err := json.Marshal(bill)
	if err != nil {
		return fmt.Errorf("failed to marshal bill into JSON: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(billCollection, billInput.ID, billJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put bill into private data collecton: %v", err)
	}

	billPrivateDetails := BillPrivateDetails{
		ID:         billInput.ID,
		BillAmount: billInput.BillAmount,
	}

	billPrivateDetailsAsBytes, err := json.Marshal(billPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(orgCollection, billInput.ID, billPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put bill private details: %v", err)
	}
	return nil
}

func (s *SmartContract) BillExists(ctx contractapi.TransactionContextInterface, billID string) (bool, error) {
	billJSON, err := ctx.GetStub().GetPrivateData(billCollection, billID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return billJSON != nil, nil
}

func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	orgCollection := clientMSPID + "PrivateCollection"

	return orgCollection, nil
}

func verifyClientOrgMatchesPeerOrg(ctx contractapi.TransactionContextInterface) error {
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the client's MSPID: %v", err)
	}
	peerMSPID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting the peer's MSPID: %v", err)
	}

	if clientMSPID != peerMSPID {
		return fmt.Errorf("client from org %v is not authorized to read or write private data from an org %v peer", clientMSPID, peerMSPID)
	}

	return nil
}

func submittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil

}
