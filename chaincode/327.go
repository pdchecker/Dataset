package main

import (
	"encoding/base64"
	"os"
	"strings"
	"time"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const (
	// OK constant - status code less than 400, endorser will endorse it.
	// OK means init or invoke successfully.
	OK = 200

	// ERRORTHRESHOLD constant - status code greater than or equal to 400 will be considered an error and rejected by endorser.
	ERRORTHRESHOLD = 400

	// ERROR constant - default error value
	ERROR = 500
)

//GetTimestamp ...
func GetTimestamp(ctx contractapi.TransactionContextInterface) (created time.Time, err error) {
	epochTime, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return
	}

	created = time.Unix(epochTime.GetSeconds(), 0)
	return
}

//GetCallerID ...
func GetCallerID(ctx contractapi.TransactionContextInterface) (string, error) {
	if os.Getenv("MODE") == "TEST" {
		return "Test Caller", nil
	}

	callerIDBase64, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", err
	}

	callerID, err := base64.StdEncoding.DecodeString(callerIDBase64)
	if err != nil {
		return "", err
	}

	callerIDList := strings.Split(string(callerID), "::")
	return callerIDList[1], nil
}
