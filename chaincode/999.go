/*
 * @Author: guiguan
 * @Date:   2020-08-12T15:22:41+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2020-08-13T13:00:16+10:00
 */

package main

import (
	"fmt"

	"github.com/SouthbankSoftware/provendb-hyperledger/chaincode/common"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for ProvenDB
type SmartContract struct {
	contractapi.Contract
}

// EmbedData embeds the data to the Hyperledger network
func (s *SmartContract) EmbedData(
	ctx contractapi.TransactionContextInterface,
	data string,
) (rp *common.EmbedDataReply, er error) {
	stub := ctx.GetStub()

	err := stub.PutState(common.StateKeyData, []byte(data))
	if err != nil {
		er = err
		return
	}

	tt, err := stub.GetTxTimestamp()
	if err != nil {
		er = err
		return
	}

	ct, err := ptypes.Timestamp(tt)
	if err != nil {
		er = err
		return
	}

	rp = &common.EmbedDataReply{
		TxnID:      stub.GetTxID(),
		CreateTime: ct,
	}
	return
}

func main() {
	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create provendb chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting provendb chaincode: %s", err.Error())
	}
}
