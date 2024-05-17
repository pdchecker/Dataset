package main

import (
	"encoding/json"
	"hyperledger_dapp/model"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func SaveERC20Metadata(stub shim.ChaincodeStubInterface, tokenName, symbol, owner string, amountUint uint) error {

	// make metadata
	metadata := model.NewERC20Metadata(tokenName, symbol, owner, uint(amountUint))
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return model.NewCustomError(model.MarshalErrorType, "metadata", err.Error())
	}

	// save token meta data
	err = stub.PutState(tokenName, metadataBytes)
	if err != nil {
		return model.NewCustomError(model.PutStateErrorType, "putstate", err.Error())
	}

	return nil
}

func GetERC20TotalSupply(stub shim.ChaincodeStubInterface, tokenName string) (*uint64, error) {

	// get metadata
	metadata := &model.ERC20Metadata{}

	metadataBytes, err := stub.GetState(tokenName)
	if err != nil {
		return nil, model.NewCustomError(model.GetStateErrorType, "balance", err.Error())
	}

	err = json.Unmarshal(metadataBytes, metadata)
	if err != nil {
		return nil, model.NewCustomError(model.UnmarshalErrorType, "unmarshal", err.Error())
	}

	return metadata.GetTotalSupply(), nil
}

func SaveBalance(stub shim.ChaincodeStubInterface, owner, balance string) error {

	// save owner balance
	err := stub.PutState(owner, []byte(balance))
	if err != nil {
		return model.NewCustomError(model.PutStateErrorType, "balance", err.Error())
	}
	return nil
}

func GetBalance(stub shim.ChaincodeStubInterface, owner string, isZero bool) (*int, error) {
	// get caller amount
	AmountBytes, err := stub.GetState(owner)
	if err != nil {
		return nil, model.NewCustomError(model.GetStateErrorType, "balance", err.Error())
	}

	if isZero && AmountBytes == nil {
		AmountBytes = []byte("0")
	}

	amount, err := strconv.Atoi(string(AmountBytes))
	if err != nil {
		return nil, model.NewCustomError(model.ConvertErrorType, "amount", err.Error())
	}

	return &amount, nil
}

func GetERC20Metadata(stub shim.ChaincodeStubInterface, tokenName string) (*model.ERC20Metadata, error) {

	metadata := &model.ERC20Metadata{}

	metadataBytes, err := stub.GetState(tokenName)
	if err != nil {
		return nil, model.NewCustomError(model.GetStateErrorType, "balance", err.Error())
	}

	err = json.Unmarshal(metadataBytes, metadata)
	if err != nil {
		return nil, model.NewCustomError(model.UnmarshalErrorType, "unmarshal", err.Error())
	}

	return metadata, nil
}
