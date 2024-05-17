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
	"github.com/Akachain/gringotts/errorcode"
	"github.com/Akachain/gringotts/glossary"
	"github.com/Akachain/gringotts/glossary/doc"
	"github.com/Akachain/gringotts/glossary/sidechain"
	"github.com/Akachain/gringotts/glossary/transaction"
	"github.com/Akachain/gringotts/helper"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/services/base"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strings"
)

type tokenService struct {
	*base.Base
}

func NewTokenService() *tokenService {
	return &tokenService{
		base.NewBase(),
	}
}

func (t *tokenService) TransferWithNote(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId, tokenId,
	amount, note string) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Token Service - TransferWithNote-----------")
	return t.transferToken(ctx, fromWalletId, toWalletId, tokenId, amount, note)
}

func (t *tokenService) Transfer(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId, tokenId, amount string) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Token Service - Transfer-----------")
	return t.transferToken(ctx, fromWalletId, toWalletId, tokenId, amount, "")
}

func (t *tokenService) TransferSideChain(ctx contractapi.TransactionContextInterface, walletId, tokenId string,
	fromChain, toChain sidechain.SideName, amount string) error {
	glogger.GetInstance().Info(ctx, "-----------Token Service - TransferSideChain-----------")

	if err := t.validateTransfer(ctx, walletId, walletId); err != nil {
		glogger.GetInstance().Errorf(ctx, "TransferSideChain - Validation transfer failed with error (%v)", err)
		return err
	}

	note := fromChain + "_" + toChain
	// create new transfer transaction
	txEntity := entity.NewTransaction(ctx)
	txEntity.SpenderWallet = walletId
	txEntity.FromWallet = walletId
	txEntity.ToWallet = walletId
	txEntity.FromTokenId = tokenId
	txEntity.ToTokenId = tokenId
	txEntity.FromTokenAmount = amount
	txEntity.ToTokenAmount = amount
	txEntity.TxType = transaction.SideChainTransfer
	txEntity.Note = string(note)

	if err := t.Repo.Create(ctx, txEntity, doc.Transactions, helper.TransactionKey(txEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "TransferSideChain - Create transfer transaction failed with error (%v)", err)
		return helper.RespError(errorcode.BizUnableCreateTX)
	}
	glogger.GetInstance().Infof(ctx, "-----------Token Service - TransferSideChain succeed (%s)-----------", txEntity.Id)

	return nil
}

func (t *tokenService) Mint(ctx contractapi.TransactionContextInterface, walletId, tokenId, amount string) error {
	glogger.GetInstance().Info(ctx, "-----------Token Service - Mint-----------")

	// validate wallet exited
	if _, err := t.GetActiveWallet(ctx, walletId); err != nil {
		glogger.GetInstance().Errorf(ctx, "Mint - Get wallet mint failed with error (%s)", err.Error())
		return err
	}

	// check total supply with max supply

	// create tx mint token
	txMint := entity.NewTransaction(ctx)
	txMint.SpenderWallet = walletId
	txMint.FromWallet = glossary.SystemWallet
	txMint.ToWallet = walletId
	txMint.FromTokenId = tokenId
	txMint.ToTokenId = tokenId
	txMint.FromTokenAmount = amount
	txMint.ToTokenAmount = amount
	txMint.TxType = transaction.Mint

	if err := t.Repo.Create(ctx, txMint, doc.Transactions, helper.TransactionKey(txMint.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Mint - Create mint transaction failed with error (%s)", err.Error())
		return helper.RespError(errorcode.BizUnableCreateTX)
	}
	glogger.GetInstance().Infof(ctx, "-----------Token Service - Mint succeed (%s)-----------", txMint.Id)

	return nil
}

func (t *tokenService) Burn(ctx contractapi.TransactionContextInterface, walletId, tokenId, amount string) error {
	glogger.GetInstance().Info(ctx, "-----------Token Service - Burn-----------")

	// validate burn wallet exist
	_, err := t.GetActiveWallet(ctx, walletId)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "Burn - Get wallet mint failed with error (%v)", err)
		return err
	}

	// get balance of token
	//balanceToken, err := t.GetBalanceOfToken(ctx, doc.SpotBalances, wallet.Id, tokenId)
	//if err != nil {
	//	glogger.GetInstance().Errorf(ctx, "Burn - Get balance of token failed with error (%v)", err)
	//	return err
	//}
	//
	//// check balance enough to burn
	//if helper.CompareStringBalance(balanceToken.Balances, amount) < 0 {
	//	glogger.GetInstance().Error(ctx, "Burn - Wallet balance is insufficient", err)
	//	return helper.RespError(errorcode.BizBalanceNotEnough)
	//}

	// create tx burn token
	txBurn := entity.NewTransaction(ctx)
	txBurn.SpenderWallet = walletId
	txBurn.FromWallet = walletId
	txBurn.ToWallet = glossary.SystemWallet
	txBurn.FromTokenId = tokenId
	txBurn.ToTokenId = tokenId
	txBurn.FromTokenAmount = amount
	txBurn.ToTokenAmount = amount
	txBurn.TxType = transaction.Burn

	if err := t.Repo.Create(ctx, txBurn, doc.Transactions, helper.TransactionKey(txBurn.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Burn - Create burn transaction failed with error (%v)", err)
		return helper.RespError(errorcode.BizUnableCreateTX)
	}
	glogger.GetInstance().Infof(ctx, "-----------Token Service - Burn succeed (%s)-----------", txBurn.Id)

	return nil
}

func (t *tokenService) CreateType(ctx contractapi.TransactionContextInterface, name, tickerToken, maxSupply string) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Token Service - CreateType-----------")

	tokenEntity := entity.NewToken(ctx)
	tokenEntity.Name = name
	tokenEntity.TickerToken = tickerToken
	tokenEntity.MaxSupply = maxSupply

	if err := t.Repo.Create(ctx, tokenEntity, doc.Tokens, helper.TokenKey(tokenEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "CreateType - Create token type failed with error (%s)", err.Error())
		return "", helper.RespError(errorcode.BizUnableCreateToken)
	}
	glogger.GetInstance().Infof(ctx, "-----------Token Service - CreateType succeed (%s)-----------", tokenEntity.Id)

	return tokenEntity.Id, nil
}

func (t *tokenService) Exchange(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId, fromTokenId,
	toTokenId, fromTokenAmount, toTokenAmount string) error {
	glogger.GetInstance().Info(ctx, "-----------Token Service - Exchange-----------")

	// validate from wallet and to wallet have active or not
	if _, _, err := t.ValidatePairWallet(ctx, fromWalletId, toWalletId); err != nil {
		glogger.GetInstance().Errorf(ctx, "Exchange - Validation swap failed with error (%s)", err.Error())
		return err
	}

	// TODO: validate token

	// create new swap transaction
	txEntity := entity.NewTransaction(ctx)
	txEntity.SpenderWallet = fromWalletId
	txEntity.FromWallet = fromWalletId
	txEntity.ToWallet = toWalletId
	txEntity.FromTokenId = fromTokenId
	txEntity.ToTokenId = toTokenId
	txEntity.FromTokenAmount = fromTokenAmount
	txEntity.ToTokenAmount = toTokenAmount
	txEntity.TxType = transaction.Exchange

	if err := t.Repo.Create(ctx, txEntity, doc.Transactions, helper.TransactionKey(txEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Exchange - Create exchange transaction failed with error (%v)", err)
		return helper.RespError(errorcode.BizUnableCreateTX)
	}
	glogger.GetInstance().Infof(ctx, "-----------Token Service - Exchange succeed (%s)-----------", txEntity.Id)

	return nil
}

func (t *tokenService) Issue(ctx contractapi.TransactionContextInterface, walletId string, fromTokenId, toTokenId, fromTokenAmount, toTokenAmount string) error {
	glogger.GetInstance().Info(ctx, "-----------Token Service - Issue-----------")

	// validate wallet active or not
	wallet, err := t.GetWallet(ctx, walletId)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "Issue - Get wallet failed with err (%s)", err.Error())
		return err
	}

	// check enrollment policy
	enrollment, isExisted, err := t.GetAndCheckExistEnrollment(ctx, toTokenId)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "Issue - Check exit enrollment failed with err (%s)", err.Error())
		return helper.RespError(errorcode.BizUnableGetEnrollment)
	}
	if isExisted {
		// check permission to issue new token
		if enrollment.FromWalletId != "" {
			if !strings.Contains(enrollment.FromWalletId, wallet.Id) {
				glogger.GetInstance().Errorf(ctx, "Issue - From wallet do not have permission issue token (%s)", toTokenId)
				return helper.RespError(errorcode.BizIssueNotPermission)
			}
		}
		if enrollment.ToWalletId != "" {
			if !strings.Contains(enrollment.ToWalletId, wallet.Id) {
				glogger.GetInstance().Errorf(ctx, "Issue - To wallet do not have permission issue token (%s)", toTokenId)
				return helper.RespError(errorcode.BizIssueNotPermission)
			}
		}
	}

	// validate total supply of AT token
	atToken, err := t.GetTokenType(ctx, toTokenId)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "Issue - Get AT token failed with err (%s)", err.Error())
		return err
	}

	if atToken.MaxSupply != "" {
		newTotalSupply, err := helper.AddBalance(atToken.TotalSupply, toTokenAmount)
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "Issue - Calculate new total supply failed with err (%s)", err.Error())
			return helper.RespError(errorcode.BizUnableToIssue)
		}

		if helper.CompareStringBalance(atToken.MaxSupply, newTotalSupply) < 0 {
			glogger.GetInstance().Error(ctx, "Issue - Number of new token over the max supply of the token")
			return helper.RespError(errorcode.BizOverMaxSupply)
		}
	}

	// create new swap transaction
	txEntity := entity.NewTransaction(ctx)
	txEntity.SpenderWallet = wallet.Id
	txEntity.FromWallet = wallet.Id
	txEntity.ToWallet = wallet.Id
	txEntity.FromTokenId = fromTokenId
	txEntity.ToTokenId = toTokenId
	txEntity.FromTokenAmount = fromTokenAmount
	txEntity.ToTokenAmount = toTokenAmount
	txEntity.TxType = transaction.Issue

	if err := t.Repo.Create(ctx, txEntity, doc.Transactions, helper.TransactionKey(txEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Issue - Create swap transaction failed with error (%v)", err)
		return helper.RespError(errorcode.BizUnableCreateTX)
	}
	glogger.GetInstance().Infof(ctx, "-----------Token Service - Issue succeed (%s)-----------", txEntity.Id)

	return nil
}

func (t *tokenService) validateTransfer(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId string) error {
	if _, _, err := t.ValidatePairWallet(ctx, fromWalletId, toWalletId); err != nil {
		return err
	}

	//balanceToken, err := t.GetBalanceOfToken(ctx, doc.SpotBalances, walletFrom.Id, tokenId)
	//if err != nil {
	//	return err
	//}

	// check balance enough to transfer
	//if helper.CompareStringBalance(balanceToken.Balances, amount) < 0 {
	//	glogger.GetInstance().Error(ctx, "ValidateTransfer - Balance of from wallet is insufficient")
	//	return helper.RespError(errorcode.BizBalanceNotEnough)
	//}
	return nil
}

func (t *tokenService) transferToken(ctx contractapi.TransactionContextInterface, fromWalletId, toWalletId, tokenId, amount, note string) (string, error) {
	if err := t.validateTransfer(ctx, fromWalletId, toWalletId); err != nil {
		glogger.GetInstance().Errorf(ctx, "Transfer - Validation transfer failed with error (%v)", err)
		return "", err
	}

	// create new transfer transaction
	txEntity := entity.NewTransaction(ctx)
	txEntity.SpenderWallet = fromWalletId
	txEntity.FromWallet = fromWalletId
	txEntity.ToWallet = toWalletId
	txEntity.FromTokenId = tokenId
	txEntity.ToTokenId = tokenId
	txEntity.FromTokenAmount = amount
	txEntity.ToTokenAmount = amount
	txEntity.TxType = transaction.Transfer
	txEntity.Note = note

	if err := t.Repo.Create(ctx, txEntity, doc.Transactions, helper.TransactionKey(txEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Transfer - Create transfer transaction failed with error (%v)", err)
		return "", helper.RespError(errorcode.BizUnableCreateTX)
	}
	glogger.GetInstance().Infof(ctx, "-----------Token Service - Transfer succeed (%s)-----------", txEntity.Id)

	return txEntity.Id, nil
}
