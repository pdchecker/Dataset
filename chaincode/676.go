package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type ProviderInfo struct {
	ProviderID   string `json:"provider_id"`
	ProviderName string `json:"provider_name"`
	Address      string `json:"address"`
}

func (s *SmartContract) UpdateProviderInfo(ctx contractapi.TransactionContextInterface, providerName string, address string) error {
	err := s.IsProvider(ctx)
	if err != nil {
		return err
	}

	clientID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	providerCompositeKey, err := ctx.GetStub().CreateCompositeKey("provider", []string{clientID})
	if err != nil {
		return err
	}

	providerInfoJSON, err := ctx.GetStub().GetState(providerCompositeKey)
	if err != nil {
		return err
	}

	providerInfo := new(ProviderInfo)
	if providerInfoJSON != nil {
		err = json.Unmarshal(providerInfoJSON, providerInfo)
		if err != nil {
			return err
		}
	} else {
		providerInfo.ProviderID = clientID
	}

	shouldUpdate := false
	if providerName != "" {
		providerInfo.ProviderName = providerName
		shouldUpdate = true
	}

	if address != "" {
		providerInfo.Address = address
		shouldUpdate = true
	}

	if !shouldUpdate {
		return nil
	}

	providerInfoJSON, err = json.Marshal(providerInfo)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(providerCompositeKey, providerInfoJSON)
	if err != nil {
		return err
	}

	return nil
}
func (s *SmartContract) GetProviderInfoByProviderID(ctx contractapi.TransactionContextInterface, providerID string) (*ProviderInfo, error) {
	err := s.IsIntegrator(ctx)
	if err != nil {
		return nil, err
	}

	providerCompositeKey, err := ctx.GetStub().CreateCompositeKey("provider", []string{providerID})
	if err != nil {
		return nil, err
	}

	providerInfoJSON, err := ctx.GetStub().GetState(providerCompositeKey)
	if err != nil {
		return nil, err
	}

	if providerInfoJSON == nil {
		return nil, fmt.Errorf("provider info is not yet initilized")
	}

	providerInfo := new(ProviderInfo)
	err = json.Unmarshal(providerInfoJSON, providerInfo)
	if err != nil {
		return nil, err
	}

	return providerInfo, nil
}

func (s *SmartContract) GetAllProviders(ctx contractapi.TransactionContextInterface) ([]*ProviderInfo, error) {
	err := s.IsIntegrator(ctx)
	if err != nil {
		return nil, err
	}

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("provider", []string{})
	if err != nil {
		return nil, err
	}

	providerInfos := []*ProviderInfo{}
	for resultsIterator.HasNext() {
		providerInfoJSON, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		providerInfo := new(ProviderInfo)
		err = json.Unmarshal(providerInfoJSON.Value, providerInfo)
		if err != nil {
			return nil, err
		}

		providerInfos = append(providerInfos, providerInfo)
	}

	return providerInfos, nil
}

func (s *SmartContract) GetMyProviderInfo(ctx contractapi.TransactionContextInterface) (*ProviderInfo, error) {
	err := s.IsProvider(ctx)
	if err != nil {
		return nil, err
	}

	clientID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, err
	}

	providerCompositeKey, err := ctx.GetStub().CreateCompositeKey("provider", []string{clientID})
	if err != nil {
		return nil, err
	}

	providerInfoJSON, err := ctx.GetStub().GetState(providerCompositeKey)
	if err != nil {
		return nil, err
	}

	if providerInfoJSON == nil {
		return nil, fmt.Errorf("provider info is not yet initilized")
	}

	providerInfo := new(ProviderInfo)
	err = json.Unmarshal(providerInfoJSON, providerInfo)
	if err != nil {
		return nil, err
	}

	return providerInfo, nil
}
