package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"math/rand"
	"time"
)

type SmartContract struct { 
	contractapi.Contract
}

type User struct {
	Id string `json:"id"`
	Password string `json:"password"`
	Money uint `json:"money"`
}

func (s *SmartContract) InitUser(ctx contractapi.TransactionContextInterface) error {
	users := []User{
		User{Id: "admin", Password: "adminadmin", Money: 999999},
		User{Id: "guest", Password: "guestguest", Money: 10000},
	}
	for _, user := range users {
		userJSON, err := json.Marshal(user)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(user.Id, userJSON)
		if err != nil {
			return fmt.Errorf("Failed to put to world state. %v", err)
		}
	}
	return nil
}

func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, id string, password string) error {
	err := s.UserExists(ctx, id)
	if err != nil{
		return fmt.Errorf("User exists.")
	}
	user := User{
		Id:       id,
		Password: password,
		Money:    10000,
	}
	userAsBytes, _ := json.Marshal(user)
	return ctx.GetStub().PutState(id, userAsBytes)
}

func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, id string) error {
	_, err := ctx.GetStub().GetState(id)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) FindUser(ctx contractapi.TransactionContextInterface, id string) (*User, error) {
	userAsBytes, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if userAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", id)
	}

	user := new(User)
	_ = json.Unmarshal(userAsBytes, user)

	return user, nil
}

func (s *SmartContract) PlayGame(ctx contractapi.TransactionContextInterface, id string, betMoney uint, betNumber uint) error {
	rand.Seed(time.Now().UnixNano())
	randomNumber := rand.Intn(3) + 1
	user, err := s.FindUser(ctx, id)
	if err != nil {
		return err
	}
	if user.Money < betMoney {
		userAsBytes, _ := json.Marshal(user)
		return ctx.GetStub().PutState(id, userAsBytes)
	}
	if betNumber == uint(randomNumber) {
		user.Money += betMoney
		userAsBytes, _ := json.Marshal(user)
		return ctx.GetStub().PutState(id, userAsBytes)
	} else {
		user.Money -= betMoney
		userAsBytes, _ := json.Marshal(user)
		return ctx.GetStub().PutState(id, userAsBytes)
	}
	
	return nil
}

func main() {
	userChaincode, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		fmt.Println("Error %v",err)
	}
	if err := userChaincode.Start(); err != nil {
		fmt.Println("Error, %v",err)
	}
}
