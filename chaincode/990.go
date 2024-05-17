package main

import (
	// "github.com/hyperledger/fabric-contract-api-go/contractapi"

	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Patient struct {
	Nik          string `json:"nik,omitempty"  bson:"nik"  form:"nik"  binding:"nik"`
	Date         string `json:"date,omitempty"  bson:"date"  form:"date"  binding:"date"`
	Gender       string `json:"gender,omitempty"  bson:"gender"  form:"gender"  binding:"gender"`
	Age          string `json:"age,omitempty"  bson:"age"  form:"age"  binding:"age"`
	Location     string `json:"location,omitempty"  bson:"location"  form:"location"  binding:"location"`
	HealthRecord string `json:"health_record,omitempty"  bson:"health_record"  form:"health_record"  binding:"health_record"`
	Status       string `json:"status,omitempty"  bson:"status"  form:"status"  binding:"status"`
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, nik string) (bool, error) {
	assetJson, err := ctx.GetStub().GetState(nik)
	if err != nil {
		return false, fmt.Errorf("failed to read from state database: %v", err)
	}
	return assetJson != nil, nil
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface,
	nik string,
	date string,
	gender string,
	age string,
	location string,
	healthRecord string,
	status string) error {
	exists, err := s.AssetExists(ctx, nik)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the patient %s is already exist", nik)
	}

	patient := Patient{
		Nik:          nik,
		Date:         date,
		Gender:       gender,
		Age:          age,
		Location:     location,
		HealthRecord: healthRecord,
		Status:       status,
	}

	assetJson, err := json.Marshal(patient)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(nik, assetJson)
}

func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface,
	nik string,
	date string,
	gender string,
	age string,
	location string,
	healthRecord string,
	status string) error {
	exists, err := s.AssetExists(ctx, nik)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the patient %s is does not exist", nik)
	}

	patient := Patient{
		Nik:          nik,
		Date:         date,
		Gender:       gender,
		Age:          age,
		Location:     location,
		HealthRecord: healthRecord,
		Status:       status,
	}

	assetJson, err := json.Marshal(patient)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(nik, assetJson)
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, nik string) (*Patient, error) {
	assetJson, err := ctx.GetStub().GetState(nik)
	if err != nil {
		return nil, fmt.Errorf("faield to read from state database: %v", err)
	}
	if assetJson == nil {
		return nil, fmt.Errorf("the user %s does not exist", nik)
	}

	var patient Patient
	err = json.Unmarshal(assetJson, &patient)
	if err != nil {
		return nil, err
	}
	return &patient, nil
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, nik string) error {
	exists, err := s.AssetExists(ctx, nik)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the patient %s does not exist", nik)
	}
	return ctx.GetStub().DelState(nik)
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Patient, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Patient
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Patient
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
