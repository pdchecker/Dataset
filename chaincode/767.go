package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type ExpireResponse struct {
	TxId string `json:"txId"`
}

func (s *SmartContract) ExpireAsset(ctx contractapi.TransactionContextInterface, data string) (*ExpireResponse, error) {

	cc := new(VoidAssetReq)
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

	if cc.ImmutableContract.Contract.SchemaVersion != cc.ImmutableContract.Contract.Definition.SchemaVersion {
		return nil, errors.New("contract schema version does not match with definition")
	}

	if err := cc.ImmutableContract.Contract.Validate(); err != nil {
		return nil, err
	}

	// todo: validate permission of the calling party

	asset, err := s.ReadAsset(ctx, fmt.Sprint(cc.ContractId))
	if err != nil {
		return nil, err
	}

	switch asset.State {
	case ContractStateVoided:
		return nil, errors.New("contract voided, cannot expire")

	case ContractStateExpired:
		return nil, errors.New("contract already expired")

	case ContractStateReleased:
		return nil, errors.New("contract released, cannot expire")

	}

	/*
		todo: Validate that the change of state is allowed as per contract rules.
		For example, if content is to be released, that if using a notary, then the notary is the caller (through a manged process), or if verifiers, then a common key is used.
		If contract is to be voided, that voiding is allowed.
		If to be set as expired, that the expiration data has been reached, etc.
	*/

	t := time.Now().Format(time.RFC3339)
	asset.State = ContractStateExpired
	asset.UpdatedAt = t

	asset.Changes = append(asset.Changes, Change{
		PackageID:   cc.PackageId,
		PackageHash: cc.PackageHash,
		PackageDate: t,
		Action:      "expire",
		NewState:    asset.State,
	})

	contractJSON, err := json.Marshal(asset)
	if err != nil {
		return nil, err
	}

	if err := ctx.GetStub().PutState(fmt.Sprint(cc.ContractId), contractJSON); err != nil {
		return nil, err
	}

	return &ExpireResponse{
		ctx.GetStub().GetTxID(),
	}, nil
}
