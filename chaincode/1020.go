package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"poc/contract/model"
	"poc/contract/service"
)

func Sign(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a key")
	}
	key := args[0]

	user, err := service.NewAuthService(stub).GetUser()
	if err != nil {
		return "", err
	}
	if !user.IsHospitalWorker() {
		return "", errors.New("user is not a Pediatrician")
	}

	var agreement model.Agreement
	agreementService := service.NewAgreementService(stub)
    err = agreementService.FindAndUnmarshal(key, &agreement)

	if err != nil {
		return "", err
	}

	agreementService.Sign(&agreement)
	jsonBytes, err := json.Marshal(agreement)
	if err != nil {
		return  "", fmt.Errorf("Failed to marshall agreement obj", key)
	}

	err = agreementService.Put(key, jsonBytes)
	if err != nil {
		return  "", err
	}

	return string(jsonBytes), nil
}