package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// UserCredential represents the data structure for user credentials
type UserCredential struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// CredentialContract represents the smart contract for managing user credentials
type CredentialContract struct {
	contractapi.Contract
}

// SignUp registers a new user
func (cc *CredentialContract) SignUp(ctx contractapi.TransactionContextInterface, userID, username, password, email, role string) error {
	userCredential := UserCredential{
		UserID:   userID,
		Username: username,
		Password: password,
		Email:    email,
		Role:     role,
	}

	userCredentialBytes, err := json.Marshal(userCredential)
	if err != nil {
		return fmt.Errorf("failed to marshal user credential: %v", err)
	}

	return ctx.GetStub().PutState(userID, userCredentialBytes)
}

// Login verifies user credentials
func (cc *CredentialContract) Login(ctx contractapi.TransactionContextInterface, userID, password string) (*UserCredential, error) {
	userCredentialBytes, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to read user credential: %v", err)
	}
	if userCredentialBytes == nil {
		return nil, fmt.Errorf("user credential not found")
	}

	var userCredential UserCredential
	err = json.Unmarshal(userCredentialBytes, &userCredential)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user credential: %v", err)
	}

	if userCredential.Password != password {
		return nil, fmt.Errorf("invalid password")
	}

	return &userCredential, nil
}

// DeleteUser removes a user from the system (only accessible by admin)
func (cc *CredentialContract) DeleteUser(ctx contractapi.TransactionContextInterface, userID string) error {

	isAdmin := false
	// Check if caller is admin
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
	if err == nil {
		isAdmin = true
	}
	if !isAdmin {
		return fmt.Errorf("only admin can delete users")
	}

	return ctx.GetStub().DelState(userID)
}

// UpdateUserDetails updates user details (only accessible by admin)
func (cc *CredentialContract) UpdateUserDetails(ctx contractapi.TransactionContextInterface, userID, newUsername, newPassword, newEmail string) error {
	// Check if caller is admin
	isAdmin := false
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "admin")
	if err == nil {
		isAdmin = true
	}
	if !isAdmin {
		return fmt.Errorf("only admin can update user details")
	}

	userCredentialBytes, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return fmt.Errorf("failed to read user credential: %v", err)
	}
	if userCredentialBytes == nil {
		return fmt.Errorf("user credential not found")
	}

	var userCredential UserCredential
	err = json.Unmarshal(userCredentialBytes, &userCredential)
	if err != nil {
		return fmt.Errorf("failed to unmarshal user credential: %v", err)
	}

	// Update user details
	if newUsername != "" {
		userCredential.Username = newUsername
	}
	if newPassword != "" {
		userCredential.Password = newPassword
	}
	if newEmail != "" {
		userCredential.Email = newEmail
	}

	updatedUserCredentialBytes, err := json.Marshal(userCredential)
	if err != nil {
		return fmt.Errorf("failed to marshal updated user credential: %v", err)
	}

	return ctx.GetStub().PutState(userID, updatedUserCredentialBytes)
}

// Init initializes the chaincode
func (cc *CredentialContract) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

func main() {
	credentialContract := new(CredentialContract)

	chaincode, err := contractapi.NewChaincode(credentialContract)
	if err != nil {
		fmt.Printf("Error creating credential chaincode: %v", err)
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting credential chaincode: %v", err)
	}
}
