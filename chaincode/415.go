package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"time"
)

// alias name for TransactionContextInterface
type Context contractapi.TransactionContextInterface

type FactoryContract struct {
	contractapi.Contract
}

type QueryResult struct {
	Report *MilkReport `json:"milk_report"`
	Timestamp string `json:"timestamp"`
}

type MilkReport struct {
	MilkID    string `json:"milk_id"`
	CowID     string `json:"cow_id"`
	MachineID string `json:"machine_id"`
	ReportData string `json:"report_data"`
}

func (c *FactoryContract) Instantiate() {
	fmt.Print("Instantiating contract")
}

func (c *FactoryContract) PutMilkReport(ctx Context, milkID, cowID, machineID, operation string) error {
	milkReport := MilkReport{
		MilkID:    milkID,
		CowID:     cowID,
		MachineID: machineID,
		ReportData: operation,
	}
	jsonStr, err := json.Marshal(milkReport)
	if err != nil {
		return fmt.Errorf("json marshal failed, err: %+v", err)
	}
	return ctx.GetStub().PutState(milkID, jsonStr)
}

func (c *FactoryContract) GetMilkHistory(ctx Context, milkID string) ([]QueryResult, error) {
	histories, err := ctx.GetStub().GetHistoryForKey(milkID)
	if err != nil {
		return nil, err
	}

	var results []QueryResult

	for histories.HasNext() {
		history, err := histories.Next()
		if err != nil {
			return nil, err
		}

		var milkReport MilkReport
		if err := json.Unmarshal(history.GetValue(), &milkReport); err != nil {
			return nil, err
		}

		results = append(results, QueryResult{
			Report: &milkReport,
			Timestamp: timestampToStr(history.GetTimestamp()),
		})
	}

	return results, nil
}

func timestampToStr(ts *timestamp.Timestamp) string {
	tm := time.Unix(ts.Seconds, 0)
	return tm.Format("2006-01-02 03:04:05 PM")
}
