/**
* Copyright 2017 HUAWEI. All Rights Reserved.
*
* SPDX-License-Identifier: Apache-2.0
*
 */

package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//"github.com/hyperledger/fabric/core/chaincode/shim"
//pb "github.com/hyperledger/fabric/protos/peer"

const errorSystem = "{\"code\":300, \"reason\": \"system error: %s\"}"
const errorWrongFormat = "{\"code\":301, \"reason\": \"command format is wrong\"}"
const errorAccountExisting = "{\"code\":302, \"reason\": \"account already exists\"}"
const errorAccountAbnormal = "{\"code\":303, \"reason\": \"abnormal account\"}"
const errorMoneyNotEnough = "{\"code\":304, \"reason\": \"account's money is not enough\"}"

type Simple struct {
}

// The init function of the Chaincode.
// Param: the Chaincode stub
// Return: the result of the initiation of the Chaincode
func (t *Simple) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// nothing to do
	return shim.Success(nil)
}

// The invoke function of the Chaincode.
// Param: the Chaincode stub
// Return: the result of the execution of the function
func (t *Simple) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()

	if function == "open" {
		return t.Open(stub, args)
	}
	if function == "delete" {
		return t.Delete(stub, args)
	}
	if function == "query" {
		return t.Query(stub, args)
	}
	if function == "transfer" {
		return t.Transfer(stub, args)
	}

	return shim.Error(errorWrongFormat)
}

// open an account, should be [open account money]
func (t *Simple) Open(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error(errorWrongFormat)
	}

	account := args[0]
	money, err := stub.GetState(account)
	if money != nil {
		return shim.Error(errorAccountExisting)
	}

	_, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error(errorWrongFormat)
	}

	err = stub.PutState(account, []byte(args[1]))
	if err != nil {
		s := fmt.Sprintf(errorSystem, err.Error())
		return shim.Error(s)
	}

	return shim.Success(nil)
}

// delete an account, should be [delete account]
func (t *Simple) Delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(errorWrongFormat)
	}

	err := stub.DelState(args[0])
	if err != nil {
		s := fmt.Sprintf(errorSystem, err.Error())
		return shim.Error(s)
	}

	return shim.Success(nil)
}

// query current money of the account,should be [query account]
func (t *Simple) Query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error(errorWrongFormat)
	}

	money, err := stub.GetState(args[0])
	if err != nil {
		s := fmt.Sprintf(errorSystem, err.Error())
		return shim.Error(s)
	}

	if money == nil {
		return shim.Error(errorAccountAbnormal)
	}

	//return shim.Success(money)
	return shim.Success(nil)
}

// transfer money from account1 to account2, should be [transfer account1 account2 money]
func (t *Simple) Transfer(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 3 {
		return shim.Error(errorWrongFormat)
	}
	money, err := strconv.Atoi(args[2])
	if err != nil {
		return shim.Error(errorWrongFormat)
	}

	moneyBytes1, err1 := stub.GetState(args[0])
	moneyBytes2, err2 := stub.GetState(args[1])
	if err1 != nil || err2 != nil {
		s := fmt.Sprintf(errorSystem, err.Error())
		return shim.Error(s)
	}
	if moneyBytes1 == nil || moneyBytes2 == nil {
		return shim.Error(errorAccountAbnormal)
	}

	money1, _ := strconv.Atoi(string(moneyBytes1))
	money2, _ := strconv.Atoi(string(moneyBytes2))
	if money1 < money {
		return shim.Error(errorMoneyNotEnough)
	}

	money1 -= money
	money2 += money

	err = stub.PutState(args[0], []byte(strconv.Itoa(money1)))
	if err != nil {
		s := fmt.Sprintf(errorSystem, err.Error())
		return shim.Error(s)
	}

	err = stub.PutState(args[1], []byte(strconv.Itoa(money2)))
	if err != nil {
		s := fmt.Sprintf(errorSystem, err.Error())
		return shim.Error(s)
	}

	return shim.Success(nil)
}

// The main function.
func main() {
	err := shim.Start(new(Simple))
	if err != nil {
		fmt.Printf("Error starting chaincode: %v \n", err)
	}

}
