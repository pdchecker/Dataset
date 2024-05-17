/*
 SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// Asset struct and properties must be exported (start with capitals) to work with contract api metadata
type Asset struct {
	Name     string `json:"name"`
	NumItems int    `json:"numItems"`
	Price    int    `json:"price"`
	OwnerOrg string `json:"ownerOrg"`
}

type transientAsset struct {
	Name     string `json:"name"`
	NumItems int    `json:"numItems"`
	Price    int    `json:"price"`
}

type transientBalance struct {
	Balance int `json:"balance"`
}

// AddItem creates an asset, sets it as owned by the client's org and returns its id
// the id of the asset corresponds to the hash of the properties of the asset that are  passed by transiet field
// The asset is stored in implicit Private Data Collection of the client's org
func (s *SmartContract) AddItem(ctx contractapi.TransactionContextInterface) (string, error) {
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties must be retrieved from the transient field as they are private
	immutablePropertiesJSON, ok := transientMap["asset_properties"]
	if !ok {
		return "", fmt.Errorf("asset_properties key not found in the transient map")
	}

	// AssetID will be the hash of the asset's properties
	// hash := sha256.New()
	// hash.Write(immutablePropertiesJSON)
	// assetID := hex.EncodeToString(hash.Sum(nil))

	// Get the clientOrgId from the input, will be used for implicit collection, owner, and state-based endorsement policy
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return "", err
	}

	// Retrieve Name, NumItems, and Price from the transient map
	var transAsset transientAsset
	err = json.Unmarshal(immutablePropertiesJSON, &transAsset)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Check if asset already exists, in private data collection of the client's org
	collection := buildCollectionName(clientOrgID)
	assetBytes, err := ctx.GetStub().GetPrivateData(collection, transAsset.Name)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	var asset Asset
	var updatedAssetBytes []byte
	if assetBytes != nil {
		// Unmarshal the assetBytes
		err = json.Unmarshal(assetBytes, &asset)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal asset: %v", err)
		}
		// Update the asset with the new NumItems and Price
		asset.NumItems += transAsset.NumItems
		asset.Price = transAsset.Price
		updatedAssetBytes, err = json.Marshal(asset)
		if err != nil {
			return "", fmt.Errorf("failed to marshal asset: %v", err)
		}
	} else {
		asset = Asset{
			// ObjectType: "asset",
			// ID:         assetID,
			Name:     transAsset.Name,
			NumItems: transAsset.NumItems,
			Price:    transAsset.Price,
			OwnerOrg: clientOrgID,
		}
		updatedAssetBytes, err = json.Marshal(asset)
		if err != nil {
			return "", fmt.Errorf("failed to create asset JSON: %v", err)
		}
	}

	// err = ctx.GetStub().PutState(assetID, assetBytes)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to put asset in public data: %v", err)
	// }

	// Persist private immutable asset properties to owner's private data collection
	err = ctx.GetStub().PutPrivateData(collection, asset.Name, updatedAssetBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put Asset private details: %v", err)
	}
	// Return a concatenated string of asset Name, NumItems, and Price
	return fmt.Sprintf("Added item to Private Data:-\nName: %s  NumItems: %d  Price: %d  OwnerOrg: %s", asset.Name, asset.NumItems, asset.Price, asset.OwnerOrg), nil
	// return asset.Name, nil
}

// AddBalance adds the balance to the client's public data, reading the current balance from the ledger and
// adding the balance from transient field to it.
func (s *SmartContract) AddBalance(ctx contractapi.TransactionContextInterface) (string, error) {
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}

	// Asset properties must be retrieved from the transient field as they are private
	immutablePropertiesJSON, ok := transientMap["balance"]
	if !ok {
		return "", fmt.Errorf("balance key not found in the input transient map")
	}

	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return "", err
	}
	balanceKey, _ := ctx.GetStub().CreateCompositeKey("balance", []string{clientOrgID})

	// Retrieve Balance from the transient map
	var transBalance transientBalance
	err = json.Unmarshal(immutablePropertiesJSON, &transBalance)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	// Check if balance already exists, in public data collection of the client's org
	assetBytes, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	var balance int
	if assetBytes != nil {
		// Unmarshal the assetBytes
		err = json.Unmarshal(assetBytes, &balance)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal asset: %v", err)
		}
		// Update the balance
		balance += transBalance.Balance
	} else {
		balance = transBalance.Balance
	}

	// Put the balance in public data collection of the client's org
	err = ctx.GetStub().PutState(balanceKey, []byte(fmt.Sprintf("%d", balance)))
	if err != nil {
		return "", fmt.Errorf("failed to put asset in public data: %v", err)
	}

	return fmt.Sprintf("Added balance to Public Data:-\nBalance: %d  OwnerOrg: %s", balance, clientOrgID), nil
}

// AddToMarket adds the asset to the market, checks the private data collection of the client's org for the asset and
// Add to public ledger and remove from private data collection
func (s *SmartContract) AddToMarket(ctx contractapi.TransactionContextInterface, name string, price int) (string, error) {
	// Check if asset already exists, in private data collection of the client's org
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return "", err
	}

	collection := buildCollectionName(clientOrgID)
	assetBytes, err := ctx.GetStub().GetPrivateData(collection, name)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	var asset Asset
	if assetBytes != nil {
		// Asset is present in private data, remove 1 quantity from it and add to public ledger with price
		// Unmarshal the assetBytes
		err = json.Unmarshal(assetBytes, &asset)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal asset: %v", err)
		}
		// Update the asset with the new NumItems
		asset.NumItems -= 1

		// Put the updated asset back in private data collection of the client's org, if NumItems is 0, delete the asset
		if asset.NumItems == 0 {
			err = ctx.GetStub().DelPrivateData(collection, name)
			if err != nil {
				return "", fmt.Errorf("failed to delete asset from implicit private data collection for seller: %v", err)
			}
		} else {
			updatedAssetBytes, err := json.Marshal(asset)
			if err != nil {
				return "", fmt.Errorf("failed to marshal asset: %v", err)
			}
			err = ctx.GetStub().PutPrivateData(collection, name, updatedAssetBytes)
			if err != nil {
				return "", fmt.Errorf("failed to put Asset private details: %v", err)
			}
		}

		// Add the asset to public ledger
		// Check if asset already exists, in public data, if yes update the quantity and price
		// Create a composite key for the asset with name and clientOrgID as attributes
		assetKey, err := ctx.GetStub().CreateCompositeKey("asset", []string{name, clientOrgID})
		if err != nil {
			return "", fmt.Errorf("failed to create composite key: %v", err)
		}

		// Check if asset already exists, in public data collection of the client's org
		assetBytes, err = ctx.GetStub().GetState(assetKey)
		if err != nil {
			return "", fmt.Errorf("failed to read from world state: %v", err)
		}
		var updatedAssetBytes []byte
		if assetBytes != nil {
			// Unmarshal the assetBytes
			err = json.Unmarshal(assetBytes, &asset)
			if err != nil {
				return "", fmt.Errorf("failed to unmarshal asset: %v", err)
			}
			// Update the asset with the new NumItems and Price
			asset.NumItems += 1
			asset.Price = price
			updatedAssetBytes, err = json.Marshal(asset)
			if err != nil {
				return "", fmt.Errorf("failed to marshal asset: %v", err)
			}
		} else {
			asset = Asset{
				Name:     name,
				NumItems: 1,
				Price:    price,
				OwnerOrg: clientOrgID,
			}
			updatedAssetBytes, err = json.Marshal(asset)
			if err != nil {
				return "", fmt.Errorf("failed to create asset JSON: %v", err)
			}
		}

		err = ctx.GetStub().PutState(assetKey, updatedAssetBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put asset in public data: %v", err)
		}
		// Emit an event for the asset that is added to market
		payloadAsBytes := []byte(asset.Name)
		err = ctx.GetStub().SetEvent("AddToMarket_"+clientOrgID, payloadAsBytes)
		if err != nil {
			return "", fmt.Errorf("failed to set event: %v", err)
		}
	} else {
		return "", fmt.Errorf("asset does not exist in the private data collection of the client's org")
	}

	return fmt.Sprintf("Enlisted item to Market:-\nName: %s  NumItems: %d  Price: %d  OwnerOrg: %s", asset.Name, asset.NumItems, asset.Price, asset.OwnerOrg), nil
}

// BuyFromMarket buys the asset from the market, checks the public data collection of the client's org for the asset and balance of buyer and
// Add asset to private ledger of buyer and remove from public data collection, also update the balance of buyer and seller
func (s *SmartContract) BuyFromMarket(ctx contractapi.TransactionContextInterface, name string) (string, error) {
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return "", err
	}

	// Check whether asset exists or not in public data
	assetIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("asset", []string{name})
	if err != nil {
		return "", fmt.Errorf("failed to get asset: %v", err)
	}
	defer assetIterator.Close()

	var asset Asset
	var assetKey string
	var foundAsset bool = false
	for assetIterator.HasNext() {
		response, err := assetIterator.Next()
		if err != nil {
			return "", err
		}
		assetKey = response.Key
		assetBytes := response.Value
		// Verify that the buyer and seller are from different orgs
		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(assetKey)
		if err != nil {
			return "", fmt.Errorf("failed to split composite key: %v", err)
		}
		if compositeKeyParts[1] == clientOrgID {
			continue
		}
		// Unmarshal the assetBytes
		err = json.Unmarshal(assetBytes, &asset)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal asset: %v", err)
		}
		foundAsset = true
		break
	}

	// Check whether asset avaialble or not
	if !foundAsset || asset.OwnerOrg == clientOrgID {
		return "", fmt.Errorf("no seller is selling the Item in market")
	} else if asset.NumItems == 0 {
		// Delete the asset from public data
		err = ctx.GetStub().DelState(assetKey)
		if err != nil {
			return "", fmt.Errorf("failed to delete asset from public data collection for seller: %v", err)
		}
		return "", fmt.Errorf("zero quantity Item detected, deleted the Item from market")
	}

	// Check whether buyer has enough balance or not
	balanceKey, _ := ctx.GetStub().CreateCompositeKey("balance", []string{clientOrgID})
	balanceBytes, err := ctx.GetStub().GetState(balanceKey)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}

	var balance int
	if balanceBytes != nil {
		// Unmarshal the balanceBytes
		err = json.Unmarshal(balanceBytes, &balance)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal balance: %v", err)
		}
	} else {
		return "", fmt.Errorf("buyer does not have a balance account")
	}

	// Check whether buyer has enough balance or not
	if balance < asset.Price {
		return "", fmt.Errorf("buyer does not have enough balance")
	}

	// Seller balance
	sellerBalanceKey, _ := ctx.GetStub().CreateCompositeKey("balance", []string{asset.OwnerOrg})
	sellerBalanceBytes, err := ctx.GetStub().GetState(sellerBalanceKey)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}

	var sellerBalance int
	if sellerBalanceBytes != nil {
		// Unmarshal the balanceBytes
		err = json.Unmarshal(sellerBalanceBytes, &sellerBalance)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal balance: %v", err)
		}
	} else {
		return "", fmt.Errorf("seller does not have a balance account")
	}

	// Check if asset already exists, in private data collection of the client's org
	collection := buildCollectionName(clientOrgID)
	privateAssetBytes, err := ctx.GetStub().GetPrivateData(collection, name)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	var privateAsset Asset
	if privateAssetBytes != nil {
		// Asset is present in private data, add 1 quantity to it and remove from public ledger
		// Unmarshal the assetBytes
		err = json.Unmarshal(privateAssetBytes, &privateAsset)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal asset: %v", err)
		}
		// Update the asset with the new NumItems
		privateAsset.NumItems += 1
	} else {
		// Asset is not present in private data, create it and add 1 quantity to it
		privateAsset = Asset{
			Name:     name,
			NumItems: 1,
			Price:    asset.Price,
			OwnerOrg: clientOrgID,
		}
	}
	privateUpdatedAssetBytes, err := json.Marshal(privateAsset)
	if err != nil {
		return "", fmt.Errorf("failed to create asset JSON: %v", err)
	}

	// Update Public ledger
	// Update the balance of buyer and seller
	balance -= asset.Price
	err = ctx.GetStub().PutState(balanceKey, []byte(fmt.Sprintf("%d", balance)))
	if err != nil {
		return "", fmt.Errorf("failed to put asset in public data: %v", err)
	}

	sellerBalance += asset.Price
	err = ctx.GetStub().PutState(sellerBalanceKey, []byte(fmt.Sprintf("%d", sellerBalance)))
	if err != nil {
		return "", fmt.Errorf("failed to put asset in public data: %v", err)
	}
	// Update the asset with the new NumItems
	asset.NumItems -= 1
	if asset.NumItems == 0 {
		err = ctx.GetStub().DelState(assetKey)
		if err != nil {
			return "", fmt.Errorf("failed to delete asset from public data collection for seller: %v", err)
		}
	} else {
		updatedAssetBytes, err := json.Marshal(asset)
		if err != nil {
			return "", fmt.Errorf("failed to marshal asset: %v", err)
		}
		err = ctx.GetStub().PutState(assetKey, updatedAssetBytes)
		if err != nil {
			return "", fmt.Errorf("failed to put Asset private details: %v", err)
		}
	}

	// Put the updated asset back in private data collection of the client's org
	err = ctx.GetStub().PutPrivateData(collection, name, privateUpdatedAssetBytes)
	if err != nil {
		return "", fmt.Errorf("failed to put Asset private details: %v", err)
	}

	return fmt.Sprintf("Successfully bought item from Market:-\nPrivate State --> Name: %s  NumItems: %d  Price: %d  OwnerOrg: %s\nMarket State --> Name: %s  NumItems: %d  Price: %d  OwnerOrg: %s", privateAsset.Name, privateAsset.NumItems, privateAsset.Price, privateAsset.OwnerOrg, asset.Name, asset.NumItems, asset.Price, asset.OwnerOrg), nil

}

// getClientOrgID gets the client org ID.
func getClientOrgID(ctx contractapi.TransactionContextInterface) (string, error) {
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return "", fmt.Errorf("failed getting client's orgID: %v", err)
	}

	return clientOrgID, nil
}

// verifyClientOrgMatchesPeerOrg checks that the client is from the same org as the peer
func verifyClientOrgMatchesPeerOrg(clientOrgID string) error {
	peerOrgID, err := shim.GetMSPID()
	if err != nil {
		return fmt.Errorf("failed getting peer's orgID: %v", err)
	}

	if clientOrgID != peerOrgID {
		return fmt.Errorf("client from org %s is not authorized to read or write private data from an org %s peer",
			clientOrgID,
			peerOrgID,
		)
	}

	return nil
}

// buildCollectionName returns the implicit collection name for an org
func buildCollectionName(clientOrgID string) string {
	return fmt.Sprintf("_implicit_org_%s", clientOrgID)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error create transfer asset chaincode: %v", err)
	}

	if err := chaincode.Start(); err != nil {
		log.Panicf("Error starting asset chaincode: %v", err)
	}
}
