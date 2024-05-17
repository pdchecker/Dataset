package main

import (
	"encoding/json"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type GetStateError struct {
	err error
}

func (e GetStateError) Error() string {
	return e.err.Error()
}

type JsonUnmarshalError struct {
	err error
}

func (e JsonUnmarshalError) Error() string {
	return e.err.Error()
}

func readFromLedger(ctx contractapi.TransactionContextInterface, key string) error {
	log.Println("Entering readFromLedger with key: " + key + "...")
	assetJSON, err := ctx.GetStub().GetState(key)
	if err != nil || len(assetJSON) == 0 {
		return GetStateError{err}
	}
	log.Println(string(assetJSON))

	if key == kittyIndexToOwnerNAME {
		err = json.Unmarshal(assetJSON, &kittyIndexToOwner)
	}
	if key == kittyIndexToApprovedNAME {
		err = json.Unmarshal(assetJSON, &kittyIndexToApproved)
	}
	if key == sireAllowedToAddressNAME {
		err = json.Unmarshal(assetJSON, &sireAllowedToAddress)
	}
	if key == kittiesNAME {
		err = json.Unmarshal(assetJSON, &kitties)
	}
	if key == pregnantKittiesNAME {
		err = json.Unmarshal(assetJSON, &pregnantKitties)
	}
	if err != nil {
		return JsonUnmarshalError{err}
	}

	return nil
}

func writeToLedger(ctx contractapi.TransactionContextInterface, key string) error {
	log.Println("Entering writeToLedger with key" + key + "...")
	var assetJSON []byte
	var err error

	if key == kittyIndexToOwnerNAME {
		assetJSON, err = json.Marshal(kittyIndexToOwner)
	}
	if key == kittyIndexToApprovedNAME {
		assetJSON, err = json.Marshal(kittyIndexToApproved)
	}
	if key == sireAllowedToAddressNAME {
		assetJSON, err = json.Marshal(sireAllowedToAddress)
	}
	if key == kittiesNAME {
		assetJSON, err = json.Marshal(kitties)
	}
	if key == pregnantKittiesNAME {
		assetJSON, err = json.Marshal(pregnantKitties)
	}
	if err != nil {
		return err
	}

	log.Println(string(assetJSON))
	if err = ctx.GetStub().PutState(key, assetJSON); err != nil {
		return err
	}

	return nil
}
