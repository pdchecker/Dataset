package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// alias name for TransactionContextInterface
type Context contractapi.TransactionContextInterface

type FarmContract struct {
	contractapi.Contract
}

type CowReport struct {
	CowID      string `json:"cow_id"`
	ReportData string `json:"report_data"`
}

type QueryResult struct {
	Report *CowReport `json:"cow_report"`
	Timestamp string `json:"timestamp"`
}

func (c *FarmContract) Instantiate() {
	fmt.Println("Instantiated")
}

func (c *FarmContract) PutCowReport(ctx Context, cowID, reportData string) error {
	report := CowReport{cowID, reportData}
	jsonStr, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("json marshal failed, err: %+v", err)
	}
	return ctx.GetStub().PutState(cowID, jsonStr)
}

func (c *FarmContract) GetCowHistory(ctx Context, cowID string) ([]QueryResult, error) {
	histories, err := ctx.GetStub().GetHistoryForKey(cowID)
	if err != nil {
		return nil, fmt.Errorf("GetHistoryForKey failed, cowID: %s, err: %+v", cowID, err)
	}

	results := make([]QueryResult, 0)
	for histories.HasNext() {
		history, err := histories.Next()
		if err != nil {
			return nil, fmt.Errorf("GetHistoryForKey failed, cowID:%s err:%+v", cowID, err)
		}
		value := history.GetValue()

		var report CowReport
		if err := json.Unmarshal(value, &report); err != nil {
			return nil, fmt.Errorf("unmarshal failed, err: %+v", err)
		}
		ts := history.GetTimestamp()
		results = append(results, QueryResult{Report: &report, Timestamp: timestampToStr(ts)})
	}

	return results, nil
}

func timestampToStr(ts *timestamp.Timestamp) string {
	tm := time.Unix(ts.Seconds, 0)
	return tm.Format("2006-01-02 03:04:05 PM")
}
