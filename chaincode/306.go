package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Asset struct {
	ID       		string		`json:"ID"`
	Temperature		[]int 		`json:"Temperature"`
	Humidity		[]int 		`json:"Humidity"`
	Date        	[]string 	`json:"Date"`
	PM				[]int		`json:"PM"`
	CO2				[]int		`json:"CO2"`
	VOC				[]int		`json:"VOC"`
	ThermostatOn 	int			`json:"thermostaton"`
	HumidifierOn 	int			`json:"humidifieron"`
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, temperature int, humidity int, date string, pm int, co2 int, voc int, thermostaton int, humidifieron int) error {
    exists, err := s.AssetExists(ctx, id)

	if err != nil {
		return err
	}

	if exists {
		assetJSON, err := ctx.GetStub().GetState(id)
		var asset Asset
		err = json.Unmarshal(assetJSON, &asset)

		if err != nil {
			return err
		}

		asset.Temperature = append(asset.Temperature, temperature)
    	asset.Humidity = append(asset.Humidity, humidity)
    	asset.Date = append(asset.Date, date)
		asset.PM = append(asset.PM, pm)
		asset.CO2 = append(asset.CO2,co2)
		asset.VOC = append(asset.VOC,voc)	

		asset.ThermostatOn = thermostaton
		asset.HumidifierOn = humidifieron

		updatedAssetJSON, err := json.Marshal(asset)
    	if err != nil {
        	return err
    	}

    	return ctx.GetStub().PutState(id, updatedAssetJSON)
	}else{
		asset := Asset{
			ID:             id,
			Temperature: 	[]int{temperature},
            Humidity:    	[]int{humidity},
            Date:        	[]string{date},
			PM:				[]int{pm},
			CO2:			[]int{co2},
			VOC:			[]int{voc},
			ThermostatOn: 	thermostaton,
			HumidifierOn:	humidifieron,
		}
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}
	
		return ctx.GetStub().PutState(id, assetJSON)
	}
}



func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
