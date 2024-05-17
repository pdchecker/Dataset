/******************************* Step1 *****************************
Import dependencies and Define smart Contract
*/

package main

import (
  "fmt"
  "log"
  "encoding/json"
  "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
   type SmartContract struct {
   contractapi.Contract
   }






/******************************* Step2 *****************************
Declare a structure to store data on ledger
*/

// OurStruct describes basic details
   type OurStruct struct {
    ID  string `json:"ID"`
   }






/******************************* Step3 *****************************
Write functions to interact with the ledger using our Smart Contract
*/

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string) error {
    // Create an object of type OurStruct
    asset := OurStruct{
        ID: id,
    }
    
    // Convert this object into json
    assetJSON, err := json.Marshal(asset)
    if err != nil {
        return err
    }
    
    key := "Key"
	
	// Store This Json into the Ledger using the PutState function
	// PutSatate Function takes input as key, value(JSON)
    return ctx.GetStub().PutState(key, assetJSON)
}





// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*OurStruct, error) {
    
    // This will amke a query to the ledger and check if there is an entry with this id or not
    // On sucess, GetState(id) will return the JSON object contains the data corrsponding to the key:id
    assetJSON, err := ctx.GetStub().GetState(id)
    if err != nil {
      return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if assetJSON == nil {
      return nil, fmt.Errorf("the asset %s does not exist", id)
    }

    
    // Convert That JSON to object of type OurStruct
    var asset OurStruct
    err = json.Unmarshal(assetJSON, &asset)
    if err != nil {
      return nil, err
    }
	
	// Return the object
    return &asset, nil
}



// The main function which will create the chaincode and start it
func main(){
	// NewChaincode creates a new chaincode using contracts passed.
    assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
    if err != nil {
      log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
    }

	// Start starts the chaincode in the fabric
    if err := assetChaincode.Start(); err != nil {
      log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
    }
}



