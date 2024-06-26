package main

import (
	. "github.com/davidkhala/goutils"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

type CommonChaincode struct {
	Name      string
	TxID      string
	ChannelID string
	CCAPI     shim.ChaincodeStubInterface // chaincode API
}

func (cc *CommonChaincode) Prepare(ccAPI shim.ChaincodeStubInterface) {
	cc.CCAPI = ccAPI
	cc.ChannelID = ccAPI.GetChannelID()
	cc.TxID = ccAPI.GetTxID()
}

// GetChaincodeID return empty for if no record.
func (cc *CommonChaincode) GetChaincodeID() string {
	var iterator, _ = cc.GetStateByRangeWithPagination("", "", 1, "")
	if !iterator.HasNext() {
		return ""
	}
	var kv, err = iterator.Next()
	PanicError(err)
	var name = kv.GetNamespace()
	cc.Name = name
	return name
}
