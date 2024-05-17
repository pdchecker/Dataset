package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Define key names for options
const totalSupplyKey = "totalSupply"

// Define objectType names for prefix
const allowancePrefix = "allowance"

type SmartContract struct {
	contractapi.Contract
}


type Userinfo struct{
	UserID string `json:"userid"`
	username string `json:"username"`
	Password string `json:"password"`
	PostingList []Postinginfo `json:"postinglist"`
	TokenBalance float64 `json:"tokenbalance"`
}

type Postinginfo struct{
	UserID string  `json:"projecttitle"`
	Postingtitle string  `json:"projecttitle"`
	Postingcontents string  `json:"projecttitle"`
	Postingdate string  `json:"projecttitle"`
	Score float64 `json:"score"`
}

type event struct {
	From  string `json:"from"`
	To    string `json:"to"`
	Value float64    `json:"value"`
}

type UserRating struct{
	User string `json:"user"`
	Average float64 `json:"average"`
	Rates []Rate `json:"rates"`
}
type Rate struct{
	ProjectTitle string  `json:"projecttitle"`
	Score float64 `json:"score"`
}
 
func (s *SmartContract) AddUser(ctx contractapi.TransactionContextInterface, username string) error {

	var user = UserRating{User: username, Average: 0}
	userAsBytes, _ := json.Marshal(user)	

	return ctx.GetStub().PutState(username, userAsBytes)
}

func (s *SmartContract) AddRating(ctx contractapi.TransactionContextInterface, username string, prjTitle string, prjscore string) error {
	
	// getState User 
	userAsBytes, err := ctx.GetStub().GetState(username)	

	if err != nil{
		return err
	} else if userAsBytes == nil{ // no State! error
		return fmt.Errorf("\"Error\":\"User does not exist: "+ username+"\"")
	}
	// state ok
	user := UserRating{}
	err = json.Unmarshal(userAsBytes, &user)
	if err != nil {
		return err
	}
	// create rate structure
	newRate, _ := strconv.ParseFloat(prjscore,64) 
	var Rate = Rate{ProjectTitle: prjTitle, Score: newRate}

	rateCount := float64(len(user.Rates))

	user.Rates=append(user.Rates,Rate)

	user.Average = (rateCount*user.Average+newRate)/(rateCount+1)
	// update to User World state
	userAsBytes, err = json.Marshal(user);
	if err != nil {
		return fmt.Errorf("failed to Marshaling: %v", err)
	}	

	err = ctx.GetStub().PutState(username, userAsBytes)
	if err != nil {
		return fmt.Errorf("failed to AddRating: %v", err)
	}	
	return nil
}

func (s *SmartContract) ReadRating(ctx contractapi.TransactionContextInterface, username string) (string, error) {

	UserAsBytes, err := ctx.GetStub().GetState(username)

	if err != nil {
		return "", fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if UserAsBytes == nil {
		return "", fmt.Errorf("%s does not exist", username)
	}

	// user := new(UserRating)
	// _ = json.Unmarshal(UserAsBytes, &user)
	
	return string(UserAsBytes[:]), nil	
}

func (s *SmartContract) UserInit(ctx contractapi.TransactionContextInterface) error {
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}
	if clientMSPID != "Org3MSP" {
		return fmt.Errorf("client is not authorized to mint new tokens%s", clientMSPID)
	}

	// Get ID of submitting client identity
	minter, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	currentBalanceBytes, err := ctx.GetStub().GetState(minter)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", minter, err)
	}

	var currentBalance int

	// If minter current balance doesn't yet exist, we'll create it with a current balance of 0
	if currentBalanceBytes == nil {
		currentBalance = 0
	} else {
		currentBalance, _ = strconv.Atoi(string(currentBalanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.
	}

	err = ctx.GetStub().PutState(minter, []byte(strconv.Itoa(currentBalance)))
	if err != nil {
		return err
	}

	return nil
}


func (s *SmartContract) Mint(ctx contractapi.TransactionContextInterface) error {

	const amount = 0.9

	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to mint new tokens
	clientMSPID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get MSPID: %v", err)
	}
	if clientMSPID != "Org3MSP" {
		return fmt.Errorf("client is not authorized to mint new tokens%s", clientMSPID)
	}

	// Get ID of submitting client identity
	minter, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	if amount <= 0 {
		return fmt.Errorf("mint amount must be a positive integer")
	}

	currentBalanceBytes, err := ctx.GetStub().GetState(minter)
	if err != nil {
		return fmt.Errorf("failed to read minter account %s from world state: %v", minter, err)
	}

	var currentBalance float64
	// var currentBalance int

	// If minter current balance doesn't yet exist, we'll create it with a current balance of 0
	if currentBalanceBytes == nil {
		currentBalance = 0
	} else {
		currentBalance, _ = strconv.ParseFloat(string(currentBalanceBytes),64) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.
	}

	updatedBalance := currentBalance + amount
	
	err = ctx.GetStub().PutState(minter, []byte(strconv.FormatFloat(updatedBalance, 'f', -1, 64)))
	// err = ctx.GetStub().PutState(minter, []byte(strconv.Itoa(updatedBalance)))
	if err != nil {
		return err
	}

	// Update the totalSupply
	totalSupplyBytes, err := ctx.GetStub().GetState(totalSupplyKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve total token supply: %v", err)
	}

	// var totalSupply int
	var totalSupply float64
	// If no tokens have been minted, initialize the totalSupply
	if totalSupplyBytes == nil {
		totalSupply = 0
	} else {
		// totalSupply, _ = strconv.Atoi(string(totalSupplyBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
		totalSupply, _ = strconv.ParseFloat(string(totalSupplyBytes),64) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
	}

	// Add the mint amount to the total supply and update the state
	totalSupply += amount
	err = ctx.GetStub().PutState(totalSupplyKey, []byte(strconv.FormatFloat(totalSupply, 'f', -1, 64)))
	if err != nil {
		return err
	}

	// Emit the Transfer event
	transferEvent := event{"0x0", minter, amount}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return fmt.Errorf("failed to set event: %v", err)
	}

	fmt.Printf("minter account %s balance updated from %d to %d", minter, currentBalance, updatedBalance)

	return nil
}

// func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, recipient string, amount float64) error {

// 	// Get ID of submitting client identity
// 	clientID, err := ctx.GetClientIdentity().GetID()
// 	if err != nil {
// 		return fmt.Errorf("failed to get client id: %v", err)
// 	}

// 	err = transferHelper(ctx, clientID, recipient, amount)
// 	if err != nil {
// 		return fmt.Errorf("failed to transfer: %v", err)
// 	}

// 	// Emit the Transfer event
// 	transferEvent := event{clientID, recipient, amount}
// 	transferEventJSON, err := json.Marshal(transferEvent)
// 	if err != nil {
// 		return fmt.Errorf("failed to obtain JSON encoding: %v", err)
// 	}
// 	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
// 	if err != nil {
// 		return fmt.Errorf("failed to set event: %v", err)
// 	}

// 	return nil
// }

// BalanceOf returns the balance of the given account
func (s *SmartContract) BalanceOf(ctx contractapi.TransactionContextInterface, account string) (int, error) {
	balanceBytes, err := ctx.GetStub().GetState(account)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if balanceBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", account)
	}

	balance, _ := strconv.Atoi(string(balanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

	return balance, nil
}

// ClientAccountBalance returns the balance of the requesting client's account
func (s *SmartContract) ClientAccountBalance(ctx contractapi.TransactionContextInterface) (float64, error) {

	// Get ID of submitting client identity
	clientID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return 0, fmt.Errorf("failed to get client id: %v", err)
	}

	balanceBytes, err := ctx.GetStub().GetState(clientID)
	if err != nil {
		return 0, fmt.Errorf("failed to read from world state: %v", err)
	}
	if balanceBytes == nil {
		return 0, fmt.Errorf("the account %s does not exist", clientID)
	}
	
	balance, _ := strconv.ParseFloat(string(balanceBytes),64)
	// balance, _ := strconv.Atoi(string(balanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

	return balance, nil
}

// ClientAccountID returns the id of the requesting client's account
// In this implementation, the client account ID is the clientId itself
// Users can use this function to get their own account id, which they can then give to others as the payment address
func (s *SmartContract) ClientAccountID(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get ID of submitting client identity
	clientAccountID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get client id: %v", err)
	}

	return clientAccountID, nil
}

// transferHelper is a helper function that transfers tokens from the "from" address to the "to" address
// Dependant functions include Transfer and TransferFrom
// func transferHelper(ctx contractapi.TransactionContextInterface, from string, to string, value float64) error {

// 	if from == to {
// 		return fmt.Errorf("cannot transfer to and from same client account")
// 	}

// 	if value < 0 { // transfer of 0 is allowed in ERC-20, so just validate against negative amounts
// 		return fmt.Errorf("transfer amount cannot be negative")
// 	}

// 	fromCurrentBalanceBytes, err := ctx.GetStub().GetState(from)
// 	if err != nil {
// 		return fmt.Errorf("failed to read client account %s from world state: %v", from, err)
// 	}

// 	if fromCurrentBalanceBytes == nil {
// 		return fmt.Errorf("client account %s has no balance", from)
// 	}

// 	fromCurrentBalance, _ := strconv.Atoi(string(fromCurrentBalanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

// 	if fromCurrentBalance < value {
// 		return fmt.Errorf("client account %s has insufficient funds", from)
// 	}

// 	toCurrentBalanceBytes, err := ctx.GetStub().GetState(to)
// 	if err != nil {
// 		return fmt.Errorf("failed to read recipient account %s from world state: %v", to, err)
// 	}

// 	var toCurrentBalance int
// 	// If recipient current balance doesn't yet exist, we'll create it with a current balance of 0
// 	if toCurrentBalanceBytes == nil {
// 		toCurrentBalance = 0
// 	} else {
// 		toCurrentBalance, _ = strconv.Atoi(string(toCurrentBalanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.
// 	}

// 	fromUpdatedBalance := fromCurrentBalance - value
// 	toUpdatedBalance := toCurrentBalance + value

// 	err = ctx.GetStub().PutState(from, []byte(strconv.Itoa(fromUpdatedBalance)))
// 	if err != nil {
// 		return err
// 	}

// 	err = ctx.GetStub().PutState(to, []byte(strconv.Itoa(toUpdatedBalance)))
// 	if err != nil {
// 		return err
// 	}

// 	fmt.Printf("client %s balance updated from %d to %d", from, fromCurrentBalance, fromUpdatedBalance)
// 	fmt.Printf("recipient %s balance updated from %d to %d", to, toCurrentBalance, toUpdatedBalance)

// 	return nil
// }


func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create teamate chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting teamate chaincode: %s", err.Error())
	}
}