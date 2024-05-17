package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Log struct {
	LogID       string `json:"LogID"`
	FileName    string `json:"FileName"`
	ContentHash string `json:"ContentHash"`
	Timestamp   string `json:"Timestamp"`
	Owner       string `json:"Owner"`
}

var authorizedUsers = []string{"Atlas", "Leonardo", "Spot"}

func (sc *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	contentHash := sha256.Sum256([]byte("content"))

	logFile := &Log{
		LogID:       "0",
		FileName:    "data.csv",
		ContentHash: fmt.Sprintf("%x", contentHash),
		Timestamp:   "2006-01-02 15:04:05",
		Owner:       "Test",
	}

	logFileJSON, err := json.Marshal(logFile)
	if err != nil {
		return fmt.Errorf("failed to marshal log file data: %w", err)
	}

	err = ctx.GetStub().PutState("1", logFileJSON)
	if err != nil {
		return fmt.Errorf("failed to put log file on the ledger: %w", err)
	}

	return nil
}

func (sc *SmartContract) CreateLog(ctx contractapi.TransactionContextInterface, logID, fileName, content, timestamp, owner string) error {
	
	err := isAuthorized(ctx)
	if err != nil {
			return err
	} 

	contentHash := sha256.Sum256([]byte(content))

	logFile := &Log{
		LogID:       logID,
		FileName:    fileName,
		ContentHash: fmt.Sprintf("%x", contentHash),
		Timestamp:   timestamp,
		Owner:       owner,
	}

	logFileJSON, err := json.Marshal(logFile)
	if err != nil {
		return fmt.Errorf("failed to marshal log file data: %w", err)
	}

	return ctx.GetStub().PutState(logID, logFileJSON)
}

func (sc *SmartContract) GetLog(ctx contractapi.TransactionContextInterface, logID string) (*Log, error) {

	err := isAuthorized(ctx)
	if err != nil {
			return nil, err
	} 

	logFileJSON, err := ctx.GetStub().GetState(logID)
	if err != nil {
		return nil, fmt.Errorf("Failed to read log file from the ledger: %w", err)
	}

	if logFileJSON == nil {
		return nil, fmt.Errorf("Log file with ID '%s' does not exist", logID)
	}

	logFile := &Log{}
	err = json.Unmarshal(logFileJSON, logFile)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal log file data: %w", err)
	}

	return logFile, nil
}

func isAuthorized(ctx contractapi.TransactionContextInterface) error {
	
	clientID, err := ctx.GetClientIdentity().GetID()
    if err != nil {
        return fmt.Errorf("failed to get client identity: %w", err)
	}

	for _, user := range authorizedUsers {
			if user == clientID {
					return nil
			}
	}

	return fmt.Errorf("unauthorized access: %s is not an authorized user", clientID)
}