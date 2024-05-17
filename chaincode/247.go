package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartBank struct {
	contractapi.Contract
}

type Website struct {
	ID   string `json:"ID"`
	Link string `json:"link"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartBank) InitLedger(ctx contractapi.TransactionContextInterface) error {
	websites := []Website{
		{ID: "WhatsappWeb", Link: "https://web.whatsapp.com/"},
		{ID: "Office365", Link: "https://www.office.com/"},
		{ID: "BOCHK", Link: "https://www.bochk.com/tc/home.html"},
		{ID: "HKGOV", Link: "https://www.gov.hk/tc/residents/"},
	}

	for _, website := range websites {
		websiteJSON, err := json.Marshal(website)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(website.ID, websiteJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

func (s *SmartBank) CreateWebsite(ctx contractapi.TransactionContextInterface, id string, link string) error {
	exists, err := s.WebsiteExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the website %s is already exists", id)
	}

	website := Website{
		ID:   id,
		Link: link,
	}
	websiteJSON, err := json.Marshal(website)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, websiteJSON)
}

func (s *SmartBank) ReadWebsite(ctx contractapi.TransactionContextInterface, id string) (*Website, error) {
	websiteJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if websiteJSON == nil {
		return nil, fmt.Errorf("the website %s have not been checked", id)
	}

	var website Website
	err = json.Unmarshal(websiteJSON, &website)
	if err != nil {
		return nil, err
	}

	return &website, nil
}

func (s *SmartBank) UpdateWebsite(ctx contractapi.TransactionContextInterface, id string, link string) error {
	exists, err := s.WebsiteExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the website %s does not exist", id)
	}

	website := Website{
		ID:   id,
		Link: link,
	}
	websiteJSON, err := json.Marshal(website)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, websiteJSON)
}

func (s *SmartBank) DeleteWebsite(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.WebsiteExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the website %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartBank) WebsiteExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	websiteJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return websiteJSON != nil, nil
}

func (s *SmartBank) GetAllWebsite(ctx contractapi.TransactionContextInterface) ([]*Website, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var websites []*Website
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var website Website
		err = json.Unmarshal(queryResponse.Value, &website)
		if err != nil {
			return nil, err
		}
		websites = append(websites, &website)
	}

	return websites, nil
}

func main() {
	websiteChaincode, err := contractapi.NewChaincode(&SmartBank{})
	if err != nil {
		log.Panicf("Error creating website bank chaincode: %v", err)
	}

	if err := websiteChaincode.Start(); err != nil {
		log.Panicf("Error starting website bank chaincode: %v", err)
	}
}
