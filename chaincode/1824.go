package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric/core/chaincode/shim" //1.4
	"github.com/hyperledger/fabric/protos/peer"         //1.4
	// "github.com/hyperledger/fabric-chaincode-go/shim" //新版
	// "github.com/hyperledger/fabric-protos-go/peer" //新版
)

type PrivacyChaincode struct {
}

type PrivacyData struct {
	Identifier string                 `json:"identifier"` // 唯一标识, client1_to_rsu1_epoch1
	Data       map[string]interface{} `json:"data"`       // 存储数据
}

func (t *PrivacyChaincode) Init(stub shim.ChaincodeStubInterface) peer.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println(function)
	fmt.Println(args[0])
	fmt.Println(" ==== Init ====")
	return shim.Success(nil)
}

// ============================================================================================================================
// Invoke - Our entry point for Invocations
// ============================================================================================================================
func (t *PrivacyChaincode) Invoke(stub shim.ChaincodeStubInterface) peer.Response {
	// 获取用户意图
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	if function == "add" {
		fmt.Println(function + "start")
		return t.add(stub, args)
	} else if function == "query" {
		return t.query(stub, args)
	} else if function == "update" {
		return t.update(stub, args)
	} else if function == "delete" {
		return t.delete(stub, args)
	}

	fmt.Println("invoke did not find func: " + function) //error

	return shim.Error("Received unknown function invocation")
}

func PutPrivacyData(stub shim.ChaincodeStubInterface, privacyData PrivacyData) ([]byte, bool) {

	fmt.Println("序列化数据")
	b, err := json.Marshal(privacyData)
	if err != nil {
		return nil, false
	}

	// 保存数据状态
	fmt.Println("保存数据")
	err = stub.PutState(privacyData.Identifier, b)
	if err != nil {
		return nil, false
	}

	return b, true
}

func GetPrivacyDataInfo(stub shim.ChaincodeStubInterface, Identifier string) (PrivacyData, bool) {
	var privacyData PrivacyData
	// 根据唯一标识查询信息状态
	b, err := stub.GetState(Identifier)
	if err != nil {
		return privacyData, false
	}

	if b == nil {
		return privacyData, false
	}

	// 对查询到的状态进行反序列化
	err = json.Unmarshal(b, &privacyData)
	if err != nil {
		return privacyData, false
	}

	// 返回结果
	return privacyData, true
}

func (t *PrivacyChaincode) add(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	fmt.Println("进入add函数")
	if len(args) != 1 {
		return shim.Error("给定的参数个数不符合要求")
	}

	var privacyData PrivacyData
	fmt.Println("反序列化中")
	err := json.Unmarshal([]byte(args[0]), &privacyData)
	if err != nil {
		return shim.Error("反序列化信息时发生错误")
	}

	fmt.Println("根据Id获取值")
	result, exist := GetPrivacyDataInfo(stub, privacyData.Identifier)
	if exist {
		fmt.Println("需要添加的ID已经存在, 更新该ID值")
		// return t.update(stub, args)
		result.Identifier = privacyData.Identifier
		result.Data = privacyData.Data
		_, bl := PutPrivacyData(stub, result)
		if !bl {
			return shim.Error("保存信息时发生错误")
		}
		fmt.Println("发送事件")
		err = stub.SetEvent("added", []byte{})
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("信息更新成功"))
	} else {
		fmt.Println("添加数据")
		_, bl := PutPrivacyData(stub, privacyData)
		if !bl {
			return shim.Error("保存信息时发生错误")
		}
		fmt.Println("发送事件")
		err = stub.SetEvent("added", []byte{})
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success([]byte("信息添加成功"))
	}
}

// 根据Identifier查询详情
func (t *PrivacyChaincode) query(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("给定的参数个数不符合要求")
	}

	b, err := stub.GetState(args[0])
	if err != nil {
		return shim.Error("根据Identifier查询信息失败")
	}

	if b == nil {
		return shim.Error("根据Identifier没有查询到相关的信息")
	}

	// 对查询到的状态进行反序列化
	var privacyData PrivacyData
	err = json.Unmarshal(b, &privacyData)
	if err != nil {
		return shim.Error("反序列化privacyData信息失败")
	}

	// 返回
	result, err := json.Marshal(privacyData)
	if err != nil {
		return shim.Error("序列化privacyData信息时发生错误")
	}
	return shim.Success(result)
}

// 根据Identifier更新信息
func (t *PrivacyChaincode) update(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("给定的参数个数不符合要求")
	}

	var privacyData PrivacyData
	err := json.Unmarshal([]byte(args[0]), &privacyData)
	if err != nil {
		return shim.Error("反序列化privacyData信息失败")
	}

	// 根据Identifier查询信息
	result, bl := GetPrivacyDataInfo(stub, privacyData.Identifier)
	if !bl {
		return shim.Error("根据Identifier查询信息时发生错误")
	}

	result.Identifier = privacyData.Identifier
	result.Data = privacyData.Data

	_, bl = PutPrivacyData(stub, result)
	if !bl {
		return shim.Error("保存信息信息时发生错误")
	}

	err = stub.SetEvent("updated", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("信息更新成功"))
}

// 根据Identifier删除信息
func (t *PrivacyChaincode) delete(stub shim.ChaincodeStubInterface, args []string) peer.Response {
	if len(args) != 1 {
		return shim.Error("给定的参数个数不符合要求")
	}

	err := stub.DelState(args[0])
	if err != nil {
		return shim.Error("删除信息时发生错误")
	}

	err = stub.SetEvent("deleted", []byte{})
	if err != nil {
		return shim.Error(err.Error())
	}

	return shim.Success([]byte("信息删除成功"))
}

func main() {
	err := shim.Start(new(PrivacyChaincode))
	if err != nil {
		fmt.Printf("启动PrivacyChaincode时发生错误: %s", err)
	}
}
