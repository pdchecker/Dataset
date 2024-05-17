package main

import (
	"encoding/json"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
	logger "github.com/sirupsen/logrus"
)

// ToChaincodeArgs Prepare chaincode arguments for invoke chaincode

func ToChaincodeArgs(args ...string) [][]byte {
	bargs := make([][]byte, len(args))
	for i, arg := range args {
		bargs[i] = []byte(arg)
	}
	return bargs
}

// InvokeChaincodeOnChannel invokes a specified chaincode function on a different channel except the current channel
func InvokeChaincodeOnChannel(stub shim.ChaincodeStubInterface, chanicodeName string, funcName string, payload string, channelID string) pb.Response {
	var invokeResponse pb.Response
	currentChannelID := stub.GetChannelID()
	logger.Debug("InvokeChaincodeOnChannel: currentChannelID > ", currentChannelID)
	logger.Debug("InvokeChaincodeOnChannel: Called Channel Id > ", channelID)
	// Do not execute on the same channel
	var channels Channels
	channels.ChannelIDs = make([]string, 1)
	channels.ChannelIDs[0] = channelID
	channelsReqJson, _ := json.Marshal(channels)
	invokeRequest := ToChaincodeArgs(funcName, payload, string(channelsReqJson))
	invokeResponse = stub.InvokeChaincode(chanicodeName, invokeRequest, channelID)
	return invokeResponse
}
