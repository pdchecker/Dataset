/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// EVObjectType for composite key
const EVObjectType = "EV-Owner"

// EV stores data
type EV struct {
	EVID        string  `json:"EVID"`
	CSOID       string  `json:"CSOID"`
	Model       string  `json:"model"`
	PowerFlow   float64 `json:"power_flow"`
	TotalMoney  float64 `json:"total_money"`
	RecentMoney float64 `json:"recent_money"`
	BatteryAge  int     `json:"battery_age"`
	Temperature float64 `json:"temperature"`
	SoC         float64 `json:"state_of_charge"`
	SoH         float64 `json:"state_of_health"`
	ChargerID   int     `json:"charger_ID"`
}

// Each update contains:
// EVID CSOID ChargeID Powerflow SOC Money BatteryTemp SOH

type Charger struct {
	EVID      string `json:"EVID"` // null if no EV connected
	ChargerID int    `json:"charger_ID"`
	PowerFlow int    `json:"power_flow"` // must be same as EV powerflow
}

// CSO stores data
type CSO struct {
	CSOID          string    `json:"ID"`
	TotalPowerFlow float64   `json:"total_power_flow"`
	EVCount        int       `json:"ev_count"`
	Chargers       []Charger `json:"charger"`
}

// ToCompositeKey returns a composite key based on the ID and accountType
func (c *EV) ToCompositeKey(ctx contractapi.TransactionContextInterface) (string,
	error) {
	attributes := []string{

		c.EVID,
	}
	return ctx.GetStub().CreateCompositeKey(EVObjectType, attributes)
}

// ToLedgerValue creates a JSON-encoded account
func (c *EV) ToLedgerValue() ([]byte, error) {
	return json.Marshal(c)
}

// SaveState saves the accounts into the ledger
func (c *EV) SaveState(ctx contractapi.TransactionContextInterface) error {
	compositeKey, err := c.ToCompositeKey(ctx)
	if err != nil {
		message := fmt.Sprintf("Unable to create a composite key: %s", err.Error())
		return errors.New(message)
	}

	ledgerValue, err := c.ToLedgerValue()

	if err != nil {
		message := fmt.Sprintf("Unable to  compose a ledger value: %s", err.Error())
		return errors.New(message)
	}
	return ctx.GetStub().PutState(compositeKey, ledgerValue)
}

// LoadState loads the data from the ledger into the EV object if the data is found
// Returns false if an Account object wasn't found in the ledger; otherwise
//returns true.
func (c *EV) LoadState(ctx contractapi.TransactionContextInterface) (bool, error) {
	compositeKey, err := c.ToCompositeKey(ctx)
	if err != nil {
		message := fmt.Sprintf("Unable to create a composite key: %s", err.Error())
		return false, errors.New(message)
	}

	ledgerValue, err := ctx.GetStub().GetState(compositeKey)
	if err != nil {
		message := fmt.Sprintf("Unable to  compose a ledger value: %s", err.Error())
		return false, errors.New(message)
	}

	if ledgerValue == nil {
		return false, nil
	}

	return true, json.Unmarshal(ledgerValue, &c)
}
