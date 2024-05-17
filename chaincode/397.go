/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// EVContract contract for managing CRUD for EVs
type EVContract struct {
	contractapi.Contract
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Record    *EV
	TxId      string    `json:"txId"`
	Timestamp time.Time `json:"timestamp"`
}

// InitLedger creates the initial set of EVs in the ledger.
func (c *EVContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	EVUsers := []EV{
		{EVID: "1", Model: "Tesla", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "2", Model: "Mclaren", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "3", Model: "Mazda", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "4", Model: "BMW", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "5", Model: "Mercedes", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "6", Model: "Aston", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "7", Model: "Renault", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "8", Model: "Tesla", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "9", Model: "Mclaren", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "10", Model: "Mazda", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "11", Model: "BMW", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "12", Model: "Mercedes", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "13", Model: "Aston", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "14", Model: "Renault", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "15", Model: "Tesla", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "16", Model: "Mclaren", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "17", Model: "Mazda", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "18", Model: "BMW", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "19", Model: "Mercedes", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
		{EVID: "20", Model: "Aston", PowerFlow: 0, TotalMoney: 0, RecentMoney: 0, BatteryAge: 0, Temperature: 20, SoC: 100.0, SoH: 100.0},
	}

	for _, User := range EVUsers {
		err := c.CreateEVUser(ctx, User.EVID, User.Model, User.BatteryAge, User.Temperature, User.SoC, User.SoH)
		if err != nil {
			return err
		}
	}

	return nil
}

// EVUserExists checks if a given EV exists
func (c *EVContract) EVUserExists(ctx contractapi.TransactionContextInterface, ID string) (bool, error) {
	evUser := EV{EVID: ID}
	exists, err := evUser.LoadState(ctx)
	if err != nil {
		return false, fmt.Errorf("Could not read from world state. %s", err)
	} else if exists {
		return true, nil
	} else {
		return false, nil
	}
}

// CreateEVUser creates a new instance of EV
func (c *EVContract) CreateEVUser(ctx contractapi.TransactionContextInterface, EVID string, model string, age int, temperature float64, SoC float64, SoH float64) error {
	evUser := new(EV)
	evUser.EVID = EVID
	exists, err := evUser.LoadState(ctx)

	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if exists {
		return fmt.Errorf("The EV %s already exists", EVID)
	}

	newEV := new(EV)
	newEV.Model = model
	newEV.EVID = EVID
	newEV.BatteryAge = age
	newEV.Temperature = temperature
	newEV.SoC = SoC
	newEV.SoH = SoH
	newEV.ChargerID = -1 // init only

	return newEV.SaveState(ctx)
}

// ReadEVData retrieves an instance of EV from the world state
func (c *EVContract) ReadEVData(ctx contractapi.TransactionContextInterface, EVID string) (*EV, error) {
	evUser := new(EV)
	evUser.EVID = EVID
	exists, err := evUser.LoadState(ctx)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return nil, fmt.Errorf("The EV User %s does not exist", EVID)
	}

	return evUser, nil
}

// DeleteEVUser deletes an EV from the world state
func (c *EVContract) DeleteEVUser(ctx contractapi.TransactionContextInterface, EVID string) error {
	evUser := new(EV)
	evUser.EVID = EVID
	exists, err := evUser.LoadState(ctx)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The EV User %s does not exist.", EVID)
	}

	return ctx.GetStub().DelState(EVID)
}

// UpdateEVData is only called by the CSOContract (Invoke Chaincode)
// maybe add powerflow charger
func (c *EVContract) UpdateEVData(ctx contractapi.TransactionContextInterface, EVID string, CSOID string, ChargerID int, PowerFlow float64, RecentMoney float64, Temperature float64, SoC float64, SoH float64) error {
	evUser := new(EV)
	evUser.EVID = EVID
	exists, err := evUser.LoadState(ctx)
	if err != nil {
		return fmt.Errorf("Could not read from world state. %s", err)
	} else if !exists {
		return fmt.Errorf("The EV User %s does not exist", EVID)
	}
	evUser.PowerFlow = PowerFlow
	evUser.CSOID = CSOID
	evUser.ChargerID = ChargerID
	evUser.RecentMoney = RecentMoney
	evUser.TotalMoney += RecentMoney
	evUser.Temperature = Temperature
	evUser.SoC = SoC
	evUser.SoH = SoH
	return evUser.SaveState(ctx)
}

// QueryAll returns a JSON of all the EVs on the blockchain
func (c *EVContract) QueryAll(ctx contractapi.TransactionContextInterface) ([]*EV, error) {
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(EVObjectType, []string{})
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	}
	defer resultsIterator.Close()

	i := 1
	var EVData []*EV
	for resultsIterator.HasNext() {
		fmt.Printf("Ran, %v time", i)
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, fmt.Errorf("unable to get the next element: %s", err.Error())
		}

		var singleEV EV
		if err := json.Unmarshal(response.Value, &singleEV); err != nil {
			return nil, fmt.Errorf("unable to parse the response: %s", err.Error())
		}
		EVData = append(EVData, &singleEV)
		i++
	}
	return EVData, nil
}

//QueryByFields allows users to query with optional fields
//Performs a rich query on couchDB with indexing (paramters can be tuned in the future)
//Parameters and the selectors can be tuend accordingly:
// model -> model of car
// age   -> age of car
// op    -> operator for comparison (i.e. $eq, $gt, $gte, $lt, $lte)
// QueryByFields(Tesla, $gt, 2) will return all Teslas that have age greater than 2
func (c *EVContract) QueryByFields(ctx contractapi.TransactionContextInterface, model string, op string, age int) ([]*EV, error) {
	queryString := fmt.Sprintf(`{"selector":{"model": "%s", "age": {"%s": %v}}, "use_index": ["_design/indexEVDoc", "indexEV"]}`, model, op, age)
	//queryString := fmt.Sprintf(`{"selector":{"model":"%s", "age": %v}}`, model, age)
	println(queryString)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, fmt.Errorf("Could not read from world state. %s", err)
	}
	defer resultsIterator.Close()

	var EVS []*EV
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var ev EV
		err = json.Unmarshal(queryResult.Value, &ev)
		if err != nil {
			return nil, err
		}
		EVS = append(EVS, &ev)
	}

	return EVS, nil
}

// QueryAssetHistory returns the chain of custody for a asset since issuance
func (c *EVContract) QueryAssetHistory(ctx contractapi.TransactionContextInterface, EVID string) ([]QueryResult, error) {
	evUser := new(EV)
	evUser.EVID = EVID
	compositeKey, err := evUser.ToCompositeKey(ctx)
	resultsIterator, err := ctx.GetStub().GetHistoryForKey(compositeKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []QueryResult
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var ev *EV
		err = json.Unmarshal(queryResult.Value, &ev)
		if err != nil {
			return nil, err
		}

		timestamp, err := ptypes.Timestamp(queryResult.Timestamp)
		if err != nil {
			return nil, err
		}
		record := QueryResult{
			TxId:      queryResult.TxId,
			Timestamp: timestamp,
			Record:    ev,
		}
		results = append(results, record)
	}

	return results, nil
}
