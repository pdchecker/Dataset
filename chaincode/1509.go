package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"

)

type User struct {

	// id
	ID string `json:"id,omitempty"`

	// name
	Name string `json:"name,omitempty"`
}

type Molecule struct {

	// calculated properties
	CalculatedProperties []string `json:"calculated-properties"`

	// comments
	Comments string `json:"comments,omitempty"`

	// documented properties
	DocumentedProperties []string `json:"documented-properties"`

	// inchi
	Inchi string `json:"inchi,omitempty"`

	// institution
	Institution string `json:"institution,omitempty"`

	// name
	// Required: true
	Name *string `json:"name"`

	// smiles
	Smiles string `json:"smiles,omitempty"`

	// uploader
	Uploader string `json:"uploader,omitempty"`

	// value
	Value int64 `json:"value,omitempty"`

	Owner string
}

// Chaincode is the definition of the chaincode structure.
type Chaincode struct {
}

// Init is called when the chaincode is instantiated by the blockchain network.
func (cc *Chaincode) Init(stub shim.ChaincodeStubInterface) sc.Response {
	_, _ = stub.GetFunctionAndParameters()

	return shim.Success(nil)
}

// Invoke is called as a result of an application request to run the chaincode.
func (cc *Chaincode) Invoke(stub shim.ChaincodeStubInterface) sc.Response {
	fcn, params := stub.GetFunctionAndParameters()

	switch fcn {
	case "TransferOwnership":
		return cc.TransferOwnership(stub, params)
	case "GetHistoryForAsset":
		return cc.GetHistoryForAsset(stub, params)
	case "UploadMolecule":
		return cc.UploadMolecule(stub, params)
	case "CreateUser":
		return cc.CreateUser(stub, params)
	case "UpdateUser":
		return cc.UpdateUser(stub, params)
	case "QueryMolecules":
		return cc.QueryMolecules(stub, params)
	default:
		return shim.Error("No match in function name")
	}
}

func (cc *Chaincode) CreateUser(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	id := args[0]
	name := args[1]

	userExists, err := stub.GetState(id)
	if err != nil {
		return shim.Error("Failed to verify if user already exists")
	}

	if userExists != nil {
		return shim.Error("User already exists")
	}

	newUser := User{
		ID:   id,
		Name: name,
	}

	newUserAsJSONBytes, err := json.Marshal(newUser)
	if err != nil {
		return shim.Error("Failed to marshall new user to bytes")
	}

	err = stub.PutState(id, newUserAsJSONBytes)
	if err != nil {
		return shim.Error("Failed to update state for new user")
	}

	return shim.Success([]byte("New user was created"))
}

func (cc *Chaincode) UpdateUser(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	id := args[0]
	name := args[1]

	userAsBytes, err := stub.GetState(id)
	if err != nil {
		return shim.Error("Failed to retrieve old user")
	}

	if userAsBytes == nil {
		return shim.Error("User does not exist")
	}

	user := &User{}
	err = json.Unmarshal(userAsBytes, user)

	if err != nil {
		return shim.Error("Failed to unmarshall user")
	}

	user.Name = name

	userAsJSONBytes, err := json.Marshal(user)
	if err != nil {
		return shim.Error("Failed to Marhsal user as Json bytes")
	}

	err = stub.PutState(id, userAsJSONBytes)
	if err != nil {
		return shim.Error("Failed to update state for user")
	}

	return shim.Success(userAsJSONBytes)
}

func (cc *Chaincode) QueryMolecules(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	queryString := args[0]

	resultsIterator, err := stub.GetQueryResult(queryString)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryRecords
	var buffer bytes.Buffer

	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {
		queryResponse,
			err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten {
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
	fmt.Printf("- getQueryResultForQueryString queryResult:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (cc *Chaincode) TransferOwnership(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//@TODO: make sure stub.GetCreator is equal to current owner
	index := args[0]
	newOwnerID := args[1]

	assetAsBytes, err := stub.GetState(index)
	if err != nil {
		return shim.Error("Error in retrieving asset")
	}

	if assetAsBytes == nil {
		return shim.Error("Asset does not exist")
	}

	newOwnerAsbytes, err := stub.GetState(newOwnerID)
	if err != nil {
		return shim.Error("Failed to retrieve new owner")
	}

	if newOwnerAsbytes == nil {
		return shim.Error("User does not exist")
	}

	asset := Molecule{}
	err = json.Unmarshal(assetAsBytes, &asset)

	if err != nil {
		return shim.Error("Failed to unmarshal assetAsBytes " + err.Error())
	}

	newOwner := User{}
	err = json.Unmarshal(newOwnerAsbytes, &newOwner)

	if err != nil {
		return shim.Error("Failed to unmarshal newOwnerAsBytes")
	}

	asset.Owner = newOwner.ID

	assetAsJSONBytes, err := json.Marshal(asset)
	if err != nil {
		return shim.Error("failed to marshal asset as bytes")
	}

	err = stub.PutState(index, assetAsJSONBytes)
	if err != nil {
		return shim.Error("Failed to update asset state")
	}

	return shim.Success(assetAsJSONBytes)
}

func (cc *Chaincode) GetHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	key := args[0]

	history, err := stub.GetHistoryForKey(key)

	if err != nil {
		return shim.Error("Failed to retrieve History for key")
	}
	defer history.Close()

	var result []byte

	for history.HasNext() {
		modification, err := history.Next()
		if err != nil {
			return shim.Error("error in iter of history")
		}

		result = append(result, modification.Value...)
	}

	return shim.Success(result)
}

func (cc *Chaincode) UploadMolecule(stub shim.ChaincodeStubInterface, args []string) sc.Response {
	//@TODO: make sure that index does not already exist!
	argumentMap := args[0]

	newAsset := Molecule{}

	err := json.Unmarshal([]byte(argumentMap), &newAsset)
	if err != nil {
		return shim.Error("Failed to unmarshall bytes: " + err.Error())
	}

	newAssetAsJSONBytes, err := json.Marshal(newAsset)
	if err != nil {
		return shim.Error("Failed to marshall bytes: " + err.Error())
	}

	err = stub.PutState(*newAsset.Name, newAssetAsJSONBytes)
	if err != nil {
		return shim.Error("Failed to put state for new asset: " + err.Error())
	}

	return shim.Success(newAssetAsJSONBytes)
}
