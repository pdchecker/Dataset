package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"strings"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/shopspring/decimal"
	"github.com/tokenERC20/base"
	"github.com/tokenERC20/util"
)

// Define key names for options
// const totalSupplyKey = "totalSupply"

// Define objectType names for prefix
const allowancePrefix = "allowance"

// name of the token
const namePrefix = "namePrefix"

// symbol of the token
const symbolPrefix = "SYB"

//decimal of the token
const decimalPrefix = "decimal"

//owner of the contract
const ownerPrefix = "owner"

// SmartContract provides functions for transferring tokens between accounts
type SmartContract struct {
	contractapi.Contract
}

// event provides an organized struct for emitting events
type Event struct {
	From  string `json:"From"`
	To    string `json:"To"`
	Value string `json:"Value"`
}

// respnse struct
type Response struct {
	Success   bool                 `json:"Success"`
	Func      *Fcn                 `json:"Func,omitempty"`
	TxID      string               `json:"TxID"`
	Timestamp *timestamp.Timestamp `json:"Timestamp"`
}

// function response struct
type Fcn struct {
	Minter string `json:"Minter,omitempty"`
	From   string `json:"From,omitempty"`
	To     string `json:"To,omitempty"`
	Value  string `json:"Value,omitempty"`
	Total  string `json:"Total,omitempty"`
}

// info response struct
type Info struct {
	Owner       string `json:"Owner"`
	TokenName   string `json:"TokenName"`
	Symbol      string `json:"Symbol"`
	Decimal     string `json:"Decimal"`
	TotalSupply string `json:"TotalSupply"`
}

// Tx Details struct
type DetailsTx struct {
	From   string `json:"From"`
	To     string `json:"To"`
	Action string `json:"Action"`
	Value  string `json:"Value"`
}

/*
	Init declares chaincode details

	@param {Context} ctx the transaction context
	@param {string} contract owner address

	Return success interface or error
*/
func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface, owner string) (interface{}, error) {

	exists, err := ctx.GetStub().GetState(ownerPrefix)
	if err != nil || exists != nil {
		return nil, fmt.Errorf("Contract already initalized by %s error:%s", string(exists), err)
	}
	err = ctx.GetStub().PutState(namePrefix, []byte("BANANACOIN"))
	err = ctx.GetStub().PutState(symbolPrefix, []byte("BNC"))
	err = ctx.GetStub().PutState(decimalPrefix, []byte(strconv.Itoa(18)))
	err = ctx.GetStub().PutState(ownerPrefix, []byte(owner))
	if err != nil {
		return nil, fmt.Errorf("error setting values %s", err)
	}
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	res := &Response{
		Success:   true,
		TxID:      ctx.GetStub().GetTxID(),
		Timestamp: txTime,
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return string(content), nil
}

/*

 */

/*
	Mint creates new tokens and adds them to contract owners account balance
	 // this function triggers a Transfer event

	@param {Context} ctx the transaction context
	@param {string} the contract owner address

	Return success interface or error
*/
func (s *SmartContract) MintAndTransfer(ctx contractapi.TransactionContextInterface, user, amount, msg, signature string, swapId string) (interface{}, error) {

	var IncrAmount decimal.Decimal

	// retrieve contract owner address
	minterByte, err := ctx.GetStub().GetState(ownerPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed while getting minterAddress %s", err)
	} else if minterByte == nil {
		return nil, fmt.Errorf("contract is not initialized yet")
	}
	minter := string(minterByte)
	// check if contract caller is contract owner
	if verify, err := util.VerifyMsgAddr(minter, signature, msg); !verify || err != nil {
		return nil, err
	}

	if IncrAmount, err = util.ParsePositive(amount); err != nil {
		return nil, fmt.Errorf("amount must be a positive integer")
	}

	bigAmount := big.NewInt(0)
	if _, ok := bigAmount.SetString(amount, 10); !ok {
		return nil, fmt.Errorf("error parsing amount")
	}

	hashedMsg, err := util.GetMsgForSign(user, swapId, bigAmount)
	if err != nil {
		return nil, err
	}

	if c := strings.Compare(hashedMsg, msg); c != 0 {
		return nil, fmt.Errorf("integrity check failed")
	}

	if err = base.AddToken(ctx, amount, user); err != nil {
		return nil, err
	}

	// Emit the Transfer event
	transferEvent := &Event{
		From: "0x0", To: minter, Value: IncrAmount.String()}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to set event: %v", err)
	}

	tokenName, err := ctx.GetStub().GetState(namePrefix)
	if err != nil {
		return nil, err
	}
	details := &DetailsTx{
		From:   minter,
		To:     string(tokenName),
		Action: "Mint",
		Value:  IncrAmount.String(),
	}

	dtl, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), dtl)
	if err != nil {
		return nil, err
	}

	txTime, _ := ctx.GetStub().GetTxTimestamp()
	mintResp := &Fcn{
		Minter: user,
		Value:  IncrAmount.String(),
	}
	res := &Response{
		Success:   true,
		Func:      mintResp,
		TxID:      ctx.GetStub().GetTxID(),
		Timestamp: txTime,
	}
	context, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return string(context), nil
}

/*
	Burn redeems tokens the contract owner's account balance
	// this function triggers a Transfer event

	@param {Context} ctx the transaction context
	@param {string} the contract owner address

	Return success interface or error
*/
func (s *SmartContract) BurnFrom(ctx contractapi.TransactionContextInterface, user, amount, msg, signature string, swapId string) (interface{}, error) {

	var BurningAmount decimal.Decimal
	// Check minter authorization - this sample assumes Org1 is the central banker with privilege to burn new tokens
	// retrieve contract owner address
	minterByte, err := ctx.GetStub().GetState(ownerPrefix)
	if err != nil {
		return nil, fmt.Errorf("failed while getting minterAddress %s", err)
	} else if minterByte == nil {
		return nil, fmt.Errorf("contract is not initialized yet")
	}
	minter := string(minterByte)

	if verify, err := util.VerifyMsgAddr(minter, signature, msg); !verify || err != nil {
		return nil, err
	}

	if BurningAmount, err = util.ParsePositive(amount); err != nil {
		return nil, fmt.Errorf("burn amount must be integer string")
	}

	bigAmount := big.NewInt(0)
	if _, ok := bigAmount.SetString(amount, 10); !ok {
		return nil, fmt.Errorf("error parsing amount")
	}

	hashedMsg, err := util.GetMsgForSign(user, swapId, bigAmount)
	if err != nil {
		return nil, err
	}

	if c := strings.Compare(hashedMsg, msg); c != 0 {
		return nil, fmt.Errorf("integrity check failed")
	}

	if err = base.SubstractToken(ctx, amount, user); err != nil {
		return nil, err
	}

	// Emit the Transfer event
	transferEvent := &Event{From: minter, To: "0x0", Value: BurningAmount.String()}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to set event: %v", err)
	}

	tokenName, err := ctx.GetStub().GetState(namePrefix)
	if err != nil {
		return nil, err
	}
	details := &DetailsTx{
		From:   minter,
		To:     string(tokenName),
		Action: "Burn",
		Value:  BurningAmount.String(),
	}

	dtl, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), dtl)
	if err != nil {
		return nil, err
	}

	txTime, _ := ctx.GetStub().GetTxTimestamp()
	mintResp := &Fcn{
		Minter: minter,
		Value:  amount,
	}
	resp := &Response{
		Success:   true,
		Func:      mintResp,
		TxID:      ctx.GetStub().GetTxID(),
		Timestamp: txTime,
	}
	content, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json %s", err)
	}
	return string(content), nil
}

/*
   Transfer transfers tokens from client account to recipient account
   // recipient account must be a valid clientID
   // this function triggers a Transfer event

   @param {Context} ctx the transcation context
   @param {string} client account address
   @param {string} recipient account address

   Returns success interface or error
*/
func (s *SmartContract) Transfer(ctx contractapi.TransactionContextInterface, from, recipient, amount, msg, signature string) (interface{}, error) {
	var decimalAmount decimal.Decimal
	var err error
	// // verify user wallet

	if verify, err := util.VerifyMsgAddr(from, signature, msg); !verify || err != nil {
		return nil, err
	}

	// from to should not be same wallet address
	if from == recipient {
		return nil, fmt.Errorf("from address and to address must be different values")
	}
	// verify for postive integer
	if decimalAmount, err = util.ParsePositive(amount); err != nil {
		return nil, fmt.Errorf("%s is not positive integer", amount)
	}

	bigAmount := big.NewInt(0)
	if _, ok := bigAmount.SetString(amount, 10); !ok {
		return nil, fmt.Errorf("error parsing amount")
	}

	hashedMsg, err := util.GetMsgForSignTransfer(recipient, bigAmount)
	if err != nil {
		return nil, err
	}

	if c := strings.Compare(hashedMsg, msg); c != 0 {
		return nil, fmt.Errorf("integrity check failed")
	}
	// move token between wallets
	err = base.MoveToken(ctx, from, recipient, amount)
	if err != nil {
		return nil, fmt.Errorf("failed to transfer: %v", err)
	}

	// Emit the Transfer event
	transferEvent := &Event{From: from, To: recipient, Value: decimalAmount.String()}
	transferEventJSON, err := json.Marshal(transferEvent)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain JSON encoding: %v", err)
	}
	err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to set event: %v", err)
	}
	txTime, _ := ctx.GetStub().GetTxTimestamp()

	details := &DetailsTx{
		From:   from,
		To:     recipient,
		Action: "Transfer",
		Value:  decimalAmount.String(),
	}

	dtl, err := json.Marshal(details)
	if err != nil {
		return nil, err
	}
	err = ctx.GetStub().PutState(ctx.GetStub().GetTxID(), dtl)
	if err != nil {
		return nil, err
	}

	mintResp := &Fcn{
		From:  from,
		To:    recipient,
		Value: decimalAmount.String(),
	}
	resp := &Response{
		Success:   true,
		Func:      mintResp,
		TxID:      ctx.GetStub().GetTxID(),
		Timestamp: txTime,
	}
	content, err := json.Marshal(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal json %s", err)
	}

	return string(content), nil
}

// BalanceOf returns the balance of the given account
func (s *SmartContract) BalanceOf(ctx contractapi.TransactionContextInterface, account string) (string, error) {
	var BalanceDecimal decimal.Decimal
	balanceBytes, err := ctx.GetStub().GetState(account)
	if err != nil {
		return "0", fmt.Errorf("failed to read from world state: %v", err)
	}
	if balanceBytes == nil {
		return "0", nil
	}
	BalanceDecimal, _ = decimal.NewFromString(string(balanceBytes))

	return BalanceDecimal.String(), nil
}

/*
	ClientAccountBalance returns the balance of the requesting client's account

	@oaram {Context} ctx the transaction context

	Returns int value of the balance or error

*/

func (s *SmartContract) ClientAccountBalance(ctx contractapi.TransactionContextInterface) (int, error) {

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

	balance, _ := strconv.Atoi(string(balanceBytes)) // Error handling not needed since Itoa() was used when setting the account balance, guaranteeing it was an integer.

	return balance, nil
}

/*
	ClientAccountID returns the id of the requesting client's account
	// in this implementation, the client account ID is the clientId itself
	// users can use this function to get their own account id, which they can then give to otherss as the payment address

	@param {Context} ctx the transaction context

	Returns string user adress or error
*/
func (s *SmartContract) ClientAccountID(ctx contractapi.TransactionContextInterface) (string, error) {

	// Get ID of submitting client identity
	clientAccountID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("failed to get client id: %v", err)
	}

	return clientAccountID, nil
}

func (s *SmartContract) GetDetails(ctx contractapi.TransactionContextInterface) (interface{}, error) {

	deployer, err := ctx.GetStub().GetState(ownerPrefix)
	tokenName, err := ctx.GetStub().GetState(namePrefix)
	symbol, err := ctx.GetStub().GetState(symbolPrefix)
	decimalType, err := ctx.GetStub().GetState(decimalPrefix)

	if err != nil {
		return nil, err
	}
	if decimalType == nil || tokenName == nil || symbol == nil || deployer == nil {
		return nil, fmt.Errorf("Init is not declared %s,%s,%s", string(decimalType), string(tokenName), string(deployer))
	}

	res := &Info{
		Owner:     string(deployer),
		TokenName: string(tokenName),
		Symbol:    string(symbol),
		Decimal:   string(decimalType),
	}
	content, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return string(content), nil
}

/*
	Approve allows the spender to withdraw from the calling client's token account
	// the spender can withdraw multiple times if neccessary, up to the value amount
	// this function triggers an Approval event

	@param {Context} ctx the transaction context
	@param {string} spender the spender address
	@param {int} value the amount to approve

	Return success interface or error
*/
func (s *SmartContract) Approve(ctx contractapi.TransactionContextInterface, spender string, value int) error {

	// Get ID of submitting client identity
	owner, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	// Create allowanceKey
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
	}

	// Update the state of the smart contract by adding the allowanceKey and value
	err = ctx.GetStub().PutState(allowanceKey, []byte(strconv.Itoa(value)))
	if err != nil {
		return fmt.Errorf("failed to update state of smart contract for key %s: %v", allowanceKey, err)
	}

	// Emit the Approval event
	// approvalEvent := &Event{From: owner, To: spender, Value: value}
	// approvalEventJSON, err := json.Marshal(approvalEvent)
	// if err != nil {
	// 	return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	// }
	// err = ctx.GetStub().SetEvent("Approval", approvalEventJSON)
	// if err != nil {
	// 	return fmt.Errorf("failed to set event: %v", err)
	// }

	log.Printf("client %s approved a withdrawal allowance of %d for spender %s", owner, value, spender)

	return nil
}

/*
	Allowance returns the amount still available for the spender to withdraw from the owner

	@param {Context} ctx the transaction context
	@param {string} owner the owner address
	@param {spender} spender the spender address

	Returns int amount or error
*/
func (s *SmartContract) Allowance(ctx contractapi.TransactionContextInterface, owner string, spender string) (int, error) {

	// Create allowanceKey
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{owner, spender})
	if err != nil {
		return 0, fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
	}

	// Read the allowance amount from the world state
	allowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
	if err != nil {
		return 0, fmt.Errorf("failed to read allowance for %s from world state: %v", allowanceKey, err)
	}

	var allowance int

	// If no current allowance, set allowance to 0
	if allowanceBytes == nil {
		allowance = 0
	} else {
		allowance, _ = strconv.Atoi(string(allowanceBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.
	}

	log.Printf("The allowance left for spender %s to withdraw from owner %s: %d", spender, owner, allowance)

	return allowance, nil
}

/*
	TransferFrom transfers the value amount from the "from" address to the "to" address
	// this function triggers a Transfer event

	@param {string} from the from client address
	@param {string} to the to client address
	@param {int} value the amount to transfer

	Returns success interface or error
*/
func (s *SmartContract) TransferFrom(ctx contractapi.TransactionContextInterface, from string, to string, value string) error {

	// Get ID of submitting client identity
	spender, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return fmt.Errorf("failed to get client id: %v", err)
	}

	// Create allowanceKey
	allowanceKey, err := ctx.GetStub().CreateCompositeKey(allowancePrefix, []string{from, spender})
	if err != nil {
		return fmt.Errorf("failed to create the composite key for prefix %s: %v", allowancePrefix, err)
	}

	// Retrieve the allowance of the spender
	currentAllowanceBytes, err := ctx.GetStub().GetState(allowanceKey)
	if err != nil {
		return fmt.Errorf("failed to retrieve the allowance for %s from world state: %v", allowanceKey, err)
	}

	currentAllowanceDecimal, _ := decimal.NewFromString(string(currentAllowanceBytes)) // Error handling not needed since Itoa() was used when setting the totalSupply, guaranteeing it was an integer.

	valueDecimal, _ := decimal.NewFromString(value)
	// Check if transferred value is less than allowance
	if currentAllowanceDecimal.Cmp(valueDecimal) < 0 {
		return fmt.Errorf("spender does not have enough allowance for transfer")
	}

	// Initiate the transfer
	err = base.MoveToken(ctx, from, to, value)
	if err != nil {
		return fmt.Errorf("failed to transfer: %v", err)
	}

	// Decrease the allowance
	updatedAllowance := currentAllowanceDecimal.Sub(valueDecimal)
	err = ctx.GetStub().PutState(allowanceKey, []byte(updatedAllowance.String()))
	if err != nil {
		return err
	}

	// Emit the Transfer event
	// transferEvent := &Event{From: from, To: to, Value: value}
	// transferEventJSON, err := json.Marshal(transferEvent)
	// if err != nil {
	// 	return fmt.Errorf("failed to obtain JSON encoding: %v", err)
	// }
	// err = ctx.GetStub().SetEvent("Transfer", transferEventJSON)
	// if err != nil {
	// 	return fmt.Errorf("failed to set event: %v", err)
	// }

	return nil
}

func AddressHelper(encodedAdr, client string) (bool, error) {

	decodedAdr, err := base64.StdEncoding.DecodeString(encodedAdr)
	if err != nil {
		return false, err
	} else if strings.Contains(string(decodedAdr), client) {
		return true, nil
	}
	return false, nil

}
