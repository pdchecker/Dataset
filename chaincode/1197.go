package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	//"reflect"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	//"github.com/hyperledger/fabric/common/util"
	//"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

type SampleRecord struct{
	Key string `json:Key`
	Value string `json:Value`
}

var logger = flogging.MustGetLogger("models_cc")


// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	logger.Infof("Chiancode : datacollection initiated")
	return shim.Success(nil)
}


// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))
	logger.Errorf("Function name is:  %d", function)
	switch function {
		case "createSampleRecord":
			return s.createSampleRecord(APIstub, args)
		case "traceSampleRecord":
			return s.traceSampleRecord(APIstub, args)
		case "getSampleRecord":
			return s.getSampleRecord(APIstub, args)
			
	}
	return shim.Error("Invoke Function Not Success.")
}


//add a new model 
func (s *SmartContract) createSampleRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	var sample = SampleRecord{Key: args[0], Value: args[1]}


	RecordAsBytes, _ := json.Marshal(sample)
	APIstub.PutState(args[0], RecordAsBytes)

	logger.Infof("Sample Record Addded Successfully")
	return shim.Success(RecordAsBytes)
}

//retrieve the sample record
func (s *SmartContract) getSampleRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	RecordAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(RecordAsBytes)
}


func (t *SmartContract) traceSampleRecord(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	ID := args[0]
	logger.Infof("searching for the Perf Contract with id %s", args[0])
	resultsIterator, err := stub.GetHistoryForKey(ID)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()

		logger.Infof(string(response.TxId))
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	logger.Infof("- getHistoryforDEC returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}




