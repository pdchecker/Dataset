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
	"github.com/Akachain/gringotts/helper/glogger"
	"github.com/Akachain/gringotts/services"
	"github.com/Akachain/gringotts/services/health_check"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type HealthCheckHandler struct {
	hcService services.HealthCheck
}

func NewHealthCheckHandler() HealthCheckHandler {
	hcService := health_check.NewHealthCheckService()
	return HealthCheckHandler{hcService: hcService}
}

// Transfer to transfer token between wallet.
func (hc *HealthCheckHandler) CreateHealthCheck(ctx contractapi.TransactionContextInterface) (string, error) {
	glogger.GetInstance().Info(ctx, "-----------Token Handler - Transfer-----------")

	return hc.hcService.CreateHealthCheck(ctx)
}
