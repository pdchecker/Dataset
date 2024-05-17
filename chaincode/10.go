/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	err := ctx.GetStub().PutState("snapshot_exists", []byte("0"))
	if err != nil {
		return fmt.Errorf("Failed to put to world state. %s", err.Error())
	}

	return nil
}

var (
	WarningLogger *log.Logger
	InfoLogger    *log.Logger
	ErrorLogger   *log.Logger
)

const NgoMSP = "Org3MSP"
const CorporateMSP = "Org2MSP"
const CreditsAuthorityMSP = "Org1MSP"

const corporate = "org2"
const ngo = "org3"
const creditsauthority = "org1"
const domain = "example.com"
const guest = "guest"
const ca = "ca"

func main() {
	//initialize custom loggers
	InfoLogger = log.New(os.Stdout, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLogger = log.New(os.Stdout, "WARNING: ", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLogger = log.New(os.Stdout, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile)

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error creating csr chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting csr chaincode: %s", err.Error())
	}
}
