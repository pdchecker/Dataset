package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type PowerTXContract struct {
	contractapi.Contract
}

// Compact 电能交易结构体
type Compact struct {
	CompactId		string  	`json:"compact_id"`
	State	    	string  	`json:"state"`
	PowerPlantName  string 		`json:"power_plant_name"`
	PowerUserName   string		`json:"power_user_name"`
	AdminName	    string		`json:"admin_name"`
	Transaction 	int     	`json:"transaction"`
	Price           float32 	`json:"price"`
	StartTime 		string		`json:"start_time"`
	EndTime 		string		`json:"end_time"`
}

// Commit powerUser提交compact
func (p *PowerTXContract) Commit(
	ctx contractapi.TransactionContextInterface,
	compactId string,
	powerUserName string,
	transaction int,
	price float32,
	startTime string,
	endTime string) (*Compact, error) {
	// 1.判断时间是否符合规范
	var t TimeContract
	if !t.CompareTime(startTime, endTime) {
		return nil, fmt.Errorf("End time earlier than Start time ! ")
	}

	// 2.判断compact是否存在
	if p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact existed ! ")
	}

	// 3.查看powerUser是否存在
	var r RoleContract
	powerUser, err := r.QueryUser(ctx, powerUserName)

	if err != nil {
		return nil, fmt.Errorf("Query poweruser false, %s ", err.Error())
	}

	// 4.查看powerUser信用值， 若小于某个额度，则拒绝交易
	if powerUser.UserCredit - CreditBorder < 0 {
		return nil, fmt.Errorf("PowerUser credit less than %d ", CreditBorder)
	}

	// 5.结构体赋值
	compact := Compact{
		CompactId: compactId,
		State: "Committing",
		PowerPlantName: "",
		PowerUserName: powerUserName,
		AdminName: "",
		Transaction: transaction,
		Price: price,
		StartTime: startTime,
		EndTime: endTime,
	}

	compactAsBytes, _ := json.Marshal(compact)

	// 6.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return &compact, nil
}

// Bid powerPlant 竞价
func (p *PowerTXContract) Bid(
	ctx contractapi.TransactionContextInterface,
	compactId string,
	powerPlantName string,
	price float32) (*Compact, error) {
	// 1.判断compact是否存在
	if !p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact not existed ! ")
	}

	// 2.判断powerPlant是否存在
	var r RoleContract
	powerPlant, errOfPowerPlant := r.QueryUser(ctx, powerPlantName)

	if errOfPowerPlant != nil {
		return nil, errOfPowerPlant
	}

	// 3.查看powerPlant信用值， 若小于某个额度，则拒绝交易
	if powerPlant.UserCredit - CreditBorder < 0 {
		return nil, fmt.Errorf("PowerPlant credit less than %d ", CreditBorder)
	}

	// 4.获取compact交易信息
	compact, err := p.QueryCompact(ctx, compactId)

	// 5.判断获取compact交易信息是否成功
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 6.判断compact的状态
	if compact.State != "Committing" {
		return nil, fmt.Errorf("Compact state is not committing ! ")
	}

	// 7.判断是否在交易时间
	//var t TimeContract
	//if !t.CompareWithNow(compact.StartTime) || t.CompareWithNow(compact.EndTime) {
	//	return nil, fmt.Errorf("It is not time to transaction ! ")
	//}

	// 8.compact交易结构体赋值
	compact.PowerPlantName = powerPlantName
	compact.Price = price
	compact.State = "Biding"

	compactAsBytes, _ := json.Marshal(compact)

	// 9.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, err
	}

	return compact, nil
 }

// Deal admin参与交易，达成三方交易
func (p *PowerTXContract) Deal(
	ctx contractapi.TransactionContextInterface,
	compactId string,
	adminName string) (*Compact, error) {
	// 1.判断compact是否存在
	if !p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact not existed ! ")
	}

	// 2.判断admin是否存在
	var r RoleContract
	admin, errOfAdmin := r.QueryUser(ctx, adminName)

	if errOfAdmin != nil {
		return nil, fmt.Errorf(errOfAdmin.Error())
	}

	// 3.查看admin信用值， 若小于某个额度，则拒绝交易
	if admin.UserCredit - CreditBorder < 0 {
		return nil, fmt.Errorf("Admin credit less than %d ", CreditBorder)
	}

	// 4.获取compact交易信息
	compact, err := p.QueryCompact(ctx, compactId)

	// 5.判断获取compact交易信息是否成功
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 6.判断是否在交易时间
	//var t TimeContract
	//if !t.CompareWithNow(compact.StartTime) || t.CompareWithNow(compact.EndTime) {
	//	return nil, fmt.Errorf("It is not time to transaction ! ")
	//}

	// 7.判断compact的状态
	if compact.State != "Accepted" {
		return nil, fmt.Errorf("Compact state is not Accepted ! ")
	}

	// 8.compact交易结构体赋值
	compact.AdminName = adminName
	compact.State = "Deal"

	compactAsBytes, _ := json.Marshal(compact)
	// 9.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return compact, nil
}

func (p *PowerTXContract) CheckCompact(
	ctx contractapi.TransactionContextInterface,
	compactId string,
	powerUsed int,
	powerPlant int) (*Compact, error) {
	// 1.判断compact是否存在
	if !p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact not existed ! ")
	}

	// 2.获取compact交易信息
	compact, err := p.QueryCompact(ctx, compactId)

	// 3.判断获取compact交易信息是否成功
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 4.判断compact的状态
	if compact.State != "Deal" {
		return nil, fmt.Errorf("Compact state is not Accepted ! ")
	}

	// 5.判断交易是否到达预期时间
	//var t TimeContract
	//if !t.CompareWithNow(compact.EndTime) {
	//	return nil, fmt.Errorf("The compact is not end! ")
	//}

	// 6.检查交易情况
	// 6.1更新信用值和交易额度
	var r RoleContract
	var v VarChangeContract
	if compact.Transaction - powerUsed < 0 {
		_ = r.ChangeCredit(ctx, compact.PowerUserName, v.AwardCredit(compact.Transaction))
	} else {
		_ = r.ChangeCredit(ctx, compact.PowerUserName, v.AwardCredit(powerUsed - compact.Transaction))
	}

	if compact.Transaction - powerPlant < 0 {
		_ = r.ChangeCredit(ctx, compact.PowerPlantName, v.AwardCredit(compact.Transaction))
	} else {
		_ = r.ChangeCredit(ctx, compact.PowerPlantName, v.AwardCredit(powerPlant - compact.Transaction))
	}

	_ = r.ChangePower(ctx, compact.PowerUserName,powerUsed)
	_ = r.ChangePower(ctx, compact.PowerPlantName,powerPlant)
	_ = r.ChangePower(ctx, compact.AdminName,powerPlant + powerUsed)

	compact.State = "Done"
	compactAsBytes, _ := json.Marshal(compact)

	// 7.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return compact, nil
}

// Reject powerUser 拒绝报价，并重新提出报价
func (p *PowerTXContract) Reject(
	ctx contractapi.TransactionContextInterface,
	compactId string,
	newPrice float32) (*Compact, error) {
	// 0.判断用户
	// 1.判断compact是否存在
	if !p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact not existed ! ")
	}

	// 2.获取compact交易信息
	compact, err := p.QueryCompact(ctx, compactId)

	// 3.判断获取compact交易信息是否成功
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 4.判断是否在交易时间
	//var t TimeContract
	//if !t.CompareWithNow(compact.StartTime) || t.CompareWithNow(compact.EndTime) {
	//	return nil, fmt.Errorf("It is not time to transaction ! ")
	//}

	// 5.判断compact的状态
	if compact.State != "Biding" {
		return nil, fmt.Errorf("Compact state is not biding ! ")
	}

	// 6.compact交易结构体赋值
	compact.PowerPlantName = ""
	compact.Price = newPrice
	compact.State = "Committing"

	compactAsBytes, _ := json.Marshal(compact)
	// 7.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return compact, nil
}

// Accept powerUser接受报价
func (p *PowerTXContract) Accept(
	ctx contractapi.TransactionContextInterface,
	compactId string) (*Compact, error) {
	// 1.判断compact是否存在
	if !p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact not existed ! ")
	}

	// 2.获取compact交易信息
	compact, err := p.QueryCompact(ctx, compactId)

	// 3.判断获取compact交易信息是否成功
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 4.判断是否在交易时间
	//var t TimeContract
	//if !t.CompareWithNow(compact.StartTime) || t.CompareWithNow(compact.EndTime) {
	//	return nil, fmt.Errorf("It is not time to transaction ! ")
	//}

	// 5.判断compact的状态
	if compact.State != "Biding" {
		return nil, fmt.Errorf("Compact state is not biding ! ")
	}

	// 6.compact交易结构体赋值
	compact.State = "Accepted"

	compactAsBytes, _ := json.Marshal(compact)
	// 7.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return compact, nil
}

// CancelCommit powerUser取消提交commit
func (p *PowerTXContract) CancelCommit(
	ctx contractapi.TransactionContextInterface,
	compactId string) (*Compact, error) {
	// 1.判断compact是否存在
	if !p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact not existed ! ")
	}

	// 2.获取compact交易信息
	compact, err := p.QueryCompact(ctx, compactId)

	// 3.判断获取compact交易信息是否成功
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 4.判断是否在交易时间
	//var t TimeContract
	//if !t.CompareWithNow(compact.StartTime) || t.CompareWithNow(compact.EndTime) {
	//	return nil, fmt.Errorf("It is not time to transaction ! ")
	//}

	// 5.判断compact的状态
	if compact.State != "Committing" {
		return nil, fmt.Errorf("Compact state is not committing ! ")
	}

	// 6.compact交易结构体赋值
	compact.State = "CancelCommit"

	compactAsBytes, _ := json.Marshal(compact)
	// 7.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return compact, nil
}

// CancelBid powerPlant取消报价
func (p *PowerTXContract) CancelBid(
	ctx contractapi.TransactionContextInterface,
	compactId string) (*Compact, error) {
	// 1.判断compact是否存在
	if !p.CompactExist(ctx, compactId) {
		return nil, fmt.Errorf("Compact not existed ! ")
	}

	// 2.获取compact交易信息
	compact, err := p.QueryCompact(ctx, compactId)

	// 3判断获取compact交易信息是否成功
	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	// 4.判断是否在交易时间
	//var t TimeContract
	//if !t.CompareWithNow(compact.StartTime) || t.CompareWithNow(compact.EndTime) {
	//	return nil, fmt.Errorf("It is not time to transaction ! ")
	//}

	// 5.判断compact的状态
	if compact.State != "Biding" {
		return nil, fmt.Errorf("Compact state is not biding ! ")
	}

	// 6.compact交易结构体赋值
	compact.PowerPlantName = ""
	compact.State = "Committing"

	compactAsBytes, _ := json.Marshal(compact)

	// 7.上链
	err = ctx.GetStub().PutState(compactId, compactAsBytes)

	if err != nil {
		return nil, fmt.Errorf(err.Error())
	}

	return compact, nil
}

// QueryCompact 获取compact信息
func (p *PowerTXContract) QueryCompact(
	ctx contractapi.TransactionContextInterface,
	compactId string) (*Compact, error) {
	// 1.获取compact交易信息
	compactAsBytes, errOfCompact := ctx.GetStub().GetState(compactId)

	// 1.1判断获取交易是否错误
	if errOfCompact != nil {
		return nil, fmt.Errorf("Failed to query Compact Info from world state. %s ", errOfCompact.Error())
	}

	// 1.2判断是否存在交易compact
	if compactAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", compactId)
	}

	// 2.赋值
	compact := new(Compact)
	_ = json.Unmarshal(compactAsBytes, compact)

	return compact, nil
}

// CompactExist 判断compact是否存在
func (p *PowerTXContract) CompactExist(
	ctx contractapi.TransactionContextInterface,
	compactId string) bool {
	// 1.获取compact
	compactAsBytes, _ := ctx.GetStub().GetState(compactId)

	// 2.如果compact不存在， 返回false
	if compactAsBytes == nil {
		return false
	}

	// 3. 如果compact存在，返回true
	return true
}