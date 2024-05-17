package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// DataChaincode ...
type DataChaincode struct {
	contractapi.Contract
}

// DataType ...
type DataType struct {
	Type        string `json:"type"`
	Uploader    string `json:"uploader"`
	Name        string `json:"name"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Downloaded  int    `json:"downloaded"`
	Owner       string `json:"owner"`
	Contents    string `json:"contents"`
	Timestamp   string `json:"timestamp"`
}

type DataCount struct {
	Type  string `json:"type"`
	List  string `json:"list"`
	Count int    `json:"count"`
}

// InitLedger ...
func (d *DataChaincode) InitLedger(ctx contractapi.TransactionContextInterface) error {
	isInitBytes, err := ctx.GetStub().GetState("isInit")
	if err != nil {
		return fmt.Errorf("failed GetState('isInit')")
	} else if isInitBytes == nil {
		initCount := DataCount{
			Type:  "DataCount",
			List:  "",
			Count: 0,
		}
		initDataCountAsBytes, err := json.Marshal(initCount)
		ctx.GetStub().PutState(makeDataCountKey("DC"), initDataCountAsBytes)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
		if err != nil {
			return fmt.Errorf("failed to json.Marshal(). %v", err)
		}
		return nil
	} else {
		return fmt.Errorf("already initialized")
	}
}

// DataInsert ...
func (d *DataChaincode) PutCommonData(ctx contractapi.TransactionContextInterface, uploader string, name string, version string, description string, owner string, contents string, timestamp string) error {
	exists, err := d.dataExists(ctx, uploader, name, version)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the data %s already exists", name)
	}

	download := 0
	dataInfo := DataType{
		Type:        "data",
		Uploader:    uploader,
		Name:        name,
		Version:     version,
		Description: description,
		Downloaded:  download,
		Owner:       owner,
		Contents:    contents,
		Timestamp:   timestamp,
	}
	dataAsBytes, err := json.Marshal(dataInfo)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal(). %v", err)
	}
	dataKey := makeDataKey(uploader, name, version)
	ctx.GetStub().PutState(dataKey, dataAsBytes)
	if err != nil {
		return fmt.Errorf("failed to put to world state. %v", err)
	}

	currentDataCount, err := d.GetCommonDataCount(ctx, "DC")
	if err != nil {
		return fmt.Errorf("failed to get current meow. %v", err)
	}
	currentDataCount.Count++

	currentDataCountAsBytes, err := json.Marshal(currentDataCount)
	if err != nil {
		return fmt.Errorf("failed to json.Marshal(). %v", err)
	}
	ctx.GetStub().PutState(makeDataCountKey("DC"), currentDataCountAsBytes)
	return nil
}

func makeDataCountKey(key string) string {
	var sb strings.Builder

	sb.WriteString("Count_D_")
	sb.WriteString(key)
	return sb.String()
}

func makeDataKey(uploader string, name string, version string) string {
	var sb strings.Builder

	sb.WriteString("D_")
	sb.WriteString(uploader)
	sb.WriteString("_")
	sb.WriteString(name)
	sb.WriteString("_")
	sb.WriteString(version)
	return sb.String()
}

func (d *DataChaincode) dataExists(ctx contractapi.TransactionContextInterface, uploader string, name string, version string) (bool, error) {
	dataAsBytes, err := ctx.GetStub().GetState(makeDataKey(uploader, name, version))
	if err != nil {
		return false, fmt.Errorf("data is exist...: %v", err)
	}

	return dataAsBytes != nil, nil
}

func (d *DataChaincode) GetAllCommonDataInfo(ctx contractapi.TransactionContextInterface) ([]*DataType, error) {
	queryString := fmt.Sprintf(`{"selector":{"type":"data"}}`)
	return getQueryResultForQueryString(ctx, queryString)
}

func (d *DataChaincode) GetCommonDataInfo(ctx contractapi.TransactionContextInterface, uploader string, name string, version string) (*DataType, error) {
	dataInfo := &DataType{}
	dataAsBytes, err := ctx.GetStub().GetState(makeDataKey(uploader, name, version))
	if err != nil {
		return nil, err
	} else if dataAsBytes == nil {
		dataInfo.Type = "empty"
		dataInfo.Uploader = "empty"
		dataInfo.Name = "empty"
		dataInfo.Version = "empty"
		dataInfo.Description = "empty"
		dataInfo.Downloaded = 0
		dataInfo.Owner = "empty"
		dataInfo.Contents = "empty"
		dataInfo.Timestamp = "empty"
	} else {
		err = json.Unmarshal(dataAsBytes, &dataInfo)
		if err != nil {
			return nil, err
		}
	}
	return dataInfo, nil
}

func (d *DataChaincode) GetCommonDataInfoWithKey(ctx contractapi.TransactionContextInterface, dataKey string) (*DataType, error) {
	dataInfo := &DataType{}
	dataAsBytes, err := ctx.GetStub().GetState(dataKey)
	if err != nil {
		return nil, err
	} else if dataAsBytes == nil {
		dataInfo.Type = "empty"
		dataInfo.Uploader = "empty"
		dataInfo.Name = "empty"
		dataInfo.Version = "empty"
		dataInfo.Description = "empty"
		dataInfo.Downloaded = 0
		dataInfo.Owner = "empty"
		dataInfo.Contents = "empty"
		dataInfo.Timestamp = "empty"
	} else {
		err = json.Unmarshal(dataAsBytes, &dataInfo)
		if err != nil {
			return nil, err
		}
	}
	return dataInfo, nil
}

func (d *DataChaincode) GetCommonDataContents(ctx contractapi.TransactionContextInterface, uploader string, name string, version string, downloader string) (string, error) {
	dataInfo, err := d.GetCommonDataInfo(ctx, uploader, name, version)
	if err != nil {
		return "failed to get Info", err
	}
	dataInfo.Downloaded++
	dataAsBytes, err := json.Marshal(dataInfo)
	if err != nil {
		return "failed to json.Marshal().", err
	}
	dataKey := makeDataKey(uploader, name, version)
	ctx.GetStub().PutState(dataKey, dataAsBytes)
	if err != nil {
		return "failed to put to world state.", err
	}

	currentDataCount, err := d.GetCommonDataCount(ctx, downloader)
	if err != nil {
		return "failed to get count", err
	}
	currentDataCount.Count++
	currentDataCount.List += "/" + name

	currentDataCountAsBytes, err := json.Marshal(currentDataCount)
	if err != nil {
		return "failed to json.Marshal()", err
	}
	ctx.GetStub().PutState(makeDataCountKey(downloader), currentDataCountAsBytes)

	return dataInfo.Contents, nil
}

func (d *DataChaincode) GetCommonDataCount(ctx contractapi.TransactionContextInterface, key string) (*DataCount, error) {
	currentDataCount := &DataCount{}
	currentDataCountAsBytes, err := ctx.GetStub().GetState(makeDataCountKey(key))
	if err != nil {
		return nil, err
	} else if currentDataCountAsBytes == nil {
		currentDataCount.Type = "DataCount"
		currentDataCount.List = ""
		currentDataCount.Count = 0
	} else {
		err = json.Unmarshal(currentDataCountAsBytes, currentDataCount)
		if err != nil {
			return nil, err
		}
	}
	return currentDataCount, nil
}

func getQueryResultForQueryString(ctx contractapi.TransactionContextInterface, queryString string) ([]*DataType, error) {
	resultsIterator, err := ctx.GetStub().GetQueryResult(queryString)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var transferHistorys []*DataType
	for resultsIterator.HasNext() {
		queryResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}
		var transferHistory DataType
		err = json.Unmarshal(queryResult.Value, &transferHistory)
		if err != nil {
			return nil, err
		}
		transferHistorys = append(transferHistorys, &transferHistory)
	}

	return transferHistorys, nil
}
