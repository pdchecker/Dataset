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
	nft2 "github.com/Akachain/gringotts/dto/nft"
	"github.com/Akachain/gringotts/handler"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/smartcontract"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type nft struct {
	nftHandler handler.NftHandler
}

func NewNFT() smartcontract.Erc721 {
	return &nft{
		nftHandler: handler.NewNftHandler(),
	}
}

func (n *nft) MintNft(ctx contractapi.TransactionContextInterface, mintNFT nft2.MintNFT) (string, error) {
	glogger.GetInstance().Info(ctx, "------------Mint NFT SmartContract------------")
	return n.nftHandler.Mint(ctx, mintNFT)
}

func (n *nft) OwnerOf(ctx contractapi.TransactionContextInterface, ownerNFT nft2.OwnerNFT) (string, error) {
	glogger.GetInstance().Info(ctx, "------------OwnerOf NFT SmartContract------------")
	return n.nftHandler.OwnerOf(ctx, ownerNFT)
}

func (n *nft) BalanceOf(ctx contractapi.TransactionContextInterface, balanceOfNFT nft2.BalanceOfNFT) (int, error) {
	glogger.GetInstance().Info(ctx, "------------BalanceOf NFT SmartContract------------")
	return n.nftHandler.BalanceOf(ctx, balanceOfNFT)
}

func (n *nft) TransferFrom(ctx contractapi.TransactionContextInterface, transferNFT nft2.TransferNFT) error {
	glogger.GetInstance().Info(ctx, "------------TransferFrom NFT SmartContract------------")
	return n.nftHandler.TransferNFT(ctx, transferNFT)
}
