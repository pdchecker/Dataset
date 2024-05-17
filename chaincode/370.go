/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a transcript
type SmartContract struct {
	contractapi.Contract
}

// Transcript describes basic details of what ids up a transcript
type Transcript struct {
	Id   string `json:"id"`
	Email  string `json:"email"`
	Pass string `json:"pass"`
	Grade  string `json:"grade"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Transcript
}

// InitLedger adds a base set of transcripts to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	transcripts := []Transcript{
		Transcript{Id: "19101362", Email: "ikram@g.bracu.ac.bd", Pass: "studentpw", Grade: "3.21"},
		Transcript{Id: "19101011", Email: "taher@g.bracu.ac.bd", Pass: "studentpw", Grade: "3.99"},
		Transcript{Id: "19101578", Email: "shadab@g.bracu.ac.bd", Pass: "studnetpw", Grade: "3.93"},
		Transcript{Id: "19101117", Email: "srijan@g.bracu.ac.bd", Pass: "studnetpw", Grade: "3.92"},
		Transcript{Id: "19101118", Email: "dihan@g.bracu.ac.bd", Pass: "studnetpw", Grade: "3.54"},
	}

	for i, transcript := range transcripts {
		transcriptAsBytes, _ := json.Marshal(transcript)
		err := ctx.GetStub().PutState(transcript.Id, transcriptAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %d %s", i, err.Error())
		}
	}

	return nil
}

// CreateTranscript adds a new transcript to the world state with given details
func (s *SmartContract) CreateTranscript(ctx contractapi.TransactionContextInterface, id string, email string, pass string) error {
	transcript := Transcript{
		Id:   id,
		Email:  email,
		Pass: pass,
	}

	transcriptAsBytes, _ := json.Marshal(transcript)

	return ctx.GetStub().PutState(id, transcriptAsBytes)
}

// QueryTranscript returns the transcript stored in the world state with given id
func (s *SmartContract) QueryTranscript(ctx contractapi.TransactionContextInterface, id string) (*Transcript, error) {
	transcriptAsBytes, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if transcriptAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", id)
	}

	transcript := new(Transcript)
	_ = json.Unmarshal(transcriptAsBytes, transcript)

	return transcript, nil
}

// QueryAllTranscripts returns all transcripts found in world state
func (s *SmartContract) QueryAllTranscripts(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		transcript := new(Transcript)
		_ = json.Unmarshal(queryResponse.Value, transcript)

		queryResult := QueryResult{Key: queryResponse.Key, Record: transcript}
		results = append(results, queryResult)
	}

	return results, nil
}

// ChangeTranscriptGrade updates the grade field of transcript with given id in world state
func (s *SmartContract) ChangeTranscriptGrade(ctx contractapi.TransactionContextInterface, id string, newGrade string) error {
	transcript, err := s.QueryTranscript(ctx, id)

	if err != nil {
		return err
	}

	transcript.Grade = newGrade

	transcriptAsBytes, _ := json.Marshal(transcript)

	return ctx.GetStub().PutState(id, transcriptAsBytes)
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabtranscript chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabtranscript chaincode: %s", err.Error())
	}
}
