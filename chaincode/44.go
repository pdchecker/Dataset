package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric/common/util"
)

// SmartContract defines the structure of a smart contract
type SmartContract struct {
	contractapi.Contract
}

type Malware struct {
	MalwareID         string `json:"malwareID"`
	InfectionDuration string `json:"duration"`
	DeviceID          string `json:"deviceID"`
	Checksum          string `json:"checksum"`
}

type Backup struct {
	BackupID     string   `json:"backupID"`
	DeviceID     string   `json:"deviceID"`
	Hash         string   `json:"hash"`
	PreviousHash string   `json:"previousHash"`
	Timestamp    int64    `json:"timestamp"`
	IsValid      bool     `json:"isValid"`
	Signature    string   `json:"signature"`
	Paths        []string `json:"paths"`
}

// InitLedger adds a base set of malware entries to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Init Ledger")
	return nil
}

// CreateMalware issues a new malware entry to the world state with given details.
func (s *SmartContract) CreateMalware(ctx contractapi.TransactionContextInterface, malwareID string, duration string, deviceID string, checksum string) error {
	exists, err := s.MalwareExists(ctx, malwareID)
	if err != nil {
		return err
	}

	if exists {
		return fmt.Errorf("the malware %s already exists", malwareID)
	}

	malware := Malware{
		MalwareID:         malwareID,
		InfectionDuration: duration,
		DeviceID:          deviceID,
		Checksum:          checksum,
	}

	malwareJSON, err := json.Marshal(malware)
	if err != nil {
		return err
	}

	estimatedDuration, err := strconv.ParseInt(duration, 10, 64)
	if err != nil {
		return err
	}

	endTimestamp := fmt.Sprintf("%d", time.Now().Unix())
	startTimestamp := fmt.Sprintf("%d", (time.Now().Unix() - estimatedDuration))

	chainCodeArgs := util.ToChaincodeArgs("QueryBackupsByTimestamps", deviceID, startTimestamp, endTimestamp)
	response := ctx.GetStub().InvokeChaincode("backup", chainCodeArgs, "mychannel")

	var backups []Backup
	json.Unmarshal([]byte(response.Payload), &backups)

	for i := 0; i < len(backups); i++ {
		chainCodeArgs := util.ToChaincodeArgs("InvalidateBackup", backups[i].BackupID)
		ctx.GetStub().InvokeChaincode("backup", chainCodeArgs, "mychannel")
		fmt.Printf("backup: %s is invalidated!\n", backups[i].BackupID)
	}

	return ctx.GetStub().PutState(malwareID, malwareJSON)
}

// ReadMalware returns the malware entry stored in the world state with given id.
func (s *SmartContract) ReadMalware(ctx contractapi.TransactionContextInterface, malwareID string) (*Malware, error) {
	malwareJSON, err := ctx.GetStub().GetState(malwareID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if malwareJSON == nil {
		return nil, fmt.Errorf("the malware %s does not exist", malwareID)
	}

	var malware Malware
	err = json.Unmarshal(malwareJSON, &malware)
	if err != nil {
		return nil, err
	}

	return &malware, nil
}

// MalwareExists returns true when a malware entry with given ID exists in world state
func (s *SmartContract) MalwareExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	malwareJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return malwareJSON != nil, nil
}
