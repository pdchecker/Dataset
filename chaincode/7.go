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
	"github.com/Akachain/gringotts/dto/iao"
	statusIao "github.com/Akachain/gringotts/glossary/iao"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Iao interface {
	// CreateAsset to create new asset and token type
	CreateAsset(ctx contractapi.TransactionContextInterface, code, name, ownerWallet, tokenName, tickerToken, maxSupply, totalValue, documentUrl string) (string, error)

	// CreateIao to create new iao of asset
	CreateIao(ctx contractapi.TransactionContextInterface, assetId, assetTokenAmount, startDate, endDate string, rate int64) (string, error)

	// UpdateStatusIao to update status of IAO
	UpdateStatusIao(ctx contractapi.TransactionContextInterface, iaoId string, status statusIao.Status) error

	// BuyBatchAsset to handle multiple request buy asset
	BuyBatchAsset(ctx contractapi.TransactionContextInterface, req []iao.BuyAsset) (string, error)

	// FinalizeIao to finish IAO and distribute AT to investor
	FinalizeIao(ctx contractapi.TransactionContextInterface, iaoId []string) error

	// CancelIao to cancel iao and return ST to investor
	CancelIao(ctx contractapi.TransactionContextInterface, lstInvestorBook []string) error
}
