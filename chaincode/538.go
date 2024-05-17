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
	"github.com/Akachain/gringotts/glossary/sidechain"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Token interface {
	// Transfer to transfer token between wallet.
	// But state balance of wallet not update at the time.
	// It will be update when accounting job start
	Transfer(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId, tokenId, amount string) (string, error)

	// TransferWithNote same with Transfer function but add note in the transaction
	TransferWithNote(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId, tokenId, amount, note string) (string, error)

	// TransferSideChain to transfer token from main chain to side chain of wallet
	TransferSideChain(ctx contractapi.TransactionContextInterface, walletId, tokenId string, fromChain, toChain sidechain.SideName, amount string) error

	// Mint to init token in the system
	Mint(ctx contractapi.TransactionContextInterface, walletId, tokenId, amount string) error

	// Burn to delete token in the system
	Burn(ctx contractapi.TransactionContextInterface, walletId, tokenId, amount string) error

	// CreateType to create new token type in the system.
	CreateType(ctx contractapi.TransactionContextInterface, name, tickerToken, maxSupply string) (string, error)

	// Exchange to swap between token type.
	Exchange(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId, fromTokenId, toTokenId, fromTokenAmount, toTokenAmount string) error

	// Issue to issue new token type from stable token.
	Issue(ctx contractapi.TransactionContextInterface, walletId, fromTokenId, toTokenId, fromTokenAmount, toTokenAmount string) error
}
