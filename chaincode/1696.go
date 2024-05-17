package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"poc_kinder/contract/model"
	"poc_kinder/contract/service"
)

func Create(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 4 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key and report data, card id, parent")
	}
	key, reportDataString, cardId, parent := args[0], args[1], args[2], args[3]
	user, err := service.NewAuthService(stub).GetUser()
	if err != nil {
		return "", err
	}
	if !user.IsHospitalWorker() {
		return "", errors.New("only hospital worker can create reports")
	}

	reportService := service.NewReportService(stub)
	exists, err := reportService.Exists(key)
	if err != nil {
		return "", err
	}
	if exists {
		return "", errors.New("report with same id already exists")
	}

	var reportData model.ReportData
	err = json.Unmarshal([]byte(reportDataString), &reportData)
	if err != nil {
		return "", err
	}

	report := model.CreateReport(key, cardId, reportData, parent, user.Id)
	jsonBytes, err := json.Marshal(report)
	if err != nil {
		return "", fmt.Errorf("Failed to marshall report obj", args[0])
	}

	err = reportService.Put(key, jsonBytes)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
