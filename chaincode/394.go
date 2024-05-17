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

// Crops 作物信息
type Crops struct {
	//作物ID
	CropsId string `json:"crops_id"`
	//作物名称
	CropsName string `json:"crops_name"`
	//所在地
	Address string `json:"address"`
	//生长开始日期
	RegisterTime string `json:"register_time"`
	//年度
	Year string `json:"year"`
	//农户名字
	FarmerName string `json:"farmer_name"`
	//农户ID
	FarmerID string `json:"farmer_id"`
	//联系电话
	FarmerTel string `json:"farmer_tel"`
	//肥料名称
	FertilizerName string `json:"fertilizer_name"`
	//种植方式
	PlatMode string `json:"plant_mode"`
	//是否套袋种植
	BaggingStatus string `json:"bagging_status"`
	//育苗周期
	GrowSeedlingsCycle string `json:"grow_seedlings_cycle"`
	//灌溉周期
	IrrigationCycle string `json:"irrigation_cycle"`
	//施肥周期
	ApplyFertilizerCycle string `json:"apply_fertilizer_cycle"`
	//除草周期
	WeedCycle string `json:"weed_cycle"`
	//备注
	Remarks string `json:"remarks"`
}

// CropsGrowInfo 作物生长信息
type CropsGrowInfo struct {
	//生长情况唯一ID
	CropsGrowId string `json:"crops_grow_id"`
	//作物ID
	CropsBakId string `json:"crops_bak_id"`
	//记录时间
	RecordTime string `json:"record_time"`
	//作物生长图片URL
	CropsGrowPhotoUrl string `json:"crops_grow_photo_url"`
	//温度
	Temperature string `json:"temperature"`
	//生长情况
	GrowStatus string `json:"grow_status"`
	//水分
	WaterContent string `json:"water_content"`
	//光照情况
	IlluminationStatus string `json:"illumination_status"`
	//备注
	Remarks string `json:"remarks"`
}

// CreateCrops 新增作物记录
func (s *SmartContract) CreateCrops(ctx contractapi.TransactionContextInterface, CropsId, CropsName, Address, RegisterTime, Year, FarmerName, FarmerID, FarmerTel, FertilizerName, PlatMode, BaggingStatus, GrowSeedlingsCycle, IrrigationCycle, ApplyFertilizerCycle, WeedCycle, Remarks string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	ok, err := IsExist(ctx, CropsId)
	if err != nil {
		resp.Code = "106"
		resp.Msg = err.Error()
		return &resp
	}
	if ok {
		resp.Code = "300"
		resp.Msg = fmt.Sprintf("crops_id already exist")
		return &resp
	}
	var crops = Crops{CropsId, CropsName, Address, RegisterTime, Year, FarmerName, FarmerID, FarmerTel, FertilizerName, PlatMode, BaggingStatus, GrowSeedlingsCycle, IrrigationCycle, ApplyFertilizerCycle, WeedCycle, Remarks}
	err = WriteLedger(ctx, crops, CropsId)
	if err != nil {
		resp.Code = "101"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// RecordCropsGrow 作物生长过程记录
func (s *SmartContract) RecordCropsGrow(ctx contractapi.TransactionContextInterface, CropsGrowId, CropsBakId, RecordTime, CropsGrowPhotoUrl, Temperature, GrowStatus, WaterContent, IlluminationStatus, Remarks string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	ok, err := IsExist(ctx, CropsGrowId)
	if err != nil {
		resp.Code = "106"
		resp.Msg = err.Error()
		return &resp
	}
	if ok {
		resp.Code = "300"
		resp.Msg = fmt.Sprintf("crops_grow_id already exist")
		return &resp
	}
	var cropsGrowInfo = CropsGrowInfo{CropsGrowId, CropsBakId, RecordTime, CropsGrowPhotoUrl, Temperature, GrowStatus, WaterContent, IlluminationStatus, Remarks}
	err = WriteLedger(ctx, cropsGrowInfo, CropsGrowId)
	if err != nil {
		resp.Code = "101"
		resp.Msg = err.Error()
		return &resp
	}
	return &resp
}

// QueryCropsById 根据作物ID查询作物信息
func (s *SmartContract) QueryCropsById(ctx contractapi.TransactionContextInterface, CropsID string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	results, err := GetState0(ctx, CropsID)
	if err != nil {
		resp.Code = "102"
		resp.Msg = err.Error()
		return &resp
	}
	resp.Data = results
	return &resp
}

// QueryCropsProcessByCropsId 根据CropsID溯源所有生长记录过程
func (s *SmartContract) QueryCropsProcessByCropsId(ctx contractapi.TransactionContextInterface, CropsBakId string) *Response {
	resp := Response{Code: "200", Msg: "OK"}
	results, err := GetState1(ctx, "crops_bak_id", CropsBakId)
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
