package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	// "time"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Bank struct {
	BankName       string  `json:"bankName"`
	Address        string  `json:"address"`
	TotalAmount    float64 `json:"totalAmount"`
	UserList       []User  `json:"userList"`
	AccountCounter int     `json:"accountCounter"`
	Version        int     `json:"version"`
}

type User struct {
	Name          string  `json:"name"`
	Address       string  `json:"address"`
	Amount        float64 `json:"amount"`
	AccountNumber int     `json:"accountNumber"`
	Version       int     `json:"version"`
}

/*const (
	MaxRetries     = 1000
	RetryInterval  = 100 * time.Millisecond
)*/

func (sc *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {

	bank := Bank{
		BankName:       "MyBank",
		Address:        "BankAddress",
		TotalAmount:    10000000,
		UserList:       []User{},
		AccountCounter: 0,
		Version:        1,
	}

	// Serialize and store the bank on the ledger
	bankJSON, err := json.Marshal(bank)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState("bank", bankJSON)
	if err != nil {
		return err
	}

	return nil
}

func (sc *SmartContract) CreateUser(ctx contractapi.TransactionContextInterface, name, address string) error {

	bankJSON, err := ctx.GetStub().GetState("bank")
	if err != nil {
		return fmt.Errorf("failed to read bank state in create user: %v", err);
	}

	bank := Bank{}

	if bankJSON != nil {
		err = json.Unmarshal(bankJSON, &bank)
		if err != nil {
			return fmt.Errorf("failed to unmarshal bank data in create user: %v", err);
		}
	}

	// Increment the account counter
	bank.AccountCounter++
	accountNumber := bank.AccountCounter

	newUser := User{
		Name:          name,
		Address:       address,
		Amount:        0.0,
		AccountNumber: accountNumber,
		Version:       1,
	}

	bank.UserList = append(bank.UserList, newUser)

	bankJSON, err = json.Marshal(bank)
	if err != nil {
		return fmt.Errorf("failed to marshal bank data in create user: %v", err);
	}

	userJSON, err := json.Marshal(newUser)
	if err != nil {
		return fmt.Errorf("failed to marshal user data in create user: %v", err);
	}

	err = ctx.GetStub().PutState("bank", bankJSON)
	if err != nil {
		return fmt.Errorf("failed to put bank data in create user: %v", err);
	}

	err = ctx.GetStub().PutState(strconv.Itoa(accountNumber), userJSON)
	if err != nil {
		return fmt.Errorf("failed to put user data in create user: %v", err);
	}

	return nil
}

func (sc *SmartContract) TransferMoney(ctx contractapi.TransactionContextInterface, fromAccountNumber, toAccountNumber int, amount float64) error {
	
	bankJSON, err := ctx.GetStub().GetState("bank")
	if err != nil {
		return fmt.Errorf("failed to read bank state in transfer money: %v", err);
	}

	bank := Bank{}
	err = json.Unmarshal(bankJSON, &bank)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bank data in transfer money: %v", err);
	}

	var fromUser, toUser *User

	for i, user := range bank.UserList {
		if user.AccountNumber == fromAccountNumber {
			fromUser = &bank.UserList[i]
		}
		if user.AccountNumber == toAccountNumber {
			toUser = &bank.UserList[i]
		}
	}

	if fromUser == nil {
		return fmt.Errorf("Sender user not found")
	}
	if toUser == nil {
		return fmt.Errorf("Receiver user not found")
	}

	if fromUser.Amount < amount {
		return fmt.Errorf("Insufficient balance")
	}

	fromUser.Amount -= amount
	toUser.Amount += amount

	bankJSON, err = json.Marshal(bank)
	if err != nil {
		return fmt.Errorf("failed to marshal bank data in transfer money: %v", err);
	}

	fromUserJSON, err := json.Marshal(fromUser)
	if err != nil {
		return fmt.Errorf("failed to marshal from user data in transfer money: %v", err);
	}

	toUserJSON, err := json.Marshal(toUser)
	if err != nil {
		return fmt.Errorf("failed to marshal to user data in transfer money: %v", err);
	}

	err = ctx.GetStub().PutState("bank", bankJSON)
	if err != nil {
		return fmt.Errorf("failed to put bank data in transfer money: %v", err);
	}

	err = ctx.GetStub().PutState(strconv.Itoa(fromAccountNumber), fromUserJSON)
	if err != nil {
		return fmt.Errorf("failed to put from user data in transfer money: %v", err);
	}

	err = ctx.GetStub().PutState(strconv.Itoa(toAccountNumber), toUserJSON)
	if err != nil {
		return fmt.Errorf("failed to put to user data in transfer money: %v", err);
	}

	return nil
}

// First version:
func (sc *SmartContract) IssueMoney(ctx contractapi.TransactionContextInterface, accountNumber int, amount float64) error {

	bankJSON, err := ctx.GetStub().GetState("bank")
	if err != nil {
		return fmt.Errorf("failed to read bank state in issue money: %v", err);
	}

	bank := Bank{}
	err = json.Unmarshal(bankJSON, &bank)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bank data in issue money: %v", err);
	}

	var user *User

	for i, u := range bank.UserList {
		if u.AccountNumber == accountNumber {
			user = &bank.UserList[i]
		}
	}

	if user == nil {
		return fmt.Errorf("User not found")
	}

	user.Amount += amount
	bank.TotalAmount -= amount

	bankJSON, err = json.Marshal(bank)
	if err != nil {
		return fmt.Errorf("failed to marshal bank data in issue money: %v", err);
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data in issue money: %v", err);
	}

	err = ctx.GetStub().PutState("bank", bankJSON)
	if err != nil {
		return fmt.Errorf("failed to put bank data in issue money: %v", err);
	}

	err = ctx.GetStub().PutState(strconv.Itoa(accountNumber), userJSON)
	if err != nil {
		return fmt.Errorf("failed to put user data in issue money: %v", err);
	}

	return nil
}

// Function handling concurrency update.
/*func (sc *SmartContract) IssueMoney(ctx contractapi.TransactionContextInterface, accountNumber int, amount float64) error {

	// Step 1: Read the current state of the bank
	bankJSON, err := ctx.GetStub().GetState("bank")
	if err != nil {
		return fmt.Errorf("failed to read bank state in issue money: %v", err)
	}

	bank := Bank{}
	err = json.Unmarshal(bankJSON, &bank)
	if err != nil {
		return fmt.Errorf("failed to unmarshal bank data in issue money: %v", err)
	}

	// Check if the version matches the expected version (assuming an optimistic concurrency control approach)
	expectedBankVersion := bank.Version
	if bank.Version != expectedBankVersion {
		return fmt.Errorf("concurrent update detected in bank data, please retry the transaction")
	}

	// Step 2: Read the current state of the user
	var user *User
	for i, u := range bank.UserList {
		if u.AccountNumber == accountNumber {
			user = &bank.UserList[i]
			break
		}
	}

	if user == nil {
		return fmt.Errorf("User not found")
	}

	// Check if the version matches the expected version
	expectedUserVersion := user.Version
	if user.Version != expectedUserVersion {
		return fmt.Errorf("concurrent update detected in user data, please retry the transaction")
	}

	// Step 3: Modify the state (Update user's balance and increment the versions)
	user.Amount += amount
	user.Version++

	bank.TotalAmount -= amount
	bank.Version++

	// Step 4: Attempt to write the updated states back to the ledger
	bankJSON, err = json.Marshal(bank)
	if err != nil {
		return fmt.Errorf("failed to marshal bank data in issue money: %v", err)
	}

	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user data in issue money: %v", err)
	}

	// Step 5: Put the updated states in the ledger
	err = ctx.GetStub().PutState("bank", bankJSON)
	if err != nil {
		return fmt.Errorf("failed to put bank data in issue money: %v", err)
	}

	err = ctx.GetStub().PutState(strconv.Itoa(accountNumber), userJSON)
	if err != nil {
		return fmt.Errorf("failed to put user data in issue money: %v", err)
	}

	// Step 6: Transaction completed successfully
	return nil
}

// Third version:
/*func (sc *SmartContract) IssueMoney(ctx contractapi.TransactionContextInterface, accountNumber int, amount float64) (int, error) {

    var failedTransactions int

    // Attempt to fetch the current state of the "bank" key
    bankJSON, err := ctx.GetStub().GetState("bank")
    if err != nil {
        return 1, err
    }

    // Save the existing state
    existingBankJSON := make([]byte, len(bankJSON))
    copy(existingBankJSON, bankJSON)

    // Unmarshal the existing state
    var existingBank Bank
    err = json.Unmarshal(existingBankJSON, &existingBank)
    if err != nil {
        return 1, err
    }

    // Find the user in the bank
    var user *User
    for i, u := range existingBank.UserList {
        if u.AccountNumber == accountNumber {
            user = &existingBank.UserList[i]
        }
    }

    // Check if the user was found
    if user == nil {
        return failedTransactions, nil
    }

    // Attempt to fetch the current state of the user
    userJSON, err := ctx.GetStub().GetState(strconv.Itoa(accountNumber))
    if err != nil {
        return 1, err
    }

    // Save the existing user state
    existingUserJSON := make([]byte, len(userJSON))
    copy(existingUserJSON, userJSON)

    var existingUser User
    err = json.Unmarshal(existingUserJSON, &existingUser)
    if err != nil {
        return 1, err
    }

    // Update the user and bank
    user.Amount += amount
    existingBank.TotalAmount -= amount

    currentUserJSON, err := ctx.GetStub().GetState(strconv.Itoa(accountNumber))
    if err != nil {
        return 1, err
    }

    if !bytes.Equal(existingUserJSON, currentUserJSON) {
        // Concurrent update detected, retry the transaction
        failedTransactions++
        return failedTransactions, nil
    }

    // Marshal the updated bank state
    bankJSON, err = json.Marshal(existingBank)
    if err != nil {
        return 1, err
    }

    // Check for concurrent updates during PutState
    currentBankJSON, err := ctx.GetStub().GetState("bank")
    if err != nil {
        return 1, err
    }

    if !bytes.Equal(existingBankJSON, currentBankJSON) {
        // Concurrent update detected, retry the transaction
        failedTransactions++
        return failedTransactions, nil
    }

    // Put the new state
    err = ctx.GetStub().PutState("bank", bankJSON)
    if err != nil {
        return 1, err
    }

    // Put the user state
    userJSON, err = json.Marshal(user)
    if err != nil {
        return 1, err
    }

    err = ctx.GetStub().PutState(strconv.Itoa(accountNumber), userJSON)
    if err != nil {
        return 1, err
    }

    return failedTransactions, nil
}*/

// Second version:
/*func (sc *SmartContract) IssueMoney(ctx contractapi.TransactionContextInterface, accountNumber int, amount float64) error {

    for retry := 0; retry < MaxRetries; retry++ {
        // Attempt to fetch the current state of the "bank" key
        bankJSON, err := ctx.GetStub().GetState("bank")
        if err != nil {
            return err
        }

        // Save the existing state
		existingBankJSON := make([]byte, len(bankJSON))
		copy(existingBankJSON, bankJSON)

        // Unmarshal the existing state
        var existingBank Bank
        err = json.Unmarshal(existingBankJSON, &existingBank)
        if err != nil {
            return err
        }

        // Find the user in the bank
        var user *User
        for i, u := range existingBank.UserList {
            if u.AccountNumber == accountNumber {
                user = &existingBank.UserList[i]
            }
        }

        // Check if the user was found
        if user == nil {
            return fmt.Errorf("User not found")
        }

		// Attempt to fetch the current state of the user
        userJSON, err := ctx.GetStub().GetState(strconv.Itoa(accountNumber))
        if err != nil {
            return err
        }

		// Save the existing user state
        existingUserJSON := make([]byte, len(userJSON))
        copy(existingUserJSON, userJSON)

		var existingUser User
        err = json.Unmarshal(existingUserJSON, &existingUser)
        if err != nil {
            return err
        }

        // Update the user and bank
        user.Amount += amount
        existingBank.TotalAmount -= amount

		currentUserJSON, err := ctx.GetStub().GetState(strconv.Itoa(accountNumber))
        if err != nil {
            return err
        }

		if !bytes.Equal(existingUserJSON, currentUserJSON) {
            // Concurrent update detected, retry the transaction
            time.Sleep(RetryInterval)
            continue
        }

        // Marshal the updated bank state
        bankJSON, err = json.Marshal(existingBank)
        if err != nil {
            return err
        }

        // Check for concurrent updates during PutState
        currentBankJSON, err := ctx.GetStub().GetState("bank")
        if err != nil {
            return err
        }

        if !bytes.Equal(existingBankJSON, currentBankJSON) {
            // Concurrent update detected, retry the transaction
            time.Sleep(RetryInterval)
            continue
        }

        // Put the new state
        err = ctx.GetStub().PutState("bank", bankJSON)
        if err != nil {
            return err
        }

        // Put the user state
        userJSON, err = json.Marshal(user)
        if err != nil {
            return err
        }

        err = ctx.GetStub().PutState(strconv.Itoa(accountNumber), userJSON)
        if err != nil {
            return err
        }
		
        break
    }

    return nil
}*/

func (sc *SmartContract) QueryUser(ctx contractapi.TransactionContextInterface, accountNumber int) (*User, error) {

	userJSON, err := ctx.GetStub().GetState(strconv.Itoa(accountNumber))
	if err != nil {
		return nil, fmt.Errorf("failed to get user data in query user: %v", err);
	}

	if userJSON == nil {
		return nil, fmt.Errorf("User not found")
	}

	user := User{}
	err = json.Unmarshal(userJSON, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal user data in query user: %v", err);
	}

	return &user, nil
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {
	// Create a new Smart Contract
	smartContract, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating SmartContract chaincode: %v\n", err)
		return
	}

	if err := smartContract.Start(); err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %v\n", err)
		return
	}
}