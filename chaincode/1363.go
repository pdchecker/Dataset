package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type dprChaincode struct {
}

type dpr struct {
	Id             string    `json:"id"`
	DprNo          string    `json:"dprNo"`
	ShipperNo      string    `json:"shipperNo"`
	From           string    `json:"from"`
	To             string    `json:"to"`
	Products       string    `json:"products"`
	DocumentNo     string    `json:"documentNo"`
	ReferenceSOPNo string    `json:"referenceSOPNo"`
	Department     string    `json:"department"`
	PickingListNo  string    `json:"pickingListNo"`
	Version        string    `json:"version"`
	LegacyDocNo    string    `json:"legacyDocNo"`
	EffectiveDate  time.Time `json:"effectiveDate"`
	CcdrStatus     string    `json:"ccdrStatus"`
	TransportMode  string    `json:"transportMode"`
	OrgId          string    `json:"orgId"`
	IsDelete       bool      `json:"isDelete"`
	CreatedBy      string    `json:"createdBy"`
	CreatedOn      time.Time `json:"createdOn"`
	PackingList    string    `json:"packingList"`
	Notes          string    `json:"notes"`
}

func (cc *dprChaincode) create(stub shim.ChaincodeStubInterface, arg []string) peer.Response {

	args := strings.Split(arg[0], "^^")

	if len(args) != 21 {
		return shim.Error("Incorrect number arguments. Expecting 20")
	}
	createdOn, err1 := time.Parse(time.RFC3339, args[18])

	if err1 != nil {
		return shim.Error("Error converting string to date: " + err1.Error())
	}

	effectiveDate, err2 := time.Parse(time.RFC3339, args[12])
	if err2 != nil {
		return shim.Error("Error converting string to date: " + err2.Error())
	}

	isDeleteBool, err3 := strconv.ParseBool(args[16])

	if err3 != nil {
		return shim.Error("Error converting string to bool: " + err3.Error())
	}

	data := dpr{
		Id:             args[0],
		DprNo:          args[1],
		ShipperNo:      args[2],
		From:           args[3],
		To:             args[4],
		Products:       args[5],
		DocumentNo:     args[6],
		ReferenceSOPNo: args[7],
		Department:     args[8],
		PickingListNo:  args[9],
		Version:        args[10],
		LegacyDocNo:    args[11],
		EffectiveDate:  effectiveDate,
		CcdrStatus:     args[13],
		TransportMode:  args[14],
		OrgId:          args[15],
		IsDelete:       isDeleteBool,
		CreatedBy:      args[17],
		CreatedOn:      createdOn,
		PackingList:    args[19],
		Notes:          args[20],
	}

	dataBytes, errMarshal := json.Marshal(data)

	if errMarshal != nil {
		return shim.Error("Error converting data as bytes: " + errMarshal.Error())
	}

	errPut := stub.PutState(args[0], dataBytes)

	if errPut != nil {
		return shim.Error("Error putting the state: " + errPut.Error())
	}

	return shim.Success(nil)
}

func (cc *dprChaincode) get(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number arguments. Expecting 1")
	}

	stateBytes, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error("Error getting the state: " + err.Error())
	}

	return shim.Success(stateBytes)
}
func (cc *dprChaincode) update(stub shim.ChaincodeStubInterface, arg []string) peer.Response {

	args := strings.Split(arg[0], "^^")

	if len(args) != 21 {
		return shim.Error("Incorrect number arguments. Expecting 20")
	}
	createdOn, err1 := time.Parse(time.RFC3339, args[18])

	if err1 != nil {
		return shim.Error("Error converting string to date: " + err1.Error())
	}

	effectiveDate, err2 := time.Parse(time.RFC3339, args[12])
	if err2 != nil {
		return shim.Error("Error converting string to date: " + err2.Error())
	}

	isDeleteBool, err3 := strconv.ParseBool(args[16])

	if err3 != nil {
		return shim.Error("Error converting string to bool: " + err3.Error())
	}

	data := dpr{
		Id:             args[0],
		DprNo:          args[1],
		ShipperNo:      args[2],
		From:           args[3],
		To:             args[4],
		Products:       args[5],
		DocumentNo:     args[6],
		ReferenceSOPNo: args[7],
		Department:     args[8],
		PickingListNo:  args[9],
		Version:        args[10],
		LegacyDocNo:    args[11],
		EffectiveDate:  effectiveDate,
		CcdrStatus:     args[13],
		TransportMode:  args[14],
		OrgId:          args[15],
		IsDelete:       isDeleteBool,
		CreatedBy:      args[17],
		CreatedOn:      createdOn,
		PackingList:    args[19],
		Notes:          args[20],
	}

	dataBytes, errMarshal := json.Marshal(data)

	if errMarshal != nil {
		return shim.Error("Error converting data as bytes: " + errMarshal.Error())
	}

	errPut := stub.PutState(args[0], dataBytes)

	if errPut != nil {
		return shim.Error("Error putting the data state: " + errPut.Error())
	}

	return shim.Success(nil)
}
func (cc *dprChaincode) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number arguments. Expecting 1")
	}

	dataBytes, err := stub.GetState(args[0])

	if err != nil {
		return shim.Error("Error getting the state: " + err.Error())
	}

	data := dpr{}

	json.Unmarshal(dataBytes, &data)

	data.IsDelete = true

	updateDataBytes, err1 := json.Marshal(data)

	if err1 != nil {
		return shim.Error("Error converting data as bytes: " + err1.Error())
	}

	err2 := stub.PutState(args[0], updateDataBytes)

	if err2 != nil {
		return shim.Error("Error putting the data state: " + err2.Error())
	}

	return shim.Success(nil)
}

func (cc *dprChaincode) history(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryResult, err := stub.GetHistoryForKey(args[0])

	if err != nil {
		return shim.Error("Error getting history results: " + err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("[")

	isDataAdded := false
	for queryResult.HasNext() {
		queryResponse, err1 := queryResult.Next()
		if err1 != nil {
			return shim.Error(err1.Error())
		}

		if isDataAdded == true {
			buffer.WriteString(",")
		}

		buffer.WriteString(string(queryResponse.Value))

		isDataAdded = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}

func (cc *dprChaincode) querystring(stub shim.ChaincodeStubInterface, args []string) peer.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryResult, err := stub.GetQueryResult(args[0])

	if err != nil {
		return shim.Error("Error getting query string results: " + err.Error())
	}

	var buffer bytes.Buffer
	buffer.WriteString("[")

	isDataAdded := false
	for queryResult.HasNext() {
		queryResponse, err1 := queryResult.Next()
		if err1 != nil {
			return shim.Error(err1.Error())
		}

		if isDataAdded == true {
			buffer.WriteString(",")
		}

		buffer.WriteString(string(queryResponse.Value))

		isDataAdded = true
	}
	buffer.WriteString("]")

	return shim.Success(buffer.Bytes())
}
func (cc *dprChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}

func (cc *dprChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {

	function, args := stub.GetFunctionAndParameters()

	if function == "create" {
		return cc.create(stub, args)
	} else if function == "get" {
		return cc.get(stub, args)
	} else if function == "update" {
		return cc.update(stub, args)
	} else if function == "delete" {
		return cc.delete(stub, args)
	} else if function == "history" {
		return cc.history(stub, args)
	} else if function == "querystring" {
		return cc.querystring(stub, args)
	}

	return shim.Error("Invalid invoke function name")
}

func main() {
	var _ = strconv.FormatInt(1234, 10)
	var _ = time.Now()
	var _ = strings.ToUpper("test")
	var _ = bytes.ToUpper([]byte("test"))

	err := shim.Start(new(dprChaincode))
	if err != nil {
		fmt.Printf("Error starting BioMetric chaincode: %s", err)
	}
}
