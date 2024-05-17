package main

import (
	"encoding/json"
	"fmt"
    "sort"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)




// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
//type Asset struct {
//	Requests int    `json:"Requests"`
//	Endpoint          string `json:"Endpoint"`
//	ID             string `json:"ID"`
//	Owner          string `json:"Owner"`
//	Description           int    `json:"Description"`
//}
type Asset struct {
	Requests    map[string]string `json:"Requests"`
	Endpoint    string `json:"Endpoint"`
	ID          string `json:"ID"`
	Owner       string `json:"Owner"`
	Description string  `json:"Description"`
}



// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}


// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
    clientIdentity := ctx.GetClientIdentity()

    creator, err := clientIdentity.GetID()
    if err != nil {
        return err
    }

	assets := []Asset{
		{ID: "asset1", Endpoint: "192.168.1.1:8080", Description: "Data 1", Owner: string(creator), Requests: map[string]string{"Manuel": "Accepted"}},
		{ID: "asset2", Endpoint: "192.168.1.2:8080", Description: "Data 2", Owner: string(creator), Requests: map[string]string{"Manuel": "Accepted"}},
		{ID: "asset3", Endpoint: "192.168.1.3:8080", Description: "Data 3", Owner: string(creator), Requests: map[string]string{"Manuel": "Accepted"}},
		{ID: "asset4", Endpoint: "192.168.1.4:8080", Description: "Data 4", Owner: string(creator), Requests: map[string]string{"Manuel": "Accepted"}},
		{ID: "asset5", Endpoint: "192.168.1.5:8080", Description: "Data 5", Owner: string(creator), Requests: map[string]string{"Manuel": "Accepted"}},
		{ID: "asset6", Endpoint: "192.168.1.6:8080", Description: "Data 6", Owner: string(creator), Requests: map[string]string{"Manuel": "Accepted"}},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}


// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}


// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, endpoint string, description string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

    clientIdentity := ctx.GetClientIdentity()
    creator, err := clientIdentity.GetID()
    if err != nil {
        return err
    }


	asset := Asset{
		ID:             id,
		Endpoint:       endpoint,
		Description:    description,
		Owner:          string(creator),
		Requests:       map[string]string{},
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}




// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) RequestAsset(ctx contractapi.TransactionContextInterface, id string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("the asset %s does not exist", id)
	}
    clientIdentity := ctx.GetClientIdentity()
    creator, err := clientIdentity.GetID()
    if err != nil {
        return err
    }
	asset.Requests[string(creator)]  = "Pending"


	aux_request := map[string]string{}


	// Obtener las claves y ordenarlas
	keys := make([]string, 0, len(asset.Requests))
	for key := range asset.Requests {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Acceder a los valores en el orden de las claves
	for _, key := range keys {
		aux_request[key] = asset.Requests[key]
	}

	asset.Requests = aux_request

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}
    return ctx.GetStub().PutState(id, assetJSON)
}



// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) AcceptRequestAsset(ctx contractapi.TransactionContextInterface, id string, requester string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("the asset %s does not exist", id)
	}
    clientIdentity := ctx.GetClientIdentity()
    creator, err := clientIdentity.GetID()
    if err != nil {
        return err
    }


    if string(creator) != asset.Owner {
        return fmt.Errorf("You are not the owner of %s", id)
    }

    _, found := asset.Requests[requester]
    if !found {
        return fmt.Errorf("%s not request acces to %s", requester,id)
    }

	asset.Requests[requester]  = "Accepted"
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}
    return ctx.GetStub().PutState(id, assetJSON)
}




// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) DenyRequestAsset(ctx contractapi.TransactionContextInterface, id string, requester string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("the asset %s does not exist", id)
	}
    clientIdentity := ctx.GetClientIdentity()
    creator, err := clientIdentity.GetID()
    if err != nil {
        return err
    }


    if string(creator) != asset.Owner {
        return fmt.Errorf("You are not the owner of %s", id)
    }

    _, found := asset.Requests[requester]
    if !found {
        return fmt.Errorf("%s not request acces to %s", requester,id)
    }

	asset.Requests[requester]  = "Denied"
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}
    return ctx.GetStub().PutState(id, assetJSON)
}


// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id string, endpoint string, description string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("the asset %s does not exist", id)
	}
    clientIdentity := ctx.GetClientIdentity()
    creator, err := clientIdentity.GetID()
    if err != nil {
        return err
    }

    if string(creator) != asset.Owner {
        return fmt.Errorf("You are not the owner of %s", id)
    }

	// overwriting original asset with new asset
	asset.Endpoint = endpoint
	asset.Description = description
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	asset, err := s.ReadAsset(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if asset == nil {
		return fmt.Errorf("the asset %s does not exist", id)
	}
    clientIdentity := ctx.GetClientIdentity()
    creator, err := clientIdentity.GetID()

    if string(creator) != asset.Owner {
        return fmt.Errorf("You are not the owner of %s", id)
    }

    if err != nil {
        return err
    }

	return ctx.GetStub().DelState(id)
}


