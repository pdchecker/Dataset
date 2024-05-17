package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SimpleContract struct {
	contractapi.Contract
}

type Status string

const (
	ACTIVE   Status = "ACTIVE"
	INACTIVE Status = "INACTIVE"
)

type Asset struct {
	ID        string `json:"ID"`
	Owner     string `json:"Owner"`
	Status    Status `json:"Active"`
	Value     int    `json:"Value"`
	UpdatedAt string `json:"UpdatedAt"`
	CreatedAt string `json:"CreatedAt"`
	DeletedAt string `json:"DeletedAt"`
}

func (sc *SimpleContract) InitLedger(ctx contractapi.TransactionContextInterface) (err error) {

	assets := [5]Asset{
		{}, {}, {}, {}, {},
	}

	for _, asset := range assets {
		assetJson, err := json.Marshal(asset)
		if err != nil {
			continue
		}

		if err = ctx.GetStub().PutState(asset.ID, assetJson); err != nil {
			continue
		}
	}

	return
}

func (sc *SimpleContract) Create(ctx contractapi.TransactionContextInterface, id, owner string, value int) (err error) {

	_, exists := sc.exists(ctx, id)
	if !exists {
		return fmt.Errorf("cannot create duplicate asset with id %s", id)
	}

	asset := Asset{
		ID:        id,
		Owner:     owner,
		Value:     value,
		Status:    ACTIVE,
		CreatedAt: time.Now().String(),
		UpdatedAt: time.Now().String(),
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return
	}

	err = ctx.GetStub().PutState(id, assetJSON)

	return
}

func (sc *SimpleContract) Query(ctx contractapi.TransactionContextInterface, id string) (asset *Asset, err error) {

	assetJSON, exists := sc.exists(ctx, id)
	if !exists {
		return
	}

	err = json.Unmarshal(assetJSON, asset)

	return
}

func (sc *SimpleContract) Update(ctx contractapi.TransactionContextInterface, id, owner string) (err error) {

	assetJSON, exists := sc.exists(ctx, id)
	if !exists {
		return
	}

	asset := Asset{}
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return
	}

	if strings.EqualFold(asset.Owner, owner) {
		return
	}

	asset.Owner = owner
	asset.UpdatedAt = time.Now().String()

	assetJSON, err = json.Marshal(asset)
	if err != nil {
		return
	}

	err = ctx.GetStub().PutState(id, assetJSON)

	return
}

func (sc *SimpleContract) Delete(ctx contractapi.TransactionContextInterface, id string) (err error) {

	assetJSON, exists := sc.exists(ctx, id)
	if !exists {
		return
	}

	asset := Asset{}
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return
	}

	if strings.EqualFold(string(asset.Status), string(INACTIVE)) {
		return
	}

	asset.Status = INACTIVE
	asset.UpdatedAt = time.Now().String()
	asset.DeletedAt = time.Now().String()

	assetJSON, err = json.Marshal(asset)
	if err != nil {
		return
	}

	err = ctx.GetStub().PutState(id, assetJSON)

	return
}

func (sc *SimpleContract) exists(ctx contractapi.TransactionContextInterface, id string) (assetJSON []byte, exists bool) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return
	}

	return assetJSON, assetJSON != nil
}
