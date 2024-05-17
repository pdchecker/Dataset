package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (t *SimpleChaincode) createTransientFund(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	// moveOutInFundID
	err := checkArgArrayLength(args, 2)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("MoveOutInFund ID must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, fmt.Errorf("Source channel must be a non-empty string")
	}

	moveOutInFundID := args[0]
	sourceChannel := args[1]

	currTime, err := getTxTimeStampAsTime(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	queryArgs := [][]byte{[]byte("getState"), []byte(moveOutInFundID)}
	moveOutInFundAsBytes, err := crossChannelQuery(ctx,
		queryArgs,
		sourceChannel,
		bilateralChaincodeName)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	moveOutInFund := &TransientFund{}
	err = json.Unmarshal(moveOutInFundAsBytes, moveOutInFund)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	transientFundAsBytes, err := ctx.GetStub().GetState(moveOutInFund.RefID)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	} else if transientFundAsBytes != nil {
		errMsg := fmt.Sprintf(
			"Error: Transient fund already exists (%s)",
			moveOutInFund.RefID)
		return nil, fmt.Errorf(errMsg)
	}

	// Owner verification
	ownerID := moveOutInFund.AccountID
	err = verifyIdentity(ctx, ownerID)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// Update fields before storing
	moveOutInFund.ObjectType = transientFundObjectType
	moveOutInFund.CreateTime = currTime

	transientFundAsBytes, err = json.Marshal(moveOutInFund)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(moveOutInFundID, transientFundAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return transientFundAsBytes, nil
}
