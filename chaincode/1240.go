package main

import (
	"poc_kinder/contract/action/report"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-protos-go/peer"
)

// Invoke is called per transaction on the chaincode. Each transaction is
// either a 'get' or a 'set' on the asset created by Init function. The Set
// method may create a new asset by specifying a new key-value pair.
func (t *KinderContract) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// Extract the function and args from the transaction proposal
	fn, args := stub.GetFunctionAndParameters()

	var result string
	var err error

	switch fn {
	case "createReport":
		result, err = report.Create(stub, args)
	case "getReports":
		result, err = report.GetAll(stub, args)
	case "approveReport":
		result, err = report.Approve(stub, args)
	}

	if err != nil {
		return shim.Error(err.Error())
	}

	// Return the result as success payload
	return shim.Success([]byte(result))
}
