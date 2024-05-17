/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const assetCollection = "assetCollection"
const transferAgreementObjectType = "transferAgreement"

// SmartContract of this fabric sample
type SmartContract struct {
	contractapi.Contract
}

// Asset describes main asset details that are visible to all organizations
type Asset struct {
	Name    string `json:"vehicleName"` //Name is used to distinguish the various types of objects in state database
	RegNum  string `json:"vehicleNumber"`
	Company string `json:"vehicleCompany"`
	MfgYear int    `json:"vehicleMfgYear"`
	Owner   string `json:"owner"`
}

// AssetPrivateDetails describes details that are private to owners
type AssetPrivateDetails struct {
	RegNum string `json:"vehicleNumber"`
	Life   int    `json:"vehicleLife"`
}

// TransferAgreement describes the buyer agreement returned by ReadTransferAgreement
type TransferAgreement struct {
	RegNum  string `json:"vehicleNumber"`
	BuyerID string `json:"buyerID"`
}

// Notes:
// > CTX
// 1. GetStub -> GetTransient
// 			  -> GetPrivateData
// 			  -> CreateCompositeKey

// 2. GetClientIdentity -> GetID
// 						-> GetMSPID
// > SHIM
// 1. GetMSPID

// CreateAsset creates a new asset by placing the main asset details in the assetCollection
// that can be read by both organizations. The appraisal value is stored in the owners org specific collection.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface) error {

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		// log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type assetTransientInput struct {
		Name    string `json:"vehicleName"` //Name is used to distinguish the various types of objects in state database
		RegNum  string `json:"vehicleNumber"`
		Company string `json:"vehicleCompany"`
		MfgYear int    `json:"vehicleMfgYear"`
		Life    int    `json:"vehicleLife"`
	}

	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetInput.Name) == 0 {
		return fmt.Errorf("vehicleName field must be a non-empty string")
	}
	if len(assetInput.RegNum) == 0 {
		return fmt.Errorf("vehicleNumber field must be a non-empty string")
	}
	if len(assetInput.Company) == 0 {
		return fmt.Errorf("vehicleCompany field must be a non-empty string")
	}
	if assetInput.MfgYear <= 0 {
		return fmt.Errorf("vehicleMfgYear field must be a positive integer")
	}
	if assetInput.Life <= 0 {
		return fmt.Errorf("vehicleLife field must be a positive integer")
	}

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetInput.RegNum)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + assetInput.RegNum)
		return fmt.Errorf("this asset already exists: " + assetInput.RegNum)
	}

	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("CreateAsset cannot be performed: Error %v", err)
	}

	// Make submitting client the owner
	asset := Asset{
		Name:    assetInput.Name,
		RegNum:  assetInput.RegNum,
		Company: assetInput.Company,
		MfgYear: assetInput.MfgYear,
		Owner:   clientID,
	}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to private data collection
	// Typical logger, logs to stdout/file in the fabric managed docker container, running this chaincode
	// Look for container name like dev-peer0.org1.example.com-{chaincodename_version}-xyz
	log.Printf("CreateAsset Put: collection %v, ID %v, owner %v", assetCollection, assetInput.RegNum, clientID)

	err = ctx.GetStub().PutPrivateData(assetCollection, assetInput.RegNum, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}

	// Save asset details to collection visible to owning organization
	assetPrivateDetails := AssetPrivateDetails{
		RegNum: assetInput.RegNum,
		Life:   assetInput.Life,
	}

	assetPrivateDetailsAsBytes, err := json.Marshal(assetPrivateDetails) // marshal asset details to JSON
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	// Get collection name for this organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", orgCollection, assetInput.RegNum)
	err = ctx.GetStub().PutPrivateData(orgCollection, assetInput.RegNum, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	return nil
}

// AgreeToTransfer is used by the potential buyer of the asset to agree to the
// asset value. The agreed to appraisal value is stored in the buying orgs
// org specifc collection, while the the buyer client ID is stored in the asset collection
// using a composite key
func (s *SmartContract) AgreeToTransfer(ctx contractapi.TransactionContextInterface) error {

	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	// Value is private, therefore it gets passed in transient field
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Persist the JSON bytes as-is so that there is no risk of nondeterministic marshaling.
	valueJSONasBytes, ok := transientMap["asset_value"]
	if !ok {
		return fmt.Errorf("asset_value key not found in the transient map")
	}

	// Unmarshal the tranisent map to get the asset ID.
	var valueJSON AssetPrivateDetails
	err = json.Unmarshal(valueJSONasBytes, &valueJSON)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Do some error checking since we get the chance
	if len(valueJSON.RegNum) == 0 {
		return fmt.Errorf("vehicleNumber field must be a non-empty string")
	}
	if valueJSON.Life <= 0 {
		return fmt.Errorf("vehicleLife field must be a positive integer")
	}

	// Read asset from the private data collection
	asset, err := s.ReadAsset(ctx, valueJSON.RegNum)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("%v does not exist", valueJSON.RegNum)
	}
	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("AgreeToTransfer cannot be performed: Error %v", err)
	}

	// Get collection name for this organization. Needs to be read by a member of the organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	log.Printf("AgreeToTransfer Put: collection %v, ID %v", orgCollection, valueJSON.RegNum)
	// Put agreed value in the org specifc private data collection
	err = ctx.GetStub().PutPrivateData(orgCollection, valueJSON.RegNum, valueJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	// Create agreeement that indicates which identity has agreed to purchase
	// In a more realistic transfer scenario, a transfer agreement would be secured to ensure that it cannot
	// be overwritten by another channel member
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{valueJSON.RegNum})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	log.Printf("AgreeToTransfer Put: collection %v, ID %v, Key %v", assetCollection, valueJSON.RegNum, transferAgreeKey)
	err = ctx.GetStub().PutPrivateData(assetCollection, transferAgreeKey, []byte(clientID))
	if err != nil {
		return fmt.Errorf("failed to put asset bid: %v", err)
	}

	return nil
}

// TransferAsset transfers the asset to the new owner by setting a new owner ID
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientTransferJSON, ok := transientMap["asset_owner"]
	if !ok {
		return fmt.Errorf("asset owner not found in the transient map")
	}

	type assetTransferTransientInput struct {
		RegNum   string `json:"vehicleNumber"`
		BuyerMSP string `json:"buyerMSP"`
	}

	var assetTransferInput assetTransferTransientInput
	err = json.Unmarshal(transientTransferJSON, &assetTransferInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetTransferInput.RegNum) == 0 {
		return fmt.Errorf("vehicleNumber field must be a non-empty string")
	}
	if len(assetTransferInput.BuyerMSP) == 0 {
		return fmt.Errorf("buyerMSP field must be a non-empty string")
	}
	log.Printf("TransferAsset: verify asset exists ID %v", assetTransferInput.RegNum)
	// Read asset from the private data collection
	asset, err := s.ReadAsset(ctx, assetTransferInput.RegNum)
	if err != nil {
		return fmt.Errorf("error reading asset: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("%v does not exist", assetTransferInput.RegNum)
	}
	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("TransferAsset cannot be performed: Error %v", err)
	}

	// Verify transfer details and transfer owner
	err = s.verifyAgreement(ctx, assetTransferInput.RegNum, asset.Owner, assetTransferInput.BuyerMSP)
	if err != nil {
		return fmt.Errorf("failed transfer verification: %v", err)
	}

	transferAgreement, err := s.ReadTransferAgreement(ctx, assetTransferInput.RegNum)
	if err != nil {
		return fmt.Errorf("failed ReadTransferAgreement to find buyerID: %v", err)
	}
	if transferAgreement.BuyerID == "" {
		return fmt.Errorf("BuyerID not found in TransferAgreement for %v", assetTransferInput.RegNum)
	}

	// Transfer asset in private data collection to new owner
	asset.Owner = transferAgreement.BuyerID

	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed marshalling asset %v: %v", assetTransferInput.RegNum, err)
	}

	log.Printf("TransferAsset Put: collection %v, ID %v", assetCollection, assetTransferInput.RegNum)
	err = ctx.GetStub().PutPrivateData(assetCollection, assetTransferInput.RegNum, assetJSONasBytes) //rewrite the asset
	if err != nil {
		return err
	}

	// Get collection name for this organization
	ownersCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	// Delete the asset appraised value from this organization's private data collection
	err = ctx.GetStub().DelPrivateData(ownersCollection, assetTransferInput.RegNum)
	if err != nil {
		return err
	}

	// Delete the transfer agreement from the asset collection
	transferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetTransferInput.RegNum})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().DelPrivateData(assetCollection, transferAgreeKey)
	if err != nil {
		return err
	}

	return nil

}

// verifyAgreement is an internal helper function used by TransferAsset to verify
// that the transfer is being initiated by the owner and that the buyer has agreed
// to the same appraisal value as the owner
func (s *SmartContract) verifyAgreement(ctx contractapi.TransactionContextInterface, RegNum string, owner string, buyerMSP string) error {

	// Check 1: verify that the transfer is being initiatied by the owner

	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != owner {
		return fmt.Errorf("error: submitting client identity does not own asset")
	}

	// Check 2: verify that the buyer has agreed to the appraised value

	// Get collection names
	collectionOwner, err := getCollectionName(ctx) // get owner collection from caller identity
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	collectionBuyer := buyerMSP + "PrivateCollection" // get buyers collection

	// Get hash of owners agreed to value
	ownerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionOwner, RegNum)
	if err != nil {
		return fmt.Errorf("failed to get hash of appraised value from owners collection %v: %v", collectionOwner, err)
	}
	if ownerAppraisedValueHash == nil {
		return fmt.Errorf("hash of appraised value for %v does not exist in collection %v", RegNum, collectionOwner)
	}

	// Get hash of buyers agreed to value
	buyerAppraisedValueHash, err := ctx.GetStub().GetPrivateDataHash(collectionBuyer, RegNum)
	if err != nil {
		return fmt.Errorf("failed to get hash of appraised value from buyer collection %v: %v", collectionBuyer, err)
	}
	if buyerAppraisedValueHash == nil {
		return fmt.Errorf("hash of appraised value for %v does not exist in collection %v. AgreeToTransfer must be called by the buyer first", RegNum, collectionBuyer)
	}

	// Verify that the two hashes match
	if !bytes.Equal(ownerAppraisedValueHash, buyerAppraisedValueHash) {
		return fmt.Errorf("hash for appraised value for owner %x does not value for seller %x", ownerAppraisedValueHash, buyerAppraisedValueHash)
	}

	return nil
}

// DeleteAsset can be used by the owner of the asset to delete the asset
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientDeleteJSON, ok := transientMap["asset_delete"]
	if !ok {
		return fmt.Errorf("asset to delete not found in the transient map")
	}

	type assetDelete struct {
		RegNum string `json:"vehicleNumber"`
	}

	var assetDeleteInput assetDelete
	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetDeleteInput.RegNum) == 0 {
		return fmt.Errorf("vehicleNumber field must be a non-empty string")
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("DeleteAsset cannot be performed: Error %v", err)
	}

	log.Printf("Deleting Asset: %v", assetDeleteInput.RegNum)
	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetDeleteInput.RegNum) //get the asset from chaincode state
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset not found: %v", assetDeleteInput.RegNum)
	}

	ownerCollection, err := getCollectionName(ctx) // Get owners collection
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	// Check the asset is in the caller org's private collection
	valAsbytes, err = ctx.GetStub().GetPrivateData(ownerCollection, assetDeleteInput.RegNum)
	if err != nil {
		return fmt.Errorf("failed to read asset from owner's Collection: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset not found in owner's private Collection %v: %v", ownerCollection, assetDeleteInput.RegNum)
	}

	// delete the asset from state
	err = ctx.GetStub().DelPrivateData(assetCollection, assetDeleteInput.RegNum)
	if err != nil {
		return fmt.Errorf("failed to delete state: %v", err)
	}

	// Finally, delete private details of asset
	err = ctx.GetStub().DelPrivateData(ownerCollection, assetDeleteInput.RegNum)
	if err != nil {
		return err
	}

	return nil

}

// PurgeAsset can be used by the owner of the asset to delete the asset
// Trigger removal of the asset
func (s *SmartContract) PurgeAsset(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("Error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientDeleteJSON, ok := transientMap["asset_purge"]
	if !ok {
		return fmt.Errorf("asset to purge not found in the transient map")
	}

	type assetPurge struct {
		RegNum string `json:"vehicleNumber"`
	}

	var assetPurgeInput assetPurge
	err = json.Unmarshal(transientDeleteJSON, &assetPurgeInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetPurgeInput.RegNum) == 0 {
		return fmt.Errorf("vehicleNumber field must be a non-empty string")
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("PurgeAsset cannot be performed: Error %v", err)
	}

	log.Printf("Purging Asset: %v", assetPurgeInput.RegNum)

	// Note that there is no check here to see if the id exist; it might have been 'deleted' already
	// so a check here is pointless. We would need to call purge irrespective of the result
	// A delete can be called before purge, but is not essential

	ownerCollection, err := getCollectionName(ctx) // Get owners collection
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	// delete the asset from state
	err = ctx.GetStub().PurgePrivateData(assetCollection, assetPurgeInput.RegNum)
	if err != nil {
		return fmt.Errorf("failed to purge state from asset collection: %v", err)
	}

	// Finally, delete private details of asset
	err = ctx.GetStub().PurgePrivateData(ownerCollection, assetPurgeInput.RegNum)
	if err != nil {
		return fmt.Errorf("failed to purge state from owner collection: %v", err)
	}

	return nil

}

// DeleteTranferAgreement can be used by the buyer to withdraw a proposal from
// the asset collection and from his own collection.
func (s *SmartContract) DeleteTranferAgreement(ctx contractapi.TransactionContextInterface) error {

	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field
	transientDeleteJSON, ok := transientMap["agreement_delete"]
	if !ok {
		return fmt.Errorf("asset to delete not found in the transient map")
	}

	type assetDelete struct {
		RegNum string `json:"vehicleNumber"`
	}

	var assetDeleteInput assetDelete
	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetDeleteInput.RegNum) == 0 {
		return fmt.Errorf("transient input ID field must be a non-empty string")
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("DeleteTranferAgreement cannot be performed: Error %v", err)
	}
	// Delete private details of agreement
	orgCollection, err := getCollectionName(ctx) // Get proposers collection.
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}
	tranferAgreeKey, err := ctx.GetStub().CreateCompositeKey(transferAgreementObjectType, []string{assetDeleteInput.
		RegNum}) // Create composite key
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, tranferAgreeKey) //get the transfer_agreement
	if err != nil {
		return fmt.Errorf("failed to read transfer_agreement: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset's transfer_agreement does not exist: %v", assetDeleteInput.RegNum)
	}

	log.Printf("Deleting TranferAgreement: %v", assetDeleteInput.RegNum)
	err = ctx.GetStub().DelPrivateData(orgCollection, assetDeleteInput.RegNum) // Delete the asset
	if err != nil {
		return err
	}

	// Delete transfer agreement record
	err = ctx.GetStub().DelPrivateData(assetCollection, tranferAgreeKey) // remove agreement from state
	if err != nil {
		return err
	}

	return nil

}

// getCollectionName is an internal helper function to get collection of submitting client identity.
func getCollectionName(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get the MSP ID of submitting client identity
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	// Create the collection name
	orgCollection := clientMSPID + "PrivateCollection"

	return orgCollection, nil
}

// verifyClientOrgMatchesPeerOrg is an internal function used verify client org id and matches peer org id.
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
