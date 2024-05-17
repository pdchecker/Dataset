package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

type SmartContract struct {
	contractapi.Contract
}

// Response 统一消息返回
type Response struct {
	Code string      `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// Machining 加工信息
type Machining struct {
	//加工ID
	MachiningId string `json:"machining_id"`
	//原料厂商负责人
	Leader string `json:"leader"`
	//货物ID
	CropsId string `json:"crops_id"`
	//厂商负责人tel
	LeaderTel string `json:"leader_tel"`
	//厂商名称
	FactoryName string `json:"factory_name"`
	//检测结果
	TestingResult string `json:"testing_result"`
	//入库时间
	InFactoryTime string `json:"in_factory_time"`
	//出库时间
	OutFactoryTime string `json:"out_factory_time"`
	//质检过程图片
	TestingPhotoUrl string `json:"testing_photo_url"`
	//备注
	Remarks string `json:"remarks"`
}

// CreateMachining 加工数据上链
func (s *SmartContract) CreateMachining(ctx contractapi.TransactionContextInterface, MachiningId, Leader, CropsId, LeaderTel, FactoryName, TestingResult, InFactoryTime, OutFactoryTime, TestingPhotoUrl, Remarks string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	ok, err := IsExist(ctx, MachiningId)
	if err != nil {
		resp.Code = "106"
		resp.Msg = err.Error()
		return &resp
	}
	if ok {
		resp.Code = "300"
		resp.Msg = fmt.Sprintf("machining_id already exist")
		return &resp
	}
	var machining = Machining{MachiningId, Leader, CropsId, LeaderTel, FactoryName, TestingResult, InFactoryTime, OutFactoryTime, TestingPhotoUrl, Remarks}
	err = WriteLedger(ctx, machining, MachiningId)
	if err != nil {
		resp.Code = "101"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// QueryMachiningById 根据加工ID查询加工信息
func (s *SmartContract) QueryMachiningById(ctx contractapi.TransactionContextInterface, MachiningId string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	result, err := GetState0(ctx, MachiningId)
	if err != nil {
		resp.Code = "102"
		resp.Msg = err.Error()
		return &resp
	}
	resp.Data = result
	return &resp
}

// QueryMachiningByCropsId 根据作物ID查询所有加工信息
func (s *SmartContract) QueryMachiningByCropsId(ctx contractapi.TransactionContextInterface, CropsId string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	results, err := GetState1(ctx, "crops_id", CropsId)
	if err != nil {
		resp.Code = "103"
		resp.Msg = err.Error()
		return &resp
	}
	resp.Data = results
	return &resp
}
func main() {
	Chaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}
	if err = Chaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}

// WriteLedger 写入账本	传入 ctx, 结构体, 键值
func WriteLedger(ctx contractapi.TransactionContextInterface, obj interface{}, key string) error {
	jsonBytes, err := json.Marshal(obj)
	if err != nil {
		return errors.New(fmt.Sprintf("json error: %s", err))
	}
	//写入区块链账本
	if err = ctx.GetStub().PutState(key, jsonBytes); err != nil {
		return errors.New(fmt.Sprintf("Error writing to blockchain ledger: %s", err))
	}
	return nil
}

// GetState0 根据键查询,一般查询	传入 ctx, 键值
func GetState0(ctx contractapi.TransactionContextInterface, key string) ([]byte, error) {
	result, err := ctx.GetStub().GetState(key)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Data acquisition error: %s", err))
	}
	if result == nil {
		return nil, errors.New(fmt.Sprintf("The Key of the query does not exist: %s", key))
	}
	return result, nil
}

// GetState1 富查询数据,针对单一键查询	传入 ctx, 查询json格式, 键值
func GetState1(ctx contractapi.TransactionContextInterface, keyName string, key string) ([]byte, error) {
	queryString := fmt.Sprintf(`{"selector":{"%s":"%s"}}`, keyName, key)
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Data acquisition error: %s", err))
	}
	if !resultsIterator.HasNext() {
		return nil, errors.New(fmt.Sprintf("The Key of the query does not exist: %s", key))
	}
	defer resultsIterator.Close()
	var buffer bytes.Buffer
	bArrayMemberAlreadyWritten := false
	buffer.WriteString(`{"result":[`)

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next() //获取迭代器中的每一个值
		if err != nil {
			return nil, errors.New(fmt.Sprintf("Data acquisition from iterator error: %s", err))
		}
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString(string(queryResponse.Value)) //将查询结果放入Buffer中
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString(`]}`)

	return buffer.Bytes(), nil
}

// IsExist 根据键查询,一般查询	传入 ctx, 键值
func IsExist(ctx contractapi.TransactionContextInterface, key string) (bool, error) {
	result, err := ctx.GetStub().GetState(key)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Data acquisition error: %s", err))
	}
	if result == nil {
		return false, nil
	}
	return true, nil
}
