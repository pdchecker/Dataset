package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric/common/flogging"
	"github.com/zhj0811/dbzl/common/define"
)

var logger = flogging.MustGetLogger("handler")

//func SavePolicy(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
//	logger.Infof("Enter .....%s", function)
//	var policy define.Policy
//	txId := stub.GetTxID()
//	err := json.Unmarshal([]byte(args[0]), &policy)
//	if err != nil {
//		return nil, err
//	}
//	err = stub.PutState(txId, []byte(args[0]))
//	if err != nil {
//		return nil, err
//	}
//	//p := Policy{PId: txId}
//	p, err := json.Marshal(&Policy{PId: txId})
//	if err != nil {
//		return nil, err
//	}
//	err = stub.PutState(policy.Number, p)
//	if err != nil {
//		return nil, err
//	}
//	return nil, nil
//}
//
//func SaveService(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
//	logger.Infof("Enter .....%s", function)
//	var req define.Service
//	txId := stub.GetTxID()
//	err := json.Unmarshal([]byte(args[0]), &req)
//	if err != nil {
//		return nil, err
//	}
//	err = stub.PutState(txId, []byte(args[0]))
//	if err != nil {
//		return nil, err
//	}
//
//	pBytes, err := stub.GetState(req.Number)
//	if err != nil {
//		return nil, err
//	}
//	policy := &Policy{}
//	err = json.Unmarshal(pBytes, policy)
//	if err != nil {
//		return nil, err
//	}
//	policy.SIds = append(policy.SIds, txId)
//	p, err := json.Marshal(policy)
//	if err != nil {
//		return nil, err
//	}
//	err = stub.PutState(req.Number, p)
//	if err != nil {
//		return nil, err
//	}
//	return nil, nil
//}

func QueryByTxID(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Infof("Get value by key of tx id %s", args[0])
	return stub.GetState(args[0])
}

func QueryByKey(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Infof("Get value by key %s", args[0])
	return stub.GetState(args[0])
}

func SavePolicy(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Infof("Enter .....%s", function)
	policy := &define.Policy{}
	txId := stub.GetTxID()
	err := json.Unmarshal([]byte(args[0]), policy)
	if err != nil {
		return nil, err
	}
	policy.TxID = txId
	bytes, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}
	err = stub.PutState(policy.ID, bytes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func SaveService(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Infof("Enter .....%s", function)
	req := &define.Service{}
	txId := stub.GetTxID()
	err := json.Unmarshal([]byte(args[0]), &req)
	if err != nil {
		return nil, err
	}
	req.TxID = txId
	bytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = stub.PutState(req.ID, bytes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func SaveCompany(stub shim.ChaincodeStubInterface, function string, args []string) ([]byte, error) {
	logger.Infof("Enter .....%s", function)
	req := &define.Company{}
	txId := stub.GetTxID()
	err := json.Unmarshal([]byte(args[0]), &req)
	if err != nil {
		return nil, err
	}
	req.TxID = txId
	bytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = stub.PutState(req.ID, bytes)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
