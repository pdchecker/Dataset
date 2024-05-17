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

package main

import (
	"encoding/json"
	"github.com/Akachain/gringotts/dto/token"
	"github.com/Akachain/gringotts/entity"
	"github.com/Akachain/gringotts/errorcode"
	"github.com/Akachain/gringotts/glossary/doc"
	"github.com/Akachain/gringotts/glossary/transaction"
	"github.com/Akachain/gringotts/helper"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/pkg/query"
	txHandler "github.com/Akachain/gringotts/pkg/tx"
	"github.com/Akachain/gringotts/services/base"
	"github.com/davecgh/go-spew/spew"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strings"
)

type accountingService struct {
	*base.Base
}

func NewAccountingService() *accountingService {
	return &accountingService{base.NewBase()}
}

func (a *accountingService) GetTx(ctx contractapi.TransactionContextInterface) ([]string, error) {
	glogger.GetInstance().Debugf(ctx, "GetTx - Get Query String %s", query.GetPendingTransactionQueryString())
	var txIdList []string
	resultsIterator, err := a.Repo.GetQueryStringWithPagination(ctx, query.GetPendingTransactionQueryString())
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "GetTx - Get query with paging failed with error (%v)", err)
		return nil, helper.RespError(errorcode.BizUnableGetTx)
	}
	defer resultsIterator.Close()

	// Check data response after query in database
	if !resultsIterator.HasNext() {
		// Return with list transaction id empty
		return txIdList, nil
	}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "GetTx - Start query failed with error (%v)", err)
			return nil, helper.RespError(errorcode.BizUnableGetTx)
		}
		tx := entity.NewTransaction()
		err = json.Unmarshal(queryResponse.Value, tx)
		if err != nil {
			glogger.GetInstance().Error(ctx, "GetTx - Unable to unmarshal transaction")
			continue
		}
		txIdList = append(txIdList, tx.Id)
	}
	glogger.GetInstance().Infof(ctx, "GetTx - List transaction have status pending: (%s)", strings.Join(txIdList, ","))

	return txIdList, nil
}

func (a *accountingService) CalculateBalance(ctx contractapi.TransactionContextInterface, accountingDto token.AccountingBalance) error {
	glogger.GetInstance().Infof(ctx, "CalculateBalance - List transaction: (%s)", strings.Join(accountingDto.TxId, ","))
	// map temp balance
	mapCurrentBalance := make(map[string]*entity.BalanceCache, len(accountingDto.TxId)*2)
	lstTx := make([]*entity.Transaction, 0, len(accountingDto.TxId))

	for _, id := range accountingDto.TxId {
		tx, err := a.GetTransaction(ctx, id)
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "CalculateBalance - Get transaction (%s) failed with error (%v)", id, err)
			continue
		}
		glogger.GetInstance().Debugf(ctx, "CalculateBalance - Transaction get from db: %s", spew.Sdump(tx))

		// check status of transaction
		if tx.Status != transaction.Pending {
			glogger.GetInstance().Errorf(ctx, "CalculateBalance - Transaction (%s) has status (%s)", id, tx.Status)
			continue
		}

		handler := txHandler.GetTxHandler(tx.TxType)
		if handler == nil {
			glogger.GetInstance().Errorf(ctx, "CalculateBalance -  Unable to get tx handler with transaction type (%s)", tx.TxType)
			continue
		}
		txUpdate, err := handler.AccountingTx(ctx, tx, mapCurrentBalance)
		if err != nil {
			txUpdate.Reason = err.Error()
			glogger.GetInstance().Errorf(ctx, "CalculateBalance - Handle transaction (%s) failed with error (%s)", id, err.Error())
		}
		lstTx = append(lstTx, txUpdate)

	}

	// Update transaction status
	if err := a.updateTransaction(ctx, lstTx); err != nil {
		return err
	}

	// Update balance of wallet after calculate total balance update of wallet
	if err := a.UpdateBalance(ctx, mapCurrentBalance); err != nil {
		return err
	}

	return nil
}

// updateTransaction update status of transaction after handle transaction
func (a *accountingService) updateTransaction(ctx contractapi.TransactionContextInterface, lstTx []*entity.Transaction) error {
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	for _, tx := range lstTx {
		tx.UpdatedAt = helper.TimestampISO(txTime.Seconds)
		if err := a.Repo.Update(ctx, tx, doc.Transactions, helper.TransactionKey(tx.Id)); err != nil {
			glogger.GetInstance().Errorf(ctx, "UpdateTransaction - Update transaction (%s) failed with err (%v)", tx.Id, err)
			return helper.RespError(errorcode.BizUnableUpdateTX)
		}
		glogger.GetInstance().Infof(ctx, "UpdateTransaction - Update transaction (%s) succeed", tx.Id)
	}

	return nil
}
