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
	"github.com/Akachain/gringotts/dto/token"
	"github.com/Akachain/gringotts/handler"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/smartcontract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type baseToken struct {
	contractapi.Contract
	tokenHandler       *handler.TokenHandler
	walletHandler      *handler.WalletHandler
	healthCheckHandler handler.HealthCheckHandler
	accountingHandler  handler.AccountingHandler
}

func NewBaseToken() smartcontract.BasicToken {
	return &baseToken{
		tokenHandler:       handler.NewTokenHandler(),
		walletHandler:      handler.NewWalletHandler(),
		healthCheckHandler: handler.NewHealthCheckHandler(),
		accountingHandler:  handler.NewAccountingHandler(),
	}
}

func (b *baseToken) InitLedger(ctx contractapi.TransactionContextInterface) error {
	glogger.GetInstance().Info(ctx, "------------Init ChainCode------------")
	return nil
}

// Wallet feature
func (b *baseToken) CreateWallet(ctx contractapi.TransactionContextInterface, createWallet token.CreateWallet) (string, error) {
	return b.walletHandler.CreateWallet(ctx, createWallet)
}

func (b *baseToken) UpdateWallet(ctx contractapi.TransactionContextInterface, updateWallet token.UpdateWallet) error {
	return b.walletHandler.UpdateWallet(ctx, updateWallet)
}

func (b *baseToken) GetBalance(ctx contractapi.TransactionContextInterface, balance token.Balance) (string, error) {
	return b.walletHandler.BalanceOf(ctx, balance)
}

// Token feature
func (b *baseToken) Mint(ctx contractapi.TransactionContextInterface, mintDto token.MintToken) error {
	return b.tokenHandler.Mint(ctx, mintDto)
}

func (b *baseToken) Burn(ctx contractapi.TransactionContextInterface, burnDto token.BurnToken) error {
	return b.tokenHandler.Burn(ctx, burnDto)
}

func (b *baseToken) Transfer(ctx contractapi.TransactionContextInterface, transferDto token.TransferToken) error {
	return b.tokenHandler.Transfer(ctx, transferDto)
}

func (b *baseToken) CreateTokenType(ctx contractapi.TransactionContextInterface, createTokenTypeDto token.CreateTokenType) (string, error) {
	return b.tokenHandler.CreateTokenType(ctx, createTokenTypeDto)
}

func (b *baseToken) TransferSideChain(ctx contractapi.TransactionContextInterface, transferChain token.TransferSideChain) error {
	return b.tokenHandler.TransferSideChain(ctx, transferChain)
}

// API healthcheck
func (b *baseToken) CreateHealthCheck(ctx contractapi.TransactionContextInterface, arg string) (string, error) {
	return b.healthCheckHandler.CreateHealthCheck(ctx)
}

// Accounting feature
func (b *baseToken) GetAccountingTx(ctx contractapi.TransactionContextInterface) ([]string, error) {
	return b.accountingHandler.GetAccountingTx(ctx)
}

func (b *baseToken) CalculateBalance(ctx contractapi.TransactionContextInterface, accountingDto token.AccountingBalance) error {
	return b.accountingHandler.CalculateBalance(ctx, accountingDto)
}

func (b *baseToken) Exchange(ctx contractapi.TransactionContextInterface, exchangeToken token.ExchangeToken) error {
	return b.tokenHandler.Exchange(ctx, exchangeToken)
}

func (b *baseToken) Issue(ctx contractapi.TransactionContextInterface, issueDto token.IssueToken) error {
	return b.tokenHandler.Issue(ctx, issueDto)
}

func (b *baseToken) EnrollToken(ctx contractapi.TransactionContextInterface, enrollmentDto token.Enrollment) error {
	return b.walletHandler.EnrollToken(ctx, enrollmentDto)
}
