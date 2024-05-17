package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Campaign struct {
	Id        		string 			`json:"id"`
	Name      		string 			`json:"name"`
	StartTime 		string 			`json:"startTime"`
	EndTime			string 			`json:"endTime"`
}

func (s *CampaignSmartContract) CreateCampaign(ctx contractapi.TransactionContextInterface, id string, name string, startTime string, endTime string) error {
	exists, err := s.CampaignExists(ctx, id)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("Campaign %s already exists", id)
    }

	campaign := Campaign{
		Id:       			id,
		Name:      			name,
		StartTime: 			startTime,
		EndTime:   			endTime,
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

func (s *CampaignSmartContract) DeleteCampaign(ctx contractapi.TransactionContextInterface, id string) error {
    exists, err := s.CampaignExists(ctx, id)
    if err != nil {
        return err
    }
    if !exists {
        return fmt.Errorf("Error while deleting campaign: the campaign %s does not exist", id)
    }

    return ctx.GetStub().DelState(id)
}

func (s *CampaignSmartContract) CampaignExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	campaignBytes, err := ctx.GetStub().GetState(id)
    if err != nil {
        return false, fmt.Errorf("Failed to read campaign %s from world state. %v", id, err)
    }
	if campaignBytes == nil {
		return false, nil
	}
	return true, nil
}

func (s *CampaignSmartContract) QueryCampaign(ctx contractapi.TransactionContextInterface, id string) Campaign {
	campaign, _ := s.getCampaign(ctx, id)

	return *campaign
}

func (s *CampaignSmartContract) getCampaign(ctx contractapi.TransactionContextInterface, id string) (*Campaign, error) {
	campaignBytes, err := ctx.GetStub().GetState(id)
    if err != nil {
        return nil, fmt.Errorf("Failed to read campaign %s from world state. %v", id, err)
    }
	if campaignBytes == nil {
		return nil, fmt.Errorf("Campaign %s does not exist", id)
	}

	campaign := new(Campaign)
	_ = json.Unmarshal(campaignBytes, campaign)

    return campaign, nil
}