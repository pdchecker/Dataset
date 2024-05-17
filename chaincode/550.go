package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Voter represents a voter in the system
type Voter struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	HasVoted bool   `json:"hasVoted"`
}

// Candidate represents a candidate in the election
type Candidate struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Votes int    `json:"votes"`
}

// VotingContract for handling voting logic
type VotingContract struct {
	contractapi.Contract
}

// InitLedger adds a base set of candidates to the ledger
func (vc *VotingContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	candidates := []Candidate{
		{ID: "C1", Name: "Candidate 1", Votes: 0},
		{ID: "C2", Name: "Candidate 2", Votes: 0},
	}

	for _, candidate := range candidates {
		candidateJSON, err := json.Marshal(candidate)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(candidate.ID, candidateJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

// RegisterVoter adds a new voter to the ledger
func (vc *VotingContract) RegisterVoter(ctx contractapi.TransactionContextInterface, voterID string, name string) error {
	exists, err := vc.VoterExists(ctx, voterID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the voter %s already exists", voterID)
	}

	voter := Voter{
		ID:       voterID,
		Name:     name,
		HasVoted: false,
	}

	voterJSON, err := json.Marshal(voter)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(voterID, voterJSON)
}

// CastVote records a vote for a candidate
func (vc *VotingContract) CastVote(ctx contractapi.TransactionContextInterface, voterID string, candidateID string) error {
	voter, err := vc.GetVoter(ctx, voterID)
	if err != nil {
		return err
	}

	if voter.HasVoted {
		return fmt.Errorf("voter %s has already voted", voterID)
	}

	candidateJSON, err := ctx.GetStub().GetState(candidateID)
	if err != nil {
		return fmt.Errorf("failed to get candidate %s", candidateID)
	}
	if candidateJSON == nil {
		return fmt.Errorf("candidate %s does not exist", candidateID)
	}

	var candidate Candidate
	err = json.Unmarshal(candidateJSON, &candidate)
	if err != nil {
		return err
	}

	candidate.Votes++

	candidateJSON, err = json.Marshal(candidate)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(candidateID, candidateJSON)
	if err != nil {
		return err
	}

	voter.HasVoted = true
	voterJSON, err := json.Marshal(voter)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(voterID, voterJSON)
}

// QueryAllVoters returns all voters found in the world state
func (vc *VotingContract) QueryAllVoters(ctx contractapi.TransactionContextInterface) ([]*Voter, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var voters []*Voter
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var voter Voter
		err = json.Unmarshal(queryResponse.Value, &voter)
		if err != nil {
			return nil, err
		}

		voters = append(voters, &voter)
	}

	return voters, nil
}

// VoterExists checks if a voter exists in the ledger
func (vc *VotingContract) VoterExists(ctx contractapi.TransactionContextInterface, voterID string) (bool, error) {
	voterJSON, err := ctx.GetStub().GetState(voterID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return voterJSON != nil, nil
}

// GetVoter returns the voter information from the ledger
func (vc *VotingContract) GetVoter(ctx contractapi.TransactionContextInterface, voterID string) (*Voter, error) {
	voterJSON, err := ctx.GetStub().GetState(voterID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if voterJSON == nil {
		return nil, fmt.Errorf("the voter %s does not exist", voterID)
	}

	var voter Voter
	err = json.Unmarshal(voterJSON, &voter)
	if err != nil {
		return nil, err
	}

	return &voter, nil
}

// GetResults returns the results of the election
func (vc *VotingContract) GetResults(ctx contractapi.TransactionContextInterface) ([]*Candidate, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var candidates []*Candidate
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var candidate Candidate
		err = json.Unmarshal(queryResponse.Value, &candidate)
		if err != nil {
			return nil, err
		}

		candidates = append(candidates, &candidate)
	}

	return candidates, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&VotingContract{})
	if err != nil {
		fmt.Printf("Error creating voting chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting voting chaincode: %s", err.Error())
	}
}
