/*
 * SPDX-License-Identifier: Apache-2.0
 */

package main

import (
	"encoding/json"
	"fmt"

	"admin_contract/types"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	uuid "github.com/satori/go.uuid"
)

type Fund struct {
	DocType            string `json:"docType"`
	ID                 string `json:"id"`
	Name               string `json:"name"`
	CurrentPeriod      int    `json:"currentPeriod"`
	InceptionDate      string `json:"inceptionDate"`
	PeriodClosingValue string `json:"periodClosingValue"`
	PeriodOpeningValue string `json:"periodOpeningValue"`
	AggregateFixedFees string `json:"aggregateFixedFees"`
	AggregateDeposits  string `json:"aggregateDeposits"`
	NextInvestorNumber int    `json:"nextInvestorNumber"`
	PeriodUpdated      bool   `json:"periodUpdated"`
}

type HighWaterMark struct {
	Amount string `json:"amount"`
	Date   string `json:"date"`
}

type CapitalAccount struct {
	DocType             string        `json:"docType"`
	ID                  string        `json:"id"`
	Fund                string        `json:"fund"`
	Investor            string        `json:"name"`
	Number              int           `json:"number"`
	CurrentPeriod       int           `json:"currentPeriod"`
	PeriodClosingValue  string        `json:"periodClosingValue"`
	PeriodOpeningValue  string        `json:"periodOpeningValue"`
	FixedFees           string        `json:"fixedFees"`
	Deposits            string        `json:"deposits"`
	OwnershipPercentage string        `json:"ownershipPercentage"`
	HighWaterMark       HighWaterMark `json:"highWaterMark"`
}

type CapitalAccountAction struct {
	DocType     string `json:"docType"`
	ID          string `json:"id"`
	Type        string `json:"type"`
	Amount      string `json:"amount"`
	Full        bool   `json:"full"`
	Status      string `json:"status"`
	Description string `json:"description"`
	Date        string `json:"Date"`
	Period      int    `json:"period"`
}

type Investor struct {
	DocType string `json:"docType"`
	ID      string `json:"id"`
	Name    string `json:"name"`
}

type Portfolio struct {
	DocType    string     `json:"docType"`
	ID         string     `json:"id"`
	Fund       string     `json:"fund"`
	Name       string     `json:"name"`
	Securities []Security `json:"securities"`
}

type Security struct {
	Name     string `json:"name"`
	CUSIP    string `json:"cusip"`
	Amount   string `json:"amount"`
	Currency string `json:"currency"`
}

type PortfolioAction struct {
	DocType     string   `json:"docType"`
	ID          string   `json:"id"`
	Fund        string   `json:"fund"`
	Portfolio   string   `json:"portfolio"`
	Security    Security `json:"security"`
	Type        string   `json:"type"`
	Date        string   `json:"date"`
	Period      int      `json:"period"`
	Status      string   `json:"status"`
	Description string   `json:"description"`
}

type AdminContractContract struct {
	contractapi.Contract
}

func (s *AdminContractContract) CreateFund(ctx contractapi.TransactionContextInterface, fundId string, name string, inceptionDate string) error {
	obj, err := ctx.GetStub().GetState(fundId)
	if err != nil {
		return fmt.Errorf(("error retrieving the world state"))
	}

	if obj != nil {
		return fmt.Errorf("an object already exists with that id")
	}

	fund := Fund{
		DocType:            types.DOCTYPE_FUND,
		ID:                 fundId,
		Name:               name,
		CurrentPeriod:      0,
		InceptionDate:      inceptionDate,
		PeriodClosingValue: "0",
		PeriodOpeningValue: "0",
		AggregateFixedFees: "0",
		AggregateDeposits:  "0",
		NextInvestorNumber: 0,
		PeriodUpdated:      false,
	}

	fundJson, err := json.Marshal(fund)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(fundId, fundJson)
}

func (s *AdminContractContract) CreateInvestor(ctx contractapi.TransactionContextInterface, name string) error {
	investors, err := s.QueryInvestorByName(ctx, name)

	if err != nil {
		return err
	}
	if investors == nil {
		return fmt.Errorf("an investor with the name '%s' already exists", name)
	}

	id := uuid.NewV4().String()
	investor := Investor{
		DocType: types.DOCTYPE_INVESTOR,
		Name:    name,
		ID:      id,
	}

	investorJson, err := json.Marshal(investor)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, investorJson)
}

func (s *AdminContractContract) CreateCapitalAccount(ctx contractapi.TransactionContextInterface, fundId string, investorId string) error {
	fund, err := s.QueryFundById(ctx, fundId)

	if err != nil {
		return err
	}
	if fund == nil {
		return fmt.Errorf("a fund with the ID '%s' does not exist", fundId)
	}

	investor, err := s.QueryInvestorById(ctx, investorId)
	if err != nil {
		return err
	}
	if investor == nil {
		return fmt.Errorf("an investor with id '%s' does not exist", investorId)
	}

	id := uuid.NewV4().String()
	capitalAccount := CapitalAccount{
		DocType:             types.DOCTYPE_CAPITALACCOUNT,
		ID:                  id,
		Fund:                fundId,
		Investor:            investorId,
		Number:              fund.NextInvestorNumber,
		CurrentPeriod:       fund.CurrentPeriod,
		PeriodClosingValue:  "0.0",
		PeriodOpeningValue:  "0.0",
		FixedFees:           "0.0",
		Deposits:            "0.0",
		OwnershipPercentage: "0.0",
		HighWaterMark:       HighWaterMark{Amount: "0.0", Date: "None"},
	}

	capitalAccountJson, err := json.Marshal(capitalAccount)
	if err != nil {
		return err
	}

	fund.NextInvestorNumber += 1
	fundJson, err := json.Marshal(fund)
	if err != nil {
		return err
	}

	ctx.GetStub().PutState(fund.ID, fundJson)
	return ctx.GetStub().PutState(id, capitalAccountJson)
}

func (s *AdminContractContract) CreatePortfolio(ctx contractapi.TransactionContextInterface, fundId, name string) error {
	portfolios, err := s.QueryPortfolioByName(ctx, fundId, name)

	if err != nil {
		return err
	}
	if portfolios == nil {
		return fmt.Errorf("an portfolio with the name '%s' already exists for fund '%s'", name, fundId)
	}

	id := uuid.NewV4().String()
	Portfolio := Portfolio{
		DocType: types.DOCTYPE_PORTFOLIO,
		Name:    name,
		ID:      id,
		Fund:    fundId,
	}

	portfolioJson, err := json.Marshal(Portfolio)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, portfolioJson)
}

func (s *AdminContractContract) CreatePortfolioAction(ctx contractapi.TransactionContextInterface, fundId string, portfolioId string, type_ string, date string, period int, name string, cusip string, amount string, currency string) error {
	if type_ != "buy" && type_ != "sell" {
		return fmt.Errorf("the specified action is invalid for a portfolio: '%s'", type_)
	}

	security := Security{
		Name:     name,
		CUSIP:    cusip,
		Amount:   amount,
		Currency: currency,
	}

	id := uuid.NewV4().String()
	portfolioAction := PortfolioAction{
		DocType:   types.DOCTYPE_PORTFOLIOACTION,
		Portfolio: portfolioId,
		Fund:      fundId,
		Type:      type_,
		Date:      date,
		ID:        id,
		Security:  security,
		Period:    period,
	}

	portfolioActionJson, err := json.Marshal(portfolioAction)

	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, portfolioActionJson)
}

func (s *AdminContractContract) CreateCapitalAccountAction(ctx contractapi.TransactionContextInterface, type_ string, amount string, full bool, date string, period int) error {
	if type_ != "deposit" && type_ != "withdrawal" {
		return fmt.Errorf("the specified type of '%s' is invalid for a CapitalAccountAction", type_)
	}

	id := uuid.NewV4().String()
	capitalAccountAction := CapitalAccountAction{
		DocType:     types.DOCTYPE_CAPITALACCOUNTACTION,
		ID:          id,
		Type:        type_,
		Amount:      amount,
		Full:        full,
		Status:      types.TX_STATUS_SUBMITTED,
		Description: "",
		Date:        date,
		Period:      period,
	}

	capitalAccountActionJson, err := json.Marshal(capitalAccountAction)

	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, capitalAccountActionJson)
}

func (s *AdminContractContract) QueryPortfolioByName(ctx contractapi.TransactionContextInterface, fundId string, name string) (*Portfolio, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType": "portfolio", "fund": "%s", "name": "%s"}}`, fundId, name)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var portfolios Portfolio
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(queryResult.Value, &portfolios)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &portfolios, nil
}

func (s *AdminContractContract) QueryPortfolioByFund(ctx contractapi.TransactionContextInterface, fundId string) (*Portfolio, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType": "portfolio", "fund": "%s"}}`, fundId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var portfolios Portfolio
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(queryResult.Value, &portfolios)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &portfolios, nil
}

func (s *AdminContractContract) QueryFundByName(ctx contractapi.TransactionContextInterface, name string) (*Fund, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"fund", "name": "%s"}}`, name)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var fund Fund
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var fund Fund
		err = json.Unmarshal(queryResult.Value, &fund)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &fund, nil
}

func (s *AdminContractContract) QueryFundById(ctx contractapi.TransactionContextInterface, fundId string) (*Fund, error) {
	data, err := ctx.GetStub().GetState(fundId)

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var fund Fund
	err = json.Unmarshal(data, &fund)
	if err != nil {
		return nil, err
	}

	return &fund, nil
}

func (s *AdminContractContract) QueryInvestorById(ctx contractapi.TransactionContextInterface, investorId string) (*Investor, error) {
	data, err := ctx.GetStub().GetState(investorId)

	if err != nil {
		return nil, err
	}

	if data == nil {
		return nil, nil
	}

	var investor Investor
	err = json.Unmarshal(data, &investor)
	if err != nil {
		return nil, err
	}

	return &investor, nil
}

func (s *AdminContractContract) QueryInvestorByName(ctx contractapi.TransactionContextInterface, name string) (*Investor, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"investor", "name": "%s"}}`, name)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var investor Investor
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(queryResult.Value, &investor)
		if err != nil {
			return nil, err
		}

		if true {
			break
		}
	}

	return &investor, nil
}

func (s *AdminContractContract) QueryCapitalAccountsByInvestor(ctx contractapi.TransactionContextInterface, fundId string, investorId string) ([]*CapitalAccount, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s", "investor": "%s"}}`, fundId, investorId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccounts []*CapitalAccount
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccount CapitalAccount
		err = json.Unmarshal(queryResult.Value, &capitalAccount)
		if err != nil {
			return nil, err
		}
		capitalAccounts = append(capitalAccounts, &capitalAccount)
	}

	return capitalAccounts, nil
}

func (s *AdminContractContract) QueryCapitalAccountsByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*CapitalAccount, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s"}}`, fundId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccounts []*CapitalAccount
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccount CapitalAccount
		err = json.Unmarshal(queryResult.Value, &capitalAccount)
		if err != nil {
			return nil, err
		}
		capitalAccounts = append(capitalAccounts, &capitalAccount)
	}

	return capitalAccounts, nil
}

func (s *AdminContractContract) QueryCapitalAccountActionsByFund(ctx contractapi.TransactionContextInterface, fundId string) ([]*CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s"}}`, fundId)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccountActions []*CapitalAccountAction
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccountAction CapitalAccountAction
		err = json.Unmarshal(queryResult.Value, &capitalAccountAction)
		if err != nil {
			return nil, err
		}
		capitalAccountActions = append(capitalAccountActions, &capitalAccountAction)
	}

	return capitalAccountActions, nil
}

func (s *AdminContractContract) QueryCapitalAccountActionsByAccountPeriod(ctx contractapi.TransactionContextInterface, fundId string, capitalAccountId string, period int) ([]*CapitalAccountAction, error) {
	queryString := fmt.Sprintf(`{"selector":{"docType":"capitalAccount", "fund": "%s", "capitalAccount": "%s", "period": "%d"}}`, fundId, capitalAccountId, period)

	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}

	defer resultsIterator.Close()

	var capitalAccountActions []*CapitalAccountAction
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var capitalAccountAction CapitalAccountAction
		err = json.Unmarshal(queryResult.Value, &capitalAccountAction)
		if err != nil {
			return nil, err
		}
		capitalAccountActions = append(capitalAccountActions, &capitalAccountAction)
	}

	return capitalAccountActions, nil
}
