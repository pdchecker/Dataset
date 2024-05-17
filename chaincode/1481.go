package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

type VerificationAttributes struct {
	Organisation   string `json:"Organisation"`
	Table_name  string `json:"Table_name"`
	User string `json:"User"`
	Time  string `json:"Time"`
}


// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("IntegrityCertificateCC")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryById":
		return s.queryById(APIstub, args)
	case "createRecord":
		return s.createRecord(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}


func (s *SmartContract) queryById(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	RecordAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(RecordAsBytes)
}

//Create new recrod using API. 
func (s *SmartContract) createRecord(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	var rec = VerificationAttributes{Organisation: args[1], Table_name: args[2], User: args[3], Time: args[4]}

	RecordAsBytes, _ := json.Marshal(rec)
	APIstub.PutState(args[0], RecordAsBytes)

	indexName := "key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{rec.Organisation, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(RecordAsBytes)
}

func main() {

	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}