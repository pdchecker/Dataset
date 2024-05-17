package main

import (
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"seller/factory"
	"seller/farm"
	"time"
)

type Context contractapi.TransactionContextInterface

type SellerContract struct {
	contractapi.Contract
}

type MilkReport struct {
	MilkID    string `json:"milk_id"`
	Operation string `json:"operation"`
}

type Report struct {
	MilkID     string `json:"milk_id"`
	CowID      string `json:"cow_id"`
	MachineID  string `json:"machine_id"`
	ReportData string `json:"report_data"`
	Timestamp  string `json:"timestamp"`
}

type QueryResult struct {
	ReportMap map[string][]Report `json:"report_map"`
}

func (c *SellerContract) Instantiate() {
	fmt.Println("Instantiated")
}

func (c *SellerContract) PutMilkReport(ctx Context, milkID, operation string) error {
	milkReport := MilkReport{
		MilkID:    milkID,
		Operation: operation,
	}

	jsonStr, err := json.Marshal(milkReport)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(milkID, jsonStr)
}

func (c *SellerContract) GetFullMilkHistory(ctx Context, milkID string) (QueryResult, error) {
	histories, err := ctx.GetStub().GetHistoryForKey(milkID)
	if err != nil {
		return QueryResult{}, err
	}

	var sellerReports []Report
	for histories.HasNext() {
		history, err := histories.Next()
		if err != nil {
			return QueryResult{}, err
		}

		var milkReport MilkReport
		if err := json.Unmarshal(history.GetValue(), &milkReport); err != nil {
			return QueryResult{}, err
		}

		sellerReports = append(sellerReports, Report{
			MilkID:     milkReport.MilkID,
			ReportData: milkReport.Operation,
			Timestamp:  timestampToStr(history.GetTimestamp()),
		})
	}

	factoryReports, err := getFactoryMilkHistory(ctx, milkID)
	if err != nil {
		return QueryResult{}, err
	}

	if len(factoryReports) == 0 {
		return QueryResult{}, fmt.Errorf("get no data from factory")
	}

	cowID := factoryReports[0].CowID
	farmReports, err := getFarmCowHistory(ctx, cowID)
	if err != nil {
		return QueryResult{}, err
	}

	var reportMap = map[string][]Report{
		"farm":    farmReports,
		"factory": factoryReports,
		"seller":  sellerReports,
	}

	return QueryResult{ReportMap: reportMap}, nil
}

func getFarmCowHistory(ctx Context, cowID string) ([]Report, error) {
	args := [][]byte{[]byte("getCowHistory"), []byte(cowID)}
	response := ctx.GetStub().InvokeChaincode("farm", args, "milkchannel")

	if response.GetStatus() != 200 { // 200 is OK
		return nil, fmt.Errorf("failed to query farm's chaincode, get error:%+v", response.GetPayload())
	}

	var farmQueryResult []farm.QueryResult
	if err := json.Unmarshal(response.GetPayload(), &farmQueryResult); err != nil {
		return nil, err
	}

	var reports []Report
	for _, fr := range farmQueryResult {
		reports = append(reports, Report{
			CowID:      fr.Report.CowID,
			ReportData: fr.Report.ReportData,
			Timestamp:  fr.Timestamp,
		})
	}

	return reports, nil
}

func getFactoryMilkHistory(ctx Context, milkID string) ([]Report, error) {
	args := [][]byte{[]byte("getMilkHistory"), []byte(milkID)}
	response := ctx.GetStub().InvokeChaincode("factory", args, "milkchannel")

	if response.GetStatus() != 200 { // 200 is OK
		return nil, fmt.Errorf("failed to query factory's chaincode, get error:%+v", response.GetPayload())
	}

	var factoryQueryResult []factory.QueryResult
	if err := json.Unmarshal(response.GetPayload(), &factoryQueryResult); err != nil {
		return nil, err
	}

	var reports []Report
	for _, fr := range factoryQueryResult {
		reports = append(reports, Report{
			MilkID:     fr.Report.MilkID,
			CowID:      fr.Report.CowID,
			MachineID:  fr.Report.MachineID,
			ReportData: fr.Report.ReportData,
			Timestamp:  fr.Timestamp,
		})
	}
	return reports, nil
}

func timestampToStr(ts *timestamp.Timestamp) string {
	tm := time.Unix(ts.Seconds, 0)
	return tm.Format("2006-01-02 03:04:05 PM")
}
