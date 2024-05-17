package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"poc_kinder/contract/service"
)

func GetAll(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	reportService := service.NewReportService(stub)
	user, err := service.NewAuthService(stub).GetUser()
	if err !=nil {
		return "", err
	}

	var response interface{}
	if user.IsParent() {
		response, err = reportService.FindAllForParent(user.Id)
	}

	if user.IsHospitalWorker() {
		response, err = reportService.FindAllForDoctor(user.Id)
	}

	if user.IsKindergartenWorker() {
		response, err = reportService.FindAllForKindergarten(user.Org)
	}

	if err !=nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(response)
	if err !=nil {
		return "", err
	}

	return string(jsonBytes), nil
}
