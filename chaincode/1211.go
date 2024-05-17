package main

import (
	"encoding/json"
	"fmt"
	"hyperledger_dapp/model"
	"hyperledger_dapp/repository"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

// TotalSupply is query function
// params is tokenName
// Returns the amount of token in existence
func (cc *Controller) TotalSupply(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// a number of params must be one
	if len(params) != 1 {
		return shim.Error("TotalSupply only 1 params")
	}

	tokenName := params[0]

	// get erc20 totalsupply
	totalSupply, err := repository.GetERC20TotalSupply(stub, tokenName)
	if err != nil {
		return shim.Error(err.Error())
	}

	// convert metadata to bytes
	totalsupplyBytes, err := json.Marshal(totalSupply)
	if err != nil {
		errMsg := fmt.Sprintf("failed to marshal from TotalSupply, error : %s", err.Error())
		return shim.Error(errMsg)
	}

	fmt.Println(tokenName + "'s totalsupply is " + string(totalsupplyBytes))

	return shim.Success(totalsupplyBytes)
}

// BalanceOf is query function
// params is address
// Returns the amount of tokens owned by address
func (cc *Controller) BalanceOf(stub shim.ChaincodeStubInterface, params []string) sc.Response {
	// a number of params must be one
	if len(params) != 1 {
		return shim.Error("BalanceOf only 1 params")
	}

	address := params[0]

	// get balance
	amountBytes, err := stub.GetState(address)
	if err != nil {
		errMsg := fmt.Sprintf("failed to getstate from balanceOf, error : %s", err.Error())
		return shim.Error(errMsg)
	}

	fmt.Println(address + "'s totalsupply is " + string(amountBytes))

	return shim.Success(amountBytes)
}

// Allowance is query fnc
// params - owner's address, spender's address
// return - the remaining amount of token to invoke (transferFrom)
func (cc *Controller) Allowance(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// check the number of params is 2
	if len(params) != 2 {
		return shim.Error("Allowance only 2 params")
	}

	ownerAddress, spenderAddress := params[0], params[1]

	// create composite key
	approvalKey, err := stub.CreateCompositeKey("approval", []string{ownerAddress, spenderAddress})
	if err != nil {
		return shim.Error("failed CreateCompositeKey for approval")
	}

	// getstate(amount)
	approvalKeyBytes, err := stub.GetState(approvalKey)
	if err != nil {
		return shim.Error("Failed to GetState")
	}

	if approvalKeyBytes == nil {
		approvalKeyBytes = []byte("0")
	}

	return shim.Success(approvalKeyBytes)
}

// ApprovalList is query fnc
// params - owner's addresss
// return - approvalList by owner
func (cc *Controller) ApprovalList(stub shim.ChaincodeStubInterface, params []string) sc.Response {

	// check the number of params is 1
	if len(params) != 1 {
		return shim.Error("allowance only 2 params")
	}

	ownerAddress := params[0]

	// get all approval (format is iterator)
	approvalIterator, err := stub.GetStateByPartialCompositeKey("approval", []string{ownerAddress})
	if err != nil {
		return shim.Error("failed to GetStateByPartialCompositeKey for approval iterator error :" + err.Error())
	}

	// make slice for return value
	approvalSlice := []model.Approval{}

	// iterator
	defer approvalIterator.Close()
	if approvalIterator.HasNext() {
		for approvalIterator.HasNext() {
			approvalKV, _ := approvalIterator.Next()
			approvalKey := approvalKV.GetKey()
			approvalValue := approvalKV.GetValue()

			// get sppender address
			_, address, err := stub.SplitCompositeKey(approvalKey)
			if err != nil {
				return shim.Error("failed to SplitCompositeKey, error :" + err.Error())
			}

			spenderAddress := address[1]

			// get amount
			amount, err := strconv.Atoi(string(approvalValue))
			if err != nil {
				return shim.Error("failed to get amount, error : " + err.Error())
			}

			// add approval result
			approvalSlice = append(approvalSlice,
				*model.NewApproval(spenderAddress, ownerAddress, amount))
		}
	}

	// marshal data
	response, err := json.Marshal(approvalSlice)
	if err != nil {
		return shim.Error("failed to Marshal approvalSlice, error : " + err.Error())
	}

	return shim.Success(response)
}
