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
	"errors"
	"github.com/Akachain/gringotts/entity"
	"github.com/Akachain/gringotts/glossary"
	"github.com/Akachain/gringotts/glossary/doc"
	"github.com/Akachain/gringotts/helper"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type CreateWallet struct {
	TokenId string          `json:"tokenId"`
	Status  glossary.Status `json:"status"`
}

func (c CreateWallet) ToEntity(ctx contractapi.TransactionContextInterface) *entity.Wallet {
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	walletEntity := new(entity.Wallet)
	walletEntity.Id = helper.GenerateID(doc.Wallets, ctx.GetStub().GetTxID())

	walletEntity.Status = c.Status
	walletEntity.CreatedAt = helper.TimestampISO(txTime.Seconds)
	walletEntity.UpdatedAt = helper.TimestampISO(txTime.Seconds)
	walletEntity.BlockChainId = ctx.GetStub().GetTxID()
	return walletEntity
}

func (c CreateWallet) IsValid() error {
	if c.TokenId == "" {
		return errors.New("token id is empty")
	}
	switch c.Status {
	case glossary.Active, glossary.InActive:
		return nil
	}
	return errors.New("invalid wallet status")
}
