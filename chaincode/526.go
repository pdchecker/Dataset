/*
	2021 Baran Kılıç <baran.kilic@boun.edu.tr>

	SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"erc1155/chaincode"
	"os"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

func main() {

	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	smartContract := new(chaincode.SmartContract)

	chaincode, err := contractapi.NewChaincode(smartContract)

	if err != nil {
		panic(err.Error())
	}

	chaincode.Info.Title = "ERC-1155 chaincode"
	chaincode.Info.Version = "0.0.1"

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		panic("Failed to start erc-1155 chaincode. " + err.Error())
	}
}
