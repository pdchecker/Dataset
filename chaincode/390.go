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
	"encoding/json"
	"fmt"
	"github.com/Akachain/gringotts/dto/iao"
	"github.com/Akachain/gringotts/entity"
	"github.com/Akachain/gringotts/errorcode"
	"github.com/Akachain/gringotts/glossary/doc"
	statusIao "github.com/Akachain/gringotts/glossary/iao"
	"github.com/Akachain/gringotts/glossary/investor_book"
	"github.com/Akachain/gringotts/glossary/transaction"
	"github.com/Akachain/gringotts/helper"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/services"
	"github.com/Akachain/gringotts/services/base"
	"github.com/Akachain/gringotts/services/token"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/pkg/errors"
	"strings"
)

type iaoService struct {
	*base.Base
	tokenService services.Token
}

func NewIaoService() services.Iao {
	return &iaoService{
		base.NewBase(),
		token.NewTokenService(),
	}
}

func (i *iaoService) CreateAsset(ctx contractapi.TransactionContextInterface, code, name, ownerWallet, tokenName, tickerToken, maxSupply, totalValue, documentUrl string) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Iao Service - CreateAsset-----------")

	tokenId, err := i.tokenService.CreateType(ctx, tokenName, tickerToken, maxSupply)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "Iao Service - Create Token type of asset failed with err (%s)", err.Error())
		return "", err
	}

	assetEntity := entity.NewAsset(ctx)
	assetEntity.Name = name
	assetEntity.Code = code
	assetEntity.OwnerWallet = ownerWallet
	assetEntity.TokenId = tokenId
	assetEntity.Documents = documentUrl
	assetEntity.TokenAmount = maxSupply
	assetEntity.TotalValue = totalValue
	assetEntity.RemainingToken = maxSupply

	if err := i.Repo.Create(ctx, assetEntity, doc.Asset, helper.AssetKey(assetEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Iao Service - Create asset failed with error (%s)", err.Error())
		return "", helper.RespError(errorcode.BizUnableCreateAsset)
	}

	result := fmt.Sprintf("{\"assetId\":\"%s\",\"tokenId\":\"%s\"}", assetEntity.Id, tokenId)
	glogger.GetInstance().Infof(ctx, "-----------Iao Service - CreateAsset succeed (%s)-----------", result)

	return result, nil
}

func (i *iaoService) CreateIao(ctx contractapi.TransactionContextInterface, assetId, assetTokenAmount, startDate, endDate string, rate int64) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Iao Service - CreateIao-----------")

	assetEntity, err := i.GetAsset(ctx, assetId)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "Iao Service - Get asset failed with err (%s)", err.Error())
		return "", err
	}

	if helper.CompareStringBalance(assetEntity.RemainingToken, assetTokenAmount) < 0 {
		glogger.GetInstance().Error(ctx, "Iao Service - Asset do not have enough token for creat Iao campaign")
		return "", helper.RespError(errorcode.BizUnableCreateIao)
	}

	iaoEntity := entity.NewIao(ctx)
	iaoEntity.AssetId = assetId
	iaoEntity.AssetTokenId = assetEntity.TokenId
	iaoEntity.AssetTokenAmount = "0"
	iaoEntity.StableTokenAmount = "0"
	iaoEntity.StartDate = startDate
	iaoEntity.EndDate = endDate
	iaoEntity.Rate = rate

	if err := i.Repo.Create(ctx, iaoEntity, doc.Iao, helper.IaoKey(iaoEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Iao Service - Create iao campaign of asset failed with error (%s)", err.Error())
		return "", helper.RespError(errorcode.BizUnableCreateIao)
	}

	// create tx transfer AT
	txEntity := entity.NewTransaction(ctx)
	txEntity.SpenderWallet = iaoEntity.Id
	txEntity.FromWallet = assetEntity.OwnerWallet
	txEntity.ToWallet = iaoEntity.Id
	txEntity.FromTokenId = assetEntity.TokenId
	txEntity.ToTokenId = assetEntity.TokenId
	txEntity.FromTokenAmount = assetTokenAmount
	txEntity.ToTokenAmount = assetTokenAmount
	txEntity.TxType = transaction.IaoDepositAT
	txEntity.Note = iaoEntity.Id

	if err := i.Repo.Create(ctx, txEntity, doc.Transactions, helper.TransactionKey(txEntity.Id)); err != nil {
		glogger.GetInstance().Errorf(ctx, "Iao Service - Create transaction failed with error (%s)", err.Error())
		return "", helper.RespError(errorcode.BizUnableCreateTX)
	}

	glogger.GetInstance().Infof(ctx, "-----------Iao Service - CreateIao succeed (%s)-----------", iaoEntity.Id)

	return iaoEntity.Id, nil
}

func (i *iaoService) BuyBatchAsset(ctx contractapi.TransactionContextInterface, batchReq []iao.BuyAsset) (string, error) {
	glogger.GetInstance().Info(ctx, "Start BuyBatchAsset")
	// calculate hash and check cache
	stringInput, _ := json.Marshal(batchReq)
	inputHash := helper.CalculateHash(string(stringInput))
	cacheEntity, isExisted, err := i.GetBuyIaoCache(ctx, inputHash)
	if err != nil {
		return "", err
	}

	if isExisted {
		return cacheEntity.Result, nil
	}

	// cache iao
	iaoMap := make(map[string]*entity.Iao, 0)
	balanceMap := make(map[string]*entity.BalanceCache, len(batchReq))
	resultHandle := make([]iao.ResultHandle, 0, len(batchReq))
	investorMap := make(map[string]*entity.InvestorBuyIao, len(batchReq))

	for _, req := range batchReq {
		res := req.CloneToResult()
		iaoEntity, err := i.getIaoInfo(ctx, iaoMap, req.IaoId)
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "BuyBatchAsset - Handle req (%s) failed get IAO", req.ReqId, err.Error())
			res.Status = transaction.Rejected
			resultHandle = append(resultHandle, res)
			continue
		}
		if helper.CompareStringBalance(iaoEntity.RemainingAssetToken, "0") <= 0 {
			glogger.GetInstance().Errorf(ctx, "BuyBatchAsset - Handle req (%s) remaining of iao is zero", req.ReqId)
			res.Status = transaction.Rejected
			resultHandle = append(resultHandle, res)
			// TODO: return not continue for optimize
			continue
		}

		var numberATBuy string
		if helper.CompareStringBalance(iaoEntity.RemainingAssetToken, req.NumberAT) > 0 {
			numberATBuy = req.NumberAT
		} else {
			numberATBuy = iaoEntity.RemainingAssetToken
		}

		stableToken, err := helper.MulBalance(numberATBuy, iaoEntity.Rate)
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "BuyBatchAsset - Handle req (%s) failed to calculate stable token", req.ReqId)
			res.Status = transaction.Rejected
			resultHandle = append(resultHandle, res)
			continue
		}

		err = i.SubAmount(ctx, balanceMap, doc.IaoBalances, req.WalletId, req.TokenId, stableToken)
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "BuyBatchAsset - Handle req (%s) failed get Balance", req.ReqId, err.Error())
			res.Status = transaction.Rejected
			resultHandle = append(resultHandle, res)
			continue
		}

		if err := i.addInvestorBook(investorMap, req, stableToken, iaoEntity.AssetTokenId); err != nil {
			glogger.GetInstance().Errorf(ctx, "BuyBatchAsset - Handle req (%s) failed get create investor book", req.ReqId, err.Error())
			res.Status = transaction.Rejected
			resultHandle = append(resultHandle, res)
			continue
		}
		updateATRemain, _ := helper.SubBalance(iaoEntity.RemainingAssetToken, numberATBuy)
		updateST, _ := helper.AddBalance(iaoEntity.StableTokenAmount, stableToken)
		iaoEntity.RemainingAssetToken = updateATRemain
		iaoEntity.StableTokenAmount = updateST
		iaoMap[iaoEntity.Id] = iaoEntity

		res.Status = transaction.Confirmed
		res.NumberATFilled = numberATBuy
		res.NumberST = stableToken
		resultHandle = append(resultHandle, res)
	}

	// log create Read/Write Set
	glogger.GetInstance().Info(ctx, "Start Create Read/Write")
	resultJson, _ := json.Marshal(resultHandle)

	// update balance
	if err := i.UpdateBalance(ctx, balanceMap); err != nil {
		return "", err
	}

	// update iao
	if err := i.updateIao(ctx, iaoMap); err != nil {
		return "", err
	}

	// insert investor book
	if err := i.insertInvestorBook(ctx, investorMap); err != nil {
		return "", err
	}

	buyCache := entity.NewIaoCache(ctx)
	buyCache.Hash = inputHash
	buyCache.Result = string(resultJson)
	if err := i.Repo.Update(ctx, buyCache, doc.BuyIaoCache, helper.ResultCacheKey(buyCache.Hash)); err != nil {
		glogger.GetInstance().Errorf(ctx, "BuyBatchAsset - Create cache buy Iao failed with err (%v)", err.Error())
		return "", helper.RespError(errorcode.BizUnableUpdateIao)
	}
	glogger.GetInstance().Info(ctx, "End BuyBatchAsset")

	return string(resultJson), nil
}

func (i *iaoService) UpdateStatusIao(ctx contractapi.TransactionContextInterface, iaoId string, status statusIao.Status) error {
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	iaoEntity, err := i.GetIao(ctx, iaoId)
	if err != nil {
		glogger.GetInstance().Errorf(ctx, "UpdateStatus - GetIao failed with err (%s)", err.Error())
		return err
	}

	if iaoEntity.Status != status {
		iaoEntity.UpdatedAt = helper.TimestampISO(txTime.Seconds)
		iaoEntity.Status = status
		if err := i.Repo.Update(ctx, iaoEntity, doc.Iao, helper.IaoKey(iaoEntity.Id)); err != nil {
			glogger.GetInstance().Errorf(ctx, "UpdateStatus - Update Iao (%s) failed with err (%v)", err.Error())
			return helper.RespError(errorcode.BizUnableUpdateIao)
		}
	}

	return nil
}

func (i *iaoService) FinalizeIao(ctx contractapi.TransactionContextInterface, lstInvestorBook []string) error {
	mapInvestor := make(map[string]entity.InvestorBuyIao, 0)
	for _, key := range lstInvestorBook {
		compositeKey := strings.Split(key, "_")
		investorBook, err := i.GetInvestorBook(ctx, compositeKey[2], compositeKey[3])
		if err != nil {
			return err
		}
		if investorBook.Status != investor_book.NotDistributed {
			glogger.GetInstance().Errorf(ctx, "FinalizeIao - Investor Book (%s) have status invalidate", key)
		}

		var lstInvestor []entity.InvestorBuyIao
		err = json.Unmarshal([]byte(investorBook.Investor), &lstInvestor)
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "FinalizeIao - Unmarshal investor failed with err (%s)", err.Error())
			return err
		}

		// calculate at token of investor
		for _, buyIao := range lstInvestor {
			keyMap := investorBook.IaoId + "_" + buyIao.WalletId
			if _, ok := mapInvestor[keyMap]; !ok {
				mapInvestor[keyMap] = buyIao
			} else {
				currentInvestor := mapInvestor[keyMap]
				balanceAT, err := helper.AddBalance(currentInvestor.AssetTokenAmount, buyIao.AssetTokenAmount)
				if err != nil {
					glogger.GetInstance().Errorf(ctx, "FinalizeIao - Sum AT balance failed with err (%s)", err.Error())
					return err
				}
				currentInvestor.AssetTokenAmount = balanceAT
				mapInvestor[keyMap] = currentInvestor
			}
		}

		// update status of investor book
		investorBook.Status = investor_book.Distributed
		if err := i.Repo.Update(ctx, investorBook, doc.InvestorBook, helper.InvestorBookKey(compositeKey[2], compositeKey[3])); err != nil {
			glogger.GetInstance().Errorf(ctx, "FinalizeIao - Update Investor book (%s) failed with err (%v)", key, err.Error())
			return helper.RespError(errorcode.BizUnableUpdateInvestorBook)
		}
	}

	for keyMap, investor := range mapInvestor {
		iaoCompositeKey := strings.Split(keyMap, "_")
		txEntity := entity.NewTransaction(ctx)
		txEntity.SpenderWallet = iaoCompositeKey[0]
		txEntity.FromWallet = iaoCompositeKey[0]
		txEntity.ToWallet = investor.WalletId
		txEntity.FromTokenId = investor.AssetTokenId
		txEntity.ToTokenId = investor.AssetTokenId
		txEntity.FromTokenAmount = investor.AssetTokenAmount
		txEntity.ToTokenAmount = investor.AssetTokenAmount
		txEntity.TxType = transaction.DistributionAT
		if err := i.Repo.Update(ctx, txEntity, doc.Transactions, helper.TransactionKey(txEntity.Id)); err != nil {
			glogger.GetInstance().Errorf(ctx, "FinalizeIao - Create distribution AT  failed with err (%v)", err.Error())
			return helper.RespError(errorcode.BizUnableUpdateInvestorBook)
		}
	}

	return nil
}

func (i *iaoService) CancelIao(ctx contractapi.TransactionContextInterface, lstInvestorBook []string) error {
	mapInvestor := make(map[string]entity.InvestorBuyIao, 0)
	for _, key := range lstInvestorBook {
		compositeKey := strings.Split(key, "_")
		investorBook, err := i.GetInvestorBook(ctx, compositeKey[2], compositeKey[3])
		if err != nil {
			return err
		}
		if investorBook.Status != investor_book.NotDistributed {
			glogger.GetInstance().Errorf(ctx, "CancelIao - Investor Book (%s) have status invalidate", key)
		}

		var lstInvestor []entity.InvestorBuyIao
		err = json.Unmarshal([]byte(investorBook.Investor), &lstInvestor)
		if err != nil {
			glogger.GetInstance().Errorf(ctx, "CancelIao - Unmarshal investor failed with err (%s)", err.Error())
			return err
		}

		// calculate at token of investor
		for _, buyIao := range lstInvestor {
			keyMap := investorBook.IaoId + "_" + buyIao.WalletId
			if _, ok := mapInvestor[keyMap]; !ok {
				mapInvestor[keyMap] = buyIao
			} else {
				currentInvestor := mapInvestor[keyMap]
				balanceAT, err := helper.AddBalance(currentInvestor.StableTokenAmount, buyIao.StableTokenAmount)
				if err != nil {
					glogger.GetInstance().Errorf(ctx, "CancelIao - Sum ST balance failed with err (%s)", err.Error())
					return err
				}
				currentInvestor.StableTokenAmount = balanceAT
				mapInvestor[keyMap] = currentInvestor
			}
		}

		// update status of investor book
		investorBook.Status = investor_book.Canceled
		if err := i.Repo.Update(ctx, investorBook, doc.InvestorBook, helper.InvestorBookKey(compositeKey[2], compositeKey[3])); err != nil {
			glogger.GetInstance().Errorf(ctx, "CancelIao - Update Investor book (%s) failed with err (%v)", key, err.Error())
			return helper.RespError(errorcode.BizUnableUpdateInvestorBook)
		}
	}

	for keyMap, investor := range mapInvestor {
		iaoCompositeKey := strings.Split(keyMap, "_")
		txEntity := entity.NewTransaction(ctx)
		txEntity.SpenderWallet = iaoCompositeKey[0]
		txEntity.FromWallet = iaoCompositeKey[0]
		txEntity.ToWallet = investor.WalletId
		txEntity.FromTokenId = investor.StableTokenId
		txEntity.ToTokenId = investor.StableTokenId
		txEntity.FromTokenAmount = investor.StableTokenAmount
		txEntity.ToTokenAmount = investor.StableTokenAmount
		txEntity.TxType = transaction.ReturnST
		if err := i.Repo.Update(ctx, txEntity, doc.Transactions, helper.TransactionKey(txEntity.Id)); err != nil {
			glogger.GetInstance().Errorf(ctx, "CancelIao - Create distribution AT  failed with err (%v)", err.Error())
			return helper.RespError(errorcode.BizUnableUpdateInvestorBook)
		}
	}

	return nil
}

func (i *iaoService) getIaoInfo(ctx contractapi.TransactionContextInterface, iaoMap map[string]*entity.Iao, iaoId string) (*entity.Iao, error) {
	if _, ok := iaoMap[iaoId]; !ok {
		iaoEntity, err := i.GetIao(ctx, iaoId)
		if err != nil {
			return nil, err
		}

		iaoMap[iaoEntity.Id] = iaoEntity
	}

	if iaoMap[iaoId].Status != statusIao.Open {
		return nil, errors.New("Iao has status not opening")
	}

	return iaoMap[iaoId], nil
}

func (i *iaoService) getBalanceWalletInfo(ctx contractapi.TransactionContextInterface, balanceMap map[string]*entity.Balance,
	walletId, tokenId string) (*entity.Balance, error) {
	if _, ok := balanceMap[walletId]; !ok {
		balanceEntity, err := i.GetBalanceOfToken(ctx, doc.IaoBalances, walletId, tokenId)
		if err != nil {
			return nil, err
		}
		balanceMap[walletId] = balanceEntity
	}

	return balanceMap[walletId], nil
}

// updateIao update remaining token of IAO after handle transaction
func (i *iaoService) updateIao(ctx contractapi.TransactionContextInterface, iaoMap map[string]*entity.Iao) error {
	txTime, _ := ctx.GetStub().GetTxTimestamp()
	for _, itemIao := range iaoMap {
		itemIao.UpdatedAt = helper.TimestampISO(txTime.Seconds)
		if err := i.Repo.Update(ctx, itemIao, doc.Iao, helper.IaoKey(itemIao.Id)); err != nil {
			glogger.GetInstance().Errorf(ctx, "UpdateIao - Update Iao (%s) failed with err (%v)", itemIao.Id, err.Error())
			return helper.RespError(errorcode.BizUnableUpdateIao)
		}
		glogger.GetInstance().Debugf(ctx, "UpdateIao - Update Iao (%s) succeed", itemIao.Id)
	}

	return nil
}

func (i *iaoService) insertInvestorBook(ctx contractapi.TransactionContextInterface, investorMap map[string]*entity.InvestorBuyIao) error {
	iaoInvestorMap := make(map[string][]*entity.InvestorBuyIao, 0)
	for key, buyIao := range investorMap {
		compositeKey := strings.Split(key, "_")
		iaoInvestorMap[compositeKey[0]] = append(iaoInvestorMap[compositeKey[0]], buyIao)
	}
	for key, value := range iaoInvestorMap {
		strJsonInvestor, _ := json.Marshal(value)
		investorBookEntity := entity.NewInvestorBook(ctx)
		investorBookEntity.Investor = string(strJsonInvestor)
		investorBookEntity.IaoId = key
		investorBookEntity.Status = investor_book.NotDistributed

		if err := i.Repo.Update(ctx, investorBookEntity, doc.InvestorBook, helper.InvestorBookKey(investorBookEntity.IaoId, ctx.GetStub().GetTxID())); err != nil {
			glogger.GetInstance().Errorf(ctx, "UpdateIao - Update Iao (%s) failed with err (%v)", investorBookEntity.IaoId, err.Error())
			return helper.RespError(errorcode.BizUnableCreateInvestorBook)
		}
	}
	return nil
}

func (i *iaoService) addInvestorBook(investorBookMap map[string]*entity.InvestorBuyIao, req iao.BuyAsset, amountST, assetTokenId string) (err error) {
	var investor *entity.InvestorBuyIao
	key := req.IaoId + "_" + req.WalletId
	if _, ok := investorBookMap[key]; !ok {
		investor = new(entity.InvestorBuyIao)
		investor.WalletId = req.WalletId
		investor.AssetTokenId = assetTokenId
		investor.StableTokenId = req.TokenId
		investor.AssetTokenAmount = "0"
		investor.StableTokenAmount = "0"

	} else {
		investor = investorBookMap[key]
	}

	balanceAT, err := helper.AddBalance(investor.AssetTokenAmount, req.NumberAT)
	if err != nil {
		return err
	}

	balanceST, err := helper.AddBalance(investor.StableTokenAmount, amountST)
	if err != nil {
		return err
	}
	investor.AssetTokenAmount = balanceAT
	investor.StableTokenAmount = balanceST

	investorBookMap[key] = investor

	return nil
}
