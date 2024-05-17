package main //包名一定要是main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi" //引入常用接口
	"strconv"
	"time"
)

// SmartContract provides functions for managing a car
// 需要定义一个链码总体的结构体，这个结构体定义了关于资产调用的常用接口
type SmartContract struct {
	contractapi.Contract
}

// Car describes basic details of what makes up a car
//定义car结构体，声明其中的字段（对应car的各种属性）
/*type Car struct {
	Make   string `json:"make"`
	Model  string `json:"model"`
	Colour string `json:"colour"`
	Owner  string `json:"owner"`
}*/

type Contract_in_company struct {
	ContractName    string `json:"contractname"`
	ContractContent string `json:"contractcontent"`
	CreaterName     string `json:"creatername"`
	CreaterSign     string `json:"creatersign"`
	CreateTime      string `json:"creatertime"`
}

type QueryResult_incompany struct {
	Key    string `json:"Key"`
	Record *Contract_in_company
}

type Contract_among_company struct {
	ContractName       string `json:"contractname"`
	ContractContent    string `json:"contractcontent"`
	CreaterCompanyName string `json:"creatercompanyname"`
	CreaterCompanySign string `json:"creatercompanysign"`
	CreateTime         string `json:"creatertime"`
}

// InitLedger adds a base set of cars to the ledger
// 首先需要init方法进行初始化
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	contracts_in_company := []Contract_in_company{
		Contract_in_company{ContractName: "Toyota", ContractContent: "Prius", CreaterName: "blue", CreaterSign: "Tomoko"},
		Contract_in_company{ContractName: "Toyo", ContractContent: "Pri", CreaterName: "bl", CreaterSign: "Tomo"},
	}

	for i, contract_in_company := range contracts_in_company {
		contractAsBytes, _ := json.Marshal(contract_in_company)
		//ctx.GetStub() 相当于 Stub()
		//PutState() : 存入数据库的API
		err := ctx.GetStub().PutState("CONTRACT"+strconv.Itoa(i), contractAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// 新增
// 输入信息为部门名，私钥签名，合同文本
func StartContract(ctx contractapi.TransactionContextInterface, contractname string, departmentname string, signature string, contractcontent string) error {

	if departmentname == "" || signature == "" || contractcontent == "" || contractname == "" {
		return fmt.Errorf("参数存在空值")
	}

	//判断私钥签名是否属于此部门
	/*
		未完成
	*/

	//时间戳
	timeUnix := time.Now().Unix() //时间戳
	createTime := time.Unix(timeUnix, 0).Format("2006-01-02 15:04:05")
	contract_in_company := Contract_in_company{
		ContractName:    contractname,
		ContractContent: contractcontent,
		CreaterName:     departmentname,
		CreaterSign:     signature,
		CreateTime:      createTime,
	}
	contractAsBytes, _ := json.Marshal(contract_in_company)
	//序列化后，存入数据库，使用putstate
	return ctx.GetStub().PutState(contractname, contractAsBytes)
}

// 第二个函数方法：querycar() 实现功能：查询车辆信息
// QueryCar returns the Contract stored in the world state with given id
func (s *SmartContract) QueryContract_incompany(ctx contractapi.TransactionContextInterface, ContractName string) (*Contract_in_company, error) {
	contractAsBytes, err := ctx.GetStub().GetState(ContractName)
	//使用getstate获取车辆信息
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if contractAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", ContractName)
	}

	contract := new(Contract_in_company)
	_ = json.Unmarshal(contractAsBytes, contract)
	//反序列化后读出数据
	return contract, nil
}

// 第三个函数方法： queryallcars() 实现功能：查询所有车辆信息
// QueryAllContract returns all cars found in world state
func (s *SmartContract) QueryALLContract_incompany(ctx contractapi.TransactionContextInterface) ([]QueryResult_incompany, error) {
	startKey := ""
	endKey := ""
	//GetStateByRange: 给定起始和终点，根据这个范围来查询信息
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()
	//定义了查询结果的切片
	results := []QueryResult_incompany{}
	//查询之后返回迭代器，GetStateByRange里的hasnext()和next()进行循环，实现功能
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		contract := new(Contract_in_company)
		_ = json.Unmarshal(queryResponse.Value, contract)

		queryResult := QueryResult_incompany{Key: queryResponse.Key, Record: contract}
		results = append(results, queryResult)
	}

	return results, nil
}

// 审核，生成放入跨公司区块的文本【需要调用公司间区块的链码ContractSanction_upload上传】
func (s *SmartContract) ContractSanction(ctx contractapi.TransactionContextInterface, company_name string, signature string, ContractName string) (*Contract_among_company, error) {
	contractAsBytes, err := ctx.GetStub().GetState(ContractName)
	//使用getstate获取车辆信息
	if err != nil {
		return nil, fmt.Errorf("Failed to read from world state. %s", err.Error())
	}

	if contractAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", ContractName)
	}

	contract := new(Contract_in_company)
	_ = json.Unmarshal(contractAsBytes, contract)
	//反序列化后读出数据

	//时间戳
	timeUnix := time.Now().Unix() //时间戳
	createTime := time.Unix(timeUnix, 0).Format("2006-01-02 15:04:05")

	contract_among_company := Contract_among_company{
		ContractName:       contract.ContractName,
		ContractContent:    contract.ContractContent,
		CreaterCompanyName: company_name,
		CreaterCompanySign: signature,
		CreateTime:         createTime,
	}
	return &contract_among_company, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create fabcar chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
	}
}
