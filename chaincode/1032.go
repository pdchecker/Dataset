package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
)

// creates extraction patten definition
func (this *SmartContract) CreateExtractionPattern(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CreateExtractionPattern: enter")
	defer logger.Debug("CreateExtractionPattern: exit")
	logger.Debug("CreateExtractionPattern json received : ", args)
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments!!")
	}

	var extractionPatten ExtPattern
	err := json.Unmarshal([]byte(args), &extractionPatten)
	if extractionPatten.ExtractionPatternID == "" {
		return shim.Error("ExtractionPatternID cant not be empty!!")
	} else if extractionPatten.DbType == "" {
		return shim.Error("DbType cant not be empty!!")
	} else if extractionPatten.ViewDefinition.Map == "" || extractionPatten.ViewDefinition.Reduce == "" {
		return shim.Error("ViewDefinition cant be empty!!")
	} else if extractionPatten.PremiumFromDate == "" {
		return shim.Error("PremiumFromDate cannot not be Empty")
	} else if extractionPatten.LossFromDate == "" {
		return shim.Error("LossFromDate cannot not be Empty")
	} else if extractionPatten.Jurisdiction == "" {
		return shim.Error("Jurisdiction cannot not be Empty")
	} else if extractionPatten.Insurance == "" {
		return shim.Error("Insurance cannot not be Empty")
	}
	if err != nil {
		logger.Error("CreateExtractionPattern: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CreateExtractionPattern: Error during json.Unmarshal").Error())
	}

	logger.Debug("Unmarshalled object ", extractionPatten)

	extractionPatten.Version = generateVersion(0)
	namespace := EXTRACTION_PATTERN_PREFIX
	extPatternKey, _ := stub.CreateCompositeKey(namespace, []string{extractionPatten.ExtractionPatternID, extractionPatten.DbType})

	// Checking the ledger to confirm that the ExtractionPattern doesn't exist
	prevExtractionPattern, _ := stub.GetState(extPatternKey)

	if prevExtractionPattern != nil {
		logger.Error("CreateExtractionPattern: Extarction Pattern already exist with ID: " + extPatternKey)
		return shim.Error("CreateExtractionPattern:Extarction Pattern already exist with ID: " + extPatternKey)
	}

	extractionPatternAsBytes, _ := json.Marshal(extractionPatten)
	err = stub.PutState(extPatternKey, extractionPatternAsBytes)
	if err != nil {
		return shim.Error("CreateExtractionPattern: Failed to Put Extarction Pattern: " + err.Error())
	}

	return shim.Success(nil)

}

// updates an existing extraction pattern definition
func (this *SmartContract) UpdateExtractionPattern(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("UpdateExtractionPattern: enter")
	defer logger.Debug("UpdateExtractionPattern: exit")
	logger.Debug("UpdateExtractionPattern json received : ", args)

	var extractionPatten ExtPattern
	err := json.Unmarshal([]byte(args), &extractionPatten)
	if err != nil {
		logger.Error("UpdateExtractionPattern: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("UpdateExtractionPattern: Error during json.Unmarshal").Error())
	}
	if extractionPatten.ExtractionPatternID == "" {
		return shim.Error("ExtractionPatternID should not be Empty")

	} else if extractionPatten.DbType == "" {
		return shim.Error("DbType should not be Empty")

	} else if extractionPatten.PremiumFromDate == "" {
		return shim.Error("PremiumFromDate cannot not be Empty")
	} else if extractionPatten.LossFromDate == "" {
		return shim.Error("LossFromDate cannot not be Empty")
	} else if extractionPatten.Jurisdiction == "" {
		return shim.Error("Jurisdiction cannot not be Empty")
	} else if extractionPatten.Insurance == "" {
		return shim.Error("Insurance cannot not be Empty")
	}
	namespace := EXTRACTION_PATTERN_PREFIX
	extPatternKey, _ := stub.CreateCompositeKey(namespace, []string{extractionPatten.ExtractionPatternID, extractionPatten.DbType})
	extractionPatternAsBytes, err := stub.GetState(extPatternKey)
	if err != nil {
		logger.Error("UpdateExtractionPattern:Error retreiving extraction pattern for key: " + extPatternKey)
		return shim.Error("UpdateExtractionPattern: Error retreiving extraction pattern for key" + extPatternKey)
	}

	var prevExtractionPattern ExtPattern
	err = json.Unmarshal(extractionPatternAsBytes, &prevExtractionPattern)
	if err != nil {
		return shim.Error("UpdateExtractionPattern: Failed to unmarshal pattern: " + err.Error())
	}
	prevExtractionPattern.PremiumFromDate = extractionPatten.PremiumFromDate
	prevExtractionPattern.LossFromDate = extractionPatten.LossFromDate
	prevExtractionPattern.Jurisdiction = extractionPatten.Jurisdiction
	prevExtractionPattern.Insurance = extractionPatten.Insurance
	prevExtractionPattern.ViewDefinition = extractionPatten.ViewDefinition
	prevExtractionPattern.UpdatedTs = extractionPatten.UpdatedTs
	prevExtractionPattern.UpdatedBy = extractionPatten.UpdatedBy
	prevExtractionPattern.IsActive = extractionPatten.IsActive
	prevExtractionPattern.ExtractionPatternName = extractionPatten.ExtractionPatternName
	prevExtractionPattern.EffectiveStartTs = extractionPatten.EffectiveStartTs
	prevExtractionPattern.EffectiveEndTs = extractionPatten.EffectiveEndTs
	prevExtractionPattern.Description = extractionPatten.Description
	prevExtractionPatternVersion, _ := strconv.Atoi(prevExtractionPattern.Version)
	prevExtractionPattern.Version = generateVersion(prevExtractionPatternVersion)

	prevExtractionPatternAsBytes, _ := json.Marshal(prevExtractionPattern)
	err = stub.PutState(extPatternKey, prevExtractionPatternAsBytes)
	if err != nil {
		return shim.Error("UpdateExtractionPattern: Failed to Update Extraction Pattern: " + err.Error())
	}

	return shim.Success(nil)

}

// function returns extraction pattern based on id and dbtype
func (this *SmartContract) GetExtractionPatternById(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetExtractionPatternById: enter")
	defer logger.Debug("GetExtractionPatternById: exit")

	var getExtractionPatternById GetExtractionPatternById
	err := json.Unmarshal([]byte(args), &getExtractionPatternById)
	if err != nil {
		logger.Error("GetExtractionPatternById: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetExtractionPatternById: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetExtractionPatternById: Unmarshalled object ", getExtractionPatternById)

	if getExtractionPatternById.ExtractionPatternID == "" || getExtractionPatternById.DbType == "" {
		return shim.Error("GetExtractionPatternById: ExtractionPatternID and DbType can not be Empty")
	}

	namespace := EXTRACTION_PATTERN_PREFIX
	extPatternKey, _ := stub.CreateCompositeKey(namespace, []string{getExtractionPatternById.ExtractionPatternID, getExtractionPatternById.DbType})
	extractionPatternAsBytes, err := stub.GetState(extPatternKey)
	if err != nil {
		logger.Error("GetExtractionPatternById: Error retreiving data for key ", extPatternKey)
		return shim.Error("GetExtractionPatternById: Error retreiving data for key" + extPatternKey)
	}
	return shim.Success(extractionPatternAsBytes)

}

// returns required data call details and extraction pattern
func (this *SmartContract) GetDataCallAndExtractionPattern(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetDataCallAndExtractionPattern: enter")
	defer logger.Debug("GetDataCallAndExtractionPattern: exit")

	var getDataCallAndExtractionPattern GetDataCallAndExtractionPattern
	err := json.Unmarshal([]byte(args), &getDataCallAndExtractionPattern)
	if err != nil {
		logger.Error("GetDataCallAndExtractionPattern: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetDataCallAndExtractionPattern: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetDataCallAndExtractionPattern: Unmarshalled object ", getDataCallAndExtractionPattern)

	if getDataCallAndExtractionPattern.DataCallID == "" {
		return shim.Error("GetDataCallAndExtractionPattern: DataCallID can not be Empty")
	} else if getDataCallAndExtractionPattern.DataCallVersion == "" {
		return shim.Error("GetDataCallAndExtractionPattern: DataCallVersion can not be Empty")
	} else if getDataCallAndExtractionPattern.DbType == "" {
		return shim.Error("GetDataCallAndExtractionPattern: DbType can not be Empty")
	}

	// invoke GetDataCallByIdAndVersion to get dataCall details
	var getDataCall GetDataCall
	getDataCall.ID = getDataCallAndExtractionPattern.DataCallID
	getDataCall.Version = getDataCallAndExtractionPattern.DataCallVersion
	getDataCallJson, _ := json.Marshal(getDataCall)
	getDataCallResponse := this.GetDataCallByIdAndVersion(stub, string(getDataCallJson))
	if getDataCallResponse.Status != 200 || len(getDataCallResponse.Payload) <= 0 {
		logger.Error("GetDataCallAndExtractionPattern: Unable to GetDataCallByIdAndVersion: ", getDataCallResponse.Message)
		return shim.Error(errors.New("GetDataCallAndExtractionPattern: Unable to GetDataCallByIdAndVersion").Error())
	}

	var dataCall DataCall
	json.Unmarshal(getDataCallResponse.Payload, &dataCall)
	logger.Debug("GetDataCallAndExtractionPattern: InvokeResponse from GetDataCallByIdAndVersion: Name ", string(dataCall.Name))

	// setting jurisdiction in dataCallAndExtractionPatternResponse
	var dataCallAndExtractionPatternResponse DataCallAndExtractionPatternResponse
	dataCallAndExtractionPatternResponse.Jurisdiction = dataCall.Jurisdiction

	//check whether extraction pattern is set or not
	if dataCall.ExtractionPatternID == "" {
		dataCallAndExtractionPatternResponse.IsSet = false
	} else {
		dataCallAndExtractionPatternResponse.IsSet = true

		//invoke GetExtractionPatternById to get extraction pattern details
		var getExtractionPatternById GetExtractionPatternById
		getExtractionPatternById.ExtractionPatternID = dataCall.ExtractionPatternID
		getExtractionPatternById.DbType = getDataCallAndExtractionPattern.DbType
		getExtractionPatternByIdJson, _ := json.Marshal(getExtractionPatternById)
		getExtractionPatternByIdResponse := this.GetExtractionPatternById(stub, string(getExtractionPatternByIdJson))
		if getExtractionPatternByIdResponse.Status != 200 || len(getExtractionPatternByIdResponse.Payload) <= 0 {
			logger.Error("GetDataCallAndExtractionPattern: Unable to GetExtractionPatternById: ", getExtractionPatternByIdResponse.Message)
			return shim.Error(errors.New("GetDataCallAndExtractionPattern: Unable to GetExtractionPatternById").Error())
		}
		var extractionPattern ExtPattern
		json.Unmarshal(getExtractionPatternByIdResponse.Payload, &extractionPattern)
		logger.Debug("GetDataCallAndExtractionPattern: InvokeResponse from GetExtractionPatternById: Id ", string(extractionPattern.ExtractionPatternID))
		dataCallAndExtractionPatternResponse.ExtractionPattern = extractionPattern
	}

	dataCallAndExtractionPatternResponseAsBytes, _ := json.Marshal(dataCallAndExtractionPatternResponse)
	return shim.Success(dataCallAndExtractionPatternResponseAsBytes)

}

// GetExtractionPatternsMap simply returns a dictionary/map that contains the pre-defined extraction-patterns
// for the chaincode.
func GetExtractionPatternsMap() map[string]ExtractionPattern {
	var patterns map[string]ExtractionPattern
	patterns = make(map[string]ExtractionPattern)

	//patterns["Pattern_01"] = ExtractionPattern{ID: "Pattern_01", Name: "Standard Annual Homeowners Report", Description: "Provides summations of written premiums and exposures within groupings of ZIP code, policy form code, and property amount of insurance. This extraction pattern is for the reporting year 2017 and returns aggregate data for policy forms 01, 02, 03, 05 and 08.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_01, Group: true}}
	//patterns["Pattern_02"] = ExtractionPattern{ID: "Pattern_02", Name: "ILCC Data Extraction Pattern", Description: "Provides summations of written premiums and exposures within groupings of ZIP code, policy form code, and property amount of insurance. This extraction pattern is for the reporting year 2017 and returns aggregate data for policy forms 01, 02, 03, 05 and 08.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_02, Group: true}}
	//patterns["Pattern_03"] = ExtractionPattern{ID: "Pattern_03", Name: "New ILCC Data Extraction Pattern", Description: "Provides summations of written premiums and exposures within groupings of ZIP code, policy form code, and property amount of insurance. This extraction pattern is for the reporting year 2017 and returns aggregate data for policy forms 01, 02, 03, 05 and 08.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_03, Group: true}}
	//patterns["Pattern_04"] = ExtractionPattern{ID: "Pattern_04", Name: "Industry Test Drive Extraction Pattern", Description: "Provides summarizations of written premiums, exposures and paid loss within the corresponding state, grouped by zip code, liability limit, and cause of loss. This extraction pattern is for the reporting years 2016 – 2018 and returns aggregate data to interrogate causes of loss over liability limits.", CouchDBView: CouchDBView{Definition: EXT_PATTERN_VIEW_04, Group: true}}
	return patterns
}

func (this *SmartContract) ListExtractionPatterns(stub shim.ChaincodeStubInterface) pb.Response {
	logger.Debug("ListExtractionPatterns: enter")
	defer logger.Debug("ListExtractionPatterns: exit")
	var patterns []ExtPattern
	queryStr := fmt.Sprintf("{\"selector\":{\"_id\":{\"$regex\":\"%s\"},\"isActive\":true}}", EXTRACTION_PATTERN_PREFIX)
	resultsIterator, err := stub.GetQueryResult(queryStr)

	if err != nil {
		logger.Error("ListExtractionPatterns: Failed to get extraction patterns")
		return shim.Error("ListExtractionPatterns: Failed to get extraction patterns : " + err.Error())
	}
	defer resultsIterator.Close()
	logger.Debug("ListExtractionPatterns: Iterating over extraction patterns")
	for resultsIterator.HasNext() {
		dataCallAsBytes, err := resultsIterator.Next()
		if err != nil {
			logger.Error("Failed to iterate over extraction patterns")
			return shim.Error("Failed to iterate over extraction patterns")
		}

		var pattern ExtPattern
		err = json.Unmarshal([]byte(dataCallAsBytes.GetValue()), &pattern)
		logger.Debug("ListExtractionPatterns: DataCall > ", pattern.ExtractionPatternID)
		if err != nil {
			return shim.Error("ListExtractionPatterns: Failed to unmarshal extraction patterns: " + err.Error())
		}
		patterns = append(patterns, pattern)

	}

	patternsAsBytes, _ := json.Marshal(patterns)
	logger.Info("ListExtractionPatterns: ExtractionPatterns", patterns)

	return shim.Success(patternsAsBytes)

}

func (this *SmartContract) GetExtractionPatternByIds(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("GetExtractionPatternByIds: enter")
	defer logger.Debug("GetExtractionPatternByIds: exit")
	var patternIds ExtractionPatternId
	var patterns []ExtractionPattern

	err := json.Unmarshal([]byte(args), &patternIds)
	logger.Debug("GetExtractionPatternByIds: Incoming args", args)
	if err != nil {
		logger.Error("GetExtractionPatternByIds: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("GetExtractionPatternByIds: Error during json.Unmarshal").Error())
	}
	logger.Debug("GetExtractionPatternByIds: Unmarshalled object ", patternIds)

	patternsMap := GetExtractionPatternsMap()

	for _, value := range patternIds.Id {
		if patternValue, found := patternsMap[value]; found {
			logger.Info("Extraction Pattern value found for key", value)
			patterns = append(patterns, patternValue)
		}
	}
	patternsAsBytes, _ := json.Marshal(patterns)
	logger.Info("GetExtractionPatternByIds: ExtractionPatterns", patterns)

	return shim.Success(patternsAsBytes)

}

func (this *SmartContract) CheckExtractionPatternIsSet(stub shim.ChaincodeStubInterface, args string) pb.Response {
	logger.Debug("CheckExtractionPatternIsSet: enter")
	defer logger.Debug("CheckExtractionPatternIsSet: exit")

	var isSet bool
	var getDataCall GetDataCall
	err := json.Unmarshal([]byte(args), &getDataCall)
	logger.Debug("CheckExtractionPatternIsSet: Incoming args", args)
	if err != nil {
		logger.Error("CheckExtractionPatternIsSet: Error during json.Unmarshal: ", err)
		return shim.Error(errors.New("CheckExtractionPatternIsSet: Error during json.Unmarshal").Error())
	}
	logger.Debug("CheckExtractionPatternIsSet: Unmarshalled object ", getDataCall)

	if getDataCall.ID == "" || getDataCall.Version == "" {
		return shim.Error("ID and Version can not be Empty")
	}

	getDataCallAsBytes, _ := json.Marshal(getDataCall)
	getDataCallReqJson := string(getDataCallAsBytes)

	var GetDataCallByIdAndVersionFunc = "GetDataCallByIdAndVersion"
	getDataCallRequest := ToChaincodeArgs(GetDataCallByIdAndVersionFunc, getDataCallReqJson)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallRequest", getDataCallRequest)
	getDataCallResponse := stub.InvokeChaincode(DEFAULT_CHAINCODE_NAME, getDataCallRequest, DEFAULT_CHANNEL)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallResponse > ", getDataCallResponse)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallResponse.Status ", getDataCallResponse.Status)
	logger.Debug("CheckExtractionPatternIsSet: getDataCallResponse.Payload", string(getDataCallResponse.Payload))
	if getDataCallResponse.Status != 200 {
		logger.Error("CheckExtractionPatternIsSet: Unable to Fetch DataCall due to Error: ", err)
		return shim.Error(errors.New("CheckExtractionPatternIsSet: Unable to Fetch DataCall due to Error").Error())
	}

	if len(getDataCallResponse.Payload) <= 0 {
		logger.Error("CheckExtractionPatternIsSet: DataCall Doesnt exist")
		return shim.Error(errors.New("CheckExtractionPatternIsSet:  DataCall Doesnt exist").Error())

	}

	/*var pks []string = []string{DATA_CALL_PREFIX, getDataCall.ID, getDataCall.Version}
	dataCallKey, _ := stub.CreateCompositeKey(DOCUMENT_TYPE, pks)
	dataCallAsBytes, err := stub.GetState(dataCallKey)
	logger.Debug("CheckExtractionPatternIsSet: key ", dataCallKey, "value", dataCallAsBytes)

	if err != nil {
		logger.Error("CheckExtractionPatternIsSet: Error retreiving data for key ", dataCallKey)
		return shim.Error("CheckExtractionPatternIsSet: Error retreiving data for key" + dataCallKey)
	} else if len(dataCallAsBytes) == 0 {
		logger.Error("CheckExtractionPatternIsSet: DataCall Doesnt exist")
		return shim.Error("CheckExtractionPatternIsSet: DataCall Doesnt exist")
	}*/
	var dataCall DataCall
	errMsg := json.Unmarshal(getDataCallResponse.Payload, &dataCall)
	if errMsg != nil {
		logger.Error("CheckExtractionPatternIsSet: Error during json.Unmarshal for response: ", errMsg)
		return shim.Error(errors.New("CheckExtractionPatternIsSet: Error during json.Unmarshal for response: ").Error())
	}
	if dataCall.ExtractionPatternID != "" {
		isSet = true
	}

	//get extraction pattern by ExtractionPatternID
	var extractionPattern ExtractionPattern
	if isSet {
		patternsMap := GetExtractionPatternsMap()
		extractionPattern = patternsMap[dataCall.ExtractionPatternID]
	}
	var extractionPatternIsSetPayload ExtractionPatternIsSetPayload
	extractionPatternIsSetPayload.IsSet = isSet
	extractionPatternIsSetPayload.ExtractionPattern = extractionPattern
	responseAsBytes, _ := json.Marshal(extractionPatternIsSetPayload)
	logger.Debug("CheckExtractionPatternIsSet: responseAsBytes ", responseAsBytes)
	return shim.Success(responseAsBytes)

}
