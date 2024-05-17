package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"poc/contract/model"
	"poc/contract/service"
	"strconv"
)

func Update(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	if len(args) != 3 {
		return "", fmt.Errorf("Incorrect arguments. Expecting a cardID, height, weight")
	}
	user, err := service.NewAuthService(stub).GetUser()
	if err != nil {
		return "", err
	}

	if !user.IsHospitalWorker() {
		return "", errors.New("user is not a parent")
	}

	var existingCard *model.Card = &model.Card{}
	err = service.NewCardService(stub).FindAndUnmarshal(args[0], existingCard)
	if err != nil {
		return "", err
	}

	height, err := strconv.Atoi(args[1])
	if err != nil {
		return "", err
	}
	weight, err := strconv.Atoi(args[2])
	if err != nil {
		return "", err
	}

	existingCard.Height = height
	existingCard.Weight = weight

	cardJsonBytes, err := json.Marshal(existingCard)
	if err != nil {
		return "", err
	}

	err = stub.PutState("CARD"+args[0], cardJsonBytes)
	if err != nil {
		return "", fmt.Errorf("Failed to set card: %s", args[0])
	}

	return args[1], nil
}
