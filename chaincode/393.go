package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Donation
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// rder when marshal golang keeps the oto json but doesn't order automatically
type Donation struct {
	AppraisedValue int    `json:"AppraisedValue"`
	DonationType   string `json:"DonationType"`
	ID             string `json:"ID"`
	Donor          string `json:"Donor"`
	Size           int    `json:"Size"`
	Timestamp      int    `jason:"Timestamp"`
	Status         string `jason:"Status"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	donations := []Donation{
		{ID: "donation1", DonationType: "money", Size: 0, Donor: "Tomoko", AppraisedValue: 300, Timestamp: 1688404431, Status: "time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)"},
		{ID: "donation2", DonationType: "money", Size: 0, Donor: "Brad", AppraisedValue: 400, Timestamp: 1688404431, Status: "time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)"},
		{ID: "donation3", DonationType: "money", Size: 0, Donor: "Jin Soo", AppraisedValue: 500, Timestamp: 1688404431, Status: "time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)"},
		{ID: "donation4", DonationType: "money", Size: 0, Donor: "Max", AppraisedValue: 600, Timestamp: 1688404431, Status: "time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)"},
		{ID: "donation5", DonationType: "ssd", Size: 1, Donor: "Adriana", AppraisedValue: 700, Timestamp: 1688404431, Status: "time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)"},
		{ID: "donation6", DonationType: "keyboard", Size: 5, Donor: "Michel", AppraisedValue: 800, Timestamp: 1688404431, Status: "time.Date(2000, 2, 1, 12, 30, 0, 0, time.UTC)"},
	}

	for _, donation := range donations {
		donationJSON, err := json.Marshal(donation)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(donation.ID, donationJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateDonation(ctx contractapi.TransactionContextInterface, id string, donationType string, size int, donor string, appraisedValue int, timestamp int, status string) error {
	exists, err := s.DonationExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the donation %s already exists", id)
	}

	donation := Donation{
		ID:             id,
		DonationType:   donationType,
		Size:           size,
		Donor:          donor,
		AppraisedValue: appraisedValue,
		Timestamp:      timestamp,
		Status:         status,
	}
	donationJSON, err := json.Marshal(donation)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, donationJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadDonation(ctx contractapi.TransactionContextInterface, id string) (*Donation, error) {
	donationJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if donationJSON == nil {
		return nil, fmt.Errorf("the donation %s does not exist", id)
	}

	var donation Donation
	err = json.Unmarshal(donationJSON, &donation)
	if err != nil {
		return nil, err
	}

	return &donation, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateDonation(ctx contractapi.TransactionContextInterface, id string, donationType string, size int, donor string, appraisedValue int, timestamp int, status string) error {
	exists, err := s.DonationExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the donation %s does not exist", id)
	}

	// overwriting original asset with new asset
	donation := Donation{
		ID:             id,
		DonationType:   donationType,
		Size:           size,
		Donor:          donor,
		AppraisedValue: appraisedValue,
		Timestamp:      timestamp,
		Status:         status,
	}
	donationJSON, err := json.Marshal(donation)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, donationJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteDonation(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.DonationExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the donation %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) DonationExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	donationJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return donationJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state, and returns the old owner.
func (s *SmartContract) TransferDonation(ctx contractapi.TransactionContextInterface, id string, newDonor string) (string, error) {
	donation, err := s.ReadDonation(ctx, id)
	if err != nil {
		return "", err
	}

	oldDonor := donation.Donor
	donation.Donor = newDonor

	donationJSON, err := json.Marshal(donation)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, donationJSON)
	if err != nil {
		return "", err
	}

	return oldDonor, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllDonations(ctx contractapi.TransactionContextInterface) ([]*Donation, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var donations []*Donation
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var donation Donation
		err = json.Unmarshal(queryResponse.Value, &donation)
		if err != nil {
			return nil, err
		}
		donations = append(donations, &donation)
	}

	return donations, nil
}
