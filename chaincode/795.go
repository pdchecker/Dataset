package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (t *SimpleChaincode) moveOutFund(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	// AccountID, targetChannel, amount, currency
	err := checkArgArrayLength(args, 4)
	if err != nil {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 4")
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("AccountID must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, fmt.Errorf("Target channel must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, fmt.Errorf("Amount must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, fmt.Errorf("Currency must be a non-empty string")
	}

	accountID := args[0]
	targetChannel := args[1]
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, fmt.Errorf("Amount must be a numeric string")
	} else if amount < 0 {
		return nil, fmt.Errorf("Amount must be a positive number")
	}
	currency := strings.ToUpper(args[3])

	currTime, err := getTxTimeStampAsTime(ctx)
	if err != nil {
		return nil,fmt.Errorf(err.Error())
	}
	currChannel, err := getChannelName(ctx)
	if err != nil {
		return nil,fmt.Errorf(err.Error())
	}

	// Access Control
	err = verifyIdentity(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// ID generation
	moveOutInFundID := sha256.New()
	moveOutInFundID.Write([]byte(currChannel + targetChannel + currTime.String()))
	moveOutInFundIDString := fmt.Sprintf("%x", moveOutInFundID.Sum(nil))

	// Check if move out request is already created
	moveOutInFundAsBytes, err := ctx.GetStub().GetState(moveOutInFundIDString)
	if err != nil {
		return nil, fmt.Errorf("Failed to get moveOutInFund: " + err.Error())
	} else if moveOutInFundAsBytes != nil {
		errMsg := fmt.Sprintf(
			"Error: moveOutInFund already exists (%s)",
			moveOutInFundIDString)
		return nil, fmt.Errorf(errMsg)
	}

	err = updateAccountBalance(ctx, accountID, currency, amount, true)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	moveOutInFund := &MoveOutInFund{}
	moveOutInFund.ObjectType = moveOutObjectType
	moveOutInFund.RefID = moveOutInFundIDString
	moveOutInFund.AccountID = accountID
	moveOutInFund.ChannelFrom = currChannel
	moveOutInFund.ChannelTo = targetChannel
	moveOutInFund.Amount = amount
	moveOutInFund.Currency = currency
	moveOutInFund.CreateTime = currTime

	moveOutInFundAsBytes, err = json.Marshal(moveOutInFund)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(moveOutInFundIDString, moveOutInFundAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return []byte(moveOutInFundIDString), nil
}
func (t *SimpleChaincode) moveInFund(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	// transientFundID
	err := checkArgArrayLength(args, 1)
	if err != nil {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 1")
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("Transient fund ID must be a non-empty string")
	}

	transientFundID := args[0]

	currTime, err := getTxTimeStampAsTime(ctx)
	if err != nil {
		return nil,fmt.Errorf(err.Error())
	}
	currChannel, err := getChannelName(ctx)
	if err != nil {
		return nil,fmt.Errorf(err.Error())
	}

	queryArgs := [][]byte{[]byte("getState"), []byte(transientFundID)}
	transientFundAsBytes, err := crossChannelQuery(ctx,
		queryArgs,
		fundingChannelName,
		fundingChaincodeName)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	transientFund := &MoveOutInFund{}
	err = json.Unmarshal(transientFundAsBytes, transientFund)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	moveOutInFundAsBytes, err := ctx.GetStub().GetState(transientFund.RefID)
	if err != nil {
		return nil, fmt.Errorf("Failed to get moveOutInFund: " + err.Error())
	} else if moveOutInFundAsBytes != nil {
		errMsg := fmt.Sprintf(
			"Error: moveOutInFund already exists (%s)",
			transientFund.RefID)
		return nil, fmt.Errorf(errMsg)
	}

	// Owner verification
	ownerID := transientFund.AccountID
	err = verifyIdentity(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// Channel verification
	targetChannel := transientFund.ChannelTo
	if targetChannel != currChannel {
		errMsg := fmt.Sprintf(
			"Error: Target channel of transient fund (%s) does not match current channel (%s)",
			targetChannel,
			currChannel)
		return nil, fmt.Errorf(errMsg)
	}

	err = updateAccountBalance(ctx,
		ownerID,
		transientFund.Currency,
		transientFund.Amount,
		false)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// Update fields before storing
	transientFund.ObjectType = moveInObjectType
	transientFund.CreateTime = currTime

	moveOutInFundAsBytes, err = json.Marshal(transientFund)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(transientFundID, moveOutInFundAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return moveOutInFundAsBytes,nil
}