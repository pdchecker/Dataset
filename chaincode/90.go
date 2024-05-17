package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"strconv"
)

type SmartContract struct {
	contractapi.Contract
}

//养殖场名称、地址、营业执照、养殖者姓名、养殖时间、出栏日期
type Farm struct {
	FarmName     string `json:"FarmName"`
	Address      string `json:"Address"`
	License      string `json:"License"`
	Owner        string `json:"owner"`
	BreedingTime string `json:"BreedingTime"`
	ReleaseDate  string `json:"ReleaseDate"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Farm
}

// 初始化账本
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	Farms := []Farm{
		Farm{"榆中肉牛养殖场", "甘肃省兰州市", "446523465768", "刘", "2022.2.1", "2022.10.23"},
	}
	for i, farm := range Farms {
		farmAsBytes, _ := json.Marshal(farm)
		err := ctx.GetStub().PutState("Farm"+strconv.Itoa(i), farmAsBytes)

		if err != nil {
			return fmt.Errorf("Failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

// 添加农场
func (s *SmartContract) Create(ctx contractapi.TransactionContextInterface, farmNumber string, farmName string, address string, license string, owner string, breedingTime string, releaseDate string) error {
	farm := Farm{
		FarmName:     farmName,
		Address:      address,
		License:      license,
		Owner:        owner,
		BreedingTime: breedingTime,
		ReleaseDate:  releaseDate,
	}

	farmAsBytes, _ := json.Marshal(farm)

	return ctx.GetStub().PutState(farmNumber, farmAsBytes)
}

// 查询一个农场
func (s *SmartContract) Query(ctx contractapi.TransactionContextInterface, farmNumber string) (*Farm, error) {
	farmAsBytes, err := ctx.GetStub().GetState(farmNumber)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if farmAsBytes == nil {
		return nil, fmt.Errorf("%s does not exist", farmNumber)
	}

	farm := new(Farm)
	_ = json.Unmarshal(farmAsBytes, farm)

	return farm, nil
}

// 查询所有的农场
func (s *SmartContract) QueryAll(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer func(resultsIterator shim.StateQueryIteratorInterface) {
		err := resultsIterator.Close()
		if err != nil {

		}
	}(resultsIterator)

	var results []QueryResult

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		farm := new(Farm)
		_ = json.Unmarshal(queryResponse.Value, farm)

		queryResult := QueryResult{Key: queryResponse.Key, Record: farm}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) Delete(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.FarmExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}
func (s *SmartContract) FarmExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create farm chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting farm chaincode: %s", err.Error())
	}
}
