package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (t *SimpleChaincode) createDestroyFund(
	ctx contractapi.TransactionContextInterface,
	args []string,
	docType string) ([]byte, error) {

	// AccountID, Currency, Amount
	err := checkArgArrayLength(args, 3)
	if err != nil {
		return nil, fmt.Errorf("Incorrect number of arguments. Expecting 3")
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("AccountID must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, fmt.Errorf("Currency must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, fmt.Errorf("Amount must be a non-empty string")
	}

	accountID := args[0]
	currency := strings.ToUpper(args[1])
	amount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return nil, fmt.Errorf("Amount must be a numeric string")
	} else if amount < 0 {
		return nil, fmt.Errorf("Amount must be a positive number")
	}
	currTime, err := getTxTimeStampAsTime(ctx)
	if err != nil {
		return nil, fmt.Errorf("Error getting transaction timestamp: %s", err.Error())
	}

	// Access Control
	err = verifyIdentity(ctx, regulatorName)
	if err != nil {
		return nil, fmt.Errorf("Error verifying identity: %s", err.Error())
	}

	if docType == pledgeObjectType || docType == nettingAddObjectType {
		err = updateAccountBalance(ctx, accountID, currency, amount, false)
	} else if docType == redeemObjectType || docType == nettingSubtractObjectType {
		err = updateAccountBalance(ctx, accountID, currency, amount, true)
	} else {
		errMsg := fmt.Sprintf("Error: Unrecognised docType (%s)", docType)
		return nil, fmt.Errorf(errMsg)
	}
	if err != nil {
		return nil, fmt.Errorf("Error updating account balance: %s", err.Error())
	}

	txID := sha256.New()
	txID.Write([]byte(accountID + currTime.String()))
	txIDString := fmt.Sprintf("%x", txID.Sum(nil))

	pledgeRedeemFund := PledgeRedeemFund{docType,
		txIDString,
		accountID,
		amount,
		currency,
		currTime}
	pledgeRedeemFundAsBytes, err := json.Marshal(pledgeRedeemFund)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	err = ctx.GetStub().PutState(txIDString, pledgeRedeemFundAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return pledgeRedeemFundAsBytes, nil
}

func validateTransaction(
	ctx contractapi.TransactionContextInterface,
	args []string) (QueuedTransaction, bool, error) {

	var err error
	newTx := QueuedTransaction{}

	err = checkArgArrayLength(args, 6)
	if err != nil {
		return newTx, false, err
	}
	if len(args[0]) <= 0 {
		return newTx, false, fmt.Errorf("sender must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return newTx, false, fmt.Errorf("receiver must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return newTx, false, fmt.Errorf("Priority must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return newTx, false, fmt.Errorf("Amount must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return newTx, false, fmt.Errorf("Currency must be a non-empty string")
	}
	if len(args[5]) <= 0 {
		return newTx, false, fmt.Errorf("isPutToQueue flag must be a non-empty string")
	}

	sender := args[0]
	receiver := args[1]
	priority, err := strconv.Atoi(args[2])
	if err != nil {
		return newTx, false, errors.New("priority must be a numeric string")
	}
	amount, err := strconv.ParseFloat(args[3], 64)
	if err != nil {
		return newTx, false, fmt.Errorf("Amount must be a numeric string")
	} else if amount < 0 {
		return newTx, false, fmt.Errorf("Amount must be a positive value")
	}
	currency := strings.ToUpper(args[4])
	isPutToQueue, err := strconv.ParseBool(strings.ToLower(args[5]))
	if err != nil {
		return newTx, false, fmt.Errorf("isPutToQueue must be a boolean string")
	}

	currTime, err := getTxTimeStampAsTime(ctx)
	if err != nil {
		return newTx, false, err
	}

	// Access Control
	err = verifyIdentity(ctx, sender)
	if err != nil {
		return newTx, false, err
	}

	txID := sha256.New()
	txID.Write([]byte(sender + receiver + currTime.String()))
	txIDString := fmt.Sprintf("%x", txID.Sum(nil))

	newTx.ObjectType = queuedTxObjectType
	newTx.RefID = txIDString
	newTx.Sender = sender
	newTx.Receiver = receiver
	newTx.Priority = priority
	newTx.Nettable = true
	newTx.Amount = amount
	newTx.Currency = currency
	newTx.IsFrozen = false
	newTx.CreateTime = currTime
	newTx.UpdateTime = currTime

	return newTx, isPutToQueue, nil
}

func (t *SimpleChaincode) fundTransfer(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	//  sender, receiver, priority, amount, currency, isPutToQueue
	newTx, isPutToQueue, err := validateTransaction(ctx, args)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	senderAccount, err := getAccountStructFromID(ctx, newTx.Sender)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	receiverAccount, err := getAccountStructFromID(ctx, newTx.Receiver)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	isParticipatingInNetting, err := checkMLNettingParticipation(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	isAccountFrozen := false
	if isParticipatingInNetting {
		isAccountFrozen = true
	}

	queryString := fmt.Sprintf(
		`{"selector":{
			"docType":"%s",
			"status":"ACTIVE",
			"sender":"%s",
			"priority":{"$gte":%d}
		}}`,
		queuedTxObjectType,
		newTx.Sender,
		newTx.Priority)
	outgoingQueueArr, err := getSortedQueues(ctx, queryString)

	queryString = fmt.Sprintf(
		`{"selector":{
			"docType":"%s",
			"status":"ACTIVE",
			"receiver":"%s"
		}}`,
		queuedTxObjectType,
		newTx.Sender,
		newTx.Priority)
	incomingQueueArr, err := getSortedQueues(ctx, queryString)

	isCompleted := false
	isNetted := false
	if isBilateralNetting &&
		len(incomingQueueArr) > 0 &&
		!isAccountFrozen { // Try Bilateral Netting

		queryString = fmt.Sprintf(
			`{"selector":{
				"docType":"%s",
				"status":"ACTIVE",
				"sender":"%s"
			}}`,
			queuedTxObjectType,
			newTx.Sender)
		fullOutgoingQueueArr, err := getSortedQueues(ctx, queryString)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		fullOutgoingQueueArr = append(fullOutgoingQueueArr, newTx)

		isNetted, err = tryBilateralNetting(ctx,
			senderAccount,
			receiverAccount,
			fullOutgoingQueueArr,
			incomingQueueArr)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

	} else if senderAccount.Amount >= newTx.Amount &&
		!isPutToQueue &&
		!isAccountFrozen &&
		len(outgoingQueueArr) == 0 { // Check for sufficient liquidity

		err = updateAccountBalance(ctx,
			newTx.Sender,
			newTx.Currency,
			newTx.Amount,
			true)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		err = updateAccountBalance(ctx,
			newTx.Receiver,
			newTx.Currency,
			newTx.Amount,
			false)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		isCompleted = true
	}

	var respMsg string
	if isNetted {
		respMsg = "Success: Bilateral netting is completed"
	} else if isCompleted {
		respMsg = "Success: Transaction is completed"
		err = moveQueuedTxStructToCompleted(ctx, newTx, "SETTLED")
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}

	} else {
		respMsg = "Success: Transaction is Queued"
		newTx.Status = "ACTIVE"

		if isAccountFrozen {
			respMsg = "Success: Transaction is queued and frozen"
			newTx.IsFrozen = true
		}
		newTxAsBytes, err := json.Marshal(newTx)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
		err = ctx.GetStub().PutState(newTx.RefID, newTxAsBytes)
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		}
	}

	respPayload := fmt.Sprintf(
		`{"msg": "%s", "refId": "%s"}`,
		respMsg,
		newTx.RefID)

	return []byte(respPayload), nil
}

func tryBilateralNetting(
	ctx contractapi.TransactionContextInterface,
	senderAccount *Account,
	receiverAccount *Account,
	outgoingQueueArr []QueuedTransaction,
	incomingQueueArr []QueuedTransaction) (bool, error) {

	isCompleted := false
	totalOutgoingAmt, err := getTotalQueuedAmount(outgoingQueueArr)

	receiverBalance := receiverAccount.Amount
	totalNettingAmt := totalOutgoingAmt + receiverBalance

	var nettableQueueArray []QueuedTransaction
	isNettingPossible := false
	for _, queueElement := range incomingQueueArr {
		if totalNettingAmt >= queueElement.Amount {
			nettableQueueArray = append(nettableQueueArray, queueElement)
			totalNettingAmt -= queueElement.Amount
			isNettingPossible = true
		} else {
			break
		}
	}

	if isNettingPossible {
		extraReceiverBalance := totalNettingAmt - receiverBalance
		nettableQueueArray = append(nettableQueueArray, outgoingQueueArr...)

		if extraReceiverBalance == 0 {
			for _, queueElement := range nettableQueueArray {
				err = moveQueuedTxStructToCompleted(ctx,
					queueElement,
					"SETTLED")
			}
			isCompleted = true

		} else if extraReceiverBalance > 0 &&
			extraReceiverBalance <= senderAccount.Amount {

			for _, queueElement := range nettableQueueArray {

				err = moveQueuedTxStructToCompleted(ctx,
					queueElement,
					"SETTLED")
			}
			err = updateAccountBalance(ctx,
				senderAccount.AccountID,
				senderAccount.Currency,
				extraReceiverBalance,
				true)
			if err != nil {
				return isCompleted, err
			}
			err = updateAccountBalance(ctx,
				receiverAccount.AccountID,
				receiverAccount.Currency,
				extraReceiverBalance,
				false)
			if err != nil {
				return isCompleted, err
			}
			isCompleted = true

		} else if extraReceiverBalance < 0 &&
			extraReceiverBalance <= receiverAccount.Amount {

			extraReceiverBalance *= -1
			for _, queueElement := range nettableQueueArray {
				err = moveQueuedTxStructToCompleted(ctx,
					queueElement,
					"SETTLED")
			}
			err = updateAccountBalance(ctx,
				senderAccount.AccountID,
				senderAccount.Currency,
				extraReceiverBalance,
				false)
			if err != nil {
				return isCompleted, err
			}
			err = updateAccountBalance(ctx,
				receiverAccount.AccountID,
				receiverAccount.Currency,
				extraReceiverBalance,
				true)
			if err != nil {
				return isCompleted, err
			}
			isCompleted = true
		}
	}
	return isCompleted, nil
}
