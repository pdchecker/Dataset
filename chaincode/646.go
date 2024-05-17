package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Collector represents the structure for a collector
type Collector struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	NIB      string   `json:"nib"`
	NIK      string   `json:"nik"`
	NoHP     string   `json:"noHP"`
	Email    string   `json:"email"`
	Address  string   `json:"address"`
	Capacity float64  `json:"capacity"`
	Partner  []string `json:"partner"`
}

// AddCollector adds a new collector to the ledger
func (pc *PalmOilContract) AddCollector(ctx contractapi.TransactionContextInterface, id string, name string, nib string, nik string, noHP string, email string, address string, capacity float64, partnersInput string) error {
	// Check if a collector with the given ID already exists
	existingCollectorJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if existingCollectorJSON != nil {
		return fmt.Errorf("a collector with ID %s already exists", id)
	}

	// Parse the partnersInput into a []string
	var partners []string
	err = json.Unmarshal([]byte(partnersInput), &partners)
	if err != nil {
		return fmt.Errorf("failed to parse partner attribute: %v", err)
	}

	collector := Collector{
		ID:       id,
		Name:     name,
		NIB:      nib,
		NIK:      nik,
		NoHP:     noHP,
		Email:    email,
		Address:  address,
		Capacity: capacity,
		Partner:  partners,
	}

	collectorJSON, err := json.Marshal(collector)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, collectorJSON)
}

// UpdateCollector updates an existing collector on the ledger
func (pc *PalmOilContract) UpdateCollector(ctx contractapi.TransactionContextInterface, id string, name string, nib string, nik string, noHP string, email string, address string, capacity float64, partnersInput string) error {
	collectorJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %v", err)
	}
	if collectorJSON == nil {
		return fmt.Errorf("the collector with ID %s does not exist", id)
	}

	// Parse the partnersInput into a []string
	var partners []string
	err = json.Unmarshal([]byte(partnersInput), &partners)
	if err != nil {
		return fmt.Errorf("failed to parse partner attribute: %v", err)
	}

	var collector Collector
	json.Unmarshal(collectorJSON, &collector)

	// Update the collector's attributes
	collector.Name = name
	collector.NIB = nib
	collector.NIK = nik
	collector.NoHP = noHP
	collector.Email = email
	collector.Address = address
	collector.Capacity = capacity
	collector.Partner = partners

	collectorJSON, err = json.Marshal(collector)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, collectorJSON)
}


// QueryCollectorByID retrieves a collector by its ID from the ledger
func (pc *PalmOilContract) QueryCollectorByID(ctx contractapi.TransactionContextInterface, id string) (*Collector, error) {
	collectorJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if collectorJSON == nil {
		return nil, fmt.Errorf("the collector with ID %s does not exist", id)
	}

	var collector Collector
	json.Unmarshal(collectorJSON, &collector)

	return &collector, nil
}

// QueryAllCollectors retrieves all collectors from the ledger
func (pc *PalmOilContract) QueryAllCollectors(ctx contractapi.TransactionContextInterface) ([]*Collector, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("COL_", "COL_zzzzzzzzzz")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var collectors []*Collector
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var collector Collector
		json.Unmarshal(queryResponse.Value, &collector)
		collectors = append(collectors, &collector)
	}

	return collectors, nil
}
