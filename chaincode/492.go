package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Chaincode structure
type DoctorChaincode struct {
	contractapi.Contract
}

// Doctor data structure
type DoctorData struct {
	DoctorID string `json:"doctorID"`
	DoctorName string `json:"DoctorName"`
	Specialization string `json:"Specialization"`
	Address string `json:"Address"`
	PhoneNumber string `json:"PhoneNumber"`
}

// Function to add doctor 
func (d *DoctorChaincode) AddDoctor(ctx contractapi.TransactionContextInterface, doctorID string, doctorName string, specialization string, address string, phoneNumber string) error {

	// Create a new doctor data object
	doctorData := DoctorData{
		DoctorID: doctorID,
		DoctorName: doctorName,
		Specialization: specialization,
		Address: address,
		PhoneNumber: phoneNumber,
	}

	doctorDataJSON, err := json.Marshal(doctorData)
	if err != nil {
		return fmt.Errorf("failed to marshal doctor data: %v", err)
	}

	err = ctx.GetStub().PutState(doctorID, doctorDataJSON)
	if err != nil {
		return fmt.Errorf("failed to store doctor data: %v", err)
	}

	return nil
}

func (d *DoctorChaincode) DoctorExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// Function to delete existing doctor
func (d *DoctorChaincode) DeleteDoctor(ctx contractapi.TransactionContextInterface,id string) error {
	exists, err := d.DoctorExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}


func (d *DoctorChaincode) GetAllCurrentDoctors(ctx contractapi.TransactionContextInterface) ([]*DoctorData, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var doctorsList []*DoctorData
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var currentDoctor DoctorData
		err = json.Unmarshal(queryResponse.Value, &currentDoctor)
		if err != nil {
			return nil, err
		}
		doctorsList = append(doctorsList, &currentDoctor)
	}

	return doctorsList, nil
}

// Main function
func main() {

	// Create a new chaincode object
	chaincode, err := contractapi.NewChaincode(&DoctorChaincode{})
	if err != nil {
		fmt.Printf("Error creating Doctor data chaincode: %s", err.Error())
		return
	}

	// Start the chaincode
	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting doctor data chaincode: %s", err.Error())
	}
}
