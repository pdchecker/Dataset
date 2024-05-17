package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"reflect"
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/msp"
	logger "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// here we are implementing basic testing for two functions ListDataCallsByCriteria and GetDataCallVersionsById
// as these functions are using GetQueryResult method which is
// Not implemented since the mock engine does not have a query engine.

//Test for CreateDataCall
//This function tests whether it returns error for empty Id.
//This function tests whether it returns 200 for Success.
//This function tests whether it creates CreateDataCall.
func Test_CreateDataCall_Should_Create_Datacall_when_Datacall_Does_not_Exists(t *testing.T) {
	fmt.Println("Test_CreateDataCall_Should_Create_Datacall_when_Datacall_Does_not_Exists")
	scc := new(SmartContract)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)
	sid := &msp.SerializedIdentity{Mspid: "aaismsp", IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}

	stub.Creator = b

	//test for CreateDataCall
	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_createDataCall := checkInvoke_forError(t, stub, "CreateDataCall", []byte(CREATE_DATA_CALL_EMPTY_ID_JSON))
	var err_message_for_create = res_err_createDataCall.Message
	if res_err_createDataCall.Status != shim.OK {
		assert.Equal(t, "Id cant not be empty!!", err_message_for_create, "Test_CreateDataCall: For Empty Id")
	} else {
		t.FailNow()
	}

	//Step-2: For SUCCESS- Check whether it returns 200
	res_createDataCall := checkInvoke(t, stub, "CreateDataCall", []byte(CREATE_DATA_CALL_FOR_UPDATE_JSON))
	if res_createDataCall.Status != shim.OK {
		fmt.Println("CreateDataCall failed with message res.Message: ", string(res_createDataCall.Message))
		t.FailNow()
	}
	var create_returnCode = int(res_createDataCall.Status)
	assert.Equal(t, 200, create_returnCode, "Test_CreateDataCall: Function's success, status code 200.")

	//Step-3: For SUCCESS- Check whether input and output are identical
	res_getByIdAndVersion := checkInvoke(t, stub, "GetDataCallByIdAndVersion", []byte(GET_DATA_CALL_BY_ID_AND_VERSION_VALID_JSON))
	var input_create DataCall
	json.Unmarshal([]byte(CREATE_DATA_CALL_FOR_UPDATE_JSON), &input_create)
	var output_create DataCall
	err_create := json.Unmarshal(res_getByIdAndVersion.Payload, &output_create)
	if err_create != nil {
		logger.Error("Test_CreateDataCall: Error during json.Unmarshal for GetDataCallByIdAndVersion: ", err_create)
		t.FailNow()
	}
	assert.True(t, reflect.DeepEqual(input_create, output_create))

}

//Test for UpdateDataCall
//This function tests whether it returns error for empty Id.
//This function tests whether it returns 200 for Success.
//This function tests whether it  updates status.
func Test_UpdateDataCall_Should_Update_An_Existing_Datacall(t *testing.T) {
	fmt.Println("Test_UpdateDataCall_Should_Update_An_Existing_Datacall")
	scc := new(SmartContract)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)
	sid := &msp.SerializedIdentity{Mspid: "aaismsp", IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}

	stub.Creator = b

	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_updateDataCall := checkInvoke_forError(t, stub, "UpdateDataCall", []byte(UPDATE_DATA_CALL_EMPTY_ID_JSON))
	var err_message = res_err_updateDataCall.Message
	if res_err_updateDataCall.Status != shim.OK {
		assert.Equal(t, "ID should not be Empty", err_message, "Test_UpdateDataCall: For Empty Id")
	} else {
		t.FailNow()
	}

	//Step-2: For SUCCESS- Check whether it returns 200
	//firstly create the DataCall for Update, else it will give error while fething the record to update
	checkInvoke(t, stub, "CreateDataCall", []byte(CREATE_DATA_CALL_FOR_UPDATE_JSON))
	res_updateDataCall := checkInvoke(t, stub, "UpdateDataCall", []byte(UPDATE_DATA_CALL_VALID_JSON))
	if res_updateDataCall.Status != shim.OK {
		fmt.Println("UpdateDataCall failed with message res.Message: ", string(res_updateDataCall.Message))
		t.FailNow()
	}

	var update_returnCode = int(res_updateDataCall.Status)
	assert.Equal(t, 200, update_returnCode, "Test_UpdateDataCall: Function's success, status code 200.")

	//Step-3: For SUCCESS- Check whether input and output are identical
	res_getByIdAndVersion := checkInvoke(t, stub, "GetDataCallByIdAndVersion", []byte(GET_DATA_CALL_BY_ID_AND_VERSION_VALID_JSON))
	var input_update DataCall
	json.Unmarshal([]byte(UPDATE_DATA_CALL_VALID_JSON), &input_update)
	var output_update DataCall
	err_update := json.Unmarshal(res_getByIdAndVersion.Payload, &output_update)
	if err_update != nil {
		logger.Error("Test_UpdateDataCall: Error during json.Unmarshal for GetDataCallByIdAndVersion: ", err_update)
		t.FailNow()
	}

	assert.True(t, reflect.DeepEqual(input_update, output_update))
}

//Test for IssueDataCall
//This function tests whether it returns error for empty Id.
//This function tests whether it returns 200 for Success.
//This function tests whether it sets status to ISSUED.
func Test_IssueDataCall_Should_Issue_A_Datacall(t *testing.T) {
	fmt.Println("Test_IssueDataCall_Should_Issue_A_Datacall")
	scc := new(SmartContract)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)
	sid := &msp.SerializedIdentity{Mspid: "aaismsp", IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b

	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_issueDataCall := checkInvoke_forError(t, stub, "IssueDataCall", []byte(ISSUE_DATA_CALL_EMPTY_ID_JSON))
	var err_message_for_issue = res_err_issueDataCall.Message
	if res_err_issueDataCall.Status != shim.OK {
		assert.Equal(t, "ID should not be Empty", err_message_for_issue, "Test_IssueDataCall : For Empty Id")
	} else {
		t.FailNow()
	}

	//Step-2: For SUCCESS- Check whether it returns 200
	checkInvoke(t, stub, "CreateDataCall", []byte(CREATE_DATA_CALL_FOR_ISSUE_JSON))
	res_issueDataCall := checkInvoke(t, stub, "IssueDataCall", []byte(ISSUE_DATA_CALL_VALID_JSON))
	if res_issueDataCall.Status != shim.OK {
		fmt.Println("IssueDataCall failed with message res.Message: ", string(res_issueDataCall.Message))
		t.FailNow()
	}
	var issue_returnCode = int(res_issueDataCall.Status)
	assert.Equal(t, 200, issue_returnCode, "Test_IssueDataCall: Function's success, status code 200.")

	//Step-3: For SUCCESS- Check whether Status is set
	res_getByIdAndVersion := checkInvoke(t, stub, "GetDataCallByIdAndVersion", []byte(GET_DATA_CALL_BY_ID_AND_VERSION_FOR_ISSUE_VALID_JSON))
	var input_issueDataCall DataCall
	json.Unmarshal([]byte(ISSUE_DATA_CALL_VALID_JSON), &input_issueDataCall)
	var output_issueDataCall DataCall
	err_issueDataCall := json.Unmarshal(res_getByIdAndVersion.Payload, &output_issueDataCall)
	if err_issueDataCall != nil {
		logger.Error("Test_IssueDataCall: Error during json.Unmarshal for GetDataCallByIdAndVersion: ", err_issueDataCall)
		t.FailNow()
	}
	assert.EqualValues(t, input_issueDataCall.Status, output_issueDataCall.Status, "Test_IssueDataCall: function's success")
	assert.True(t, reflect.DeepEqual(input_issueDataCall, output_issueDataCall))
}

//Test for SaveNewDraft
//This function tests whether it returns error for empty Id.
//This function tests whether it returns 200 for Success.
//This function tests whether it creates new version of Datacall.
func Test_SaveNewDraft_Should_Save_An_Existing_Datacall_With_New_Version(t *testing.T) {
	fmt.Println("Test_SaveNewDraft_Should_Save_An_Existing_Datacall_With_New_Version")
	scc := new(SmartContract)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)
	sid := &msp.SerializedIdentity{Mspid: "aaismsp", IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b


	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_saveNewDraft := checkInvoke_forError(t, stub, "SaveNewDraft", []byte(SAVE_NEW_DRAFT_EMPTY_ID_JSON))
	var err_message_for_saveNewDraft = res_err_saveNewDraft.Message
	if res_err_saveNewDraft.Status != shim.OK {
		assert.Equal(t, "Id can not be empty", err_message_for_saveNewDraft, "Test_SaveNewDraft: For Empty Id")
	} else {
		t.FailNow()
	}

	//Step-2: For SUCCESS- Check whether it returns 200
	checkInvoke(t, stub, "CreateDataCall", []byte(CREATE_DATA_CALL_FOR_UPDATE_JSON))
	res_saveNewDraft := checkInvoke(t, stub, "SaveNewDraft", []byte(SAVE_NEW_DRAFT_JSON))
	if res_saveNewDraft.Status != shim.OK {
		fmt.Println("SaveNewDraft failed with message res.Message: ", string(res_saveNewDraft.Message))
		t.FailNow()
	}
	var saveDraft_returnCode = int(res_saveNewDraft.Status)
	assert.Equal(t, 200, saveDraft_returnCode, "Test_SaveNewDraft: Function's success, status code 200.")

	//Step-3: For SUCCESS- Check whether Version is changed
	res_getByIdAndVersion := checkInvoke(t, stub, "GetDataCallByIdAndVersion", []byte(GET_DATA_CALL_BY_ID_AND_VERSION_FOR_DRAFT_JSON))
	var input_saveNewDraft DataCall
	json.Unmarshal([]byte(SAVE_NEW_DRAFT_JSON), &input_saveNewDraft)

	//increase the version for input_saveNewDraft else it will return error while asserting
	updatedVersion, _ := strconv.Atoi(input_saveNewDraft.Version)
	input_saveNewDraft.Version = strconv.Itoa(updatedVersion + 1)

	var output_saveNewDraft DataCall
	err_saveNewDraft := json.Unmarshal(res_getByIdAndVersion.Payload, &output_saveNewDraft)
	if err_saveNewDraft != nil {
		logger.Error("Test_UpdateDataCall: Error during json.Unmarshal for GetDataCallByIdAndVersion: ", err_saveNewDraft)
		t.FailNow()
	}

	assert.True(t, reflect.DeepEqual(input_saveNewDraft, output_saveNewDraft))
}

//Test for GetDataCallByIdAndVersion
//This function tests whether it returns error for empty Id.
//This function tests whether it retrives DataCall for paricular ID and version
func Test_GetDataCallByIdAndVersion_Should_Return_Datacall_With_Input_Datacall_ID_And_Version(t *testing.T) {
	fmt.Println("Test_GetDataCallByIdAndVersion_Should_Return_Datacall_With_Input_Datacall_ID_And_Version")
	scc := new(SmartContract)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)
	sid := &msp.SerializedIdentity{Mspid: "aaismsp", IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b

	//test for GetDataCallByIdAndVersion
	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_getByIdAndVersion := checkInvoke_forError(t, stub, "GetDataCallByIdAndVersion", []byte(GET_DATA_CALL_BY_ID_AND_VERSION_EMPTY_ID_JSON))
	var err_message_for_getByIdAndVersion = res_err_getByIdAndVersion.Message
	if res_err_getByIdAndVersion.Status != shim.OK {
		assert.Equal(t, "ID and Version can not be Empty", err_message_for_getByIdAndVersion, "Test_GetDataCallByIdAndVersion: For Empty Id")
	} else {
		t.FailNow()
	}

	//Step-2 : check whether it returns the required DataCall
	checkInvoke(t, stub, "CreateDataCall", []byte(CREATE_DATA_CALL_FOR_UPDATE_JSON))
	res_getByIdAndVersion := checkInvoke(t, stub, "GetDataCallByIdAndVersion", []byte(GET_DATA_CALL_BY_ID_AND_VERSION_VALID_JSON))
	if res_getByIdAndVersion.Status != shim.OK {
		fmt.Println("GetDataCallByIdAndVersion failed with message res.Message: ", string(res_getByIdAndVersion.Message))
		t.FailNow()
	}
	var input_getByIdAndVersion DataCall
	json.Unmarshal([]byte(CREATE_DATA_CALL_FOR_UPDATE_JSON), &input_getByIdAndVersion)
	var output_getByIdAndVersion DataCall
	err_getByIdAndVersion := json.Unmarshal(res_getByIdAndVersion.Payload, &output_getByIdAndVersion)
	if err_getByIdAndVersion != nil {
		logger.Error("Test_GetDataCallByIdAndVersion: Error during json.Unmarshal: ", err_getByIdAndVersion)
		t.FailNow()
	}

	assert.True(t, reflect.DeepEqual(input_getByIdAndVersion, output_getByIdAndVersion))

}

//Test for GetDataCallVersionsById
//This function tests whether it returns error for empty Id.
//This function tests whether it retrives all the different versions of a DataCall.
func Test_GetDataCallVersionsById_Should_Return_All_Versions_Of_Input_Datacall_Id(t *testing.T) {
	fmt.Println("Test_GetDataCallVersionsById_Should_Return_All_Versions_Of_Input_Datacall_Id")
	scc := new(SmartContract)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)
	sid := &msp.SerializedIdentity{Mspid: "aaismsp", IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b

	//Step-1: For ERROR- Check whether it returns error for empty ID
	res_err_getDataCallVersions := checkInvoke_forError(t, stub, "GetDataCallVersionsById", []byte(GET_DATA_CALL_VERSIONS_BY_ID_EMPTY_ID_JSON))
	var err_message_getDataCallVersions = res_err_getDataCallVersions.Message
	if res_err_getDataCallVersions.Status != shim.OK {
		assert.Equal(t, "Failed to get Data Calls : not implemented", err_message_getDataCallVersions, "Test_GetDataCallVersionsById: For Empty Id")
	} else {
		t.FailNow()
	}

}

//Test for ListDataCallsByCriteria
//This function tests whether it returns error for empty Status.
func Test_ListDataCallsByCriteria_Should_Return_Datacalls_Based_On_Input_Criterion(t *testing.T) {
	fmt.Println("Test_ListDataCallsByCriteria_Should_Return_Datacalls_Based_On_Input_Criterion")
	scc := new(SmartContract)
	stub := NewCouchDBMockStub("OpenIDLMockStub", scc)
	sid := &msp.SerializedIdentity{Mspid: "aaismsp", IdBytes: idbytes}
	b, err := proto.Marshal(sid)
	if err != nil {
		t.FailNow()
	}
	stub.Creator = b

	//Step-1: For ERROR- Check whether it returns error for empty Status
	res_err_listByCriteria := checkInvoke_forError(t, stub, "ListDataCallsByCriteria", []byte(LIST_DATA_CALL_BY_CRITERIA_EMPTY_STATUS_JSON))
	var err_message_listByCriteria = res_err_listByCriteria.Message
	if res_err_listByCriteria.Status != shim.OK {
		assert.Equal(t, "ListDataCallsByCriteria: Status not present, You must pass Status in agument", err_message_listByCriteria, "Test_ListDataCallsByCriteria: For Empty Id")
	} else {
		t.FailNow()
	}

}
