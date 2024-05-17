package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type DemoChaincode struct {
}

type item struct {
	Id      string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	State       string `json:"state"`
}

func main() {
	fmt.Println("demo-chaincode main function was called.")
	server := &shim.ChaincodeServer{
		CCID:    os.Getenv("CHAINCODE_CCID"),
		Address: os.Getenv("CHAINCODE_ADDRESS"),
		CC:      new(DemoChaincode),
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}
	fmt.Println("Starting demo-chaincode server...")
	err := server.Start()
	if err != nil {
		fmt.Printf("Error starting demo-chaincode: %s", err)
	}
}

func (t *DemoChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("demo-chaincode Init function was called.")
	return shim.Success(nil)
}

func (t *DemoChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("demo-chaincode Invoke function was called.")
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("demo-chaincode Invoke function: " + function)

	if function == "addItem" {
		return t.addItem(stub, args)
	} else if function == "modifyPrice" {
		return t.modifyPrice(stub, args)
	} else if function == "removeItem" {
		return t.removeItem(stub, args)
	} else if function == "queryItem" {
		return t.queryItem(stub, args)
	} else if function == "history" {
		return t.history(stub, args)
	}

	fmt.Println("demo-chaincode Invoke function does not handle: " + function)
	return shim.Error("Received unknown function invocation")
}

func (t *DemoChaincode) addItem(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5 arguments.")
	}
	if len(strings.TrimSpace(args[0])) <= 0 {
		return shim.Error("The id must be a non-empty string")
	}
	if len(strings.TrimSpace(args[1])) <= 0 {
		return shim.Error("The name must be a non-empty string")
	}
	if len(strings.TrimSpace(args[2])) <= 0 {
		return shim.Error("The description must be a non-empty string")
	}
	if len(strings.TrimSpace(args[3])) <= 0 {
		return shim.Error("The price argument must be a non-empty string")
	}
	if len(strings.TrimSpace(args[4])) <= 0 {
		return shim.Error("The state argument must be a non-empty string")
	}

	id := args[0]
	fmt.Println("demo-chaincode - start addItem ", id)
	itemAsBytes, err := stub.GetState(id)
	if err != nil {
		return shim.Error("Failed to get the Item: " + err.Error())
	} else if itemAsBytes != nil {
		fmt.Println("An Item already exists for this ID: " + id)
		return shim.Error("An Item already exists for this ID: " + id)
	}

	item := &item{args[0], args[1], args[2], args[3], args[4]}
	itemJSONasBytes, err := json.Marshal(item)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(id, itemJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("demo-chaincode - end addItem")
	return shim.Success(itemJSONasBytes)
}

func (t *DemoChaincode) queryItem(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting the ID of the Item.")
	}
	if len(strings.TrimSpace(args[0])) <= 0 {
		return shim.Error("The Item ID must be a non-empty string.")
	}

	id := args[0]
    fmt.Println("demo-chaincode - start queryItem ", id)
	itemAsBytes, err := stub.GetState(id)
	if err != nil {
		return shim.Error("{\"Error\":\"Failed to get the Item for this ID: " + id + "\"}")
	} else if itemAsBytes == nil {
		return shim.Error("{\"Error\":\"An Item does not exist for this ID: " + id + "\"}")
	}

    fmt.Println("demo-chaincode - end queryItem")
	return shim.Success(itemAsBytes)
}

func (t *DemoChaincode) removeItem(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting the ID of the Item.")
	}
	if len(strings.TrimSpace(args[0])) <= 0 {
		return shim.Error("The Item ID must be a non-empty string.")
	}

	id := args[0]
	fmt.Println("demo-chaincode - start removeItem ", id)

	itemAsBytes, err := stub.GetState(id)
	if err != nil {
		return shim.Error("Failed to get the Item: " + err.Error())
	} else if itemAsBytes == nil {
		return shim.Error("{\"Error\":\"An Item does not exist for this ID: " + id + "\"}")
	}

	modifiedItem := item{}
	err = json.Unmarshal(itemAsBytes, &modifiedItem)
	if err != nil {
		return shim.Error(err.Error())
	}
	modifiedItem.State = "REMOVED"

	itemJSONasBytes, _ := json.Marshal(modifiedItem)
	err = stub.PutState(id, itemJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("demo-chaincode - end removeItem")
	return shim.Success(itemJSONasBytes)
}

func (t *DemoChaincode) modifyPrice(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) < 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2 arguments.")
	}
	if len(strings.TrimSpace(args[0])) <= 0 {
		return shim.Error("The Item ID must be a non-empty string.")
	}
	if len(strings.TrimSpace(args[1])) <= 0 {
		return shim.Error("The Item price must be a non-empty string.")
	}

	id := args[0]
	price := args[1]
	fmt.Println("demo-chaincode - start modifyPrice ", id, price)

	itemAsBytes, err := stub.GetState(id)
	if err != nil {
		return shim.Error("Failed to get the Item: " + err.Error())
	} else if itemAsBytes == nil {
		return shim.Error("{\"Error\":\"An Item does not exist for this ID: " + id + "\"}")
	}

	modifiedItem := item{}
	err = json.Unmarshal(itemAsBytes, &modifiedItem)
	if err != nil {
		return shim.Error(err.Error())
	}
	modifiedItem.Price = price

	itemJSONasBytes, _ := json.Marshal(modifiedItem)
	err = stub.PutState(id, itemJSONasBytes)
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("demo-chaincode - end modifyPrice")
	return shim.Success(itemJSONasBytes)
}

func (t *DemoChaincode) history(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting the ID of the Item.")
	}
	if len(strings.TrimSpace(args[0])) <= 0 {
		return shim.Error("The Item ID must be a non-empty string.")
	}

	id := args[0]
	fmt.Printf("demo-chaincode - start history: %s\n", id)
	resultsIterator, err := stub.GetHistoryForKey(id)
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
		if bArrayMemberAlreadyWritten {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")
		buffer.WriteString(", \"Value\":")
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

	fmt.Printf("demo-chaincode - history returning:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}
