package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

type SmartContract struct {
	contractapi.Contract
}

// 屠宰场名称、进厂检疫、清洗消毒、屠宰时间、宰后检疫、分割信息
type Process struct {
	SlaughterhouseName      string `json:"slaughterhouseName"`
	EntryQuarantine         string `json:"entryQuarantine"`
	Cleaning                string `json:"cleaning"`
	SlaughteringTime        string `json:"slaughteringTime"`
	Quarantine              string `json:"quarantine"`
	SegmentationInformation string `json:"segmentationInformation"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Process
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	Process := []Process{
		Process{"兰州肉多多有限责任公司", "true", "true", "2022.2.1", "true", "酮体"},
	}
	for i, item := range Process {
		itemAsBytes, _ := json.Marshal(item)
		err := ctx.GetStub().PutState("Process"+strconv.Itoa(i), itemAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, processNumber string, slaughterhouseName string, entryQuarantine string, cleaning string,
	slaughteringTime string, quarantine string, segmentationInformation string) error {
	process := Process{
		SlaughterhouseName:      slaughterhouseName,
		EntryQuarantine:         entryQuarantine,
		Cleaning:                cleaning,
		SlaughteringTime:        slaughteringTime,
		Quarantine:              quarantine,
		SegmentationInformation: segmentationInformation,
	}

	itemAsBytes, _ := json.Marshal(process)

	return ctx.GetStub().PutState(processNumber, itemAsBytes)
}

func (s *SmartContract) Query(ctx contractapi.TransactionContextInterface, processNumber string) (*Process, error) {
	itemAsBytes, err := ctx.GetStub().GetState(processNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if itemAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", processNumber)
	}

	process := new(Process)
	_ = json.Unmarshal(itemAsBytes, process)

	return process, nil
}

func (s *SmartContract) QueryAll(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer func(resultsIterator shim.StateQueryIteratorInterface) {
		err := resultsIterator.Close()
		if err != nil {

		}
	}(resultsIterator)

	var results []QueryResult

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		process := new(Process)
		_ = json.Unmarshal(queryResponse.Value, process)

		queryResult := QueryResult{Key: queryResponse.Key, Record: process}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) Delete(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.ProcessExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}
func (s *SmartContract) ProcessExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create farm chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting farm chaincode: %s", err.Error())
	}
}
