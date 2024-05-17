// smart_contract_test.go

package main

import (
	"testing"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/vidalpaul/golang-chaincodes/tokens/erc20/chaincode"
)

// TestTransfer tests the transfer function of the ERC20 token.
func TestTransfer(t *testing.T) {
	// Create a new instance of the SmartContract struct
	sc := chaincode.SmartContract{}

	// Mock the transaction context using contractapi.TransactionContextInterface
	ctx := &contractapi.TransactionContext{}

	// Set up initial state, e.g., account balances
	// For simplicity, we will use a mock implementation of contractapi.StateListInterface
	stateList := &MockStateList{}
	sc.StateList = stateList

	// Set up test data
	sender := "Alice"
	recipient := "Bob"
	amount := 100

	// Mock existing sender balance
	stateList.On("GetState", sender).Return([]byte("1000"), nil)

	// Invoke the transfer function
	err := sc.Transfer(ctx, recipient, amount)
	assert.NoError(t, err, "Transfer should succeed")

	// Verify the updated balances after the transfer
	senderBalance, _ := stateList.GetState(sender)
	recipientBalance, _ := stateList.GetState(recipient)
	assert.Equal(t, "900", string(senderBalance), "Sender balance should be reduced by the transferred amount")
	assert.Equal(t, "100", string(recipientBalance), "Recipient balance should be increased by the transferred amount")
}

// TestTransferInsufficientBalance tests the transfer function with insufficient balance.
func TestTransferInsufficientBalance(t *testing.T) {
	// Create a new instance of the SmartContract struct
	sc := chaincode.SmartContract{}

	// Mock the transaction context using contractapi.TransactionContextInterface
	ctx := &contractapi.TransactionContext{}

	// Set up initial state, e.g., account balances
	// For simplicity, we will use a mock implementation of contractapi.StateListInterface
	stateList := &MockStateList{}
	sc.StateList = stateList

	// Set up test data
	sender := "Alice"
	recipient := "Bob"
	amount := 100

	// Mock existing sender balance (less than the transfer amount)
	stateList.On("GetState", sender).Return([]byte("50"), nil)

	// Invoke the transfer function
	err := sc.Transfer(ctx, recipient, amount)
	assert.EqualError(t, err, "Insufficient balance", "Transfer should fail with an error message")
}

// MockStateList is a mock implementation of contractapi.StateListInterface for testing.
type MockStateList struct{}

func (m *MockStateList) AddState(key string, value []byte) error {
	// Mock implementation of AddState
	return nil
}

func (m *MockStateList) GetState(key string) ([]byte, error) {
	// Mock implementation of GetState
	return nil, nil
}

func (m *MockStateList) UpdateState(key string, value []byte) error {
	// Mock implementation of UpdateState
	return nil
}

func (m *MockStateList) DeleteState(key string) error {
	// Mock implementation of DeleteState
	return nil
}
