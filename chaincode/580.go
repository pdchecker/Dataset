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
	iaoDto "github.com/Akachain/gringotts/dto/iao"
	"github.com/Akachain/gringotts/errorcode"
	"github.com/Akachain/gringotts/helper"
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/services"
	"github.com/Akachain/gringotts/services/iao"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type IaoHandler struct {
	iaoService services.Iao
}

func NewIaoHandler() IaoHandler {
	return IaoHandler{iao.NewIaoService()}
}

func (i IaoHandler) CreateAsset(ctx contractapi.TransactionContextInterface, asset iaoDto.CreateAsset) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Iao Handler - CreateAsset-----------")

	// checking dto validate
	if err := asset.IsValid(); err != nil {
		glogger.GetInstance().Errorf(ctx, "IaoHandler - CreateAsset Input invalidate %v", err)
		return "", helper.RespError(errorcode.InvalidParam)
	}

	return i.iaoService.CreateAsset(ctx, asset.Code, asset.Name, asset.OwnerWallet, asset.TokenName, asset.TickerToken,
		asset.MaxSupply, asset.TotalValue, asset.DocumentUrl)
}

func (i IaoHandler) CreateIao(ctx contractapi.TransactionContextInterface, assetIao iaoDto.AssetIao) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Iao Handler - CreateIao-----------")

	// checking dto validate
	if err := assetIao.IsValid(); err != nil {
		glogger.GetInstance().Errorf(ctx, "IaoHandler - CreateIao Input invalidate %v", err)
		return "", helper.RespError(errorcode.InvalidParam)
	}

	return i.iaoService.CreateIao(ctx, assetIao.AssetId, assetIao.AssetTokenAmount, assetIao.StartDate, assetIao.EndDate, assetIao.Rate)
}

func (i IaoHandler) BuyBatchAsset(ctx contractapi.TransactionContextInterface, batchAsset iaoDto.BuyBatchAsset) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Iao Handler - BuyBatchAsset-----------")

	// checking dto validate
	if err := batchAsset.IsValid(); err != nil {
		glogger.GetInstance().Errorf(ctx, "IaoHandler - BuyBatchAsset Input invalidate %v", err)
		return "", helper.RespError(errorcode.InvalidParam)
	}

	return i.iaoService.BuyBatchAsset(ctx, batchAsset.Requests)
}

func (i IaoHandler) UpdateStatusIao(ctx contractapi.TransactionContextInterface, updateIao iaoDto.UpdateIao) error {
	glogger.GetInstance().Info(ctx, "-----------Iao Handler - UpdateStatusIao-----------")

	// checking dto validate
	if err := updateIao.IsValid(); err != nil {
		glogger.GetInstance().Errorf(ctx, "IaoHandler - UpdateStatusIao Input invalidate %v", err)
		return helper.RespError(errorcode.InvalidParam)
	}

	return i.iaoService.UpdateStatusIao(ctx, updateIao.IaoId, updateIao.Status)
}

func (i IaoHandler) FinalizeIao(ctx contractapi.TransactionContextInterface, finishIao iaoDto.FinishIao) error {
	glogger.GetInstance().Info(ctx, "-----------Iao Handler - FinalizeIao-----------")

	// checking dto validate
	if err := finishIao.IsValid(); err != nil {
		glogger.GetInstance().Errorf(ctx, "IaoHandler - FinalizeIao Input invalidate %v", err)
		return helper.RespError(errorcode.InvalidParam)
	}

	return i.iaoService.FinalizeIao(ctx, finishIao.InvestorBookId)
}

func (i IaoHandler) CancelIao(ctx contractapi.TransactionContextInterface, finishIao iaoDto.FinishIao) error {
	glogger.GetInstance().Info(ctx, "-----------Iao Handler - CancelIao-----------")

	// checking dto validate
	if err := finishIao.IsValid(); err != nil {
		glogger.GetInstance().Errorf(ctx, "IaoHandler - CancelIao Input invalidate %v", err)
		return helper.RespError(errorcode.InvalidParam)
	}

	return i.iaoService.CancelIao(ctx, finishIao.InvestorBookId)
}
