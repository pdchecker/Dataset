package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/hyperledger/fabric-contract-api-go/metadata"
)

type InsuranceContract struct {
	contractapi.Contract
}

// Insurance 保险数据结构
type Insurance struct {

	// 保险单号
	Key string `json:"key"`

	// 保险名称
	InsuranceName string `json:"insuranceName"`

	// 客户姓名
	ClientName string `json:"clientName"`

	// 客户年龄
	ClientAge int32 `json:"clientAge"`

	// 客户性别
	ClientGender string `json:"clientGender"`

	// 是否报案
	Reported bool `json:"reported"`

	// 报案描述
	ReportCaseDesc string `json:"reportCaseDesc"`

	// 是否赔偿
	Compensated bool `json:"compensated"`

	// 赔偿金额
	CompensationAmount float64 `json:"compensationAmount"`
}

func (ic *InsuranceContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	fmt.Println("保险合约初始化交易 ***********************************************")
	return nil
}

// CreateInsurance 创建保险
func (ic *InsuranceContract) CreateInsurance(ctx contractapi.TransactionContextInterface, key string, insuranceName string, clientName string, clientAge int32, clientGender string) (Insurance, error) {

	insurance := Insurance{
		Key:                key,
		InsuranceName:      insuranceName,
		ClientName:         clientName,
		ClientAge:          clientAge,
		ClientGender:       clientGender,
		Reported:           false,
		ReportCaseDesc:     "",
		Compensated:        false,
		CompensationAmount: 0.0,
	}

	fmt.Println("创建保险=====> ", insurance)

	bytes, _ := json.Marshal(insurance)

	ctx.GetStub().PutState(key, bytes)

	return insurance, nil
}

// ReportCase 保险报案
func (ic *InsuranceContract) ReportCase(ctx contractapi.TransactionContextInterface, key string, reportCaseDesc string) (*Insurance, error) {

	stub := ctx.GetStub()
	bytes, err := stub.GetState(key)

	if err != nil {
		return nil, fmt.Errorf("读取保险数据失败. %s", err.Error())
	}

	if bytes == nil {
		return nil, fmt.Errorf("%s 保险数据不存在", key)
	}

	insurance := new(Insurance)
	_ = json.Unmarshal(bytes, insurance)

	insurance.Reported = true
	insurance.ReportCaseDesc = reportCaseDesc

	fmt.Println("保险报案=====> ", insurance)

	newBytes, _ := json.Marshal(insurance)
	stub.PutState(key, newBytes)

	return insurance, nil
}

// Compensate 保险赔付款
func (ic *InsuranceContract) Compensate(ctx contractapi.TransactionContextInterface, key string, compensationAmount float64) (*Insurance, error) {
	stub := ctx.GetStub()
	bytes, err := stub.GetState(key)

	if err != nil {
		return nil, fmt.Errorf("读取保险数据失败. %s", err.Error())
	}

	if bytes == nil {
		return nil, fmt.Errorf("%s 保险数据不存在", key)
	}

	insurance := new(Insurance)
	_ = json.Unmarshal(bytes, insurance)
	insurance.Compensated = true
	insurance.CompensationAmount = compensationAmount

	fmt.Println("保险赔付=====> ", insurance)

	newBytes, _ := json.Marshal(insurance)
	stub.PutState(key, newBytes)

	return insurance, nil
}

// QueryInsurance 查询保险
func (ic *InsuranceContract) QueryInsurance(ctx contractapi.TransactionContextInterface, key string) (*Insurance, error) {
	stub := ctx.GetStub()
	bytes, err := stub.GetState(key)

	if err != nil {
		return nil, fmt.Errorf("读取保险数据失败. %s", err.Error())
	}

	if bytes == nil {
		return nil, fmt.Errorf("%s 保险数据不存在", key)
	}

	insurance := new(Insurance)
	_ = json.Unmarshal(bytes, insurance)

	return insurance, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(InsuranceContract))

	if err != nil {
		fmt.Printf("Error create InsuranceContract chaincode: %s", err.Error())
		return
	}

	chaincode.Info.Version = "1.0.0"
	chaincode.Info.Description = "InsuranceContract"
	chaincode.Info.License = new(metadata.LicenseMetadata)
	chaincode.Info.License.Name = "Apache-2.0"
	chaincode.Info.Contact = new(metadata.ContactMetadata)
	chaincode.Info.Contact.Name = "InsuranceContract"

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting InsuranceContract chaincode: %s", err.Error())
	}
}
