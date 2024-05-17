// Copyright (c) 2021 akachain
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package util provides generic CRUD methods for working with Hyperledger Fabric
// golang chaincode.
package main

import (
	"encoding/json"
	"fmt"

	"github.com/Akachain/akc-go-sdk-v2/common"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
	"github.com/mitchellh/mapstructure"
)

// ChangeInfo overwrites a value in the state database by performing an insert with
// overwrite indicator
func ChangeInfo(stub shim.ChaincodeStubInterface, DocPrefix string, rowKey []string, data interface{}) error {
	_, err := InsertTableRow(stub, DocPrefix, rowKey, data, FAIL_UNLESS_OVERWRITE, nil)
	return err
}

// UpdateExistingData works similar to ChangeInfo function.
// However, it does not check if the document is already existed
// This is useful if we already query out the row before and do not want to query again.
func UpdateExistingData(stub shim.ChaincodeStubInterface, DocPrefix string, rowKey []string, data interface{}) error {
	err := UpdateTableRow(stub, DocPrefix, rowKey, data)
	return err
}

// CreateData simply inserts a new key-value pair into the state database. It will fail if the key already
// exists. The pair is formatted as follows.
//
// Key: DocPrefix_rowKey[0]_rowKey[1]_..._rowKey[n]_randomId
// Value: JSON document parsed from the data object.
func CreateData(stub shim.ChaincodeStubInterface, DocPrefix string, rowKey []string, data interface{}) error {
	var oldData interface{}
	rowWasFound, err := InsertTableRow(stub, DocPrefix, rowKey, data, FAIL_BEFORE_OVERWRITE, &oldData)
	if err != nil {
		return err
	}
	if rowWasFound {
		return fmt.Errorf("Could not create data %v because an data already exists", data)
	}
	return nil //success
}

// GetDataById get the state value of a key from the state database
// It returns a generic object.
func GetDataById(stub shim.ChaincodeStubInterface, ID string, DocPrefix string) (interface{}, error) {
	var dataStruct interface{}

	rowWasFound, err := GetTableRow(stub, DocPrefix, []string{ID}, &dataStruct, FAIL_IF_MISSING)

	if err != nil {
		return nil, err
	}
	if !rowWasFound {
		return nil, fmt.Errorf("Data with ID %s does not exist", ID)
	}
	return dataStruct, nil
}

// GetDataByIdWithResponse returns the peer.Response object directly to the caller so that
// the caller does not have to format it into Fabric response.
func GetDataByIdWithResponse(stub shim.ChaincodeStubInterface, DataID string, data interface{}, ModelTable string) peer.Response {
	rs, err := GetDataById(stub, DataID, ModelTable)
	if err != nil {
		//Get Data Fail
		resErr := common.ResponseError{ResCode: common.ERR4, Msg: fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}
	if rs != nil {
		mapstructure.Decode(rs, data)
	} else {
		data = nil
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		//Convert Json Fail
		resErr := common.ResponseError{ResCode: common.ERR3, Msg: fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}
	fmt.Printf("Response: %s\n", string(bytes))
	resSuc := common.ResponseSuccess{ResCode: common.SUCCESS, Msg: common.ResCodeDict[common.SUCCESS], Payload: string(bytes)}
	return common.RespondSuccess(resSuc)
}

// GetDataByRowKeys returns the state value of a key that is constructed with only document prefix
// and rowKeys. This is used for documents that we don't need random generated value in the key.
func GetDataByRowKeys(stub shim.ChaincodeStubInterface, rowKeys []string, DocPrefix string) (interface{}, error) {
	var dataStruct interface{}

	rowWasFound, err := GetTableRow(stub, DocPrefix, rowKeys, &dataStruct, FAIL_IF_MISSING)

	if err != nil {
		return nil, err
	}
	if !rowWasFound {
		return nil, fmt.Errorf("Data with rowKeys %s does not exist", rowKeys)
	}
	return dataStruct, nil
}

// GetDataByRowKeysWithResponse returns the peer.Response directly to the caller so that
// the caller does not have to format it into Fabric response
func GetDataByRowKeysWithResponse(stub shim.ChaincodeStubInterface, rowKeys []string, data interface{}, ModelTable string) peer.Response {

	rs, err := GetDataByRowKeys(stub, rowKeys, ModelTable)
	if err != nil {
		//Get Data Fail
		resErr := common.ResponseError{ResCode: common.ERR4, Msg: fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR4], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}
	if rs != nil {
		mapstructure.Decode(rs, data)
	} else {
		data = nil
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		//Convert Json Fail
		resErr := common.ResponseError{ResCode: common.ERR3, Msg: fmt.Sprintf("%s %s %s", common.ResCodeDict[common.ERR3], err.Error(), common.GetLine())}
		return common.RespondError(resErr)
	}
	fmt.Printf("Response: %s\n", string(bytes))
	resSuc := common.ResponseSuccess{ResCode: common.SUCCESS, Msg: common.ResCodeDict[common.SUCCESS], Payload: string(bytes)}
	return common.RespondSuccess(resSuc)
}
