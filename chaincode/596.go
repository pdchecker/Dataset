package main

import (
	"bytes"
	"crypto/sha256"
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
type User struct {
	UserID string `json:"user_id"`
	Passwd []byte `json:"passwd"`
}

// CreateAccount 新建账号
func (s *SmartContract) CreateAccount(ctx contractapi.TransactionContextInterface, UserID, Passwd string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	//是否存在
	ok, err := IsExist(ctx, UserID)
	if err != nil {
		resp.Code = "106"
		resp.Msg = err.Error()
		return &resp
	}
	if ok {
		resp.Code = "300"
		resp.Msg = fmt.Sprintf("user_id already exist")
		return &resp
	}
	//哈希
	h := sha256.New()
	h.Write([]byte(Passwd))
	passByte := h.Sum(nil)
	var userJson = User{UserID, passByte}
	err = WriteLedger(ctx, userJson, UserID)
	if err != nil {
		resp.Code = "101"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// ChangePass 更改密码
func (s *SmartContract) ChangePass(ctx contractapi.TransactionContextInterface, UserID, Passwd string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	//是否存在
	_, err := GetState0(ctx, UserID)
	if err != nil {
		resp.Code = "102"
		resp.Msg = err.Error()
		return &resp
	}
	h := sha256.New()
	h.Write([]byte(Passwd))
	passByte := h.Sum(nil)
	var userJson = User{UserID, passByte}
	err = WriteLedger(ctx, userJson, UserID)
	if err != nil {
		resp.Code = "101"
		resp.Msg = err.Error()
		return &resp
	}
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
