package main

import (
	"encoding/json"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
)

func BeforeTransaction(ctx contractapi.TransactionContextInterface) error {
	if err := readFromLedger(ctx, kittiesNAME); err != nil {
		if serr, ok := err.(GetStateError); ok {
			log.Println(serr)
		} else {
			return err
		}
	}
	if err := readFromLedger(ctx, kittyIndexToOwnerNAME); err != nil {
		if serr, ok := err.(GetStateError); ok {
			log.Println(serr)
		} else {
			return err
		}
	}
	if err := readFromLedger(ctx, kittyIndexToApprovedNAME); err != nil {
		if serr, ok := err.(GetStateError); ok {
			log.Println(serr)
		} else {
			return err
		}
	}
	if err := readFromLedger(ctx, sireAllowedToAddressNAME); err != nil {
		if serr, ok := err.(GetStateError); ok {
			log.Println(serr)
		} else {
			return err
		}
	}
	if err := readFromLedger(ctx, pregnantKittiesNAME); err != nil {
		if serr, ok := err.(GetStateError); ok {
			log.Println(serr)
		} else {
			return err
		}
	}

	g_event = make(map[string]interface{})

	return nil
}

func AfterTransaction(ctx contractapi.TransactionContextInterface) error {
	if err := writeToLedger(ctx, kittiesNAME); err != nil {
		return err
	}
	if err := writeToLedger(ctx, kittyIndexToOwnerNAME); err != nil {
		return err
	}
	if err := writeToLedger(ctx, kittyIndexToApprovedNAME); err != nil {
		return err
	}
	if err := writeToLedger(ctx, sireAllowedToAddressNAME); err != nil {
		return err
	}
	if err := writeToLedger(ctx, pregnantKittiesNAME); err != nil {
		return err
	}

	if len(g_event) > 0 {
		jsonPayload, err := json.Marshal(g_event)
		if err != nil {
			return err
		}
		if err := ctx.GetStub().SetEvent("HyperledgerEvent", jsonPayload); err != nil {
			return err
		}
	}

	return nil
}
