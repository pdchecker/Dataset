package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Bank struct {
}

// Init is called during chaincode instantiation to initialize any
// data. Note that chaincode upgrade also calls this function to reset
// or to migrate data.
func (t *Bank) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success([]byte("done"))
}

// Invoke is called per transaction on the chaincode.
func (t *Bank) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error
	if fn == "sendAmount" {
		result, err = sendAmount(stub, args)
	} else {
		if fn == "getBalance" {
			result, err = getBalance(stub, args)
		} else {
			if fn == "createAccount" {
				result, err = createAccount(stub, args)
			} else {
				if fn == "createAccounts" {
					result, err = createAccounts(stub, args)
				} else {
					return shim.Error("no such method")
				}
			}
		}
	}
	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

func sendAmount(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// check args
	if len(args) != 3 {
		return "error", fmt.Errorf("Expecting 3 arguments: account1Id, account2Id, amount.")
	}

	// convert string to float
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return "error", fmt.Errorf("Argument 3 should be a float: %s", args[2])
	}

	// get balance of account1
	data, err := stub.GetState(args[0])
	if err != nil {
		return "error", fmt.Errorf("No such account: %s", args[0])
	}
	acc1, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return "error", fmt.Errorf("Invalid account balance: %s", string(data))
	}

	// get balance of account2
	data, err = stub.GetState(args[1])
	if err != nil {
		return "error", fmt.Errorf("No such account: %s", args[1])
	}
	acc2, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return "error", fmt.Errorf("Invalid account balance: %s", string(data))
	}

	// check balance
	if acc1 < amount {
		return "error", fmt.Errorf("Not enough funds: %v (%v)", acc1, amount)
	}

	// do transfer
	acc1 = acc1 - amount
	acc2 = acc2 + amount
	str1 := strconv.FormatFloat(acc1, 'f', -1, 64)
	str2 := strconv.FormatFloat(acc2, 'f', -1, 64)

	// update accounts
	err = stub.PutState(args[0], []byte(str1))
	if err != nil {
		return "error", fmt.Errorf("Failed to set account: %s", args[0])
	}
	err = stub.PutState(args[1], []byte(str2))
	if err != nil {
		return "error", fmt.Errorf("Failed to set account: %s", args[1])
	}

	return "success", nil
}

func getBalance(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// TODO: checks
	data, err := stub.GetState(args[0])
	if err != nil {
		return "error", fmt.Errorf("No such account: %s", args[0])
	}
	return string(data), nil
}

func createAccount(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// TODO: checks
	err := stub.PutState(args[0], []byte(args[1]))
	if err != nil {
		return "error", fmt.Errorf("No such account: %s", args[0])
	}
	return "success", nil
}

func createAccounts(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// TODO: checks
	for i := 1; i < 100; i++ {
		str := fmt.Sprintf("accounta%d", i)
		stub.PutState(str, []byte("100"))
		str = fmt.Sprintf("accountb%d", i)
		stub.PutState(str, []byte("100"))
	}
	return "success", nil
}

// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(Bank)); err != nil {
		fmt.Printf("Error starting Bank chaincode: %s", err)
	}
}
