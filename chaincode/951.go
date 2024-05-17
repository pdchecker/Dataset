package main

import (
	"fmt"
	"encoding/json"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// HealthInsuranceContract represents the smart contract for a health insurance policy
type HealthInsuranceContract struct {
	contractapi.Contract
}

// Policy represents the structure of a health insurance policy
type Policy struct {
	ObjectType   string `json:"docType"`      // Type of the object stored in the ledger
	PolicyID     string `json:"policyID"`     // Unique identifier for the policy
	SumAssured   int    `json:"sumAssured"`   // Total sum assured for the policy
	PersonName   string `json:"personName"`   // Name of the insured person
	DateOfBirth  string `json:"dateOfBirth"`  // Date of birth of the insured person
	Gender       string `json:"gender"`       // Gender of the insured person
	StartDate    string `json:"startDate"`    // Start date of the policy
	EndDate      string `json:"endDate"`      // End date of the policy
	CoPay        int    `json:"coPay"`        // Co-pay percentage for the policy
	Coverages    string `json:"coverages"`    // Coverage details of the policy
	Benefits     string `json:"benefits"`     // Benefits provided by the policy
	Exclusions   string `json:"exclusions"`   // Exclusions or limitations of the policy
	ClaimedTotal int    `json:"claimedTotal"` // Total amount claimed so far
}

// CreatePolicy creates a new health insurance policy
func (c *HealthInsuranceContract) CreatePolicy(ctx contractapi.TransactionContextInterface, policyID string, sumAssured int, personName string, dateOfBirth string, gender string, startDate string, endDate string, coPay int, coverages string, benefits string, exclusions string) error {
	policy := Policy{
		ObjectType:   "policy",
		PolicyID:     policyID,
		SumAssured:   sumAssured,
		PersonName:   personName,
		DateOfBirth:  dateOfBirth,
		Gender:       gender,
		StartDate:    startDate,
		EndDate:      endDate,
		CoPay:        coPay,
		Coverages:    coverages,
		Benefits:     benefits,
		Exclusions:   exclusions,
		ClaimedTotal: 0,
	}

	policyJSON, err := json.Marshal(policy) // Convert the policy struct to JSON format
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(policyID, policyJSON) // Store the policy in the ledger using the policy ID as the key
	if err != nil {
		return err
	}

	return nil
}

// GetPolicy retrieves the details of a health insurance policy based on the policy ID
func (c *HealthInsuranceContract) GetPolicy(ctx contractapi.TransactionContextInterface, policyID string) (*Policy, error) {
	policyJSON, err := ctx.GetStub().GetState(policyID) // Retrieve the policy from the ledger using the policy ID
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if policyJSON == nil {
		return nil, fmt.Errorf("policy does not exist")
	}

	var policy Policy
	err = json.Unmarshal(policyJSON, &policy) // Convert the JSON data to a policy struct
	if err != nil {
		return nil, err
	}

	return &policy, nil
}

// SubmitClaim allows submitting a claim for a health insurance policy
func (c *HealthInsuranceContract) SubmitClaim(ctx contractapi.TransactionContextInterface, policyID string, claimAmount int) error {
	policy, err := c.GetPolicy(ctx, policyID) // Retrieve the policy details
	if err != nil {
		return err
	}

	if policy.ClaimedTotal+claimAmount > policy.SumAssured { // Check if the claim amount exceeds the sum assured
		return fmt.Errorf("claim amount exceeds sum assured")
	}

	policy.ClaimedTotal += claimAmount // Update the claimed total

	policyJSON, err := json.Marshal(policy) // Convert the updated policy struct to JSON format
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(policyID, policyJSON) // Store the updated policy in the ledger
	if err != nil {
		return err
	}

	return nil
}

// UpdatePolicy updates the details of a health insurance policy
func (c *HealthInsuranceContract) UpdatePolicy(ctx contractapi.TransactionContextInterface, policyID string, sumAssured int, personName string, dateOfBirth string, gender string, startDate string, endDate string, coPay int, coverages string, benefits string, exclusions string) error {
	policy, err := c.GetPolicy(ctx, policyID) // Retrieve the policy details
	if err != nil {
		return err
	}

	// Update the policy details with the new values
	policy.SumAssured = sumAssured
	policy.PersonName = personName
	policy.DateOfBirth = dateOfBirth
	policy.Gender = gender
	policy.StartDate = startDate
	policy.EndDate = endDate
	policy.CoPay = coPay
	policy.Coverages = coverages
	policy.Benefits = benefits
	policy.Exclusions = exclusions

	policyJSON, err := json.Marshal(policy) // Convert the updated policy struct to JSON format
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(policyID, policyJSON) // Store the updated policy in the ledger
	if err != nil {
		return err
	}

	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&HealthInsuranceContract{}) // Create a new instance of the chaincode
	if err != nil {
		fmt.Printf("Error creating health insurance chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil { // Start the chaincode server
		fmt.Printf("Error starting health insurance chaincode: %v\n", err)
	}
}
