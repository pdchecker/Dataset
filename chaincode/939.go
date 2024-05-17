// Copyright (c) 2021 akachain
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

// Package handler take request from the smart contract API and pass the
// DTO object to the service layer.
package main

import (
	"github.com/Akachain/gringotts/dto/token"
	"github.com/Akachain/gringotts/errorcode"
	"github.com/Akachain/gringotts/helper"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/services"
	"github.com/Akachain/gringotts/services/accounting"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type AccountingHandler struct {
	accountingService services.Accounting
}

func NewAccountingHandler() AccountingHandler {
	return AccountingHandler{
		accounting.NewAccountingService(),
	}
}

// GetAccountingTx return limit 10 transaction have status pending to accounting balance.
func (a *AccountingHandler) GetAccountingTx(ctx contractapi.TransactionContextInterface) ([]string, error) {
	glogger.GetInstance().Info(ctx, "-----------Accounting Handler - GetAccountingTx-----------")

	return a.accountingService.GetTx(ctx)
}

// CalculateBalance calculate balance of list transaction from client.
func (a *AccountingHandler) CalculateBalance(ctx contractapi.TransactionContextInterface, accountingBalance token.AccountingBalance) error {
	glogger.GetInstance().Info(ctx, "-----------Accounting Handler - CalculateBalance-----------")

	// checking dto validate
	if err := accountingBalance.IsValid(); err != nil {
		glogger.GetInstance().Errorf(ctx, "Accounting - CalculateBalance Input invalidate %v", err)
		return helper.RespError(errorcode.InvalidParam)
	}

	return a.accountingService.CalculateBalance(ctx, accountingBalance)
}
