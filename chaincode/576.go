package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func (t *SimpleChaincode) conductMLNetting(ctx contractapi.TransactionContextInterface, args []string) ([]byte, error) {

	err := t.checkArgArrayLength(args, 5)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	if len(args[1]) <= 0 {
		return nil, fmt.Errorf("Bank ID must be a non-empty string")
	}
	if len(args[2]) <= 0 {
		return nil, fmt.Errorf("Nettable array must be a non-empty string")
	}
	if len(args[3]) <= 0 {
		return nil, fmt.Errorf("Non-nettable array must be a non-empty string")
	}
	if len(args[4]) <= 0 {
		return nil, fmt.Errorf("nettedValue must be a non-empty string")
	}
	bankID := args[1]
	nettableAarray, err := t.convertStringToArrayOfStrings(args[2])

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	nonNettableArray, err := t.convertStringToArrayOfStrings(args[3])
	if err != nil {
		return nil, fmt.Errorf(err.Error())

	}
	nettedValue, err := strconv.ParseFloat(args[4], 64)

	if err != nil {
		return nil, fmt.Errorf("Amount must be a numeric string")
	}
	err = t.verifyIdentity(ctx, bankID)
	if err != nil {
		return nil, fmt.Errorf(err.Error())

	}
	currTime, err := t.getTxTimeStampAsTime(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	nettingCycle, err := t.getCurrentNettingCycle(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	nettingCycleID := -1
	if len(args[0]) > 0 {
		nettingCycleID, err = strconv.Atoi(args[0])
		if err != nil {
			return nil, fmt.Errorf("Netting cycle ID must be a numeric string")
		} else if nettingCycleID <= 0 {
			return nil, fmt.Errorf("Netting cycle ID must be a positive value")
		}

		if nettingCycle.CycleID != nettingCycleID {
			return nil, fmt.Errorf("Cycle ID does not match current netting cycle")
		}
	}
	bankRequestID := args[0] + bankID
	bankRequest, err := t.createBankRequest(ctx, bankRequestID, bankID, nettedValue, nettableAarray, nonNettableArray)
	if nettingCycle.Status == "ACHIEVED" || nettingCycle.Status == "INVALID" {
		return nil, fmt.Errorf("unable to start new netting cycle: last cycle not settled yet")
	}

	if len(args[0]) == 0 && nettingCycle.Status != "ONGOING" { // Start new netting cycle
		bankRequestMap := make(map[string]BankRequest)
		bankRequestMap[bankID] = *bankRequest

		nettingCycle.CycleID = nettingCycle.CycleID + 1
		nettingCycle.Status = "ONGOING"
		nettingCycle.Created = currTime
		nettingCycle.Updated = currTime
		nettingCycle.BankRequests = bankRequestMap

	} else { // Participate in ongoing cycle
		b, err := t.checkOngoingMLNettingExpiry(ctx)

		isExpired := b[0] != 0
		if err != nil {
			return nil, fmt.Errorf(err.Error())
		} else if isExpired {
			respMsg := "netting cycle is expired"
			return nil, fmt.Errorf(respMsg)
		}

		if nettingCycle.CycleID != nettingCycleID {
			return nil, fmt.Errorf(
				"netting cycle ID provided does not match current netting cycle")
		}
		nettingCycle.BankRequests[bankID] = *bankRequest
		nettingCycle.Updated = currTime

		var totalNettedValue float64
		nettableTxMap := make(map[string]int)
		for _, request := range nettingCycle.BankRequests {
			totalNettedValue += request.NetValue

			requestNettableList := request.NettableList
			for _, txID := range requestNettableList {
				nettableTxMap[txID]++
				if nettableTxMap[txID] > 2 {
					errMsg := fmt.Sprintf(
						"Error: transaction %s has been proposed more than twice",
						txID)
					return nil, fmt.Errorf(errMsg)
				}
			}
		}
		isNettable := false
		for _, txOccurance := range nettableTxMap { // Check for transaction pairs
			isNettable = true
			if txOccurance != 2 {
				isNettable = false
				break
			}
		}
		if isNettable {
			if totalNettedValue != 0 {
				nettingCycle.Status = "INVALID"
			} else {
				nettingCycle.Status = "ACHIEVED"
			}
		}
	}

	nettingCycleAsBytes, err := json.Marshal(nettingCycle)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	err = ctx.GetStub().PutState(nettingCycleObjectType, nettingCycleAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return nil, nil
}

func (t *SimpleChaincode) expireOngoingMLNetting(
	ctx contractapi.TransactionContextInterface) ([]byte, error) {

	b, err := t.checkOngoingMLNettingExpiry(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	isExpired := b[0] != 0

	respMsg := "Ongoing netting cycle is still valid"
	if isExpired {
		respMsg = "Netting cycle is now expired"
	}
	fmt.Println(respMsg)
	return ([]byte(respMsg)), nil
}

func (t *SimpleChaincode) checkOngoingMLNettingExpiry(
	ctx contractapi.TransactionContextInterface) ([]byte, error) {
	b := make([]byte, 1)
	isExpired := false
	currTime, err := t.getTxTimeStampAsTime(ctx)
	if isExpired {
		b[0] = 1
	} else {
		b[0] = 0
	}
	if err != nil {
		return b, err
	}

	nettingCycle, err := t.getCurrentNettingCycle(ctx)
	if isExpired {
		b[0] = 1
	} else {
		b[0] = 0
	}
	if err != nil {
		return b, err
	}
	cycleTimeElapsed := currTime.Sub(nettingCycle.Created)
	if cycleTimeElapsed.Minutes() >= cycleExpiryMinutes &&
		nettingCycle.Status == "ONGOING" {
		nettingCycle.Status = "EXPIRED"
		nettingCycleAsBytes, err := json.Marshal(nettingCycle)
		if isExpired {
			b[0] = 1
		} else {
			b[0] = 0
		}
		if err != nil {
			return b, err
		}
		err = ctx.GetStub().PutState(nettingCycleObjectType, nettingCycleAsBytes)
		if isExpired {
			b[0] = 1
		} else {
			b[0] = 0
		}
		if err != nil {
			return b, err
		}
		isExpired = true
	}

	if isExpired {
		b[0] = 1
	} else {
		b[0] = 0
	}

	return b, err
}

func (t *SimpleChaincode) updateOngoingMLNettingStatus(
	ctx contractapi.TransactionContextInterface,
	status string) ([]byte, error) {

	currTime, err := t.getTxTimeStampAsTime(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	nettingCycle, err := t.getCurrentNettingCycle(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	nettingCycle.Status = status
	nettingCycle.Updated = currTime
	nettingCycleAsBytes, err := json.Marshal(nettingCycle)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	err = ctx.GetStub().PutState(nettingCycleObjectType, nettingCycleAsBytes)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return nil, nil
}

func (t *SimpleChaincode) queryOngoingMLNetting(ctx contractapi.TransactionContextInterface) ([]byte, error) {
	nettingCycleAsBytes, err := ctx.GetStub().GetState(nettingCycleObjectType)
	if err != nil {
		return nil, fmt.Errorf("error: Failed to get state for current nettingcycle")
	} else if nettingCycleAsBytes == nil {
		return nil, fmt.Errorf("error: netting cycle does not exist")
	}
	return nettingCycleAsBytes, nil
}

func (t *SimpleChaincode) getBilateralNettableTxList(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	err := t.checkArgArrayLength(args, 2)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("bank 1 ID must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, fmt.Errorf("bank 2 ID must be a non-empty string")
	}

	bank1ID := args[0]
	bank2ID := args[1]

	nettingCycle, err := t.getCurrentNettingCycle(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	if nettingCycle.Status != "ACHIEVED" {
		errMsg := "error: Current netting cycle is not achieved"
		return nil, fmt.Errorf(errMsg)
	}

	bank1Request := nettingCycle.BankRequests[bank1ID]
	bank2Request := nettingCycle.BankRequests[bank2ID]

	nettableArray := bank1Request.NettableList
	nettableArray = append(bank2Request.NettableList, nettableArray...)

	nettableTxMap := make(map[string]int)
	for _, txID := range nettableArray {
		nettableTxMap[txID]++
	}

	var bilateralNettableArr []string
	for txID, txOccurance := range nettableTxMap { // Check for transaction pairs
		if txOccurance == 2 {
			bilateralNettableArr = append(bilateralNettableArr, txID)
		}
	}
	bilateralNettableArrByte, err := json.Marshal(bilateralNettableArr)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	fmt.Println(bilateralNettableArrByte)
	return bilateralNettableArrByte, nil
}

func (t *SimpleChaincode) checkParticipation(
	ctx contractapi.TransactionContextInterface,
	args []string) ([]byte, error) {

	err := t.checkArgArrayLength(args, 2)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	if len(args[0]) <= 0 {
		return nil, fmt.Errorf("bank 1 ID must be a non-empty string")
	}
	if len(args[1]) <= 0 {
		return nil, fmt.Errorf("bank 2 ID must be a non-empty string")
	}

	bank1ID := args[0]
	bank2ID := args[1]

	nettingCycle, err := t.getCurrentNettingCycle(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	_, isBank1Exist := nettingCycle.BankRequests[bank1ID]
	_, isBank2Exist := nettingCycle.BankRequests[bank2ID]

	isParticipating := false
	if (nettingCycle.Status == "ONGOING" ||
		nettingCycle.Status == "ACHIEVED") &&
		(isBank1Exist || isBank2Exist) {

		isParticipating = true
	}
	fmt.Println(isParticipating)

	return []byte(strconv.FormatBool(isParticipating)), nil
}

func (t *SimpleChaincode) getNonNettableTxList(
	ctx contractapi.TransactionContextInterface) ([]byte, error) {

	nettingCycle, err := t.getCurrentNettingCycle(ctx)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	nonNettableMap := make(map[string]int)
	for _, request := range nettingCycle.BankRequests {
		requestNonNettableList := request.NonNettableList
		for _, txID := range requestNonNettableList {
			nonNettableMap[txID]++
		}
	}
	var nonNettableArray []string
	for txID := range nonNettableMap {
		nonNettableArray = append(nonNettableArray, txID)
	}
	nonNettableArrayByte, err := json.Marshal(nonNettableArray)
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}
	fmt.Println(nonNettableArrayByte)
	return nonNettableArrayByte, nil
}

func (t *SimpleChaincode) resetNettingCycle(
	ctx contractapi.TransactionContextInterface) error {

	nettingCycle, err := t.getCurrentNettingCycle(ctx)
	if err != nil {
		return err
	}
	bankRequestMap := make(map[string]BankRequest)
	nettingCycle.CycleID = 0
	nettingCycle.Status = "SETTLED"
	nettingCycle.BankRequests = bankRequestMap

	nettingCycleAsBytes, err := json.Marshal(nettingCycle)
	if err != nil {
		return err
	}
	err = ctx.GetStub().PutState(nettingCycleObjectType, nettingCycleAsBytes)
	if err != nil {
		return err
	}
	return nil
}
