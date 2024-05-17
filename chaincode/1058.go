package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// Chanicode struct
type Chaincode struct {
}

// Asset used to track define data, publicly
type data struct {
	Id          string
	Owner       string
	Description string
}

// Asset used in the private databases to store the actual data
type dataBioPrivateDetails struct {
	Id   string
	Data string
}

// Enumeration used in assetRequest
type status int

const (
	Pending  status = iota
	Accepted status = iota
	Rejected status = iota
)

// Asset used to track the data requests and their status
type assetRequest struct {
	RequestId int64
	Applicant string
	DataId    string
	Owner     string
	Status    status
}

// ============================================================================

// Initializes chaincode
// =====================
func (c *Chaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {

	var counter int64
	counter = 0

	// Marshal counter
	dataCounter, err := json.Marshal(counter)
	if err != nil {
		return shim.Error("Error Marshaling data during Init.")
	}

	// Init the request counter
	err = stub.PutState("REQUESTCOUNTER", dataCounter)
	if err != nil {
		return shim.Error("Error during put state: " + err.Error())
	}

	return shim.Success(nil)
}

// Invoke
// ======
func (c *Chaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Invoke is running " + function)

	switch function {
	case "insertData":
		return c.insertData(stub, args)
	case "removeData":
		return c.removeData(stub, args)
	case "viewCatalogue":
		return c.viewCatalogue(stub, args)
	case "viewPersonalData":
		return c.viewPersonalData(stub, args)
	case "getPrivateData":
		return c.getPrivateData(stub, args)
	case "requestData":
		return c.requestData(stub, args)
	case "viewRequests":
		return c.viewRequests(stub, args)
	case "acceptRequest":
		return c.acceptRequest(stub, args)
	case "denyRequest":
		return c.denyRequest(stub, args)
	case "viewAllRequests":
		return c.viewAllRequests(stub, args)
	default:
		fmt.Println("Unknown function!")
		return shim.Error("Received unknown function")
	}
}

// insertData
// ==========
// Gives a set of transient data, this methods insert in the ledger the pair
// <name> - <name,descritpion,owner> and in the private collection the pair
// <name> - <data>.
func (c *Chaincode) insertData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- start insertData")

	type dataTransientInput struct {
		Name        string
		Description string
		Data        string
	}

	var err error
	var dataInput dataTransientInput

	// No parameter shall be passed outside the transient map
	if len(args) != 0 {
		return shim.Error("Incorrect number of parameters. Data must be passed in transient map.")
	}

	// Retrieve the transient data
	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	// Get the 'data' field
	dataJSONBytes, ok := transMap["data"]
	if !ok {
		return shim.Error("'data' must be a key in the transient map.")
	}

	// Decode the JSON bytes and store it in a variable
	err = json.Unmarshal(dataJSONBytes, &dataInput)
	if err != nil {
		return shim.Error("Error decoding JSON data: " + string(dataJSONBytes))
	}

	// Check input consistency
	if len(dataInput.Data) == 0 {
		return shim.Error("'Data' field must be a non-empty string.")
	}
	if len(dataInput.Name) == 0 {
		return shim.Error("'Name' field must be a non-empty string.")
	}
	if len(dataInput.Description) == 0 {
		return shim.Error("'Description' field must be a non-empty string.")
	}

	// Check if data already exists
	dataLedgerBytes, err := stub.GetState(dataInput.Name)
	if err != nil {
		return shim.Error("Failed to get data from the ledger: " + err.Error())
	} else if dataLedgerBytes != nil {
		fmt.Println("This data already exists: " + dataInput.Name)
		return shim.Error("This data already exists: " + dataInput.Name)
	}

	// Creation of a unique string for the owner
	var creator string
	// retrieve creator (Id + MSPID)
	creator = getCreator(stub)

	// Create data to be inserted into the ledger
	dataToInsert := &data{
		Id:          dataInput.Name,
		Description: dataInput.Description,
		Owner:       creator,
	}

	// Marshal the data
	dataToInsertJSON, err := json.Marshal(dataToInsert)
	if err != nil {
		return shim.Error("Error creating data:" + err.Error())
	}

	// Save data into the ledger
	err = stub.PutState("DATA"+dataInput.Name, dataToInsertJSON)
	if err != nil {
		return shim.Error("Error during put state: " + err.Error())
	}

	// Create data to be inserted into the private data
	dataBio := &dataBioPrivateDetails{
		Id:   dataInput.Name,
		Data: dataInput.Data,
	}

	dataBioJSON, err := json.Marshal(dataBio)
	if err != nil {
		return shim.Error("Error creating data: " + err.Error())
	}

	// Retrieve private collection name
	collectionName := getCollectionName(stub)

	// Insert data into the private collection
	err = stub.PutPrivateData(collectionName, "DATA"+dataInput.Name, dataBioJSON)
	if err != nil {
		return shim.Error("Error during put private state: " + err.Error())
	}

	fmt.Println("end insertData")
	return shim.Success(nil)
}

// removeData
// ==========
// Given a data name, using transient data, this method deletes the
// corresponding data from the ledger and from the private collection.
func (c *Chaincode) removeData(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("- Start removeData")

	type removeDataTransientInput struct {
		Name string
	}

	// No parameter shall be passed outside the transient map
	if len(args) != 0 {
		return shim.Error("Incorrect number of parameters. Data must be passed in transient map.")
	}

	// Retrieve the transient data
	transMap, err := stub.GetTransient()
	if err != nil {
		return shim.Error("Error getting transient: " + err.Error())
	}

	// Get the 'data' field
	dataJSONBytes, ok := transMap["data"]
	if !ok {
		return shim.Error("'data' must be a key in the transient map.")
	}

	if len(dataJSONBytes) == 0 {
		return shim.Error("'data' value in the transient map must be a non-empty JSON string.")
	}

	// Unmarshal Json data
	var removeDataInput removeDataTransientInput
	err = json.Unmarshal(dataJSONBytes, &removeDataInput)
	if err != nil {
		return shim.Error("Failed to decode JSON of: " + string(dataJSONBytes))
	}

	if len(removeDataInput.Name) == 0 {
		return shim.Error("'Name' fiels must be a non-empty string.")
	}

	// Check if data already exists
	dataBytes, err := stub.GetState(removeDataInput.Name)
	if err != nil {
		return shim.Error("Failed to get data from the ledger: " + err.Error())
	} else if dataBytes == nil {
		fmt.Println("This data does not exists: " + removeDataInput.Name)
		return shim.Error("This data cannot be removed, it does not exist!")
	}

	// Check if the one how asked for deletion is the owner
	var data data
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return shim.Error("Error during data unmarshal")
	}

	var creator string
	// retrieve creator (Id + MSPID)
	creator = getCreator(stub)

	if data.Owner != creator {
		return shim.Error("Deletion not allowed. Only the creator can delete its data.")
	}

	// Delete public state
	err = stub.DelState(removeDataInput.Name)
	if err != nil {
		return shim.Error("Error removing the key from the ledger " + err.Error())
	}

	// Retrieve private collection name
	collectionName := getCollectionName(stub)

	// Remove the data from the private collection
	err = stub.PurgePrivateData(collectionName, removeDataInput.Name)
	if err != nil {
		return shim.Error("Error removing the key from the private collection " + err.Error())
	}

	fmt.Println("end removeData")
	return shim.Success(nil)
}

// viewCatalogue
// =============
// This method returns all the data units stored in the public ledger.
func (c *Chaincode) viewCatalogue(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// Check the number of arguements
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Zero expected")
	}

	// Perform the range query
	resultIterator, err := stub.GetStateByRange("DATA", "")
	if err != nil {
		return shim.Error("Error during range query. " + err.Error())
	}
	defer resultIterator.Close()

	// Write result on the buffer
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Add a comma before array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}

		// Check if the retrieved key-value pair is about the DATA type
		if strings.Contains(queryResponse.Key, "DATA") {
			buffer.WriteString(
				fmt.Sprintf(`{"Key":"%s", "Record":"%s"}`,
					queryResponse.Key, queryResponse.Value),
			)
			bArrayMemberAlreadyWritten = true
		}

	}
	buffer.WriteString("]")

	fmt.Printf("- viewCatalogue Result:\n%s\n", buffer.String())

	return shim.Success([]byte(buffer.String()))
}

// viewPersonalData
// ================
// This method allows to view all the personal data inserted into the public
// ledger.
func (c *Chaincode) viewPersonalData(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// Check the number of arguements
	if len(args) != 0 {
		return shim.Error("Incorrect number of arguments. Zero expected")
	}

	// Perform the range query
	resultIterator, err := stub.GetStateByRange("DATA", "")
	if err != nil {
		return shim.Error("Error during range query. " + err.Error())
	}
	defer resultIterator.Close()

	var creator string
	var dataElement data
	// Write result on the buffer
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Check if the retrieved key-value pair is about the DATA type
		if strings.Contains(queryResponse.Key, "DATA") {

			// Unmarshal the incoming data
			err = json.Unmarshal(queryResponse.Value, &dataElement)
			if err != nil {
				return shim.Error("Error unmashaling data from response.")
			}

			// retrieve creator (Id + MSPID)
			creator = getCreator(stub)

			if creator == dataElement.Owner {
				// Add a comma before array member
				if bArrayMemberAlreadyWritten {
					buffer.WriteString(",")
				}

				fmt.Printf(`{"Key":"%s", "Record":"%s"}`, queryResponse.Key, queryResponse.Value)

				buffer.WriteString(
					fmt.Sprintf(`{"Key":"%s", "Record":"%s"}`,
						queryResponse.Key, queryResponse.Value),
				)
				bArrayMemberAlreadyWritten = true

			}
		}

	}
	buffer.WriteString("]")

	fmt.Printf("- viewPersonalData Result:\n%s\n", buffer.String())

	return shim.Success([]byte(buffer.String()))
}

// getPrivateData
// ==============
// This method allows to see the secret data stored in the private collection.
func (c *Chaincode) getPrivateData(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	collectionName := getCollectionName(stub)
	fmt.Println("Collection name: " + collectionName)

	resultIterator, err := stub.GetPrivateDataByRange(collectionName, "", "")
	if err != nil {
		return shim.Error("Error retrieving the data from the collection!")
	}
	defer resultIterator.Close()

	// Write result on the buffer
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultIterator.HasNext() {
		queryResponse, err := resultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Add a comma before array member
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}

		// Check if the retrieved key-value pair is about the DATA type
		if strings.Contains(queryResponse.Key, "DATA") {
			buffer.WriteString(
				fmt.Sprintf(`{"Key":"%s", "Record":"%s"}`,
					queryResponse.Key, queryResponse.Value),
			)
			bArrayMemberAlreadyWritten = true
		}

	}
	buffer.WriteString("]")

	fmt.Printf("- getPrivateData Result:\n%s\n", buffer.String())

	// CANNOT PRINT THE BUFFER, GIVES ERROR
	return shim.Success(nil)
}

// requestData
// ===========
// This method allows to create a request for a data owned by another
// organization. The name of the data must be specified in the arguements during
// the call.
func (c *Chaincode) requestData(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// Check arguement consistency
	if len(args) != 1 {
		return shim.Error("Incorrect number of parameters. Expecting 1.")
	}

	fmt.Println("Args: " + args[0])

	// Check existence of the data
	dataBytes, err := stub.GetState("DATA" + args[0])
	if err != nil {
		return shim.Error("Failed to get data from the ledger: " + err.Error())
	} else if dataBytes == nil {
		fmt.Println("This data does not exists: " + args[0])
		return shim.Error("This data cannot be requested, it does not exist!")
	}

	var dataRequested assetRequest
	err = json.Unmarshal(dataBytes, &dataRequested)
	if err != nil {
		return shim.Error("Error during data unmarshal!")
	}

	// Get value of the global counter of the requests
	counterBytes, err := stub.GetState("REQUESTCOUNTER")
	if err != nil {
		return shim.Error("Failed to get Counter from the ledger!")
	}
	if counterBytes == nil {
		fmt.Println("The counter does not exist!")
		return shim.Error("Counter does not exist!")
	}

	var counter int64

	// Unmarshal counter data
	err = json.Unmarshal(counterBytes, &counter)
	if err != nil {
		return shim.Error("Error during data unmarshal")
	}

	// Increment counter
	counter++

	// Get applicant
	applicant := getCreator(stub)

	assetRequestProposal := assetRequest{
		RequestId: counter,
		Applicant: applicant,
		DataId:    args[0],
		Owner:     dataRequested.Owner,
		Status:    Pending,
	}

	// Marshal data
	assetRequestProposalJSON, err := json.Marshal(assetRequestProposal)
	if err != nil {
		return shim.Error("Error during data marshal.")
	}

	// Insert data into the ledger
	err = stub.PutState("REQUEST"+strconv.FormatInt(counter, 10), assetRequestProposalJSON)
	if err != nil {
		// If something does wrong, restore the counter to the previous value
		return shim.Error("Error during putState: " + err.Error())
	}

	counterJSON, err := json.Marshal(counter)
	if err != nil {
		return shim.Error("Error during data marshal.")
	}

	// Update counter value into the ledger
	err = stub.PutState("REQUESTCOUNTER", counterJSON)
	if err != nil {
		return shim.Error("Error during putState: " + err.Error())
	}

	return shim.Success(nil)
}

// viewRequests
// ============
// This method allows to see all the requests of data sharing belonging to the
// caller.
func (c *Chaincode) viewRequests(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 0 {
		return shim.Error("Incorrect number of parameters. Expected 0.")
	}

	// Perform the range query for the requests
	resultIteratorRequests, err := stub.GetStateByRange("REQUEST", "")
	if err != nil {
		return shim.Error("Error during range query. " + err.Error())
	}
	defer resultIteratorRequests.Close()

	// Retrieve caller
	creator := getCreator(stub)

	var requestElement assetRequest
	// Write result on the buffer
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultIteratorRequests.HasNext() {
		queryResponse, err := resultIteratorRequests.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		if strings.Contains(queryResponse.Key, "REQUEST") &&
			!strings.Contains(queryResponse.Key, "COUNTER") {
			// Unmarshal the incoming data
			err = json.Unmarshal(queryResponse.Value, &requestElement)
			if err != nil {
				return shim.Error("Error unmashaling data from response.")
			}

			// Check the owner
			if requestElement.Owner == creator {
				// Add a comma before array member
				if bArrayMemberAlreadyWritten {
					buffer.WriteString(",")
				}

				fmt.Printf(`{"Key":"%s", "Record":"%s"}`, queryResponse.Key, queryResponse.Value)

				buffer.WriteString(
					fmt.Sprintf(`{"Key":"%s", "Record":"%s"}`,
						queryResponse.Key, queryResponse.Value),
				)
				bArrayMemberAlreadyWritten = true

			}
		}

	}
	buffer.WriteString("]")

	fmt.Printf("- viewRequests Result:\n%s\n", buffer.String())

	return shim.Success([]byte(buffer.String()))
}

// viewAllRequests
// ===============
// This function allows to see all the sharing requests present in the ledger.
func (c *Chaincode) viewAllRequests(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	// Perform the range query for the requests
	resultIteratorRequests, err := stub.GetStateByRange("REQUEST", "")
	if err != nil {
		return shim.Error("Error during range query. " + err.Error())
	}
	defer resultIteratorRequests.Close()

	// Write result on the buffer
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultIteratorRequests.HasNext() {
		queryResponse, err := resultIteratorRequests.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		// Check if the key-value pair is a request
		if strings.Contains(queryResponse.Key, "REQUEST") &&
			!strings.Contains(queryResponse.Key, "COUNTER") {
			// Add a comma before array member
			if bArrayMemberAlreadyWritten {
				buffer.WriteString(",")
			}

			fmt.Printf(`{"Key":"%s", "Record":"%s"}`, queryResponse.Key, queryResponse.Value)

			buffer.WriteString(
				fmt.Sprintf(`{"Key":"%s", "Record":"%s"}`,
					queryResponse.Key, queryResponse.Value),
			)
			bArrayMemberAlreadyWritten = true

		}

	}
	buffer.WriteString("]")

	fmt.Printf("- viewAllRequests Result:\n%s\n", buffer.String())

	return shim.Success([]byte(buffer.String()))
}

// acceptRequest
// =============
// This method allow the user to accept a request of data sharing. Given a
// requestId this methods puts the request from Pending to Accepted and copy the
// secret data from a private collection to the other.
func (c *Chaincode) acceptRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of parameters. Expecting 1.")
	}

	// Convert string to int64
	requestIdInput, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Error converting requestId from string to int64")
	}

	// Retrieve the request counter to know the max request number
	requestCounterBytes, err := stub.GetState("REQUESTCOUNTER")
	if err != nil {
		return shim.Error("Error retrieving request counter")
	}

	var requestCounter int64
	err = json.Unmarshal(requestCounterBytes, &requestCounter)
	if err != nil {
		return shim.Error("Error unmarshaling request counter")
	}

	// Check in the request Id is valid
	if requestIdInput > requestCounter || requestIdInput <= 0 {
		return shim.Error("Request ID out of range!")
	}

	// Retrieve the request
	key := "REQUEST" + strconv.FormatInt(requestIdInput, 10)
	requestBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Error retrieving data from the ledger!")
	} else if requestBytes == nil {
		fmt.Println("This request does not exists: " + key)
		return shim.Error("This request does not exist! " + key)
	}

	var requestToAccept assetRequest
	err = json.Unmarshal(requestBytes, &requestToAccept)
	if err != nil {
		return shim.Error("Error unmarshaling data")
	}

	// Check if the caller is the owner of the requested resource
	creator := getCreator(stub)
	if creator != requestToAccept.Owner {
		return shim.Error("You are not allowed to manage this request!")
	}

	// Change the state to ACCEPTED
	requestToAccept.Status = Accepted

	requestBytes, err = json.Marshal(requestToAccept)
	if err != nil {
		return shim.Error("Error marshaling data")
	}

	// Retrieve personal collection name
	personalCollectionName := getCollectionName(stub)

	// Retrieve the private data
	privateDataBytes, err := stub.GetPrivateData(personalCollectionName, requestToAccept.DataId)
	if err != nil {
		return shim.Error("Error retrieving data from personal collection")
	}

	// Retrieve applicant collection name
	applicant := strings.Split(requestToAccept.Applicant, "#")
	if len(applicant) != 2 {
		return shim.Error("Error splitting applicant name!")
	}

	// Create applicant complete collection name
	applicantCollectionName := applicant[1] + "PrivateCollection"

	// Copy the secret data to applicant secret collection
	err = stub.PutPrivateData(applicantCollectionName, requestToAccept.DataId, privateDataBytes)
	if err != nil {
		return shim.Error("Error coping data into the applicante secret collection")
	}

	// Insert back the modified data in the ledger
	err = stub.PutState(key, requestBytes)
	if err != nil {
		return shim.Error("Error in the putState")
	}

	return shim.Success(nil)
}

// denyRequest
// ===========
// This method allow the user to deny a request of data sharing. Given a
// requestId this methods puts the request from Pending to Rejected.
func (c *Chaincode) denyRequest(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of parameters. Expecting 1.")
	}

	// Convert string to int64
	requestIdInput, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Error converting requestId from string to int64")
	}

	// Retrieve the request counter to know the max request number
	requestCounterBytes, err := stub.GetState("REQUESTCOUNTER")
	if err != nil {
		return shim.Error("Error retrieving request counter")
	}

	var requestCounter int64
	err = json.Unmarshal(requestCounterBytes, &requestCounter)
	if err != nil {
		return shim.Error("Error unmarshaling request counter")
	}

	// Check in the request Id is valid
	if requestIdInput > requestCounter || requestIdInput <= 0 {
		return shim.Error("Request ID out of range!")
	}

	// Retrieve the request
	key := "REQUEST" + strconv.FormatInt(requestIdInput, 10)
	requestBytes, err := stub.GetState(key)
	if err != nil {
		return shim.Error("Error retrieving data from the ledger!")
	} else if requestBytes == nil {
		fmt.Println("This request does not exists: " + key)
		return shim.Error("This request does not exist! " + key)
	}

	var requestToReject assetRequest
	err = json.Unmarshal(requestBytes, &requestToReject)
	if err != nil {
		return shim.Error("Error unmarshaling data")
	}

	// Check if the caller is the owner of the requested resource
	creator := getCreator(stub)
	if creator != requestToReject.Owner {
		return shim.Error("You are not allowed to manage this request!")
	}

	// Change the state to REJECTED
	requestToReject.Status = Rejected

	requestBytes, err = json.Marshal(requestToReject)
	if err != nil {
		return shim.Error("Error marshaling data")
	}

	// Insert back the modified data in the ledger
	err = stub.PutState(key, requestBytes)
	if err != nil {
		return shim.Error("Error in the putState")
	}

	return shim.Success(nil)
}

// Main
// ====
func main() {
	err := shim.Start(&Chaincode{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error in chaincode start: %s", err)
		os.Exit(2)
	}
}

// #############################################################################

// getCreator
// ==========
// This function returns a string that represents the creator of the transaction
func getCreator(stub shim.ChaincodeStubInterface) (creator string) {
	Id, _ := cid.GetID(stub)
	MSPID, _ := cid.GetMSPID(stub)
	creator = Id + "#" + MSPID
	return creator
}

// getCollectionName
// =================
// This function returns the name of the private collection of the creatore
// of the transaction
func getCollectionName(stub shim.ChaincodeStubInterface) (collectionName string) {

	MSPID, _ := cid.GetMSPID(stub)
	collectionName = MSPID + "PrivateCollection"

	return collectionName
}

// sliceContains
// =============
// Contains function for slices of strings
func sliceContains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
