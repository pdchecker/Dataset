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
	"github.com/Akachain/gringotts/entity"
	"github.com/Akachain/gringotts/glossary"
	"github.com/Akachain/gringotts/glossary/doc"
	"github.com/Akachain/gringotts/glossary/transaction"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/services/base"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

type TxBase struct {
	*base.Base
}

func NewTxBase() *TxBase {
	return &TxBase{
		base.NewBase(),
	}
}

func (b TxBase) TxHandlerTransfer(ctx contractapi.TransactionContextInterface, mapBalanceToken map[string]*entity.BalanceCache, tx *entity.Transaction) (*entity.Transaction, error) {
	if tx.FromWallet == glossary.SystemWallet || tx.ToWallet == glossary.SystemWallet {
		glogger.GetInstance().Errorf(ctx, "TxHandler - Base - Transaction (%s) has from/to wallet Id is system type", tx.Id)
		tx.Status = transaction.Rejected
		return tx, errors.New("From/To wallet id invalidate")
	}

	if err := b.SubAmount(ctx, mapBalanceToken, doc.SpotBalances, tx.FromWallet, tx.FromTokenId, tx.FromTokenAmount); err != nil {
		glogger.GetInstance().Errorf(ctx, "TxHandler - Base - Transaction (%s): Unable to sub temp amount of From wallet", tx.Id)
		tx.Status = transaction.Rejected
		return tx, errors.WithMessage(err, "Sub balance of from wallet failed")
	}

	if err := b.AddAmount(ctx, mapBalanceToken, doc.SpotBalances, tx.ToWallet, tx.FromTokenId, tx.ToTokenAmount); err != nil {
		glogger.GetInstance().Errorf(ctx, "TxHandler - Base - Transaction (%s): Unable to add temp amount of To wallet", tx.Id)
		tx.Status = transaction.Rejected
		if err := b.RollbackTxHandler(ctx, tx, mapBalanceToken, transaction.SubFromWallet); err != nil {
			glogger.GetInstance().Errorf(ctx, "TxHandler - Base - Rollback handle transaction (%s) failed with error (%v)", tx.Id, err)
		}
		return tx, errors.WithMessage(err, "Add balance of to wallet failed")
	}

	tx.Status = transaction.Confirmed
	return tx, nil
}
