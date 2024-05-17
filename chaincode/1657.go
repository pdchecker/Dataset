package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
	//"time"
)

// logDataCallTransaction creates a new Transactional Log Entry for a datacall
// Success: nil
// Error: {"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) LogDataCallTransaction(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("LogDataCallTransaction: enter")
	defer logger.Debug("LogDataCallTransaction: exit")
	logger.Debug("LogDataCallTransaction json received : ", args)
	if len(args) < 1 {
		return shim.Error("LogDataCallTransaction: Incorrect number of arguments!!")
	}

	var dataCallLog DataCallLog
	dataCallLogAsBytes := []byte(args)
	err := json.Unmarshal(dataCallLogAsBytes, &dataCallLog)
	if dataCallLog.DataCallID == "" || dataCallLog.DataCallVersion == "" {
		return shim.Error("LogDataCallTransaction: DataCallID and DataCallVersion cant not be empty!!")
	}
	if err != nil {
		logger.Error("LogDataCallTransaction: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("LogDataCallTransaction: Error during json.Unmarshal").Error())
	}

	// Create a new entry in log each time
	var pks []string = []string{DATACALL_LOG_PREFIX, dataCallLog.DataCallID, dataCallLog.DataCallVersion, stub.GetTxID()}
	dataCallLogKey, _ := stub.CreateCompositeKey(DATACALL_LOG_DOCUMENT, pks)

	stub.PutState(dataCallLogKey, dataCallLogAsBytes)
	return shim.Success(nil)

}

// GetDataCallTransactionHistory retrives all data calls that match given criteria. If startindex and pageSize are not provided,
// this method returns the complete list of data calls. If version = latest, the it returns only latest version of a data call
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// params {json}: {
//  "startIndex":"optional",
//  "pageSize":"optional",
//  "version": "latest or all"
//  "status" :"DRAFT OR ISSUED OR CANCELLED"}
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) GetDataCallTransactionHistory(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallTransactionHistory: enter")
	defer logger.Debug("GetDataCallTransactionHistory: exit")
	logger.Debug("GetDataCallTransactionHistory json received : ", args)
	if len(args) < 1 {
		return shim.Error("GetDataCallTransactionHistory: Incorrect number of arguments!!")
	}
	var getTxHistoryReq DataCallLog
	err := json.Unmarshal([]byte(args), &getTxHistoryReq)
	if err != nil {
		logger.Error("GetDataCallTransactionHistory: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallTransactionHistory: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetDataCallTransactionHistory: Unmarshalled object ", getTxHistoryReq)
	var pks []string = []string{DATACALL_LOG_PREFIX, getTxHistoryReq.DataCallID, getTxHistoryReq.DataCallVersion}
	resultsIterator, errMsg := stub.GetStateByPartialCompositeKey(DATACALL_LOG_DOCUMENT, pks)
	defer resultsIterator.Close()
	if errMsg != nil {
		logger.Warning("GetDataCallTransactionHistory: Failed to get state for all the data calls")
	}

	var getTxHistoryRes []DataCallLog
	if !resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := json.Marshal(getTxHistoryRes)
		logger.Debug("GetDataCallTransactionHistory: dataCallsAsByte", getTxHistoryRes)
		return shim.Success(dataCallLogsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := resultsIterator.Next()
		var currenDatacallLog DataCallLog
		json.Unmarshal([]byte(dataCallLogsAsByte.GetValue()), &currenDatacallLog)
		getTxHistoryRes = append(getTxHistoryRes, currenDatacallLog)
	}
	getTxHistoryResAsBytes, _ := json.Marshal(getTxHistoryRes)

	return shim.Success(getTxHistoryResAsBytes)

}

func (this *SmartContract) ListDataCallTransactionHistory(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallTransactionHistory: enter")
	defer logger.Debug("GetDataCallTransactionHistory: exit")
	logger.Debug("GetDataCallTransactionHistory json received : ", args)
	if len(args) < 1 {
		return shim.Error("GetDataCallTransactionHistory: Incorrect number of arguments!!")
	}
	var getTxHistoryReq DataCallLog
	err := json.Unmarshal([]byte(args), &getTxHistoryReq)
	if err != nil {
		logger.Error("GetDataCallTransactionHistory: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallTransactionHistory: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetDataCallTransactionHistory: Unmarshalled object ", getTxHistoryReq)
	//var pks []string = []string{DATACALL_LOG_PREFIX, getTxHistoryReq.DataCallID, getTxHistoryReq.DataCallVersion}
	queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"dataCallID\":\"%s\",\"dataCallVersion\":\"%s\"},\"use_index\":[\"_design/actionTs\"],\"sort\":[{\"actionTs\":\"desc\"}]}", DATACALL_LOG_PREFIX, getTxHistoryReq.DataCallID, getTxHistoryReq.DataCallVersion)
	resultsIterator, errMsg := stub.GetQueryResult(queryStr)
	//resultsIterator, errMsg := stub.GetStateByPartialCompositeKey(DATACALL_LOG_DOCUMENT, pks)
	defer resultsIterator.Close()
	if errMsg != nil {
		logger.Warning("GetDataCallTransactionHistory: Failed to get state for all the data calls")
	}

	var getTxHistoryRes []DataCallLog
	if !resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := json.Marshal(getTxHistoryRes)
		logger.Debug("GetDataCallTransactionHistory: dataCallsAsByte", getTxHistoryRes)
		return shim.Success(dataCallLogsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallLogsAsByte, _ := resultsIterator.Next()
		var currenDatacallLog DataCallLog
		json.Unmarshal([]byte(dataCallLogsAsByte.GetValue()), &currenDatacallLog)
		getTxHistoryRes = append(getTxHistoryRes, currenDatacallLog)
	}
	getTxHistoryResAsBytes, _ := json.Marshal(getTxHistoryRes)

	return shim.Success(getTxHistoryResAsBytes)

}
