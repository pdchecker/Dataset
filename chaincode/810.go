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
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type BasicToken interface {
	contractapi.ContractInterface

	// CreateWallet to create new wallet. Each wallet belong to token type
	CreateWallet(ctx contractapi.TransactionContextInterface, createWallet token.CreateWallet) (string, error)

	// UpdateWallet to update status of wallet. Active or InActive
	UpdateWallet(ctx contractapi.TransactionContextInterface, updateWallet token.UpdateWallet) error

	// GetBalance return balance of wallet
	GetBalance(ctx contractapi.TransactionContextInterface, balance token.Balance) (string, error)

	// Mint to init base token in the system
	Mint(ctx contractapi.TransactionContextInterface, mintDto token.MintToken) error

	// Burn to delete token in the system
	Burn(ctx contractapi.TransactionContextInterface, burnDto token.BurnToken) error

	// Transfer amount of tokens from address FromWalletId to address ToWalletId
	Transfer(ctx contractapi.TransactionContextInterface, transferDto token.TransferToken) error

	// CreateTokenType to create new token type in the system
	CreateTokenType(ctx contractapi.TransactionContextInterface, createTokenTypeDto token.CreateTokenType) (string, error)

	// CreateHealthCheck check system ready to use
	CreateHealthCheck(ctx contractapi.TransactionContextInterface, arg string) (string, error)

	// GetAccountingTx return list id transaction that have status pending
	GetAccountingTx(ctx contractapi.TransactionContextInterface) ([]string, error)

	// CalculateBalance update balance of wallet. Accounting job will call this
	CalculateBalance(ctx contractapi.TransactionContextInterface, accountingDto token.AccountingBalance) error

	// Exchange to swap between token type. Example from Stable token to X token
	Exchange(ctx contractapi.TransactionContextInterface, exchangeToken token.ExchangeToken) error

	// Issue to issue new token from stable token
	Issue(ctx contractapi.TransactionContextInterface, issueDto token.IssueToken) error

	// EnrollToken to register wallet policy use to issue new token
	EnrollToken(ctx contractapi.TransactionContextInterface, enrollmentDto token.Enrollment) error

	// TransferSideChain to transfer token from main chain to side chain
	TransferSideChain(ctx contractapi.TransactionContextInterface, transferChain token.TransferSideChain) error
}
