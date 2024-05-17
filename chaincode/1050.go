package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

type Chaincode struct {
}

// bcFunctions
var bcFunction = map[string]func(shim.ChaincodeStubInterface, []string) peer.Response{
	"get_dob_cert_fileName": getDOBCert,
	"responRequest":         responRequest,
	"verifyPersonal":        verifyPersonal,
	"addQualification":      addQualification,
	"getPersonal":           getPersonal,
	"getEduc":               getEduc,
	"getHealthc":            getHealthc,
	"getStateByte":          getStateByte,
	"addHealthReports":      addReports,
	"ExecuteRichQuery":      ExecuteRichQuery,
	/*
		change name,address
	*/
}

func (c *Chaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	return shim.Success(nil)
}
func (c *Chaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	funcName, args := stub.GetFunctionAndParameters()
	bcFunc := bcFunction[funcName]
	if bcFunc == nil {
		return shim.Error("No function avaiable of given name")
	}
	return bcFunc(stub, args)
}
func main() {
	err := shim.Start(new(Chaincode))
	if err != nil {
		fmt.Println(err.Error())
	}
}
func getStateByte(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// [0]=IKey
	if len(args) != 1 {
		return shim.Error("please Provide IKey")
	}
	DByet, err := getState(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(DByet)
}
func getPersonal(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// [0]=IKey
	if len(args) != 1 {
		return shim.Error("please Provide IKey")
	}
	DByet, err := getState(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	var identity Identity
	json.Unmarshal(DByet, &identity)
	DByet, _ = getState(stub, identity.PersonalDetails)
	return shim.Success(DByet)
}
func getEduc(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// [0]=IKey
	if len(args) != 1 {
		return shim.Error("please Provide IKey")
	}
	DByet, err := getState(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	var identity Identity
	json.Unmarshal(DByet, &identity)
	DByet, _ = getState(stub, identity.EducationDetails)
	return shim.Success(DByet)

}
func getHealthc(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	// [0]=IKey
	if len(args) != 1 {
		return shim.Error("please Provide IKey")
	}
	DByet, err := getState(stub, args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	var identity Identity
	json.Unmarshal(DByet, &identity)
	DByet, _ = getState(stub, identity.HealthDetails)
	return shim.Success(DByet)
}
