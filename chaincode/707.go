package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Campaign struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Advertiser string `json:"advertiser"`
	Business   string `json:"business"`
	Custom     string `json:"custom"`
	Timestamp  string `json:"timestamp`
}

func (s *SmartContract) Test(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// Create a new campaign
func (s *SmartContract) CreateCampaign(ctx contractapi.TransactionContextInterface, id string, name string, advertiser string, business string, timestamp string) error {
	existing, err := ctx.GetStub().GetState(id)

	if err != nil {
		return errors.New("Unable to read the world state")
	}

	if existing != nil {
		return fmt.Errorf("Cannot create asset since its id %s is existed", id)
	}

	cryptoAPIAddress := "http://crypto.promark.com:3000"
	apiEndpoint := fmt.Sprintf("%s/helloworld", cryptoAPIAddress)

	response, err := http.Get(apiEndpoint)
	if err != nil {
		return fmt.Errorf("Cannot access to %s: %s", apiEndpoint, err)
	}

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("Cannot parse response data %s", response.Body)
	}
	message := string(responseData)

	campaign := Campaign{
		ID:         id,
		Name:       name,
		Advertiser: advertiser,
		Business:   business,
		Custom:     message,
		Timestamp: timestamp,
	}

	campaignJSON, err := json.Marshal(campaign)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, campaignJSON)

	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) ReadAllCampaigns(ctx contractapi.TransactionContextInterface) ([]*Campaign, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}

	var campaigns []*Campaign

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var campaign Campaign
		err = json.Unmarshal(queryResponse.Value, &campaign)
		if err != nil {
			return nil, err
		}

		campaigns = append(campaigns, &campaign)
	}

	resultsIterator.Close()

	return campaigns, nil
}

func (s *SmartContract) ReadCampaignsByTimestamp(ctx contractapi.TransactionContextInterface, startTime string, endTime string) ([]*Campaign, error) {
	queryString := fmt.Sprintf(`{"selector":{"timestamp":{"$gte": "%s","$lte": "%s"}}}`, startTime, endTime)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var campaigns []*Campaign

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var campaign Campaign
		err = json.Unmarshal(queryResponse.Value, &campaign)
		if err != nil {
			return nil, err
		}

		campaigns = append(campaigns, &campaign)
	}

	resultsIterator.Close()

	return campaigns, nil
}



func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating campaign chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting campaign chaincode: %v", err)
	}
}
