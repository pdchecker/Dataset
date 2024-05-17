package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/util"
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
	switch fn {
	case "sendAmount":
		result, err = sendAmount(stub, args)
	case "getBalance":
		result, err = getBalance(stub, args)
	case "getStatus":
		result, err = getStatus(stub, args)
	case "receiveAmount":
		result, err = receiveAmount(stub, args)
	case "complete":
		result, err = complete(stub, args)
	case "createAccount":
		result, err = createAccount(stub, args)
	case "createAccounts":
		result, err = createAccounts(stub, args)
	default:
		return shim.Error("no such method")
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}

func sendAmount(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// check args
	if len(args) != 5 {
		return "error", fmt.Errorf("Expecting 5 arguments: account1Id, account2Id, amount, channel1, channel2")
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

	// check
	if acc1 < amount {
		return "error", fmt.Errorf("Not enough funds: %v (%v)", acc1, amount)
	}

	// do partial transfer (pending)
	acc1 = acc1 - amount
	str1 := strconv.FormatFloat(acc1, 'f', -1, 64)

	err = stub.PutState(args[0], []byte(str1))
	if err != nil {
		return "error", fmt.Errorf("Failed to set account: %s", args[0])
	}

	// update states
	txid := stub.GetTxID()
	key := fmt.Sprintf("status_%s", txid)
	stub.PutState(key, []byte("pending"))
	key = fmt.Sprintf("details_%s", txid)
	val := fmt.Sprintf("%s_%s_%s_%s_%s", args[0], args[1], args[2], args[3], args[4])
	stub.PutState(key, []byte(val))
	key = fmt.Sprintf("amount_%s", txid)
	stub.PutState(key, []byte(args[2]))

	return txid, nil
}

func receiveAmount(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// check args
	if len(args) != 6 {
		return "error", fmt.Errorf("Expecting 6 arguments: account1Id, account2Id, amount, channel1, channel2, txid.")
	}
	txid := args[5]
	key := fmt.Sprintf("status_%s", txid)

	// TODO: check locally

	// convert string to float
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return "error", fmt.Errorf("Argument 3 should be a float: %s", args[2])
	}

	// get balance of account2
	data, err := stub.GetState(args[1])
	if err != nil {
		return "error", fmt.Errorf("No such account: %s", args[1])
	}
	acc2, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return "error", fmt.Errorf("Invalid account balance: %s", string(data))
	}

	// check status in the other channel (needs to be pending)
	invokeArgs := util.ToChaincodeArgs("getStatus", txid)
	response := stub.InvokeChaincode("bankcross", invokeArgs, args[3])
	status := string(response.Payload)
	if status != "pending" {
		stub.PutState(key, []byte("failed"))
		return "error", fmt.Errorf("Transaction status is not pending: %s", status)
	}

	// do transfer
	// TODO: extra checks
	acc2 = acc2 + amount
	str2 := strconv.FormatFloat(acc2, 'f', -1, 64)
	err = stub.PutState(args[1], []byte(str2))
	if err != nil {
		stub.PutState(key, []byte("failed"))
		return "error", fmt.Errorf("Failed to set account: %s", args[1])
	}
	stub.PutState(key, []byte("done"))

	return "success", nil
}

func rollBack(stub shim.ChaincodeStubInterface, account string, amount string) (string, error) {
	// check args
	data, err := stub.GetState(account)
	if err != nil {
		return "error", fmt.Errorf("No such account: %s", account)
	}
	acc1, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return "error", fmt.Errorf("Invalid account balance: %s", string(data))
	}

	val, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		return "error", fmt.Errorf("Argument 3 should be a float: %s", amount)
	}

	acc1 = acc1 + val
	str1 := strconv.FormatFloat(acc1, 'f', -1, 64)

	err = stub.PutState(account, []byte(str1))
	if err != nil {
		return "error", fmt.Errorf("Failed to set account: %s", account)
	}

	return "success", nil
}

func complete(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// check args
	if len(args) != 6 {
		return "error", fmt.Errorf("Expecting 6 arguments: account1Id, account2Id, amount, channel1, channel2, txid.")
	}
	txid := args[5]

	// check status in this channel
	key := fmt.Sprintf("status_%s", txid)
	data, _ := stub.GetState(key)
	statusHere := string(data)
	if statusHere == "done" || statusHere == "failed" {
		return "already done", nil
	}

	// check status in the other channel
	invokeArgs := util.ToChaincodeArgs("getStatus", txid)
	response := stub.InvokeChaincode("bankcross", invokeArgs, args[4])
	statusThere := string(response.Payload)

	// update states
	if statusThere == "failed" {
		stub.PutState(key, []byte("failed"))
		rollBack(stub, args[0], args[2])
	} else {
		stub.PutState(key, []byte("done"))
	}

	return "success", nil
}

func getStatus(stub shim.ChaincodeStubInterface, args []string) (string, error) {
	// TODO: checks
	key := fmt.Sprintf("status_%s", args[0])
	data, err := stub.GetState(key)
	if err != nil {
		return "error", fmt.Errorf("No such account: %s", args[0])
	}
	return string(data), nil
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
