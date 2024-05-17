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

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

type NFT interface {
	// Mint generate new NFT token
	Mint(ctx contractapi.TransactionContextInterface, gs1Number string, ownerWalletId string, metaData string, hashData string) (string, error)

	// OwnerOf return owner wallet id of nft token
	OwnerOf(ctx contractapi.TransactionContextInterface, nftTokenId string) (string, error)

	// BalanceOf return number nft of wallet address
	BalanceOf(ctx contractapi.TransactionContextInterface, ownerWalletId string) (int, error)

	// TransferFrom to transfer nft token from owner to other wallet
	TransferFrom(ctx contractapi.TransactionContextInterface, ownerWalletId string, toWalletId string, fromTokenId string, nftTokenId string, price float64) error
}
