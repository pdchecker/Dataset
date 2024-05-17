package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type PseudoRecord struct {
	// 直接拿PID当成索引
	PID              string
	OrigDomain       string
	DestDomain       string
	PubkeyDev2Domain string
	Valid            string
	Tag              string // 谁核验的
	Timestamp        string
	// Sig              []byte // 好像不需要签名，账本可溯源吧
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	recordJSON, _ := json.Marshal(PseudoRecord{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	})
	ctx.GetStub().PutState("redh3tALWAYS", recordJSON)
	return nil
}
func (s *SmartContract) CheckExistance(ctx contractapi.TransactionContextInterface, pid string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(pid)
	if err != nil {
		return false, fmt.Errorf("[chaincode err in CheckExistance()] get state failed.")
	} else if assetJSON != nil {
		return true, nil
	} else {
		// pid不存在
		return false, nil
	}
}
func (s *SmartContract) QueryOne(ctx contractapi.TransactionContextInterface, pid string) (PseudoRecord, error) {
	assetJSON, err := ctx.GetStub().GetState(pid)
	rec := PseudoRecord{}
	if err != nil {
		return rec, err
	}
	err = json.Unmarshal(assetJSON, &rec)
	return rec, err
}
func (s *SmartContract) AddPseudoRecord(ctx contractapi.TransactionContextInterface, jsonstr string) error {
	pseudo_record := PseudoRecord{}
	if json.Unmarshal([]byte(jsonstr), &pseudo_record) != nil {
		fmt.Println("add pseudo: json unmarshal failed.")
	}
	pseudo_record.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	recordJSON, _ := json.Marshal(pseudo_record)
	return ctx.GetStub().PutState(pseudo_record.PID, recordJSON)
}
func (s *SmartContract) QueryAll(ctx contractapi.TransactionContextInterface) ([]*PseudoRecord, error) {
	// 抄的chaincode
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	var assets []*PseudoRecord
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset PseudoRecord
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}
	return assets, nil
}

func main() {
	chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("[main] create chaincode failed.")
	}
	err = chaincode.Start()
	if err != nil {
		log.Panicf("[main] start chaincode failed.")
	}
}
