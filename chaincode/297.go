package main

import (
	"chaincode-go/model"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// 用户相关

// CreateUser issues a new user to the world state with given details.
func (s *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, request string) error {

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)

	if err != nil {
		return err
	}
	exists, err := s.UserExists(ctx, clientID)

	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s already exists", clientID)
	}

	var user model.User
	err = json.Unmarshal([]byte(request), &user)
	if err != nil {
		return err
	}

	user.ID = clientID

	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(clientID, userJSON)
}

func (s *SmartContract) DeleteUser(ctx contractapi.TransactionContextInterface) error {

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	exists, err := s.UserExists(ctx, clientID)

	if err != nil {
		return err
	}
	if exists {
		ctx.GetStub().DelState(clientID)
		return nil
	}

	return fmt.Errorf("用户不存在")
}

func (s *SmartContract) FindUserById(ctx contractapi.TransactionContextInterface, userId string) (*model.User, error) {

	userAsByte, err := ctx.GetStub().GetState(userId)
	if err != nil {
		return nil, fmt.Errorf("查询资源失败")
	}
	var user model.User
	err = json.Unmarshal(userAsByte, &user)

	return &user, err
}

// returns all users found in world state
func (s *SmartContract) GetAllUsers(ctx contractapi.TransactionContextInterface) ([]*model.User, error) {

	startKey := "user:"
	endKey := string(BytesPrefix([]byte(startKey)))
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var users []*model.User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var user model.User

		err = json.Unmarshal(queryResponse.Value, &user)
		if err != nil {
			return nil, err
		}
		users = append(users, &user)
	}

	return users, nil
}

func (s *SmartContract) GetUserHistory(ctx contractapi.TransactionContextInterface) ([]string, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	clientIdentity, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return nil, err
	}
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(clientIdentity)
	var result []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		resJson, err := json.Marshal(queryResponse)
		result = append(result, string(resJson))
	}
	return result, nil
}

// UserExists returns true when asset with given ID exists in world state
func (s *SmartContract) UserExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {

	userJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return userJSON != nil, nil
}

// GetSubmittingClientIdentity returns the name and issuer of the identity that
// invokes the smart contract. This function base64 decodes the identity string
// before returning the value to the client or smart contract.
func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {
	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	//decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	//if err != nil {
	//	return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	//}
	sum := md5.New().Sum([]byte(b64ID))
	identify := "user:" + hex.EncodeToString(sum)
	//fmt.Println(identify)
	return identify, nil
}
