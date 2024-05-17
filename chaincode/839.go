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

const assetCollection = "JCACollection"

// SmartContract of this fabric sample
type SmartContract struct {
	contractapi.Contract
}


type Asset2 struct {
	ID 		string  `json:"ID"`
	ReadFlag     bool 	`json:"ReadFlag"`
	ConsentFlag     bool 	`json:"ConsentFlag"`
	Owner string `json:"owner"`
}

type LogAsset struct {
	ID 		string	`json:"ID"`
	Action		string	`json:"Action"`
	Peer		string	`json:"Peer"`
	JCAAnswer	bool	`json:"JCAAnswer"`
}

type CounterAsset struct {
	ID		string	`json:"ID"`
	CounterVal	int	`json:"CounterVal"`
}

// CreateAsset creates a new asset by placing the main asset details in the assetCollection
// that can be read by both organizations. The appraisal value is stored in the owners org specific collection.
func (s *SmartContract) CreateAssetConsent(ctx contractapi.TransactionContextInterface, id string, readFlag bool, consentFlag bool) error {

	// Check if asset already exists
	assetAsBytes, err := ctx.GetStub().GetPrivateData(assetCollection, id)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	} else if assetAsBytes != nil {
		fmt.Println("Asset already exists: " + id)
		return fmt.Errorf("this asset already exists: " + id + string(assetAsBytes))
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
	asset := Asset2{
		ID:  id,
		ReadFlag:  readFlag,
		ConsentFlag: consentFlag,
		Owner: clientID,
	}
	
	
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}

	// Save asset to private data collection
	// Typical logger, logs to stdout/file in the fabric managed docker container, running this chaincode
	// Look for container name like dev-peer0.org1.example.com-{chaincodename_version}-xyz
	log.Printf("CreateAsset Put: collection %v, ID %v, owner %v", assetCollection, asset.ID, clientID)

	err = ctx.GetStub().PutPrivateData(assetCollection, asset.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}

	return nil
}

func (s *SmartContract) CheckOrgConsent(ctx contractapi.TransactionContextInterface, id string, action string) (bool, error) {
	log.Printf("ReadAsset: collection %v, ID %v", assetCollection, id)
	asset, err := s.ReadAsset1(ctx, id)
	
	if err != nil {
		//s.CreateLogging(ctx, action, false)
		return false, fmt.Errorf("failed to read asset: %v", err)
	}
	if asset == nil{
		//s.CreateLogging(ctx, action, false)
		return false, fmt.Errorf("no asset found: %v", err)
	}

	//if (action[:6]=="Create"){
	log.Printf("ReadAsset(JCA): WIll create log...")
	err = s.CreateLogging(ctx, action, asset.ConsentFlag)
	//}
	return asset.ConsentFlag, nil
}

func (s *SmartContract) GetCounterValue(ctx contractapi.TransactionContextInterface) (int, error) {
	log.Printf("ReadAssetCounter: collection %v, ID MainCounter", assetCollection)
	assetJSON, err := ctx.GetStub().GetPrivateData(assetCollection, "MainCounter") //get the asset from chaincode state
	if err != nil {
		return 0, fmt.Errorf("failed to read asset: %v", err)
	}
	newVal :=0
	//No Asset found, create it
	if assetJSON == nil {
		newVal = 1
	} else {
		var assetC CounterAsset
		err = json.Unmarshal(assetJSON, &assetC)
		if err != nil {
			return 0, fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		newVal = assetC.CounterVal+1	
	}
	
	assetCN := CounterAsset{
			ID:		"MainCounter",
			CounterVal:	newVal,
		}
	assetJSONasBytes, err := json.Marshal(assetCN)
	if err != nil {
		return 0,fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}
	
	err = ctx.GetStub().PutPrivateData(assetCollection, "MainCounter", assetJSONasBytes)
	if err != nil {
		return 0,fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	return newVal, err
}


//Func create Log, will create a logging asset, independently of the JCA answer.
//The asset will log the invoker and the invoke. It will also log the JCA answer.
func (s *SmartContract) CreateLogging(ctx contractapi.TransactionContextInterface, action string, jcaAnswer bool) (error) {
	var idStr string

	transientMap, err := ctx.GetStub().GetTransient()
	//Some info will now come in transient form.
	transientAssetJSON, ok := transientMap["JCA_properties"]
	if !ok {
		//log error to stdout
		//return fmt.Errorf("asset not found in the transient map input")
		fmt.Print("asset not found in the transient map input, will use GetCounterValue")
		//Get ID from CounterAsset
		id, err := s.GetCounterValue(ctx)
		if err != nil {
			return fmt.Errorf("failed to get counter val: %v", err)
		}
		idStr = "cv" + strconv.Itoa(id)
	} else {
		type assetTransientInput struct {
			//Type           string `json:"objectType"` //Type is used to distinguish the various types of objects in state database
			ID             string 	`json:"ID"`
		}
		
		var assetInput assetTransientInput
		
		err = json.Unmarshal(transientAssetJSON, &assetInput)
		if err != nil {
			return fmt.Errorf("failed to unmarshal JSON: %v", err)
		}
		idStr= assetInput.ID
	}
	
	// Get ID of submitting client identity
	clientID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
	
	asset := LogAsset{
		ID:  idStr,
		Action:  action,
		Peer: clientID,
		JCAAnswer: jcaAnswer,
	}
	
	assetJSONasBytes, err := json.Marshal(asset)
	if err != nil {
		return fmt.Errorf("failed to marshal asset into JSON: %v", err)
	}
	
	// Save asset to private data collection
	// Typical logger, logs to stdout/file in the fabric managed docker container, running this chaincode
	// Look for container name like dev-peer0.org1.example.com-{chaincodename_version}-xyz
	log.Printf("CreateAsset Put: collection %v, ID %v, owner %v", assetCollection, asset.ID, clientID)

	err = ctx.GetStub().PutPrivateData(assetCollection, asset.ID, assetJSONasBytes)
	if err != nil {
		return fmt.Errorf("failed to put asset into private data collecton: %v", err)
	}
	
	return err
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

/*func (s *SmartContract) GetConsentBasic(ctx contractapi.TransactionContextInterface) (string,error){
	stub := ctx.GetStub()
	params:= []string {"ReadAssetConsent", "asset1"}
	queryArgs := make ([][]byte, len(params))
	for i,arg :=range params {
		queryArgs[i] = []byte(arg) 
	}
	
	response := stub.InvokeChaincode("basic", queryArgs, "mychannel")
	if response.Status != shim.OK{
		return "fail", fmt.Errorf("Error: %s", response.Payload)
	}
	return string(response.Payload), nil
}

func (s *SmartContract) GetAssetBasic(ctx contractapi.TransactionContextInterface) (string,error){
	stub := ctx.GetStub()
	params:= []string {"ReadAsset", "asset1"}
	queryArgs := make ([][]byte, len(params))
	for i,arg :=range params {
		queryArgs[i] = []byte(arg) 
	}
	
	response := stub.InvokeChaincode("basic", queryArgs, "mychannel")
	if response.Status != shim.OK{
		return "fail", fmt.Errorf("Error: %s", response.Payload)
	}
	return string(response.Payload), nil
}

func (s *SmartContract) GetAssetPvtBasic(ctx contractapi.TransactionContextInterface) (string,error){
	stub := ctx.GetStub()
	collectionName, err := getCollectionName(ctx)
	if err != nil {
		return "fail", fmt.Errorf("Error On getCollection call")
	}
	params:= []string {"ReadAssetPrivateDetails", collectionName, "asset1"}
	queryArgs := make ([][]byte, len(params))
	for i,arg :=range params {
		queryArgs[i] = []byte(arg) 
	}
	
	response := stub.InvokeChaincode("basic", queryArgs, "mychannel")
	if response.Status != shim.OK{
		return "fail", fmt.Errorf("Error: %s", response.Payload)
	}
	return string(response.Payload), nil
}*/
