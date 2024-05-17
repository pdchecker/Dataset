package main

import (
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
)

func (this *SmartContract) SaveInsuranceDataHash(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Info("SaveInsuranceDataHash: enter")
	defer logger.Debug("SaveInsuranceDataHash: exit")
	var insurance InsuranceDataHash

	err := json.Unmarshal([]byte(args), &insurance)
	if err != nil {
		logger.Error("SaveInsuranceDataHash: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveInsuranceDataHash: Error during json.Unmarshal").Error())
	}
	if insurance.BatchId == "" {
		return shim.Error("BatchId should not be Empty")
	} else if insurance.Hash == "" {
		return shim.Error("Hash should not be Empty")
	} else if insurance.CarrierId == "" {
		return shim.Error("CarrierId should not be Empty")
	} else if insurance.ChunkId == "" {
		return shim.Error("ChunkId should not be Empty")
	}

	namespacePrefix := INSURANCE_HASH_PREFIX
	//var pks []string = []string{INSURANCE_HASH_PREFIX, insurance.CarrierId, insurance.BatchId}
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{insurance.CarrierId, insurance.BatchId, insurance.ChunkId})

	//key := insurance.BatchId
	insuranceDataAsBytes, _ := json.Marshal(insurance)
	err = stub.PutState(key, insuranceDataAsBytes)

	if err != nil {
		return shim.Error("SaveInsuranceDataHash: Error committing data for key: " + key)
	}

	return shim.Success(nil)
}

// function-name: GetHashById (invoke)
// params {json}: {
// "id":"mandatory"}
// Success {byte[]}: byte[]  - Report
// Error   {json}:{"message":"....","errorCode":"Sys_Err/Bus_Err"}
// Description : returns a InsuranceHashRecord of specifc batchid

func (this *SmartContract) GetHashById(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetHashById: enter")
	defer logger.Debug("GetHashById: exit")

	var insurance InsuranceDataHash
	err := json.Unmarshal([]byte(args), &insurance)
	if err != nil {
		logger.Error("GetHashById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetHashById: Error during json.Unmarshal").Error())
	}
	namespacePrefix := INSURANCE_HASH_PREFIX
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{insurance.CarrierId, insurance.BatchId, insurance.ChunkId})
	if key == "" {
		return shim.Error("GetHashById:BatchId should not be empty")
	}

	insuranceHashAsBytes, err := stub.GetState(key)
	if err != nil {
		logger.Error("Error retreiving data for key ", key)
		return shim.Error("Error retreiving data for key" + key)
	}
	return shim.Success(insuranceHashAsBytes)

}

//this function puts the Insurance data extracted based on extraction pattern into private data collection
func (this *SmartContract) SaveInsuranceData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	logger.Info("SaveInsuranceData: enter")
	defer logger.Info("SaveInsuranceData: exit")
	var insurance InsuranceData
	transientMapKey := INSURANCE_TRANSACTIONAL_RECORD_PREFIX

	//getting trnasientMap data (using INSURANCE_TRANSACTIONAL_RECORD_PREFIX as key)
	InsuranceDataTransMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("SaveInsuranceData: Error getting InsuranceDataTransMap: " + err.Error())
	}
	if _, ok := InsuranceDataTransMap[transientMapKey]; !ok {
		return shim.Error("SaveInsuranceData: Invalid key in the transient map")
	}
	err = json.Unmarshal([]byte(InsuranceDataTransMap[transientMapKey]), &insurance)
	logger.Info("SaveInsuranceData: got transient map")

	if err != nil {
		logger.Error("SaveInsuranceData: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("SaveInsuranceData: Error during json.Unmarshal").Error())
	}
	pageNumber := insurance.PageNumber
	pageNumberAsString := strconv.Itoa(pageNumber)

	if insurance.CarrierId == "" {
		return shim.Error("CarrierId should not be Empty")
	} else if insurance.DataCallId == "" {
		return shim.Error("DataCallId should not be Empty")
	} else if insurance.DataCallVersion == "" {
		return shim.Error("DataCallVersion should not be Empty")
	} else if pageNumber == 0 {
		return shim.Error("PageNumber should not be Empty")
	}

	logger.Info("SaveInsuranceData: all necessary params found")
	//Identify the pdc name based on channelName
	channelName := stub.GetChannelID()
	private_data_collection := getPDCNameByChannelName(channelName)

	namespacePrefix := INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{insurance.DataCallId, insurance.DataCallVersion, insurance.CarrierId, pageNumberAsString})

	insuranceDataAsBytes, _ := json.Marshal(insurance)
	err = stub.PutPrivateData(private_data_collection, key, insuranceDataAsBytes)

	logger.Info("SaveInsuranceData: put private data done")
	if err != nil {
		logger.Error("Error commiting pdc data:", err)
		return shim.Error("SaveInsuranceData: Error committing data for key: " + key)
	}

	//insurance data has been ingested, now creating audit record
	var auditRecord InsuranceRecordAudit
	auditRecord.DataCallId = insurance.DataCallId
	auditRecord.DataCallVersion = insurance.DataCallVersion
	auditRecord.CarrierId = insurance.CarrierId

	namespacePrefixForAudit := AUDIT_INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	auditRecordKey, _ := stub.CreateCompositeKey(namespacePrefixForAudit, []string{auditRecord.DataCallId, auditRecord.DataCallVersion, auditRecord.CarrierId})

	auditRecordAsBytes, _ := json.Marshal(auditRecord)
	err = stub.PutState(auditRecordKey, auditRecordAsBytes)
	logger.Info("SaveInsuranceData: put audit key done")

	if err != nil {
		return shim.Error("SaveInsuranceData: Creating Audit Record: Error committing data for key: " + auditRecordKey)
	}

	//audit record has been created now firing chaincode event -->TransactionalDataAvailable
	var eventPayload InsuranceRecordEventPayload
	eventPayload.ChannelName = channelName
	eventPayload.DataCallId = insurance.DataCallId
	eventPayload.DataCallVersion = insurance.DataCallVersion
	eventPayload.CarrierId = insurance.CarrierId
	eventPayload.PageNumber = insurance.PageNumber

	eventPayloadAsBytes, _ := json.Marshal(eventPayload)
	err = stub.SetEvent(INSURANCE_RECORD_AND_AUDIT_CREATED_EVENT, eventPayloadAsBytes)
	if err != nil {
		return shim.Error("SaveInsuranceData: error during creating event")
	}

	logger.Info("SaveInsuranceData: set event done")
	return shim.Success(nil)
}

// this function returns true if InsuranceData exists in pdc for a data call else returns false
func (this *SmartContract) CheckInsuranceDataExists(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CheckInsuranceDataExists: enter")
	defer logger.Debug("CheckInsuranceDataExists: exit")

	var isExist bool
	var getInsuranceData InsuranceRecordAudit
	err := json.Unmarshal([]byte(args), &getInsuranceData)
	if err != nil {
		logger.Error("CheckInsuranceDataExists: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CheckInsuranceDataExists: Error during json.Unmarshal").Error())
	}
	logger.Debug("CheckInsuranceDataExists: Unmarshalled object ", getInsuranceData)

	if getInsuranceData.CarrierId == "" {
		return shim.Error("CarrierId can not be Empty")
	} else if getInsuranceData.DataCallId == "" {
		return shim.Error("DataCallId can not be Empty")
	} else if getInsuranceData.DataCallVersion == "" {
		return shim.Error("DataCallVersion can not be Empty")
	}
	namespace := AUDIT_INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	key, _ := stub.CreateCompositeKey(namespace, []string{getInsuranceData.DataCallId, getInsuranceData.DataCallVersion, getInsuranceData.CarrierId})
	insuranceDataAsBytes, err := stub.GetState(key)
	if err != nil {
		logger.Error("CheckInsuranceDataExists: Error retreiving Insurance data for key ", key)
		return shim.Error("CheckInsuranceDataExists: Error retreiving Insurance data for key" + key)
	}
	if len(insuranceDataAsBytes) > 0 {
		isExist = true
	}
	return shim.Success([]byte(strconv.FormatBool(isExist)))
}

// this function returns Insurance Data fetching from pdc
func (this *SmartContract) GetInsuranceData(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetInsuranceData: enter")
	defer logger.Debug("GetInsuranceData: exit")

	var getInsuranceData GetInsuranceData
	//var insuranceData []InsuranceData
	logger.Debug("args", args)
	err := json.Unmarshal([]byte(args), &getInsuranceData)
	if err != nil {
		logger.Error("GetInsuranceData: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetInsuranceData: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetInsuranceData: Unmarshalled object ", getInsuranceData)

	if getInsuranceData.CarrierId == "" {
		return shim.Error("CarrierId can not be Empty")
	} else if getInsuranceData.DataCallId == "" {
		return shim.Error("DataCallId can not be Empty")
	} else if getInsuranceData.DataCallVersion == "" {
		return shim.Error("DataCallVersion can not be Empty")
	} else if getInsuranceData.ChannelName == "" {
		return shim.Error("ChannelName can not be Empty")
	}
	//startIndex := getInsuranceData.StartIndex
	//pageSize := getInsuranceData.PageSize
	pageNumber := getInsuranceData.PageNumber
	pageNumberAsString := strconv.Itoa(pageNumber)
	namespacePrefix := INSURANCE_TRANSACTIONAL_RECORD_PREFIX
	key, _ := stub.CreateCompositeKey(namespacePrefix, []string{getInsuranceData.DataCallId, getInsuranceData.DataCallVersion, getInsuranceData.CarrierId, pageNumberAsString})
	//Identify the pdc name based on channelID
	channelName := stub.GetChannelID()
	private_data_collection := getPDCNameByChannelName(channelName)

	insuranceDataResponseAsBytes, err := stub.GetPrivateData(private_data_collection, key)
	if err != nil {
		logger.Error("GetInsuranceData: Failed to get Insurance Data due to error", err)
		return shim.Error("GetInsuranceData: Failed to get Insurance Data")
	}
	return shim.Success(insuranceDataResponseAsBytes)
}

// getPDCNameByChannelName is a helper function for getting PDC name
func getPDCNameByChannelName(channelName string) string {
	pdcName := strings.Replace(channelName, "-", "_", -1) + "_pdc"
	return pdcName
}
