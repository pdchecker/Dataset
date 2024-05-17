/*
 SPDX-License-Identifier: Apache-2.0
*/

/*
====CHAINCODE EXECUTION SAMPLES (CLI) ==================

==== Invoke assets ====
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["CreateAsset","asset1","blue","5","tom","35"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["CreateAsset","asset2","red","4","tom","50"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["CreateAsset","asset3","blue","6","tom","70"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["TransferAsset","asset2","jerry"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["TransferAssetByColor","blue","jerry"]}'
peer chaincode invoke -C myc1 -n asset_transfer -c '{"Args":["DeleteAsset","asset1"]}'

==== Query assets ====
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["ReadAsset","asset1"]}'
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["GetAssetsByRange","asset1","asset3"]}'
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["GetAssetHistory","asset1"]}'

Rich Query (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssetsByOwner","tom"]}'
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssets","{\"selector\":{\"owner\":\"tom\"}}"]}'

Rich Query with Pagination (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssetsWithPagination","{\"selector\":{\"owner\":\"tom\"}}","3",""]}'

INDEXES TO SUPPORT COUCHDB RICH QUERIES

Indexes in CouchDB are required in order to make JSON queries efficient and are required for
any JSON query with a sort. Indexes may be packaged alongside
chaincode in a META-INF/statedb/couchdb/indexes directory. Each index must be defined in its own
text file with extension *.json with the index definition formatted in JSON following the
CouchDB index JSON syntax as documented at:
http://docs.couchdb.org/en/2.3.1/api/database/find.html#db-index

This asset transfer ledger example chaincode demonstrates a packaged
index which you can find in META-INF/statedb/couchdb/indexes/indexOwner.json.

If you have access to the your peer's CouchDB state database in a development environment,
you may want to iteratively test various indexes in support of your chaincode queries.  You
can use the CouchDB Fauxton interface or a command line curl utility to create and update
indexes. Then once you finalize an index, include the index definition alongside your
chaincode in the META-INF/statedb/couchdb/indexes directory, for packaging and deployment
to managed environments.

In the examples below you can find index definitions that support asset transfer ledger
chaincode queries, along with the syntax that you can use in development environments
to create the indexes in the CouchDB Fauxton interface or a curl command line utility.


Index for docType, owner.

Example curl command line to define index in the CouchDB channel_chaincode database
curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[\"docType\",\"owner\"]},\"name\":\"indexOwner\",\"ddoc\":\"indexOwnerDoc\",\"type\":\"json\"}" http://hostname:port/myc1_assets/_index


Index for docType, owner, size (descending order).

Example curl command line to define index in the CouchDB channel_chaincode database:
curl -i -X POST -H "Content-Type: application/json" -d "{\"index\":{\"fields\":[{\"size\":\"desc\"},{\"docType\":\"desc\"},{\"owner\":\"desc\"}]},\"ddoc\":\"indexSizeSortDoc\", \"name\":\"indexSizeSortDesc\",\"type\":\"json\"}" http://hostname:port/myc1_assets/_index

Rich Query with index design doc and index name specified (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssets","{\"selector\":{\"docType\":\"asset\",\"owner\":\"tom\"}, \"use_index\":[\"_design/indexOwnerDoc\", \"indexOwner\"]}"]}'

Rich Query with index design doc specified only (Only supported if CouchDB is used as state database):
peer chaincode query -C myc1 -n asset_transfer -c '{"Args":["QueryAssets","{\"selector\":{\"docType\":{\"$eq\":\"asset\"},\"owner\":{\"$eq\":\"tom\"},\"size\":{\"$gt\":0}},\"fields\":[\"docType\",\"owner\",\"size\"],\"sort\":[{\"size\":\"desc\"}],\"use_index\":\"_design/indexSizeSortDoc\"}"]}'
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/golang/protobuf/ptypes"
	//"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	//pb "github.com/hyperledger/fabric-protos-go/peer"
)


// SimpleChaincode implements the fabric-contract-api-go programming model
type SimpleChaincode struct {
	contractapi.Contract
}

type Asset struct {
	DocType        string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID             string `json:"ID"`      //the field tags are needed to keep case from bouncing around
	Class          string `json:"class"`
	Version           int `json:"version"`
	Other          string `json:"other"`
}

/*type AssetSRB struct {
	DocType        string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID             string `json:"ID"`      //the field tags are needed to keep case from bouncing around
	ItemID         string `json:"itemID"`
	FromOwner      string `json:"fromOwner"`
	FromPSB        string `json:"fromPSB"`
	ToOwner        string `json:"ToOwner"`
	ToPSB          string `json:"ToPSB"`
}*/

type AssetSRB struct {
	DocType        string `json:"docType"` //docType is used to distinguish the various types of objects in state database
	ID             string `json:"ID"`      //the field tags are needed to keep case from bouncing around
	Group          string `json:"group"`
	InPSB          string `json:"inPSB"`
}
/*
type record struct {
	Record		*AssetSRB    `json:"record"`
}
*/
// HistoryQueryResult structure used for returning result of history query
type record struct {
	Record    *AssetSRB  `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// HistoryQueryResult structure used for returning result of history query
type HistoryQueryResult struct {
	Record    *Asset    `json:"record"`
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
	IsDelete  bool      `json:"isDelete"`
}

// PaginatedQueryResult structure used for returning paginated query results and metadata
type PaginatedQueryResult struct {
	Records             []*Asset `json:"records"`
	FetchedRecordsCount int32    `json:"fetchedRecordsCount"`
	Bookmark            string   `json:"bookmark"`
}


// TransferAsset transfers an asset by setting a new owner name on the asset
func (t *SimpleChaincode) UpdateAsset(ctx contractapi.TransactionContextInterface, assetID, class, upperPSB, currPSB string) error {
	//1 check whether asst exist in current PSB
	exists, err := t.AssetExists(ctx, assetID)
	if err != nil {
		return fmt.Errorf("1.1 failed to get asset: %v", err)
	}	
	if exists {
		//1.1 if exist, update	
		return t.TransferAsset(ctx, assetID, "HR")
	}

	//2 if not, query for SRB
	/*
	asset, err := t.ReadAsset(ctx, assetID)
	if err != nil {
		return err
	}*/
	bfunc := []byte("GetAssetHistory")
	basset := []byte(class)
	args :=  [][]byte{bfunc, basset}
	SRBresp := ctx.GetStub().InvokeChaincode("srbCC", args, "srb")
	SRBstatus := SRBresp.GetStatus()	
	if SRBstatus != 200 {
		return fmt.Errorf("1.2 fail to invoke srbCC %d", SRBstatus)
	}
	
	
	//3 check whether exist uppstream in SRB
	if len(SRBresp.GetPayload()) == 0 {
		//3.1 if not, create new asset
		return t.CreateAsset(ctx, assetID, class, "hr")
	}
	
	//4 if exist, findout the upper PSB
	var records []record
	err = json.Unmarshal(SRBresp.GetPayload(), &records)
	if err != nil {
		return fmt.Errorf("1.2 failed to extract proposal payload: %s", err)
	}
	
	var targetRecord = -1
	for k := range records {
		if(records[k].Record.InPSB == currPSB && records[k+1].Record.InPSB == upperPSB) {
			targetRecord = k+1
			break
		}
	}
	if targetRecord == -1 {
		//4.1 if not, reject transaction
		return fmt.Errorf("4.1 failed to find the item %s in psb%s", assetID, upperPSB)
	}
	
	//5 if exist, check if it is ready in upper PSB	
	bfunc = []byte("ReadAsset")
	basset = []byte(assetID)
	args =  [][]byte{bfunc, basset}
	PSBresp := ctx.GetStub().InvokeChaincode("psb"+upperPSB+"CC", args, "psb"+upperPSB)
	PSBstatus := PSBresp.GetStatus()		
	if PSBstatus != 200 {
		//5.1 if not, reject
		return fmt.Errorf("5.1 fail to invoke psbCC, asset in psb%s is unexist", upperPSB)
	}
					
	//6 if is ready, create Asset history in currPSB
	return t.CreateAsset(ctx, assetID, class, "hr")
}


// CreateAsset initializes a new asset in the ledger
func (t *SimpleChaincode) CreateAsset(ctx contractapi.TransactionContextInterface, assetID string, class string, owner string) error {
	exists, err := t.AssetExists(ctx, assetID)
	if err != nil {
		return fmt.Errorf("failed to get asset: %v", err)
	}
	if exists {
		return fmt.Errorf("asset already exists: %s", assetID)
	}

	asset := &Asset{
		DocType:        "asset",
		ID:             assetID,
		Class:          class,
		Version:        0,
		Other:          owner,
	}
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(assetID, assetBytes)
	if err != nil {
		return err
	}
	
	return nil

}

// ReadAsset retrieves an asset from the ledger
func (t *SimpleChaincode) ReadAsset(ctx contractapi.TransactionContextInterface, assetID string) (*Asset, error) {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get asset %s: %v", assetID, err)
	}
	if assetBytes == nil {
		return nil, fmt.Errorf("asset %s does not exist", assetID)
	}

	var asset Asset
	err = json.Unmarshal(assetBytes, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// DeleteAsset removes an asset key-value pair from the ledger
func (t *SimpleChaincode) DeleteAsset(ctx contractapi.TransactionContextInterface, assetID string) error {
	return ctx.GetStub().DelState(assetID)
}


// TransferAsset transfers an asset by setting a new owner name on the asset
func (t *SimpleChaincode) TransferAsset(ctx contractapi.TransactionContextInterface, assetID, newOwner string) error {
	asset, err := t.ReadAsset(ctx, assetID)
	if err != nil {
		return err
	}

	asset.Other = newOwner
	asset.Version = asset.Version+1
	assetBytes, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(assetID, assetBytes)
}

// GetAssetHistory returns the chain of custody for an asset since issuance.
func (t *SimpleChaincode) GetAssetHistory(ctx contractapi.TransactionContextInterface, assetID string) ([]HistoryQueryResult, error) {
	log.Printf("GetAssetHistory: ID %v", assetID)

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(assetID)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var records []HistoryQueryResult
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		if len(response.Value) > 0 {
			err = json.Unmarshal(response.Value, &asset)
			if err != nil {
				return nil, err
			}
		} else {
			asset = Asset{
				ID: assetID,
			}
		}

		timestamp, err := ptypes.Timestamp(response.Timestamp)
		if err != nil {
			return nil, err
		}

		record := HistoryQueryResult{
			TxId:      response.TxId,
			Timestamp: timestamp,
			Record:    &asset,
			IsDelete:  response.IsDelete,
		}
		records = append(records, record)
	}

	return records, nil
}

// AssetExists returns true when asset with given ID exists in the ledger.
func (t *SimpleChaincode) AssetExists(ctx contractapi.TransactionContextInterface, assetID string) (bool, error) {
	assetBytes, err := ctx.GetStub().GetState(assetID)
	if err != nil {
		return false, fmt.Errorf("failed to read asset %s from world state. %v", assetID, err)
	}

	return assetBytes != nil, nil
}


func main() {
	chaincode, err := contractapi.NewChaincode(&SimpleChaincode{})
	if err != nil {
		log.Panicf("Error creating asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}
