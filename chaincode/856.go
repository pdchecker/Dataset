package main

import (
	"log"

	"github.com/chunsik-is-meow/blockchain/src/asset/chaincodes/trade/contract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
	tradeChaincode, err := contractapi.NewChaincode(&contract.TradeChaincode{})
	if err != nil {
		log.Panicf("Error creating tradeChaincode: %v", err)
	}

	if err := tradeChaincode.Start(); err != nil {
		log.Panicf("Error starting tradeChaincode: %v", err)
	}
}
