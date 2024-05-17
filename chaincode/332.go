package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Contract struct {
	ContractID              string    `json:"contract_id"`
	ContractToken           string    `json:"contract_token"`
	ContractType            string    `json:"contract_type"`
	ContractFileHash        string    `json:"contract_file_hash"`
	PlayerID                string    `json:"player_id"`
	TeamID                  string    `json:"team_id"`
	SponsorID               string    `json:"sponsor_id"`
	SponsorName             string    `json:"sponsor_name"`
	VendorID                string    `json:"vendor_id"`
	VendorName              string    `json:"vendor_name"`
	SeasonID                string    `json:"season_id"`
	ContractStartDate       string    `json:"contract_start_date"`
	ContractEndDate         string 	  `json:"contract_end_date"`
	ContractWith            string    `json:"contract_with"`
	ContractWithEmailID     string    `json:"contract_with_email_id"`
	ContractWithContactNum  string    `json:"contract_with_contact_number"`
	ContractFromEmailID     string    `json:"contract_from_email_id"`
	ActionBy                string    `json:"action_by"`
	ContractFromContactNum  string    `json:"contract_from_contact_number"`
	UploadedBy              string    `json:"uploaded_by"`
	ContractStatus          string    `json:"contract_status"`
	ContractComment         string    `json:"contract_comment"`
	CreatedAt               string    `json:"created_at"`
	UpdatedAt              	string    `json:"updated_at"`
	DeletedAt               string    `json:"deleted_at"`
	IsActive                bool      `json:"is_active"`
	IsContractFabricated    bool      `json:"is_contract_fabricated"`
}

type ContractChaincode struct {
	contractapi.ContractChaincode
}

func (cc *ContractChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	// Initialize any initial data if required
	fmt.Println("Contract Deploy Successfully")
	return nil
}

func (cc *ContractChaincode) addContract(ctx contractapi.TransactionContextInterface, contractID string, contractDetailsJSON string) error {

	existing, err := ctx.GetStub().GetState(contractID) //check---
	if err != nil {
		return fmt.Errorf("failed to read from world state: %w", err)
	}
	if existing != nil {
		return fmt.Errorf("the contract ID already exists")
		// Parse the contract details JSON
	contractDetails := &Contract{}
	err = json.Unmarshal([]byte(contractDetailsJSON), contractDetails)
	if err != nil {
		return fmt.Errorf("failed to unmarshal contract details: %w", err)
	}

	}
	contractJSON, err := json.Marshal(contract)
	if err != nil {
		return fmt.Errorf("failed to marshal contract JSON: %w", err)
	}
	fmt.Printf("Contract added Successfully with %s ",contract.ContractID);
	err = ctx.GetStub().PutState(contract.ContractID, contractJSON)
	if err != nil {
		return fmt.Errorf("failed to put contract details in world state: %w", err)
	}

	return nil
}

func (cc *ContractChaincode) approveContract(ctx contractapi.TransactionContextInterface, contractID string, contractStatus string, actionBy string, UpdatedAt string, comment string,is_Active bool, is_contract_fabricated bool) error {
	
	contractDetailsBytes, err := ctx.GetStub().GetState(contractID)
	if err != nil {
		return fmt.Errorf("failed to read from world state: %w", err)
	}
	if contractDetailsBytes == nil {
		return fmt.Errorf("the contract ID does not exist")
	}

	contract, err := cc.getContractByContractID(ctx, contractID)
	if err != nil {
		return err
	}

	// Update the contract details
	contract.ContractStatus = contractStatus
	contract.ActionBy = actionBy
	contract.UpdatedAt = UpdatedAt
	contract.ContractComment = comment
	contract.IsActive = is_Active
	contract.IsContractFabricated = is_contract_fabricated


	contractJSON, err := json.Marshal(contract)
	if err != nil {
		return fmt.Errorf("failed to marshal contract JSON: %w", err)
	}
	fmt.Printf("Contract status updated Successfully in this %s contractId",contractID);
	return ctx.GetStub().PutState(contract.ContractID, contractJSON)
}

func (cc *ContractChaincode) getAllContracts(ctx contractapi.TransactionContextInterface, pageLimit int, pageNumber string) ([]*Contract, string, error) {
	queryString := `{
		"selector": {
			"contract_id": {
				"$regex": ""
			}
		},
		"use_index": ["_design/indexContract", "contract_id"],
		"bookmark": "%s",
		"limit": %d
	}`

	queryString = fmt.Sprintf(queryString, pageNumber, pageLimit)

	resultsIterator, responseMetadata, err := ctx.GetStub().GetQueryResultWithPagination(queryString)
	if err != nil {
		return nil, "", err
	}
	defer resultsIterator.Close()

	var contracts []*Contract
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, "", err
		}

		var contract Contract
		err = json.Unmarshal(queryResponse.Value, &contract)
		if err != nil {
			return nil, "", fmt.Errorf("failed to unmarshal contract JSON: %w", err)
		}

		contracts = append(contracts, &contract)
	}

	return contracts, responseMetadata.Bookmark, nil
}

func (cc *ContractChaincode) getContractByQuery(ctx contractapi.TransactionContextInterface, queryString string) ([]*Contract, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var contracts []*Contract
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var contract Contract
		err = json.Unmarshal(queryResponse.Value, &contract)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal contract JSON: %w", err)
		}
		fmt.Printf("Contract details fetched Successfully by this %s queryString", queryString);
		contracts = append(contracts, &contract)
	}

	return contracts, nil
}
func (cc *ContractChaincode) deleteByContractID(ctx contractapi.TransactionContextInterface, contractID string) error {

	fmt.Printf("Contract Deleted Successfully by this %s contractId",contractID);
	return ctx.GetStub().DelState(contractID)
}

func (cc *ContractChaincode) getContractByContractID(ctx contractapi.TransactionContextInterface, contractID string) (*Contract, error) {
	contractJSON, err := ctx.GetStub().GetState(contractID)
	if err != nil {
		return nil, fmt.Errorf("failed to read contract from world state: %w", err)
	}
	if contractJSON == nil {
		return nil, fmt.Errorf("contract does not exist")
	}

	var contract Contract
	err = json.Unmarshal(contractJSON, &contract)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal contract JSON: %w", err)
	}
	fmt.Printf("Contract fetched Successfully by this %s contractId",contractID);
	return &contract, nil
}

func (cc *ContractChaincode) getContractBySeasonID(ctx contractapi.TransactionContextInterface, seasonID string) ([]*Contract, error) {
	queryString := fmt.Sprintf(`{
		"selector": {
			"season_id": "%s"
		}
	}`, seasonID)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var contracts []*Contract
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var contract Contract
		err = json.Unmarshal(queryResponse.Value, &contract)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal contract JSON: %w", err)
		}

		contracts = append(contracts, &contract)
	}
	fmt.Printf("Contract fetched Successfully by this %s seasonId",seasonID);
	return contracts, nil
}

func main() {
	cc, err := contractapi.NewChaincode(&ContractChaincode{})
	if err != nil {
		fmt.Printf("Error creating contract chaincode: %s", err.Error())
		return
	}

	if err := cc.Start(); err != nil {
		fmt.Printf("Error starting contract chaincode: %s", err.Error())
	}
}