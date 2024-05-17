package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

// TokenFest provides functions for transferring tokens between accounts
type TokenFest struct {
	contractapi.Contract
}

func main() {

	TokenFestChaincode, err := contractapi.NewChaincode(&TokenFest{}, &FTContract{})
	if err != nil {
		log.Panicf("Error creating TokenFest chaincode: %v", err)
	}

	fmt.Println("TokenFestChaincode:", TokenFestChaincode)
	if err := TokenFestChaincode.Start(); err != nil {
		log.Panicf("Error starting TokenFestChaincode : %v", err)
	}

}

// InitLedger adds a base set of assets to the ledger
func (t *TokenFest) InitLedger(ctx contractapi.TransactionContextInterface) error {

	//Initiliaze Admin for HDFCBank

	admin := User{UserID: AdminID, UserName: "u"+AdminID, Role: []string{"admin"}, Balance: AdminBalance, BankID: BankID}

	adminJSON, err := json.Marshal(admin)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState("u"+AdminID, adminJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state for admin %v", err)
	}

	//Initiliaze bank
	bank := Bank{BankID: BankID, Name: "HDFC bank", MaxNoOfToken: Maxtoken, Admins: []string{AdminID}, TotalToken: AdminBalance}

	bankJSON, err := json.Marshal(bank)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(BankID, bankJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state for bank %v", err)
	}

	FTContract := FTContract{}
	ptrFTContract := &FTContract
	// Mint 200000 tokens by Admin
	(*ptrFTContract).MintToken(ctx, AdminID, 200000, "122435abc", "festtickets")

	return nil
}

func (t *TokenFest) RegisterUser(ctx contractapi.TransactionContextInterface, userID string, bankID string) error {

	//Verify userID is already registered
	UserBytes, err := ctx.GetStub().GetState(userID)
	if err != nil {
		return fmt.Errorf("failed to read user %s : %v", userID, err)
	}

	if UserBytes != nil {
		return fmt.Errorf("'userID' %s already exists", userID)

	}

	user := User{UserID: userID, UserName: "u"+userID, Role: []string{"user"}, BankID: bankID}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState("u"+userID, userJSON)
	if err != nil {
		return fmt.Errorf("failed to put to world state for user %v", err)
	}

	return nil

}

func (t *TokenFest) QueryBankByID(ctx contractapi.TransactionContextInterface, bankID string) (Bank, error) {
	//Check bankID exists or not

	bankBytes, err := ctx.GetStub().GetState(bankID)
	if err != nil {
		return Bank{}, fmt.Errorf("failed to read bankID %s : %v", bankID, err)
	}

	var bankData Bank

	if bankBytes == nil {
		return Bank{}, fmt.Errorf("bankID %s is invalid.It does not exist", bankID)

	}

	if bankBytes != nil {
		err = json.Unmarshal(bankBytes, &bankData)
	}

	return bankData, nil

}
