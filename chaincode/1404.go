/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"fmt"
	"hyperledger_dapp/controller"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
)

// Chaincode is the definitaion of the chaincode structure
type ERC20Chaincode struct {
	Controller *controller.Controller
}

// NewChaincode is construtor function for Chaincode
func NewChaincode() *ERC20Chaincode {
	controller := controller.NewContoller()
	return &ERC20Chaincode{Controller: controller}
}

// Init is called when the chaincode is instantiated by the blockchain network.
func (cc *ERC20Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	_, params := stub.GetFunctionAndParameters()
	fmt.Println("Init called with params: ", params)

	return cc.Controller.Init(stub, params)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *ERC20Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fnc, params := stub.GetFunctionAndParameters()

	switch fnc {
	case "init":
		return cc.Controller.Init(stub, params)
	case "totalSupply":
		return cc.Controller.TotalSupply(stub, params)
	case "balanceOf":
		return cc.Controller.BalanceOf(stub, params)
	case "transfer":
		return cc.Controller.Transfer(stub, params)
	case "allowance":
		return cc.Controller.Allowance(stub, params)
	case "approve":
		return cc.Controller.Approve(stub, params)
	case "transferFrom":
		return cc.Controller.TransferFrom(stub, params)
	case "increaseAllowance":
		return cc.Controller.IncreaseAllowance(stub, params)
	case "decreaseAllowance":
		return cc.Controller.DecreaseAllowance(stub, params)
	case "approvalList":
		return cc.Controller.ApprovalList(stub, params)
	case "transferOtherToken":
		return cc.Controller.TransferOtherToken(stub, params)
	case "mint":
		return cc.Controller.Mint(stub, params)
	case "burn":
		return cc.Controller.Burn(stub, params)
	default:
		return sc.Response{Status: 404, Message: "404 Not Found", Payload: nil}
	}
}
