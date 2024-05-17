package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type CreateAssetResponse struct {
	ContractId int64  `json:"contractId"`
	TxId       string `json:"txId"`
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, data string) (*CreateAssetResponse, error) {

	cc := new(NewAssetReq)
	if err := ParseRequest(data, cc); err != nil {
		return nil, err
	}

	icHash, err := JsonHashS256(cc.ImmutableContract)
	if err != nil {
		return nil, err
	}

	if icHash != cc.ImmutableContractHash {
		return nil, errors.New("invalid immutable contract hash")
	}

	// todo: Check the schema version of the immutable contract, as well as schema version of definition, to see if any routing to different validators is needed.

	if err := cc.ImmutableContract.Contract.Validate(); err != nil {
		return nil, err
	}

	t := time.Now().Format(time.RFC3339)

	contract := Contract{
		ContractHash: cc.ImmutableContractHash,
		CreatedAt:    t,
		UpdatedAt:    t,
		State:        ContractStateActive,
		Version:      cc.ImmutableContract.Contract.SchemaVersion,
		Changes:      []Change{},
	}

	contract.ContractId = cc.ImmutableContract.Contract.ContractID
	contractIdStr := fmt.Sprint(contract.ContractId)

	exists, err := s.AssetExists(ctx, contractIdStr)
	if err != nil {
		return nil, err
	}

	if exists {
		return nil, fmt.Errorf("the contract %d already exists", contract.ContractId)
	}

	contractJSON, err := json.Marshal(contract)
	if err != nil {
		return nil, err
	}

	if err := ctx.GetStub().PutState(contractIdStr, contractJSON); err != nil {
		return nil, err
	}

	log.Println("contract instantiated:", contract.ContractId, ctx.GetStub().GetTxID())

	return &CreateAssetResponse{
		contract.ContractId,
		ctx.GetStub().GetTxID(),
	}, nil
}
