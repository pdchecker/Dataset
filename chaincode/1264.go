package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
)

// generateVersion is a helper function for generating a version number.
func generateVersion(number int) string {
	return strconv.Itoa(number + 1)
}

// CreateDataCall creates a new DataCall object. This method receives as a parameter a DataCall object in JSON format.
// Success: nil
// Error: {"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (s *SmartContract) CreateDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateDataCall: enter")
	defer logger.Debug("CreateDataCall: exit")
	logger.Debug("CreaCreateDataCall json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var dataCall DataCall
	err := json.Unmarshal([]byte(args), &dataCall)
	if dataCall.ID == "" {
		return shim.Error("Id cant not be empty!!")
	}
	if err != nil {
		logger.Error("CreateDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CreateDataCall: Error during json.Unmarshal").Error())
	}
	if dataCall.Status == STATUS_DRAFT {
		dataCall.IsLocked = false
	}
	if dataCall.Status == STATUS_ISSUED || dataCall.Status == STATUS_CANCELLED {
		dataCall.IsLocked = true
	}
	logger.Debug("Unmarshalled object ", dataCall)
	dataCall.IsLatest = true
	//dataCall.IsLocked = "false"
	dataCall.Version = generateVersion(0)

	var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID, dataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	//dataCallKey := DATA_CALL_PREFIX + dataCall.ID + dataCall.Version
	logger.Info("In data call create ", dataCallKey)
	// Checking the ledger to confirm that the dataCall doesn't exist
	prevDataCall, _ := stub.GetState(dataCallKey)

	if prevDataCall != nil {
		logger.Error("CreateDataCall: Data Call already exist for the data call with ID: " + dataCallKey)
		return shim.Error("Data Call already exist for the data call with ID: " + dataCallKey)
	}

	dataCallAsBytes, _ := json.Marshal(dataCall)
	err = stub.PutState(dataCallKey, dataCallAsBytes)

	// update datacalllog
	//var dataCallCreateLog DataCallLog
	if dataCall.Status == STATUS_ISSUED {
		dataCallIssuedLog := DataCallLog{dataCall.ID, dataCall.Version, ActionIssued.ActionID, ActionIssued.ActionDesc, dataCall.UpdatedTs, dataCall.UpdatedBy}
		dataCallIssuedLogAsBytes, _ := json.Marshal(dataCallIssuedLog)
		logger.Info((dataCallIssuedLogAsBytes))
		// this.LogDataCallTransaction(stub, string(dataCallIssuedLogAsBytes))
	}

	if err != nil {
		logger.Error("Error commiting data for key: ", dataCallKey)
		return shim.Error("Error committing data for key: " + dataCallKey)
	}

	//change the count of the statuses
	logger.Info("Toggling ", dataCall.Status)
	toggleDataCallCount := ToggleDataCallCount{"", dataCall.Status}
	datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
	logger.Info((datacallIssueLogAsBytes))
	// dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
	// logger.Info("The reply from toggle is ", dataCallCountAsBytes)

	return shim.Success(nil)

}

// ListDataCallsByCriteria retrives all data calls that match given criteria. If startindex and pageSize are not provided,
// this method returns the complete list of data calls. If version = latest, the it returns only latest version of a data call
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// params {json}: {
//  "startIndex":"optional",
//  "pageSize":"optional",
//  "version": "latest or all"
//  "status" :"DRAFT OR ISSUED OR CANCELLED"}
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) ListDataCallsByCriteria(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("ListDataCallsByCriteria: enter")
	defer logger.Debug("ListDataCallsByCriteria: exit")
	logger.Debug("ListDataCallsByCriteria json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}
	var searchCriteria SearchCriteria
	err := json.Unmarshal([]byte(args), &searchCriteria)
	if err != nil {
		logger.Error("ListDataCallsByCriteria: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("ListDataCallsByCriteria: Error during json.Unmarshal").Error())
	}
	logger.Debug("ListDataCallsByCriteria: Unmarshalled object ", searchCriteria)
	startIndex := searchCriteria.StartIndex
	pageSize := searchCriteria.PageSize
	version := searchCriteria.Version
	var isLatest string
	if version == LATEST_VERSION {
		isLatest = "true"
	} else {
		isLatest = "false"
	}
	status := searchCriteria.Status
	if status == "" {
		logger.Error("ListDataCallsByCriteria: Status not present, You must pass Status in agument")
		return shim.Error("ListDataCallsByCriteria: Status not present, You must pass Status in agument")
	}
	var queryStr string

	if status == STATUS_ISSUED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\"},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"deadline\": \"desc\"}],\"limit\":%d,\"skip\":%d}", DATA_CALL_PREFIX, status, pageSize, startIndex)
	} else if status == STATUS_DRAFT {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\", \"isLatest\":%s, \"isLocked\":false},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"updatedTs\": \"desc\"}],\"limit\":%d,\"skip\":%d}", DATA_CALL_PREFIX, status, isLatest, pageSize, startIndex)
	} else if status == STATUS_CANCELLED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\"},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"deadline\": \"desc\"}],\"limit\":%d,\"skip\":%d}", DATA_CALL_PREFIX, status, pageSize, startIndex)
	}

	logger.Info("ListDataCallsByCriteria: Selector ", queryStr)
	logger.Info("startIndex", "pageSize", "version", "status", startIndex, pageSize, version, status)

	var dataCalls []DataCall
	startTime := time.Now()
	resultsIterator, responseMetadata, err := stub.GetQueryResultWithPagination(queryStr, int32(pageSize), "")
	elapsedTime := time.Since(startTime)
	logger.Info("RESPONSE META DATA ", responseMetadata.FetchedRecordsCount)
	logger.Info("Time consumed to get Data Calls", elapsedTime)
	defer resultsIterator.Close()
	if err != nil {
		logger.Error("Failed to get state for all the data calls")
		return shim.Error("Failed to get state for all the data calls")
	}

	if !resultsIterator.HasNext() {
		dataCallsAsByte, _ := json.Marshal(dataCalls)
		logger.Debug("ListDataCallsByCriteria: dataCallsAsByte", dataCallsAsByte)
		//return shim.Error(errors.New("ListDataCallsByCriteria :DataCall not found ").Error())
		return shim.Success(dataCallsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}

		dataCalls = append(dataCalls, dataCall)
	}

	//var paginatedDataCall []DataCall
	//var paginatedDataCalls []DataCall
	//paginatedDataCalls = paginate(dataCalls, startIndex, pageSize)
	var dataCallList DataCallList

	//getting the count
	getDataCallCount := GetDataCallCount{"123456", "1"}
	datacallIssueLogAsBytes, _ := json.Marshal(getDataCallCount)
	dataCallCountAsBytes := this.GetDataCallCount(stub, string(datacallIssueLogAsBytes))
	var dataCallCount DataCallCount
	err = json.Unmarshal(dataCallCountAsBytes.Payload, &dataCallCount)
	logger.Info("The retrieved data is ", dataCallCount)

	if status == STATUS_ISSUED {
		dataCallList.TotalNoOfRecords = dataCallCount.ISSUED
	} else if status == STATUS_DRAFT {
		dataCallList.TotalNoOfRecords = dataCallCount.DRAFT
	} else if status == STATUS_CANCELLED {
		dataCallList.TotalNoOfRecords = dataCallCount.CANCELLED
	}

	var IdAndVersionMap map[string]string
	IdAndVersionMap = make(map[string]string)
	var dataCallIDs []string
	var dataCallVersions []string
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		if dataCallIndex == 0 {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `"`+dataCalls[dataCallIndex].Version+`"`)
		} else {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `,`+`"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `,`+`"`+dataCalls[dataCallIndex].Version+`"`)
		}
	}

	startTimeForAll := time.Now()
	/*
		//get ConsentCount map
		startTimeForConsent := time.Now()
		consentCounts := GetConsentsCount(stub, dataCallIDs)
		elapsedTimeForConsent := time.Since(startTimeForConsent)
		logger.Info("Time consumed for consent", elapsedTimeForConsent)

		startTimeForLike := time.Now()
		likeCounts := GetLikesCount(stub, dataCallIDs, IdAndVersionMap)
		elapsedTimeForLike := time.Since(startTimeForLike)
		logger.Info("Time consumed for Like", elapsedTimeForLike)
	*/

	startTimeForReport := time.Now()
	latestReports := GetLatestaReport(stub, dataCallIDs, dataCallVersions)
	elapsedTimeForReport := time.Since(startTimeForReport)
	logger.Info("Time consumed for Report", elapsedTimeForReport)

	elapsedTimeForAll := time.Since(startTimeForAll)
	logger.Info("Time consumed for all", elapsedTimeForAll)
	logger.Info("Final ===========", elapsedTime)
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		var dataCallExtended DataCallExtended
		dataCallExtended.DataCall = dataCalls[dataCallIndex]
		dataCallExtended.Reports = append(dataCallExtended.Reports, latestReports[dataCalls[dataCallIndex].ID])
		//dataCallExtended.NoOfConsents = consentCounts[paginatedDataCalls[dataCallIndex].ID]
		//dataCallExtended.NoOfLikes = likeCounts[paginatedDataCalls[dataCallIndex].ID]
		dataCallList.DataCalls = append(dataCallList.DataCalls, dataCallExtended)

	}

	dataCallsAsByte, _ := json.Marshal(dataCallList)
	return shim.Success(dataCallsAsByte)

}

// ListMatureDataCalls retrives all data calls that has matured in last 24 hours.
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) ListMatureDataCalls(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("ListMatureDataCalls: enter")
	defer logger.Debug("ListMatureDataCalls: exit")
	status := STATUS_ISSUED
	var queryStr string
	queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\"},\"use_index\":[\"deadline\", \"deadlineIndex\"],\"sort\":[{\"deadline\": \"desc\"}]}", DATA_CALL_PREFIX, status)

	var dataCalls []DataCall
	startTime := time.Now()
	resultsIterator, err := stub.GetQueryResult(queryStr)
	elapsedTime := time.Since(startTime)
	logger.Info("Time consumed to get Data Calls", elapsedTime)
	defer resultsIterator.Close()
	if err != nil {
		logger.Error("Failed to get state for all the data calls")
		return shim.Error("Failed to get state for all the data calls")
	}

	if !resultsIterator.HasNext() {
		dataCallsAsByte, _ := json.Marshal(dataCalls)
		logger.Debug("ListDataCallsByCriteria: dataCallsAsByte", dataCallsAsByte)
		//return shim.Error(errors.New("ListDataCallsByCriteria :DataCall not found ").Error())
		return shim.Success(dataCallsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}
		// If mature date in last 24 hours, add them to datacalls

		startDate := startTime.Truncate(24*time.Hour).AddDate(0, 0, -1)
		endDate := startTime.Truncate(24 * time.Hour)
		if (dataCall.Deadline.After(startDate) && dataCall.Deadline.Before(endDate)) || dataCall.Deadline.Equal(startTime) {
			dataCalls = append(dataCalls, dataCall)
		}
	}

	var dataCallList DataCallList

	//getting the count
	getDataCallCount := GetDataCallCount{"123456", "1"}
	datacallIssueLogAsBytes, _ := json.Marshal(getDataCallCount)
	dataCallCountAsBytes := this.GetDataCallCount(stub, string(datacallIssueLogAsBytes))
	var dataCallCount DataCallCount
	err = json.Unmarshal(dataCallCountAsBytes.Payload, &dataCallCount)
	logger.Info("The retrieved data is ", dataCallCount)

	if status == STATUS_ISSUED {
		dataCallList.TotalNoOfRecords = dataCallCount.ISSUED
	} else if status == STATUS_DRAFT {
		dataCallList.TotalNoOfRecords = dataCallCount.DRAFT
	} else if status == STATUS_CANCELLED {
		dataCallList.TotalNoOfRecords = dataCallCount.CANCELLED
	}

	var IdAndVersionMap map[string]string
	IdAndVersionMap = make(map[string]string)
	var dataCallIDs []string
	var dataCallVersions []string
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		if dataCallIndex == 0 {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `"`+dataCalls[dataCallIndex].Version+`"`)
		} else {
			IdAndVersionMap[dataCalls[dataCallIndex].ID] = dataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `,`+`"`+dataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `,`+`"`+dataCalls[dataCallIndex].Version+`"`)
		}
	}

	startTimeForAll := time.Now()
	/*
		//get ConsentCount map
		startTimeForConsent := time.Now()
		consentCounts := GetConsentsCount(stub, dataCallIDs)
		elapsedTimeForConsent := time.Since(startTimeForConsent)
		logger.Info("Time consumed for consent", elapsedTimeForConsent)

		startTimeForLike := time.Now()
		likeCounts := GetLikesCount(stub, dataCallIDs, IdAndVersionMap)
		elapsedTimeForLike := time.Since(startTimeForLike)
		logger.Info("Time consumed for Like", elapsedTimeForLike)
	*/

	startTimeForReport := time.Now()
	latestReports := GetLatestaReport(stub, dataCallIDs, dataCallVersions)
	elapsedTimeForReport := time.Since(startTimeForReport)
	logger.Info("Time consumed for Report", elapsedTimeForReport)

	elapsedTimeForAll := time.Since(startTimeForAll)
	logger.Info("Time consumed for all", elapsedTimeForAll)
	logger.Info("Final ===========", elapsedTime)
	for dataCallIndex := 0; dataCallIndex < len(dataCalls); dataCallIndex++ {
		var dataCallExtended DataCallExtended
		dataCallExtended.DataCall = dataCalls[dataCallIndex]
		dataCallExtended.Reports = append(dataCallExtended.Reports, latestReports[dataCalls[dataCallIndex].ID])
		//dataCallExtended.NoOfConsents = consentCounts[paginatedDataCalls[dataCallIndex].ID]
		//dataCallExtended.NoOfLikes = likeCounts[paginatedDataCalls[dataCallIndex].ID]
		dataCallList.DataCalls = append(dataCallList.DataCalls, dataCallExtended)

	}

	dataCallsAsByte, _ := json.Marshal(dataCallList)
	return shim.Success(dataCallsAsByte)

}

//helper function for pagination
func paginate(dataCall []DataCall, startIndex int, pageSize int) []DataCall {
	if startIndex == 0 {
		startIndex = PAGINATION_DEFAULT_START_INDEX
	}
	// no pageSize specified then return all results
	if pageSize == 0 {
		pageSize = len(dataCall)
		return dataCall
	}
	limit := func() int {
		if startIndex+pageSize > len(dataCall) {
			return len(dataCall)
		} else {
			return startIndex + pageSize - 1
		}
	}

	start := func() int {
		if startIndex > len(dataCall) {
			return len(dataCall) - 1
		} else {
			return startIndex - 1
		}
	}
	logger.Debug("ListDataCallsByCriteria: Getting Records from index", start(), " to index ", limit())
	return dataCall[start():limit()]
}

// function-name: SaveNewDraft (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Creates/Update new version of the data call.

func (this *SmartContract) SaveNewDraft(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("SaveNewDraft: enter")
	defer logger.Debug("SaveNewDraft: exit")
	logger.Debug("SaveNewDraft json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}
	var dataCalls []DataCall
	var currentDataCall DataCall
	//var prevDataCall DataCall
	err := json.Unmarshal([]byte(args), &currentDataCall)
	if err != nil {
		logger.Error("SaveNewDraft: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveNewDraft: Error during json.Unmarshal").Error())
	}

	logger.Debug("Unmarshalled object ", currentDataCall)

	if currentDataCall.ID == "" {
		return shim.Error("Id can not be empty")
	}
	var pks []string = []string{DATA_CALL_PREFIX, currentDataCall.ID}

	//Step-1: getting the previous data for the data call
	resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(DOCUMENT_TYPE, pks)
	if errmsg != nil {
		logger.Error("SaveNewDraft: Failed to get state for previous data calls")
		return shim.Error("SaveNewDraft: Failed to get state for previous data calls")
	}
	defer resultsIterator.Close()
	logger.Debug("SaveNewDraft: ", resultsIterator)
	for resultsIterator.HasNext() {
		prevDataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var prevDataCall DataCall
		err = json.Unmarshal([]byte(prevDataCallAsBytes.GetValue()), &prevDataCall)
		if err != nil {
			logger.Error("Failed to unmarshal data call ")
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}
		//fetching the latest record
		if prevDataCall.IsLatest == true {
			dataCalls = append(dataCalls, prevDataCall)
		}
	}

	//Step-2: setting the previous DataCall with isLatest to false and
	dataCalls[0].IsLatest = false
	//dataCalls[0].IsLocked = "false"
	prevDataCallVersion, _ := strconv.Atoi(dataCalls[0].Version)

	//creating composite key to save the previous DataCall
	var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, dataCalls[0].ID, dataCalls[0].Version}
	prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
	fmt.Println(prevDataCallKey)
	logger.Debug("SaveNewDraft: previousDatacallKey > ", prevDataCallKey)

	prevDataCallAsBytes, _ := json.Marshal(dataCalls[0])
	//saving the previous DataCall
	err = stub.PutState(prevDataCallKey, prevDataCallAsBytes)
	logger.Debug("SaveNewDraft: Previous Datacall saved!")
	if err != nil {
		logger.Error("Error commiting the previous DataCall")
		return shim.Error("Error commiting the previous DataCall")
	}

	//Step-4: saving the draft with updating new version and setting isLatest to true
	currentDataCall.IsLatest = true
	currentDataCall.IsLocked = false
	//currentDataCallVersion, _ := strconv.Atoi(currentDataCall.Version)
	currentDataCall.Version = generateVersion(prevDataCallVersion)

	var pkForCurrentDataCall []string = []string{DATA_CALL_PREFIX, currentDataCall.ID, currentDataCall.Version}
	currentDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForCurrentDataCall)

	currentDataCallAsBytes, _ := json.Marshal(currentDataCall)
	err = stub.PutState(currentDataCallKey, currentDataCallAsBytes)
	if err != nil {
		return shim.Error("Error committing Current DataCall for key: " + currentDataCallKey)
	}
	logger.Debug("SaveNewDraft: Latest Datacall saved!")

	if currentDataCall.Status != dataCalls[0].Status {
		//change the count of the statuses
		logger.Info("Toggling ", currentDataCall.Status, " and ", dataCalls[0].Status)
		toggleDataCallCount := ToggleDataCallCount{dataCalls[0].Status, currentDataCall.Status}
		datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
		dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
		logger.Info("The reply from toggle is ", dataCallCountAsBytes)
	}

	return shim.Success(nil)

}

// function-name: GetDataCallVersionsById(invoke)
// params {json}: {
// "id":"mandatory",
// "startIndex":"optional",
// "pageSize":"optional"}
// Success {byte[]}: byte[]  - List of data calls
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : returns a datacall of specifc version and id.
// if pageSize is blank returns all versions

func (this *SmartContract) GetDataCallVersionsById(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallVersionsById: enter")
	defer logger.Debug("GetDataCallVersionsById: exit")

	var getDataCallVersions GetDataCallVersions
	err := json.Unmarshal([]byte(args), &getDataCallVersions)
	if err != nil {
		logger.Error("GetDataCallVersionsById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallVersionsById: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getDataCallVersions)

	var dataCalls []DataCall
	if getDataCallVersions.ID == "" {
		return shim.Error("GetDataCallVersionsById: ID is Empty")
	}
	logger.Debug("GetDataCallVersionsById: Status", getDataCallVersions.Status)
	//fmt.Println(startIndex, pageSize)
	//queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"id\":\"%s\"}}", DATA_CALL_PREFIX, getDataCallVersions.ID)
	//queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"id\":\"%s\"},\"sort\":[{\"version\":\"desc\"}]}", DATA_CALL_PREFIX, getDataCallVersions.ID)
	queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"id\":\"%s\"},\"use_index\":\"_design/versionDoc\",\"sort\":[{\"version\": \"desc\"}]}", DATA_CALL_PREFIX, getDataCallVersions.ID)
	resultsIterator, err := stub.GetQueryResult(queryStr)

	if err != nil {
		logger.Error("GetDataCallVersionsById: Failed to get Data Calls")
		return shim.Error("Failed to get Data Calls : " + err.Error())
	}
	defer resultsIterator.Close()
	logger.Debug("GetDataCallVersionsById: Iterating over datacalls versions")
	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)

		logger.Debug("GetDataCallVersionsById: DataCall > ", dataCall.ID)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}

		if len(getDataCallVersions.Status) > 0 {
			// if status is present in input add as per status
			if getDataCallVersions.Status == dataCall.Status {
				dataCalls = append(dataCalls, dataCall)
			}
		} else {
			// else add all data calls in response
			dataCalls = append(dataCalls, dataCall)
		}

	}
	dataCallsAsByte, _ := json.Marshal(dataCalls)
	return shim.Success(dataCallsAsByte)

}

// function-name: GetDataCallByIdAndVersion (invoke)
// params {json}: {
// "id":"mandatory",
// "version": "mandatory"}
// Success {byte[]}: byte[]  - List of data calls
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : returns a datacall of specifc version and id.

func (s *SmartContract) GetDataCallByIdAndVersion(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallByIdAndVersion: enter")
	defer logger.Debug("GetDataCallByIdAndVersion: exit")

	var getDataCall GetDataCall
	err := json.Unmarshal([]byte(args), &getDataCall)
	if err != nil {
		logger.Error("GetDataCallVersionsById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallVersionsById: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getDataCall)

	if getDataCall.ID == "" || getDataCall.Version == "" {
		return shim.Error("ID and Version can not be Empty")
	}

	var pks []string = []string{DATA_CALL_PREFIX, getDataCall.ID, getDataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		logger.Error("Error retreiving data for key ", dataCallKey)
		return shim.Error("Error retreiving data for key" + dataCallKey)
	}
	return shim.Success(dataCallAsBytes)

}

// function-name: UpdateDataCall (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Updates the data call, without creating new version.
//   This needs to be invoked when we need to change delivery date

func (this *SmartContract) UpdateDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateDataCall: enter")
	defer logger.Debug("UpdateDataCall: exit")
	logger.Debug("UpdateDataCall json received : ", args)

	var dataCall DataCall
	var isCancelled bool
	err := json.Unmarshal([]byte(args), &dataCall)
	if err != nil {
		logger.Error("UpdateDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateDataCall: Error during json.Unmarshal").Error())
	}
	if dataCall.ID == "" {
		return shim.Error("ID should not be Empty")

	}
	if dataCall.Version == "" {
		return shim.Error("Version should not be Empty")

	}
	var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID, dataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		return shim.Error("Error retreiving data for key" + dataCallKey)
	}

	var prevDataCall DataCall
	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
	if err != nil {
		return shim.Error("Failed to unmarshal data call: " + err.Error())
	}
	if prevDataCall.Status == STATUS_DRAFT {
		if dataCall.Status == STATUS_CANCELLED {
			isCancelled = true
			prevDataCall.Status = dataCall.Status
			prevDataCall.IsLocked = true
		}
		prevDataCall.ForumURL = dataCall.ForumURL

	} else if prevDataCall.Status == STATUS_ISSUED {
		if prevDataCall.ProposedDeliveryDate != dataCall.ProposedDeliveryDate {
			// update datacalllog
			datacallIssueLog := DataCallLog{dataCall.ID, dataCall.Version, ActionDeliveryDateUpdate.ActionID,
				ActionDeliveryDateUpdate.ActionDesc, dataCall.UpdatedTs, dataCall.UpdatedBy}
			datacallIssueLogAsBytes, _ := json.Marshal(datacallIssueLog)
			this.LogDataCallTransaction(stub, string(datacallIssueLogAsBytes))
		}
		prevDataCall.ProposedDeliveryDate = dataCall.ProposedDeliveryDate
		prevDataCall.ForumURL = dataCall.ForumURL
		prevDataCall.ExtractionPatternID = dataCall.ExtractionPatternID
		prevDataCall.ExtractionPatternTs = dataCall.ExtractionPatternTs
		prevDataCall.IsLocked = true
	}

	//updating all the dataCalls with isLocked = true, when isCancelled
	if isCancelled {
		var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID}

		//Step-1: getting the previous data for the data call
		resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(DOCUMENT_TYPE, pks)
		if errmsg != nil {
			logger.Error("Failed to get state for previous data calls")
			return shim.Error("Failed to get state for previous data calls")
		}
		defer resultsIterator.Close()
		for resultsIterator.HasNext() {
			prevDataCallAsBytes, err := resultsIterator.Next()
			if err != nil {
				return shim.Error("Failed to iterate data call")
			}

			var prevDataCall DataCall
			err = json.Unmarshal([]byte(prevDataCallAsBytes.GetValue()), &prevDataCall)
			if err != nil {
				logger.Error("Failed to unmarshal data call: ")
				return shim.Error("Failed to unmarshal data call: " + err.Error())
			}

			//dont update the current update as it is already being updated
			if prevDataCall.Version != dataCall.Version {
				prevDataCall.IsLocked = true
				//creating composite key to save the previous DataCall
				var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, prevDataCall.ID, prevDataCall.Version}
				prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
				prevDataAsBytes, _ := json.Marshal(prevDataCall)
				err = stub.PutState(prevDataCallKey, prevDataAsBytes)
				if err != nil {
					logger.Error("Error commiting the previous DataCall")
					return shim.Error("Error commiting the previous DataCall")
				}
			}
		}

	}

	//prevDataCall.ForumURL = dataCall.ForumURL
	//creating composite key to save the previous DataCall
	var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, prevDataCall.ID, prevDataCall.Version}
	prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
	fmt.Println(prevDataCallKey)
	prevDataAsBytes, _ := json.Marshal(prevDataCall)
	err = stub.PutState(prevDataCallKey, prevDataAsBytes)
	if err != nil {
		logger.Error("Error commiting the previous DataCall")
		return shim.Error("Error commiting the previous DataCall")
	}

	//Prepare response to emit
	//patterns := GetExtractionPatternsMap() //GetExtractionPatternById(dataCall.ExtractionPatternID)
	var extPatternResponse ExtractionPatternPayload
	extPatternResponse.DataCallId = dataCall.ID
	extPatternResponse.DataCallVsersion = dataCall.Version
	//extPatternResponse.ExtractionPattern = patterns[dataCall.ExtractionPatternID]
	extPatternResponse.ExtractionPatternID = dataCall.ExtractionPatternID
	extPatternResponse.ExtPatternTs = dataCall.ExtractionPatternTs

	extPatternResponseAsBytes, _ := json.Marshal(extPatternResponse)

	_ = stub.SetEvent(SET_EXTRACTION_PATTERN_EVENT, extPatternResponseAsBytes)

	if dataCall.Status != prevDataCall.Status {
		//change the count of the statuses
		logger.Info("Toggling ", dataCall.Status, " and ", prevDataCall.Status)
		toggleDataCallCount := ToggleDataCallCount{prevDataCall.Status, dataCall.Status}
		datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
		dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
		logger.Info("The reply from toggle is ", dataCallCountAsBytes)
	}

	return shim.Success(nil)

}

// function-name: IssueDataCall (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Updates the status of DataCall to STATUS_ISSUED
// set all the versions of DataCall to isLocked="true"

func (this *SmartContract) IssueDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("IssueDataCall: enter")
	defer logger.Debug("IssueDataCall: exit")
	logger.Debug("IssueDataCall json received : ", args)

	var dataCall DataCall
	//var key_for_response string
	err := json.Unmarshal([]byte(args), &dataCall)
	if err != nil {
		logger.Error("IssueDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("IssueDataCall: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshelled object ", dataCall)
	if dataCall.ID == "" {
		return shim.Error("ID should not be Empty")

	}
	if dataCall.Version == "" {
		return shim.Error("Version should not be Empty")

	}
	var pks []string = []string{DATA_CALL_PREFIX, dataCall.ID}

	//Step-1: getting the previous data for the data call
	resultsIterator, errmsg := stub.GetStateByPartialCompositeKey(DOCUMENT_TYPE, pks)
	if errmsg != nil {
		logger.Error("Failed to get state for previous data calls")
		return shim.Error("Failed to get state for previous data calls")
	}
	defer resultsIterator.Close()
	for resultsIterator.HasNext() {
		prevDataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			return shim.Error("Failed to iterate data call")
		}

		var prevDataCall DataCall
		err = json.Unmarshal([]byte(prevDataCallAsBytes.GetValue()), &prevDataCall)
		if err != nil {
			logger.Error("Failed to unmarshal data call: ")
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}
		//updating the particular version which is coming in current DataCall
		if prevDataCall.IsLocked == true {
			return shim.Error("DataCall is Locked, as it has already been Issued or Cancelled")

		} else if prevDataCall.Version == dataCall.Version {
			if prevDataCall.Status == STATUS_DRAFT {
				if dataCall.Status != STATUS_ISSUED {
					return shim.Error("Invalid Status ")
				}
				prevDataCall.Status = dataCall.Status
				prevDataCall.IsLocked = true

			}
		}
		prevDataCall.IsLocked = true
		//creating composite key to save the previous DataCall
		var pkForPrevDataCall []string = []string{DATA_CALL_PREFIX, prevDataCall.ID, prevDataCall.Version}
		prevDataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pkForPrevDataCall)
		//key_for_response = prevDataCallKey
		prevDataAsBytes, _ := json.Marshal(prevDataCall)
		err = stub.PutState(prevDataCallKey, prevDataAsBytes)

		// update datacalllog
		datacallIssueLog := DataCallLog{dataCall.ID, dataCall.Version, ActionIssued.ActionID,
			ActionIssued.ActionDesc, dataCall.UpdatedTs, dataCall.UpdatedBy}
		datacallIssueLogAsBytes, _ := json.Marshal(datacallIssueLog)
		this.LogDataCallTransaction(stub, string(datacallIssueLogAsBytes))

		if err != nil {
			logger.Error("Error commiting the previous DataCall")
			return shim.Error("Error commiting the previous DataCall")
		}

		if dataCall.Status != prevDataCall.Status {
			//change the count of the statuses
			logger.Info("Toggling ", dataCall.Status, " and ", prevDataCall.Status)
			toggleDataCallCount := ToggleDataCallCount{prevDataCall.Status, dataCall.Status}
			datacallIssueLogAsBytes, _ := json.Marshal(toggleDataCallCount)
			dataCallCountAsBytes := this.ToggleDataCallCount(stub, string(datacallIssueLogAsBytes))
			logger.Info("The reply from toggle is ", dataCallCountAsBytes)
		}

	}

	return shim.Success(nil)

}

// TODO Remove this function as it must be done from API End through orchestration
// function-name: SaveAndIssueDataCall (invoke)
// params {DataCall json}
// Success :nil
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : Updates the status of DataCall to STATUS_ISSUED
// set all the versions of DataCall to isLocked="true"

func (this *SmartContract) SaveAndIssueDataCall(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("SaveAndIssueDataCall: enter")
	defer logger.Debug("SaveAndIssueDataCall: exit")
	logger.Debug("SaveAndIssueDataCall json received : ", args)

	var dataCall DataCall
	err := json.Unmarshal([]byte(args), &dataCall)

	if err != nil {
		logger.Error("SaveAndIssueDataCall: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveAndIssueDataCall: Error during json.Unmarshal").Error())
	}

	dataCall.Status = STATUS_DRAFT

	saveDraftJson, _ := json.Marshal(dataCall)
	saveDraftResponse := this.SaveNewDraft(stub, string(saveDraftJson))
	if saveDraftResponse.Status != 200 {
		logger.Error("SaveAndIssueDataCall: Unable to Save a new Draft: ", saveDraftResponse.Message)
		return shim.Error(errors.New("SaveAndIssueDataCall: Unable to Save a new Draft").Error())
	}

	logger.Debug("SaveAndIssueDataCall: Successfully Saved a new datacall, proceeding to issue data call")
	issueDataCall := dataCall
	issueVersion, err := strconv.Atoi(dataCall.Version)
	issueDataCall.Version = generateVersion(issueVersion)
	issueDataCall.Status = STATUS_ISSUED

	issueDataCallJson, _ := json.Marshal(issueDataCall)
	logger.Debug("SaveAndIssueDataCall: issueDataCallJson > ", issueDataCallJson)
	issueDataCallResponse := this.IssueDataCall(stub, string(issueDataCallJson))
	if issueDataCallResponse.Status != 200 {
		logger.Error("SaveAndIssueDataCall: Unable to Save a new Draft: ", issueDataCallResponse.Message)
		return shim.Error(errors.New("SaveAndIssueDataCall: Unable to Save a new Draft").Error())
	}
	return shim.Success(issueDataCallResponse.Payload)
}

func (this *SmartContract) toggleDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Info("REACHED INSIDE************** ")

	version := generateVersion(0)

	var pks []string = []string{DATA_CALL_PREFIX, "12345678", version}
	countPatternKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)

	fmt.Println("KEY IS [")
	fmt.Println(countPatternKey)
	logger.Info(countPatternKey)
	fmt.Println("] KEY IS")
	dataCallAsBytes, err := stub.GetState(countPatternKey)
	if err != nil {
		return shim.Error("Error retreiving data for key" + countPatternKey)
	}

	var prevDataCall DataCallCount
	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
	if err != nil {
		return shim.Error("Failed to unmarshal data call: " + err.Error())
	}
	logger.Info("GOT THE RESPONSE ", prevDataCall)

	return shim.Success(nil)

}

// func (this *openIDLCC) putDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
// 	logger.Debug("putDataCallCount: enter")
// 	defer logger.Debug("putDataCallCount: exit")
// 	logger.Info("putDataCallCount json received : ", args)
// 	if len(args) < 1 {
// 		return shim.Error("Incorrect number of arguments!!")
// 	}

// 	var pks []string = []string{DATA_CALL_PREFIX, "COUNT", "10"}
// 	countPatternKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)

// 	args3 := map[string]int{"ISSUED": 1,"DRAFT": 0, "CANCELLED": 0}
// 	var dataCallCount = DataCallCount{counts: args3}
// 	fmt.Println("REACHED INSIDE 3 " + countPatternKey)

// 	logger.Info("REACHED INSIDE 4 ", dataCallCount)

// 	prevDataCallAsBytes, _ := json.Marshal(dataCallCount)
// 	logger.Info("REACHED INSIDE 5 ", dataCallCount)
// 	fmt.Println(prevDataCallAsBytes)
// 	//saving the previous DataCall
// 	err := stub.PutState(countPatternKey, prevDataCallAsBytes)
// 	logger.Info("SaveNewDraft: Previous Datacall saved!")
// 	if err != nil {
// 		logger.Info("SaveNewDraft: Previous Datacall saved!111111")
// 		logger.Error("Error commiting the previous DataCall")
// 		return shim.Error("Error commiting the previous DataCall")
// 	}
// 	logger.Info("SaveNewDraft: Previous Datacall saved!2222222")
// 	dataCallAsBytes, err := stub.GetState(countPatternKey)
// 	fmt.Println("SaveNewDraft: Previous Datacall saved!33333333 ")
// 	fmt.Println(dataCallAsBytes)
// 	if err != nil {
// 		logger.Info("SaveNewDraft: Previous Datacall saved!44444444444")
// 		return shim.Error("Error retreiving data for key")
// 	}

// 	var prevDataCall DataCallCount
// 	logger.Info("SaveNewDraft: Previous Datacall saved!55555555")
// 	err = json.Unmarshal(dataCallAsBytes, &prevDataCall)
// 	logger.Info("SaveNewDraft: Previous Datacall saved!66666666")
// 	if err != nil {
// 		logger.Info("SaveNewDraft: Previous Datacall saved!777777")
// 		return shim.Error("Failed to unmarshal data call: " + err.Error())
// 	}
// 	logger.Info("GOT THE RESPONSE ", prevDataCall)

// 	return shim.Success(nil)

// }

func (this *SmartContract) ToggleDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("ToggleDataCallCount: enter")
	defer logger.Debug("ToggleDataCallCount: exit")
	logger.Debug("ToggleDataCallCount json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var toggleDataCallCount ToggleDataCallCount
	err := json.Unmarshal([]byte(args), &toggleDataCallCount)
	if err != nil {
		logger.Error("ToggleDataCallCount: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("ToggleDataCallCount: Error during json.Unmarshal").Error())
	}
	logger.Info("Unmarshalled object ", toggleDataCallCount)

	getDataCallCount := GetDataCallCount{"123456", "1"}
	datacallIssueLogAsBytes, _ := json.Marshal(getDataCallCount)
	dataCallCountAsBytes := this.GetDataCallCount(stub, string(datacallIssueLogAsBytes))
	var dataCallCount DataCallCount
	err = json.Unmarshal(dataCallCountAsBytes.Payload, &dataCallCount)

	logger.Info("The retrieved data is ", dataCallCount)

	if toggleDataCallCount.OriginalStatus == "ISSUED" {
		dataCallCount.ISSUED = dataCallCount.ISSUED - 1
	} else if toggleDataCallCount.OriginalStatus == "DRAFT" {
		dataCallCount.DRAFT = dataCallCount.DRAFT - 1
	} else if toggleDataCallCount.OriginalStatus == "CANCELLED" {
		dataCallCount.CANCELLED = dataCallCount.CANCELLED - 1
	}

	if toggleDataCallCount.NewStatus == "ISSUED" {
		dataCallCount.ISSUED = dataCallCount.ISSUED + 1
	} else if toggleDataCallCount.NewStatus == "DRAFT" {
		dataCallCount.DRAFT = dataCallCount.DRAFT + 1
	} else if toggleDataCallCount.NewStatus == "CANCELLED" {
		dataCallCount.CANCELLED = dataCallCount.CANCELLED + 1
	}

	dataCallCount.Version = generateVersion(0)

	var pks []string = []string{DATA_CALLCOUNT_PREFIX, dataCallCount.ID, dataCallCount.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENTCOUNT_TYPE, pks)
	//dataCallKey := DATA_CALL_PREFIX + dataCall.ID + dataCall.Version
	logger.Info("In data call create ", dataCallKey)
	// Checking the ledger to confirm that the dataCall doesn't exist

	dataCallAsBytes, _ := json.Marshal(dataCallCount)
	err = stub.PutState(dataCallKey, dataCallAsBytes)

	if err != nil {
		logger.Error("Error commiting data for key: ", dataCallKey)
		return shim.Error("Error committing data for key: " + dataCallKey)
	}

	return shim.Success(nil)

}

func (this *SmartContract) GetDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallCount: enter")
	defer logger.Debug("GetDataCallCount: exit")

	var getDataCallCount GetDataCallCount
	err := json.Unmarshal([]byte(args), &getDataCallCount)
	if err != nil {
		logger.Error("GetDataCallCount: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallCount: Error during json.Unmarshal").Error())
	}
	logger.Debug("Unmarshalled object ", getDataCallCount)

	var pks []string = []string{DATA_CALLCOUNT_PREFIX, getDataCallCount.ID, getDataCallCount.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENTCOUNT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	if err != nil {
		logger.Error("Error retreiving data for key ", dataCallKey)
		return shim.Error("Error retreiving data for key" + dataCallKey)
	}
	return shim.Success(dataCallAsBytes)

}

func (this *SmartContract) UpdateDataCallCount(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateDataCallCount: enter")
	defer logger.Debug("UpdateDataCallCount: exit")
	logger.Debug("UpdateDataCallCount json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var dataCallCount DataCallCount
	err := json.Unmarshal([]byte(args), &dataCallCount)
	if err != nil {
		logger.Error("UpdateDataCallCount: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateDataCallCount: Error during json.Unmarshal").Error())
	}
	logger.Info("Unmarshalled object ", dataCallCount)

	dataCallCount.ID = "123456"
	dataCallCount.Version = generateVersion(0)

	var pks []string = []string{DATA_CALLCOUNT_PREFIX, dataCallCount.ID, dataCallCount.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENTCOUNT_TYPE, pks)
	//dataCallKey := DATA_CALL_PREFIX + dataCall.ID + dataCall.Version
	logger.Info("In data call create ", dataCallKey)
	// Checking the ledger to confirm that the dataCall doesn't exist

	dataCallAsBytes, _ := json.Marshal(dataCallCount)
	err = stub.PutState(dataCallKey, dataCallAsBytes)

	if err != nil {
		logger.Error("Error commiting data for key: ", dataCallKey)
		return shim.Error("Error committing data for key: " + dataCallKey)
	}

	return shim.Success(nil)

}

// SearchDataCalls retrives all data calls that match given criteria. If startindex and pageSize are not provided,
// this method returns the complete list of data calls. If version = latest, the it returns only latest version of a data call
// using the specified criteria. If version = all, it returns all data calls with their versions as individual items in list.
// params {json}: {
//  "startIndex":"optional",
//  "pageSize":"optional",
//  "version": "latest or all"
//  "status" :"DRAFT OR ISSUED OR CANCELLED"}
// Success {byte[]}: byte[]
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
func (this *SmartContract) SearchDataCalls(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("SearchDataCalls: enter")
	defer logger.Debug("SearchDataCalls: exit")
	logger.Debug("SearchDataCalls json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}
	var searchCriteria SearchCriteria
	err := json.Unmarshal([]byte(args), &searchCriteria)
	if err != nil {
		logger.Error("SearchDataCalls: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SearchDataCalls: Error during json.Unmarshal").Error())
	}
	logger.Debug("SearchDataCalls: Unmarshalled object ", searchCriteria)
	startIndex := searchCriteria.StartIndex
	pageSize := searchCriteria.PageSize
	version := searchCriteria.Version
	searchKey := searchCriteria.SearchKey
	var isLatest string
	if version == LATEST_VERSION {
		isLatest = "true"
	} else {
		isLatest = "false"
	}
	status := searchCriteria.Status
	if status == "" {
		logger.Error("SearchDataCalls: Status not present, You must pass Status in agument")
		return shim.Error("SearchDataCalls: Status not present, You must pass Status in agument")
	}
	var queryStr string

	if status == STATUS_ISSUED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\",\"$or\":[{\"name\":{\"$regex\":\"%s\"}},{\"description\":{\"$regex\":\"%s\"}},{\"lineOfBusiness\":{\"$regex\":\"%s\"}},{\"jurisdiction\":{\"$regex\":\"%s\"}}]},\"use_index\":[\"_design/deadline\"],\"sort\":[{\"deadline\": \"desc\"}]}", DATA_CALL_PREFIX, status, searchKey, searchKey, searchKey, searchKey)
	} else if status == STATUS_DRAFT {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\", \"isLatest\":%s, \"isLocked\":false, \"$or\":[{\"name\":{\"$regex\":\"%s\"}},{\"description\":{\"$regex\":\"%s\"}},{\"lineOfBusiness\":{\"$regex\":\"%s\"}},{\"jurisdiction\":{\"$regex\":\"%s\"}}]},\"use_index\":[\"_design/updatedTs\"],\"sort\":[{\"updatedTs\": \"desc\"}]}", DATA_CALL_PREFIX, status, isLatest, searchKey, searchKey, searchKey, searchKey)
	} else if status == STATUS_CANCELLED {
		queryStr = fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"status\":\"%s\", \"$or\":[{\"name\":{\"$regex\":\"%s\"}},{\"description\":{\"$regex\":\"%s\"}},{\"lineOfBusiness\":{\"$regex\":\"%s\"}},{\"jurisdiction\":{\"$regex\":\"%s\"}}]},\"use_index\":[\"_design/deadline\"],\"sort\":[{\"deadline\": \"desc\"}]}", DATA_CALL_PREFIX, status, searchKey, searchKey, searchKey, searchKey)
	}

	logger.Debug("SearchDataCalls: Selector ", queryStr)
	logger.Debug("startIndex", "pageSize", "version", "status", startIndex, pageSize, version, status)

	var dataCalls []DataCall
	startTime := time.Now()
	resultsIterator, err := stub.GetQueryResult(queryStr)
	elapsedTime := time.Since(startTime)
	logger.Info("Time consumed to get Data Calls", elapsedTime)
	defer resultsIterator.Close()
	if err != nil {
		logger.Error("Failed to get state for all the data calls")
		return shim.Error("Failed to get state for all the data calls")
	}

	if !resultsIterator.HasNext() {
		dataCallsAsByte, _ := json.Marshal(dataCalls)
		logger.Debug("SearchDataCalls: dataCallsAsByte", dataCallsAsByte)
		//return shim.Error(errors.New("SearchDataCalls :DataCall not found ").Error())
		return shim.Success(dataCallsAsByte)
	}

	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate data call")
			return shim.Error("Failed to iterate data call")
		}

		var dataCall DataCall
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &dataCall)
		if err != nil {
			return shim.Error("Failed to unmarshal data call: " + err.Error())
		}

		dataCalls = append(dataCalls, dataCall)
	}

	//var paginatedDataCall []DataCall
	var paginatedDataCalls []DataCall
	paginatedDataCalls = paginate(dataCalls, startIndex, pageSize)
	var dataCallList DataCallList
	dataCallList.TotalNoOfRecords = len(dataCalls)

	var IdAndVersionMap map[string]string
	IdAndVersionMap = make(map[string]string)
	var dataCallIDs []string
	var dataCallVersions []string
	for dataCallIndex := 0; dataCallIndex < len(paginatedDataCalls); dataCallIndex++ {
		if dataCallIndex == 0 {
			IdAndVersionMap[paginatedDataCalls[dataCallIndex].ID] = paginatedDataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `"`+paginatedDataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `"`+paginatedDataCalls[dataCallIndex].Version+`"`)
		} else {
			IdAndVersionMap[paginatedDataCalls[dataCallIndex].ID] = paginatedDataCalls[dataCallIndex].Version
			dataCallIDs = append(dataCallIDs, `,`+`"`+paginatedDataCalls[dataCallIndex].ID+`"`)
			dataCallVersions = append(dataCallVersions, `,`+`"`+paginatedDataCalls[dataCallIndex].Version+`"`)
		}
	}

	startTimeForAll := time.Now()
	/*
		//get ConsentCount map
		startTimeForConsent := time.Now()
		consentCounts := GetConsentsCount(stub, dataCallIDs)
		elapsedTimeForConsent := time.Since(startTimeForConsent)
		logger.Info("Time consumed for consent", elapsedTimeForConsent)

		startTimeForLike := time.Now()
		likeCounts := GetLikesCount(stub, dataCallIDs, IdAndVersionMap)
		elapsedTimeForLike := time.Since(startTimeForLike)
		logger.Info("Time consumed for Like", elapsedTimeForLike)
	*/

	startTimeForReport := time.Now()
	latestReports := GetLatestaReport(stub, dataCallIDs, dataCallVersions)
	elapsedTimeForReport := time.Since(startTimeForReport)
	logger.Info("Time consumed for Report", elapsedTimeForReport)

	elapsedTimeForAll := time.Since(startTimeForAll)
	logger.Info("Time consumed for all", elapsedTimeForAll)

	for dataCallIndex := 0; dataCallIndex < len(paginatedDataCalls); dataCallIndex++ {
		var dataCallExtended DataCallExtended
		dataCallExtended.DataCall = paginatedDataCalls[dataCallIndex]
		dataCallExtended.Reports = append(dataCallExtended.Reports, latestReports[paginatedDataCalls[dataCallIndex].ID])
		//dataCallExtended.NoOfConsents = consentCounts[paginatedDataCalls[dataCallIndex].ID]
		//dataCallExtended.NoOfLikes = likeCounts[paginatedDataCalls[dataCallIndex].ID]
		dataCallList.DataCalls = append(dataCallList.DataCalls, dataCallExtended)

	}

	dataCallsAsByte, _ := json.Marshal(dataCallList)
	return shim.Success(dataCallsAsByte)

}
