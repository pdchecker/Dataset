/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	//"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const assetCollection = "assetCollection"
//const transferAgreementObjectType = "transferAgreement"

// SmartContract of this fabric sample
type SmartContract struct {
	contractapi.Contract
}

// Asset describes main asset details that are visible to all organizations
type Asset struct {
	//Type  string `json:"objectType"` //Type is used to distinguish the various types of objects in state database
	ID    string `json:"ID"`
	Rep   int `json:"rep"`
	//Size  int    `json:"size"`
	Owner string `json:"owner"`
}

type AssetConsent struct {
	ID 		string  `json:"ID"`
	ConsentFlag     bool 	`json:"ConsentFlag"`
	BlockedFlag	bool	`json:"BlockedFlag"`
	//ID             string `json:"ID"`
	//Owner          string `json:"Owner"`
	//Size           int    `json:"Size"`
}

// AssetPrivateDetails describes details that are private to owners
type AssetPrivateDetails struct {
	ID             string `json:"ID"`
	Name	       string `json:"name"`
	Email	       string `json:"email"`
	Salt	       string `json:"salt"`
	SCounter	   int	  `json:"sCounter"`
	SSum		   int	  `json:"sSum"`
}

type UnkAsset struct{
	ID	string `json:"ID"`
	Unk1	string `json:"unk1"`
	Unk2	string `json:"unk2"`
	Owner 	string `json:"owner"`
}

type UnkAssetPvt struct{
	ID		string `json:"ID"`
	UnkPvt1		string `json:"unkPvt1"`
	UnkPvt2 	string `json:"unkPvt2"`
	UnkPvt3		string `json:"unkPvt3"`
	Salt		string `json:"salt"`
}

// TransferAgreement describes the buyer agreement returned by ReadTransferAgreement
type TransferAgreement struct {
	ID      string `json:"assetID"`
	BuyerID string `json:"buyerID"`
}

// CreateAsset creates a new asset by placing the main asset details in the assetCollection
// that can be read by both organizations. The appraisal value is stored in the owners org specific collection.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface) error {
	validJCA, er := s.CheckJCA (ctx, "CreateAsset")
	if er != nil{
		return fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return fmt.Errorf("No valid JCA for client ORG found")
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type assetTransientInput struct {
		//Type           string `json:"objectType"` //Type is used to distinguish the various types of objects in state database
		ID             string 	`json:"ID"`
		Rep            int 	`json:"rep"`
		Name           string   `json:"name"`
		Email	       string   `json:"email"`
		Salt	       string	`json:"salt"`
	}

	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetInput.Name) == 0 {
		return fmt.Errorf("objectType field must be a non-empty string")
	}
	if len(assetInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(assetInput.Email) == 0 {
		return fmt.Errorf("color field must be a non-empty string")
	}
	if len(assetInput.Salt) == 0 {
		return fmt.Errorf("color field must be a non-empty string")
	}
	if assetInput.Rep < 0 {
		return fmt.Errorf("size field must be 0 or a positive integer")
	}
	/*if assetInput.AppraisedValue <= 0 {
		return fmt.Errorf("appraisedValue field must be a positive integer")
	}*/



	//------------------------------------------
	//------------------------------------------
	//------------------------------------------
	//------------------------------------------
	assetC, err := s.ReadAssetConsent(ctx, assetInput.ID)
	if err != nil {
		return fmt.Errorf("Error On ReadAssetConsent call")
	}
	if assetC.ConsentFlag != true{
		return fmt.Errorf("No consent given")
	} else if assetC.BlockedFlag != false{
		return fmt.Errorf("Processing is blocked (RoO, RoR)")
	}
	//------------------------------------------
	//------------------------------------------
	//------------------------------------------
	//------------------------------------------




	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + assetInput.ID)
		return fmt.Errorf("this asset already exists: " + assetInput.ID)
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
		ID:  assetInput.ID,
		Rep:    assetInput.Rep,
		Owner: clientID,
	}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to private data collection
	// Typical logger, logs to stdout/file in the fabric managed docker container, running this chaincode
	// Look for container name like dev-peer0.org1.example.com-{chaincodename_version}-xyz
	log.Printf("CreateAsset Put: collection %v, ID %v, owner %v", assetCollection, assetInput.ID, clientID)

	err = ctx.GetStub().PutPrivateData(assetCollection, assetInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	
	sCounterVal := 0 //counter for avg rep calc
	if assetInput.Rep > 0 {
		sCounterVal = 1
	}

	// Save asset details to collection visible to owning organization
	assetPrivateDetails := AssetPrivateDetails{
		ID:             assetInput.ID,
		Name: 		assetInput.Name,
		Email: 		assetInput.Email,
		Salt:		assetInput.Salt,
		SCounter:	sCounterVal,
		SSum:		assetInput.Rep,
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
	log.Printf("Put: collection %v, ID %v", orgCollection, assetInput.ID)
	err = ctx.GetStub().PutPrivateData(orgCollection, assetInput.ID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	return nil
}

func (s *SmartContract) CreateAssetConsent(ctx contractapi.TransactionContextInterface, id string, consentFlag bool) error {
	validJCA, er := s.CheckJCA (ctx, "CreateAssetConsent")
	if er != nil{
		return fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return fmt.Errorf("No valid JCA for client ORG found")
	}

	exists, err := s.AssetExistsConsent(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := AssetConsent{
		ID:          	 id,
		ConsentFlag:      consentFlag,
		BlockedFlag:	 false,
		//Size:           size,
		//Owner:          owner,
		//AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) AssetExistsConsent(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetPrivateData(assetCollection, id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return assetJSON != nil, nil
}

func (s *SmartContract) ReadAssetConsent(ctx contractapi.TransactionContextInterface, id string) (*AssetConsent, error) {
	validJCA, er := s.CheckJCA (ctx, "ReadAssetConsent")
	if er != nil{
		return nil, fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return nil, fmt.Errorf("No valid JCA for client ORG found")
	}
	
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset AssetConsent
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAssetConsent(ctx contractapi.TransactionContextInterface, id string, consentFlag bool, blockedFlag bool) error {
	validJCA, er := s.CheckJCA (ctx, "UpdateAssetConsent")
	if er != nil{
		return fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return fmt.Errorf("No valid JCA for client ORG found")
	}

	exists, err := s.AssetExistsConsent(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}
	

	// overwriting original asset with new asset
	assetC := AssetConsent{
		ID:	          id,
		ConsentFlag:      consentFlag,
		BlockedFlag:	  blockedFlag,
		//Size:           size,
		//Owner:          owner,
		//AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(assetC)
	if err != nil {
		return err
	}
	
	
	if assetC.ConsentFlag == false{
		log.Printf("Consent set to false")
		//assetU, err := s.ReadAsset(ctx, assetC.ID)
		assetAB, err := ctx.GetStub().GetPrivateData(assetCollection, assetC.ID)
		if err != nil || assetAB==nil{
			log.Printf("Error, or no assets")
			return ctx.GetStub().PutState(id, assetJSON)
		} else {
			var assetU *Asset
			err = json.Unmarshal(assetAB, &assetU)
			log.Printf("UserObject Exists. Trying to delete...")
			ownerCollection, err := getCollectionName(ctx) // Get owners collection
			if err != nil {
				return fmt.Errorf("failed to infer private collection name for the org: %v", err)
			}
			// delete the asset from state
			err = ctx.GetStub().DelPrivateData(assetCollection, assetU.ID)
			if err != nil {
				return fmt.Errorf("failed to delete state: %v", err)
			}

			// Finally, delete private details of asset
			err = ctx.GetStub().DelPrivateData(ownerCollection, assetU.ID)
			if err != nil {
				return err
			}
		}
		
			
	}
	return ctx.GetStub().PutState(id, assetJSON)
}


// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAssetScore(ctx contractapi.TransactionContextInterface, id string, rep int) error {
	validJCA, er := s.CheckJCA (ctx, "UpdateAssetScore")
	if er != nil{
		return fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return fmt.Errorf("No valid JCA for client ORG found")
	}
	
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}
	//Checks consent validity
	assetCo, err := s.ReadAssetConsent(ctx, id)
	if err != nil {
		return fmt.Errorf("Error On ReadAssetConsent call")
	}
	if assetCo.ConsentFlag != true{
		return fmt.Errorf("No consent given")
	} else if assetCo.BlockedFlag != false{
		return fmt.Errorf("Processing is blocked (RoO, RoR)")
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
	collectionName, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("Error On getCollection call")
	}
	assetC, err := s.ReadAssetPrivateDetails(ctx, collectionName, id)
	if err != nil {
		return fmt.Errorf("Error On ReadAssetPrivateDetails call")
	}
	newRep := int((rep + assetC.SSum) / (assetC.SCounter + 1))

	// Make submitting client the owner
	asset := Asset{
		ID:		id,
		Rep:	newRep,
		Owner:	clientID,
	}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}
	
	err = ctx.GetStub().PutPrivateData(assetCollection, asset.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	
	assetPrivateDetails := AssetPrivateDetails{
		ID:         id,
		Name: 		assetC.Name,
		Email: 		assetC.Email,
		Salt:		assetC.Salt,
		SCounter:	assetC.SCounter + 1,
		SSum:		assetC.SSum + rep,
	}

	assetPrivateDetailsAsBytes, err := json.Marshal(assetPrivateDetails) // marshal asset details to JSON
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", collectionName, id)
	err = ctx.GetStub().PutPrivateData(collectionName, id, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	
	return nil
}

// CreateAsset creates a new asset by placing the main asset details in the assetCollection
// that can be read by both organizations. The appraisal value is stored in the owners org specific collection.
func (s *SmartContract) UpdateAssetPersonal(ctx contractapi.TransactionContextInterface) error {
	validJCA, er := s.CheckJCA (ctx, "UpdateAssetPersonal")
	if er != nil{
		return fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return fmt.Errorf("No valid JCA for client ORG found")
	}

	// Get new asset from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["asset_properties"]
	if !ok {
		//log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type assetTransientInput struct {
		ID             string 	`json:"ID"`
		Name           string   `json:"name"`
		Email	       string   `json:"email"`
	}

	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetInput.Name) == 0 {
		return fmt.Errorf("objectType field must be a non-empty string")
	}
	if len(assetInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}
	if len(assetInput.Email) == 0 {
		return fmt.Errorf("email field must be a non-empty string")
	}

	//Check for valid consent
	assetC, err := s.ReadAssetConsent(ctx, assetInput.ID)
	if err != nil {
		return fmt.Errorf("Error On ReadAssetConsent call")
	}
	if assetC.ConsentFlag != true{
		return fmt.Errorf("No consent given")
	} else if assetC.BlockedFlag != false{
		return fmt.Errorf("Processing is blocked (RoO, RoR)")
	}

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes == nil {
		fmt.Println("Asset doesn't exist: " + assetInput.ID)
		return fmt.Errorf("this asset doesn't exist: " + assetInput.ID)
	}

	// Get ID of submitting client identity
	//clientID, err := submittingClientIdentity(ctx)
	//if err != nil {
	//	return err
	//}

	// Verify that the client is submitting request to peer in their organization
	// This is to ensure that a client from another org doesn't attempt to read or
	// write private data from this peer.
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("UpdateAssetPersonal cannot be performed: Error %v", err)
	}
	
	collectionName, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("Error On getCollection call")
	}
	
	//get previously stored personal data
	assetP, err := s.ReadAssetPrivateDetails(ctx, collectionName, assetInput.ID)
	if err != nil {
		return fmt.Errorf("Error On ReadAssetPrivateDetails call")
	}

	// Save asset details to collection visible to owning organization
	assetPrivateDetails := AssetPrivateDetails{
		ID:             assetInput.ID,
		Name: 		assetInput.Name,
		Email: 		assetInput.Email,
		Salt:		assetP.Salt,
		SCounter:	assetP.SCounter,
		SSum:		assetP.SSum,
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
	log.Printf("Put: collection %v, ID %v", orgCollection, assetInput.ID)
	err = ctx.GetStub().PutPrivateData(orgCollection, assetInput.ID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	return nil
}


// CreateUnknownAsset creates a new "unknown asset". These assets are not defined in type, can be "anything".
// PDA will decide previously to the CC invoke which fields are private.
// There is always 3 public fields (id+2), and 3 pvt fields.
func (s *SmartContract) CreateUnknownAsset(ctx contractapi.TransactionContextInterface, id string, unk1 string, unk2 string) error {
	validJCA, er := s.CheckJCA (ctx, "CreateUnknownAsset")
	if er != nil{
		return fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return fmt.Errorf("No valid JCA for client ORG found")
	}	

	// Get pvtVal from transient
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties are private, therefore they get passed in transient field, instead of func args
	transientAssetJSON, ok := transientMap["unkAsset_properties"]
	if !ok {
		//log error to stdout
		return fmt.Errorf("asset not found in the transient map input")
	}

	type assetTransientInput struct {
		//Type           string `json:"objectType"` //Type is used to distinguish the various types of objects in state database
		ID             string 	`json:"ID"`
		UnkPvt1        string   `json:"unkPvt1"`
		UnkPvt2	       string   `json:"unkPvt2"`
		UnkPvt3	       string	`json:"unkPvt3"`
		Salt	       string	`json:"salt"`
	}

	var assetInput assetTransientInput
	err = json.Unmarshal(transientAssetJSON, &assetInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetInput.ID) == 0 {
		return fmt.Errorf("ID field must be a non-empty string")
	} else if (string(assetInput.ID[0])!="u" || assetInput.ID!=id){
		return fmt.Errorf("ID does not start with \"u\", or IDs do not match")
	}
	if len(assetInput.UnkPvt1) == 0 {
		return fmt.Errorf("UnkPvt1 field must be a non-empty string")
	}
	if len(assetInput.UnkPvt2) == 0 {
		return fmt.Errorf("UnkPvt2 field must be a non-empty string")
	}
	if len(assetInput.UnkPvt3) == 0 {
		return fmt.Errorf("UnkPvt3 field must be a non-empty string")
	}
	if len(assetInput.Salt) == 0 {
		return fmt.Errorf("color field must be a non-empty string")
	}

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetInput.ID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + assetInput.ID)
		return fmt.Errorf("this asset already exists: " + assetInput.ID)
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
	asset := UnkAsset{
		ID: 	assetInput.ID,
		Unk1: 	unk1,
		Unk2:	unk2,
		Owner: 	clientID,
	}
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to private data collection
	// Typical logger, logs to stdout/file in the fabric managed docker container, running this chaincode
	// Look for container name like dev-peer0.org1.example.com-{chaincodename_version}-xyz
	log.Printf("CreateAsset Put: collection %v, ID %v, owner %v", assetCollection, assetInput.ID, clientID)

	err = ctx.GetStub().PutPrivateData(assetCollection, assetInput.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}

	/*// Save asset details to collection visible to owning organization
	assetPrivateDetails := UnkAssetPvt{
		ID:             assetInput.ID,
		UnkPvt1: 	assetInput.UnkPvt1,
		UnkPvt2: 	assetInput.UnkPvt2,
		UnkPvt3:	assetInput.UnkPvt3,
		Salt:		assetInput.Salt,
	}*/

	assetPrivateDetailsAsBytes, err := json.Marshal(assetInput) // marshal asset details to JSON
	if err != nil {
		return fmt.Errorf("failed to marshal into JSON: %v", err)
	}

	// Get collection name for this organization.
	orgCollection, err := getCollectionName(ctx)
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	// Put asset appraised value into owners org specific private data collection
	log.Printf("Put: collection %v, ID %v", orgCollection, assetInput.ID)
	err = ctx.GetStub().PutPrivateData(orgCollection, assetInput.ID, assetPrivateDetailsAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset private details: %v", err)
	}
	return nil
}

func (s *SmartContract) ReadUnknownAsset(ctx contractapi.TransactionContextInterface, id string) (*UnkAsset, error) {
	validJCA, er := s.CheckJCA (ctx, "ReadUnknownAsset")
	if er != nil{
		return nil, fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return nil, fmt.Errorf("No valid JCA for client ORG found")
	}
	log.Printf("ReadUnknownAsset: collection %v, ID %v", assetCollection, id)
	assetJSON, err := ctx.GetStub().GetPrivateData(assetCollection, id)
	
	if err != nil {
		return nil, fmt.Errorf("failed to read from assetCollection: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset UnkAsset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *SmartContract) ReadUnknownAssetPrivateDetails(ctx contractapi.TransactionContextInterface, collection string, assetID string) (*UnkAssetPvt, error) {
	validJCA, er := s.CheckJCA (ctx, "ReadUnknownAssetPrivateDetails")
	if er != nil{
		return nil, fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return nil, fmt.Errorf("No valid JCA for client ORG found")
	}
	log.Printf("ReadUnknownAssetPrivateDetails: collection %v, ID %v", collection, assetID)
	assetDetailsJSON, err := ctx.GetStub().GetPrivateData(collection, assetID) // Get the asset from chaincode state
	if err != nil {
		return nil, fmt.Errorf("failed to read asset details: %v", err)
	}
	if assetDetailsJSON == nil {
		log.Printf("AssetPrivateDetails for %v does not exist in collection %v", assetID, collection)
		return nil, nil
	}

	var assetDetails *UnkAssetPvt
	err = json.Unmarshal(assetDetailsJSON, &assetDetails)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	return assetDetails, nil
}





// DeleteAsset can be used by the owner of the asset to delete the asset
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface) error {

	validJCA, er := s.CheckJCA (ctx, "DeleteAsset")
	if er != nil{
		return fmt.Errorf("Error checking for valid JCA: %v", er)
	} else if !validJCA {
		return fmt.Errorf("No valid JCA for client ORG found")
	}

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
		ID string `json:"ID"`
	}

	var assetDeleteInput assetDelete
	err = json.Unmarshal(transientDeleteJSON, &assetDeleteInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	if len(assetDeleteInput.ID) == 0 {
		return fmt.Errorf("assetID field must be a non-empty string")
	}

	// Verify that the client is submitting request to peer in their organization
	err = verifyClientOrgMatchesPeerOrg(ctx)
	if err != nil {
		return fmt.Errorf("DeleteAsset cannot be performed: Error %v", err)
	}

	log.Printf("Deleting Asset: %v", assetDeleteInput.ID)
	valAsbytes, err := ctx.GetStub().GetPrivateData(assetCollection, assetDeleteInput.ID) //get the asset from chaincode state
	if err != nil {
		return fmt.Errorf("failed to read asset: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset not found: %v", assetDeleteInput.ID)
	}

	ownerCollection, err := getCollectionName(ctx) // Get owners collection
	if err != nil {
		return fmt.Errorf("failed to infer private collection name for the org: %v", err)
	}

	//check the asset is in the caller org's private collection
	valAsbytes, err = ctx.GetStub().GetPrivateData(ownerCollection, assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("failed to read asset from owner's Collection: %v", err)
	}
	if valAsbytes == nil {
		return fmt.Errorf("asset not found in owner's private Collection %v: %v", ownerCollection, assetDeleteInput.ID)
	}

	// delete the asset from state
	err = ctx.GetStub().DelPrivateData(assetCollection, assetDeleteInput.ID)
	if err != nil {
		return fmt.Errorf("failed to delete state: %v", err)
	}

	// Finally, delete private details of asset
	err = ctx.GetStub().DelPrivateData(ownerCollection, assetDeleteInput.ID)
	if err != nil {
		return err
	}

	return nil

}


func (s *SmartContract) CheckJCA(ctx contractapi.TransactionContextInterface, action string) (bool,error){
	
	// Get the MSP ID of submitting client identity
	orgId, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return false, fmt.Errorf("failed to get verified MSPID: %v", err)
	}

	stub := ctx.GetStub()
	params:= []string {"CheckOrgConsent", orgId, action}
	queryArgs := make ([][]byte, len(params))
	for i,arg :=range params {
		queryArgs[i] = []byte(arg) 
	}
	
	response := stub.InvokeChaincode("JCA", queryArgs, "mychannel")
	if response.Status != shim.OK{
		return false, fmt.Errorf("Error: %s", response.Payload)
	}
	
	JCABoolValue, err := strconv.ParseBool(string(response.Payload))
	
	return JCABoolValue, nil
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
