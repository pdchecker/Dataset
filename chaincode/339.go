package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// SmartContract provides functions for managing a car
type SmartContract struct {
	contractapi.Contract
}

// Init is called when chaincode is iniialized
func (sc *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("instantiate was called!")
	return nil
}

// viewRecord checks if a Doctor has proper access to a patient's EHR, returns error if not
func (sc *SmartContract) viewRecord(ctx contractapi.TransactionContextInterface, patientID string, name string, EHR []byte) pb.Response {
	fmt.Println("viewRecord was called!")

	// check if Doctor is authorized to view account
	patientAccount := AuthorizedAccounts{
		PatientID: patientID,
		Name:      name,
	}

	isFoundAccount, err := patientAccount.LoadState(ctx)
	if err != nil {
		message := err.Error()
		return shim.Error(message)
	}

	if !isFoundAccount {
		message := fmt.Sprintf("unable to find the patient Authorization %s, %s", patientAccount.PatientID, patientAccount.Name)
		return pb.Response{Status: 404, Message: message}
	}

	// Check if doctor is authorized to view patient's EHR
	patientEHR := PatientEHR{
		PatientID: patientID,
		Name:      name,
	}

	isFoundEHR, err := patientEHR.LoadState(ctx)
	if err != nil {
		message := err.Error()
		return shim.Error(message)
	}

	if !isFoundEHR {
		message := fmt.Sprintf("unable to find the patient EHR %s, %s", patientAccount.PatientID, patientAccount.Name)
		fmt.Println(message)
		return pb.Response{Status: 404, Message: message}
	}

	result, err := json.Marshal(patientEHR.SingleRecord)
	if err != nil {
		message := fmt.Sprintf("unable to marshal the result: %s", err.Error())
		fmt.Println(message)
		return shim.Error(message)
	}

	fmt.Println("viewRecords exited successfully")
	return shim.Success(result)
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
