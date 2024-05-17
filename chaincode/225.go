package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// TokenAsset represents an individual token
type TokenAsset struct {
	ID     string `json:"ID"`     // 토큰의 고유 식별자
	Owner  string `json:"Owner"`  // 토큰의 소유자
	Amount int    `json:"Amount"` // 토큰의 수량
}

// TokenContract represents the smart contract for managing tokens
type TokenContract struct {
	contractapi.Contract
}

// MintTokens issues new tokens to a specific owner
// 새로운 토큰을 발행
// 호출된 토큰 ID가 이미 존재하는지 확인하고, 없다면 새로운 토큰을 생성하여 World State에 저장
func (tc *TokenContract) MintTokens(ctx contractapi.TransactionContextInterface, owner string, tokenID string, amount int) error {
	// Check if the token already exists
	existingToken, err := ctx.GetStub().GetState(tokenID)
	if err != nil {
		return fmt.Errorf("Failed to read from world state: %v", err)
	}
	if existingToken != nil {
		return fmt.Errorf("Token with ID %s already exists", tokenID)
	}

	// Create a new token asset
	token := TokenAsset{
		ID:     tokenID,
		Owner:  owner,
		Amount: amount,
	}

	// Marshal the token into JSON
	tokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("Failed to marshal token to JSON: %v", err)
	}

	// Save the token to the world state
	err = ctx.GetStub().PutState(tokenID, tokenJSON)
	if err != nil {
		return fmt.Errorf("Failed to write to world state: %v", err)
	}
	return nil
}

// GetAllTokens retrieves all tokens stored in the world state
// World State에 저장된 모든 토큰을 검색
func (tc *TokenContract) GetAllTokens(ctx contractapi.TransactionContextInterface) ([]*TokenAsset, error) {
	// Get all tokens from the world state
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("token", []string{})
	if err != nil {
		return nil, fmt.Errorf("Failed to get tokens: %v", err)
	}
	defer resultsIterator.Close()

	// Iterate through the result set
	var tokens []*TokenAsset
	for resultsIterator.HasNext() {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Failed to iterate over tokens: %v", err)
		}

		// Unmarshal the token from JSON
		var token TokenAsset
		err = json.Unmarshal(responseRange.Value, &token)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal token from JSON: %v", err)
		}
		// Append the token to the result set
		tokens = append(tokens, &token)
	}

	return tokens, nil
}

// GetTokensByOwner retrieves tokens owned by a specific owner
// 특정 소유자가 소유한 모든 토큰을 검색
func (tc *TokenContract) GetTokensByOwner(ctx contractapi.TransactionContextInterface, owner string) ([]*TokenAsset, error) {
	// Get all tokens from the world state
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("token", []string{"owner", owner})
	if err != nil {
		return nil, fmt.Errorf("Failed to get tokens: %v", err)
	}
	defer resultsIterator.Close()

	// Iterate through the result set
	var tokens []*TokenAsset
	for resultsIterator.HasNext() {
		responseRange, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("Failed to iterate over tokens: %v", err)
		}

		// Unmarshal the token from JSON
		var token TokenAsset
		err = json.Unmarshal(responseRange.Value, &token)
		if err != nil {
			return nil, fmt.Errorf("Failed to unmarshal token from JSON: %v", err)
		}

		// Append the token to the result set
		tokens = append(tokens, &token)
	}
	return tokens, nil
}

// TransferTokens transfers tokens from one owner to another
// 특정 토큰을 한 소유자에서 다른 소유자로 이전
// 호출된 토큰 ID가 World State에 존재하는지 확인하고, 토큰의 소유자가 호출자인지 확인.
// 호출자가 토큰의 소유자가 아니라면 오류를 반환
// 소유자가 확인되면 새로운 소유자로 변경된 정보를 포함한 토큰을 World State에 업데이트.
func (tc *TokenContract) TransferTokens(ctx contractapi.TransactionContextInterface, newOwner string, tokenID string) error {
	// Retrieve the token from the world state
	tokenJSON, err := ctx.GetStub().GetState(tokenID)
	if err != nil {
		return fmt.Errorf("Failed to read from world state: %v", err)
	}
	if tokenJSON == nil {
		return fmt.Errorf("Token with ID %s does not exist", tokenID)
	}

	// Unmarshal the token from JSON
	var token TokenAsset
	err = json.Unmarshal(tokenJSON, &token)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal token from JSON: %v", err)
	}

	// Check if the caller is the current owner of the token
	callerID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("Failed to get client ID: %v", err)
	}

	if token.Owner != callerID {
		return fmt.Errorf("Caller is not the owner of the token")
	}

	// Update the owner of the token
	token.Owner = newOwner

	// Marshal the updated token into JSON
	updatedTokenJSON, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("Failed to marshal updated token to JSON: %v", err)
	}

	// Save the updated token to the world state
	err = ctx.GetStub().PutState(tokenID, updatedTokenJSON)
	if err != nil {
		return fmt.Errorf("Failed to write to world state: %v", err)
	}
	return nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&TokenContract{})
	if err != nil {
		fmt.Printf("Error creating token chaincode: %v\n", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting token chaincode: %v\n", err)
	}
}
