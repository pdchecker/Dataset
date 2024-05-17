package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

type SimpleChaincode struct {
}

/*
 * The Init method is called when the Smart Contract is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (t *SimpleChaincode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	// fmt.Println("ugachaincode - Init")
	// return shim.Success(nil)

	fmt.Println("ugachaincode - Init")
	_, args := stub.GetFunctionAndParameters()
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var err error

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	// Initialize the chaincode
	A = args[0]
	Aval, err = strconv.Atoi(args[1])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	B = args[2]
	Bval, err = strconv.Atoi(args[3])
	if err != nil {
		return shim.Error("Expecting integer value for asset holding")
	}
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// err = stub.PutState("c", []byte("VALID"))
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	return shim.Success(nil)
}

func (t *SimpleChaincode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("ugachaincode - Invoke")
	function, args := stub.GetFunctionAndParameters()
	if function == "delete" {
		return t.delete(stub, args)
	} else if function == "invoke" {
		return t.invoke(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	} else if function == "add" {
		return t.add(stub, args)
	} else if function == "setInvalid" {
		return t.setInvalid(stub, args)
	} else if function == "setValid" {
		return t.setValid(stub, args)
	} else if function == "setFraudulent" {
		return t.setFraudulent(stub, args)
	} else if function == "readAll" {
		return t.readAll(stub)
	}

	return shim.Error("Invalid invoke function name. Expecting \"delete\" \"query\" \"add\" \"setInvalid\" \"setValid\" \"setFraudulent\"")
}

// Transaction makes payment of X units from A to B
func (t *SimpleChaincode) invoke(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var A, B string    // Entities
	var Aval, Bval int // Asset holdings
	var X int          // Transaction value
	var err error

	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 3")
	}

	A = args[0]
	B = args[1]

	// Get the state from the ledger
	// TODO: will be nice to have a GetAllState call to ledger
	Avalbytes, err := stub.GetState(A)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Avalbytes == nil {
		return shim.Error("Entity not found")
	}
	Aval, _ = strconv.Atoi(string(Avalbytes))

	Bvalbytes, err := stub.GetState(B)
	if err != nil {
		return shim.Error("Failed to get state")
	}
	if Bvalbytes == nil {
		return shim.Error("Entity not found")
	}
	Bval, _ = strconv.Atoi(string(Bvalbytes))

	// Perform the execution
	X, err = strconv.Atoi(args[2])
	if err != nil {
		return shim.Error("Invalid transaction amount, expecting a integer value")
	}
	Aval = Aval - X
	Bval = Bval + X
	fmt.Printf("Aval = %d, Bval = %d\n", Aval, Bval)

	// Write the state back to the ledger
	err = stub.PutState(A, []byte(strconv.Itoa(Aval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	err = stub.PutState(B, []byte(strconv.Itoa(Bval)))
	if err != nil {
		return shim.Error(err.Error())
	}

	// err = stub.PutState("c", []byte("VALID"))
	// if err != nil {
	// 	return shim.Error(err.Error())
	// }

	return shim.Success(nil)
}

// Deletes an entity from state
func (t *SimpleChaincode) delete(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	A := args[0]
	fmt.Println("ugachaincode - delete(hash: " + A + ")")

	// Delete the key from the state in ledger
	err := stub.DelState(A)
	if err != nil {
		return shim.Error("Failed to delete state")
	}

	return shim.Success(nil)
}

// query callback representing the query of a chaincode
func (t *SimpleChaincode) query(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var hash string // Diploma hash to query
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting hash of the diploma to query")
	}

	hash = args[0]
	fmt.Println("ugachaincode - query(hash: " + hash + ")")

	// Get the state from the ledger
	hashStateBytes, err := stub.GetState(hash)
	if err != nil {
		fmt.Println("Error while getting state from the ledger: " + err.Error())
		return shim.Error(err.Error())
	}
	if hashStateBytes == nil {
		fmt.Println("NOT_FOUND")
		return shim.Success([]byte("NOT_FOUND - query(hash: " + hash + ")"))
	}

	fmt.Printf("Query Response for %s: %s\n", hash, string(hashStateBytes))
	return shim.Success(hashStateBytes)
}

// Add the diploma hash to the blockchain and validate it
func (t *SimpleChaincode) add(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var hash string // Diploma hash to store in the ledger
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting diplomas hash to add.")
	}

	hash = args[0]
	fmt.Println("ugachaincode - add(hash: " + hash + ")")

	// Checking that the diploma hash has not been added yet
	hashStateBytes, err := stub.GetState(hash)
	if err != nil {
		fmt.Println("Error while getting state from the ledger: " + err.Error())
		return shim.Error(err.Error())
	}
	if hashStateBytes != nil {
		return shim.Success([]byte("ALREADY_EXIST - add(hash: " + hash + ")"))
	}

	// Write the state valid to the ledger
	err = stub.PutState(hash, []byte("VALID"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("hash: " + hash + " STATE set to VALID"))
}

func (t *SimpleChaincode) readAll(stub shim.ChaincodeStubInterface) pb.Response {
	// Read all KV of the ledger
	iter, err := stub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	str := ""

	for iter.HasNext() {
		kv, err2 := iter.Next()
		if err2 != nil {
			fmt.Println(err2)
		}
		str += "key:"+kv.GetKey()+", value:"+string(kv.GetValue())+"; "
	}

	return shim.Success([]byte(str))
}

// Validate an existing diploma in the blockchain
func (t *SimpleChaincode) setValid(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var hash string // Hash of the diploma to set
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting diplomas hash to validate.")
	}

	hash = args[0]
	fmt.Println("ugachaincode - setValid(hash: " + hash + ")")

	// Checking that the diploma hash exists
	hashStateBytes, err := stub.GetState(hash)
	if err != nil {
		fmt.Println("Error while getting state from the ledger: " + err.Error())
		return shim.Error(err.Error())
	}
	if hashStateBytes == nil {
		return shim.Success([]byte("NOT_FOUND"))
	}

	// Checking that the state is not already set
	if string(hashStateBytes) == "VALID" {
		return shim.Success([]byte("STATE_ALREADY_SET"))
	}

	// Write the state to the ledger
	err = stub.PutState(hash, []byte("VALID"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Invalidate an existing diploma in the blockchain
func (t *SimpleChaincode) setInvalid(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var hash string // Hash of the diploma to set
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting diplomas hash to invalidate.")
	}

	hash = args[0]
	fmt.Println("ugachaincode - setInvalid(hash: " + hash + ")")

	// Checking that the diploma hash exists
	hashStateBytes, err := stub.GetState(hash)
	if err != nil {
		fmt.Println("Error while getting state from the ledger: " + err.Error())
		return shim.Error(err.Error())
	}
	if hashStateBytes == nil {
		return shim.Success([]byte("NOT_FOUND"))
	}

	// Checking that the state is not already set
	if string(hashStateBytes) == "INVALID" {
		return shim.Success([]byte("STATE_ALREADY_SET"))
	}

	// Write the state to the ledger
	err = stub.PutState(hash, []byte("INVALID"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

// Set an existing diploma to fraudulent
func (t *SimpleChaincode) setFraudulent(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var hash string // Hash of the diploma to set
	var err error

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting diplomas hash to set to fraudulent.")
	}

	hash = args[0]
	fmt.Println("ugachaincode - setFraudulent(hash: " + hash + ")")

	// Checking if the diploma hash exists
	hashStateBytes, err := stub.GetState(hash)
	if err != nil {
		fmt.Println("Error while getting state from the ledger: " + err.Error())
		return shim.Error(err.Error())
	}
	if hashStateBytes == nil {
		return shim.Success([]byte("NOT_FOUND"))
	}

	// Checking that the state is not already set
	if string(hashStateBytes) == "FRAUDULENT" {
		return shim.Success([]byte("STATE_ALREADY_SET"))
	}

	// Write the state to the ledger
	err = stub.PutState(hash, []byte("FRAUDULENT"))
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success(nil)
}

func main() {
	err := shim.Start(new(SimpleChaincode))
	if err != nil {
		fmt.Printf("Error starting UGAChaincode: %s", err)
	}
}
