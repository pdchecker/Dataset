package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract defines the Smart Contract structure
type SmartContract struct {
}

// Person defines the person structure with properties
type Person struct {
	Name      string `json:"name"`
	Birthdate string `json:"birthdate"`
	Address   string `json:"address"`
	Phone     string `json:"phone"`
}

type panCardDetails struct {
	YearlyIncome string `json:"yearlyIncome"`
	PanCardID    string `json:"panCardID"`
}

// Init initializes the smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("person_cc")

// Invoke is the entry point for all transactions
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {
	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is: %s", function)
	logger.Infof("Args length is: %d", len(args))

	switch function {
	case "queryPerson":
		return s.queryPerson(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createPerson":
		return s.createPerson(APIstub, args)
	case "queryAllPersons":
		return s.queryAllPersons(APIstub)
	case "updateAllOrganizations":
		return s.updateAllOrganizations(APIstub, args)
	case "queryPanCardDetailsOrg1":
		return s.queryPanCardDetailsOrg1(APIstub, args)
	case "queryPanCardDetailsOrg2":
		return s.queryPanCardDetailsOrg2(APIstub, args)
	case "getHistoryForPerson":
		return s.getHistoryForPerson(APIstub, args)
	case "queryPersonsByAddress":
		return s.queryPersonsByAddress(APIstub, args)
	case "restrictedMethod":
		return s.restrictedMethod(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}

func (s *SmartContract) queryPerson(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	personAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(personAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	persons := []Person{
		Person{Name: "John Doe", Birthdate: "1990-01-01", Address: "123 Main St", Phone: "555-1234"},
		Person{Name: "Jane Smith", Birthdate: "1985-05-15", Address: "456 Oak Ave", Phone: "555-5678"},
	}

	i := 0
	for i < len(persons) {
		personAsBytes, _ := json.Marshal(persons[i])
		APIstub.PutState("PERSON"+strconv.Itoa(i), personAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createPerson(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var person = Person{
		Name:      args[1],
		Birthdate: args[2],
		Address:   args[3],
		Phone:     args[4],
	}

	personAsBytes, _ := json.Marshal(person)
	APIstub.PutState(args[0], personAsBytes)

	return shim.Success(personAsBytes)
}

func (s *SmartContract) queryAllPersons(APIstub shim.ChaincodeStubInterface) sc.Response {
	startKey := "PERSON0"
	endKey := "PERSON999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllPersons:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) updateAllOrganizations(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 7 {
		return shim.Error("Incorrect number of arguments. Expecting 7")
	}

	personAsBytes, err := APIstub.GetState(args[0])
	if err != nil {
		return shim.Error("Failed to get person: " + err.Error())
	} else if personAsBytes == nil {
		fmt.Println("This person does not exist: " + args[0])
		return shim.Error("This person does not exist: " + args[0])
	}

	var person = Person{}
	json.Unmarshal(personAsBytes, &person)

	person.Name = args[1]
	person.Birthdate = args[2]
	person.Address = args[3]
	person.Phone = args[4]

	personAsBytes, err = json.Marshal(person)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutState(args[0], personAsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Update for Org1
	panCardDetailsOrg1 := &panCardDetails{YearlyIncome: args[5], PanCardID: args[6]}
	panCardDetailsOrg1AsBytes, err := json.Marshal(panCardDetailsOrg1)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionPanCardDetailsOrg1", args[0], panCardDetailsOrg1AsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	// Update for Org2
	panCardDetailsOrg2 := &panCardDetails{YearlyIncome: args[5], PanCardID: args[6]}
	panCardDetailsOrg2AsBytes, err := json.Marshal(panCardDetailsOrg2)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionPanCardDetailsOrg2", args[0], panCardDetailsOrg2AsBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}



func (s *SmartContract) queryPanCardDetailsOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	panCardDetailsAsBytes, err := APIstub.GetPrivateData("collectionPanCardDetailsOrg1", args[0])
	if err != nil {
		return shim.Error("Failed to get PAN card details from Org1: " + err.Error())
	} else if panCardDetailsAsBytes == nil {
		return shim.Error("PAN card details not found in Org1 for: " + args[0])
	}

	return shim.Success(panCardDetailsAsBytes)
}

func (s *SmartContract) queryPanCardDetailsOrg2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	panCardDetailsAsBytes, err := APIstub.GetPrivateData("collectionPanCardDetailsOrg2", args[0])
	if err != nil {
		return shim.Error("Failed to get PAN card details from Org2: " + err.Error())
	} else if panCardDetailsAsBytes == nil {
		return shim.Error("PAN card details not found in Org2 for: " + args[0])
	}

	return shim.Success(panCardDetailsAsBytes)
}



func (s *SmartContract) getHistoryForPerson(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	personAsBytes, _ := APIstub.GetState(args[0])
	if personAsBytes == nil {
		return shim.Error("Person not found")
	}

	resultsIterator, err := APIstub.GetHistoryForKey(args[0])
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"TxnID\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		buffer.WriteString(string(response.Value))

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

	fmt.Printf("- getHistoryForPerson returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) queryPersonsByAddress(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := fmt.Sprintf("{\"selector\":{\"address\":\"%s\"}}", args[0])
	resultsIterator, err := APIstub.GetQueryResult(queryString)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryPersonsByAddress:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) restrictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	clientID, err := cid.New(APIstub)
	if err != nil {
		return shim.Error("Error getting client ID: " + err.Error())
	}

	mspID, err := clientID.GetMSPID()
	if err != nil {
		return shim.Error("Error getting MSP ID: " + err.Error())
	}

	// Check if the client has the required role (e.g., "admin")
	if !hasAdminRole(mspID) {
		return shim.Error("Permission denied. Only users with 'admin' role can invoke this method.")
	}

	// Your logic for the restricted method goes here

	return shim.Success(nil)
}

// Example function to check if the client has the admin role
func hasAdminRole(mspID string) bool {
	// Add your logic to check if the user has the 'admin' role in the specified MSP
	// For example, you might have a configuration or an external system that defines roles.

	// For demonstration purposes, assume that 'admin' role is required for access.
	return mspID == "AdminMSP"
}


// main function starts up the chaincode in the container during instantiate
func main() {
	if err := shim.Start(new(SmartContract)); err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %s", err)
	}
}