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
	"github.com/Akachain/gringotts/helper"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/pkg/tx/base"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
)

type txBurn struct {
	*base.TxBase
}

func NewTxBurn() *txBurn {
	return &txBurn{base.NewTxBase()}
}

func (t *txBurn) AccountingTx(ctx contractapi.TransactionContextInterface, tx *entity.Transaction, mapBalanceToken map[string]*entity.BalanceCache) (*entity.Transaction, error) {
	if tx.ToWallet != glossary.SystemWallet {
		glogger.GetInstance().Errorf(ctx, "TxBurn - Transaction (%s): has To wallet Id is not system type", tx.Id)
		tx.Status = transaction.Rejected
		return tx, errors.New("To wallet id invalidate")
	}

	// decrease total supply of token on the blockchain
	tokenType, err := t.GetTokenType(ctx, tx.FromTokenId)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "TxBurn - Transaction (%s): Unable to get token type", tx.Id)
		tx.Status = transaction.Rejected
		return nil, err
	}
	totalSupplyUpdated, err := helper.SubBalance(tokenType.TotalSupply, tx.FromTokenAmount)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "TxBurn - Transaction (%s): Unable to decrease total supply of token", tx.Id)
		tx.Status = transaction.Rejected
		return nil, err
	}
	tokenType.TotalSupply = totalSupplyUpdated
	if err := t.Repo.Update(ctx, tokenType, doc.Tokens, helper.TokenKey(tokenType.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "TxBurn - Transaction (%s): Unable to update token type (%s)", tx.Id, err.Error())
		tx.Status = transaction.Rejected
		return tx, errors.New("Unable to decrease total of token on the blockchain")
	}

	if err := t.SubAmount(ctx, mapBalanceToken, doc.SpotBalances, tx.FromWallet, tx.FromTokenId, tx.FromTokenAmount); err != nil {
		glogger.GetInstance().Errorf(ctx, "TxBurn - Transaction (%s): add balance failed (%s)", tx.Id, err.Error())
		tx.Status = transaction.Rejected
		return tx, errors.New("Sub balance of to wallet failed")
	}

	tx.Status = transaction.Confirmed
	return tx, nil
}
