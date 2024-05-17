package main

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	mspprotos "github.com/hyperledger/fabric-protos-go/msp"
)

func (t *SimpleChaincode) getCurrentNettingCycle(
	ctx contractapi.TransactionContextInterface) (*NettingCycle, error) {

	nettingCycle := &NettingCycle{}
	nettingCycleAsBytes, err := ctx.GetStub().GetState(nettingCycleObjectType)
	if err != nil {
		return nettingCycle, errors.New("Error: Failed to get state for current nettingcycle")
	} else if nettingCycleAsBytes == nil {
		return nettingCycle, errors.New("Error: nettingcycle does not exist")
	}

	err = json.Unmarshal([]byte(nettingCycleAsBytes), nettingCycle)
	if err != nil {
		return nettingCycle, err
	}
	return nettingCycle, nil
}

func (t *SimpleChaincode) getSigner(
	ctx contractapi.TransactionContextInterface) (string, error) {

	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return "", err
	}
	id := &mspprotos.SerializedIdentity{}
	err = proto.Unmarshal(creator, id)
	if err != nil {
		return "", err
	}
	block, _ := pem.Decode(id.GetIdBytes())
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return "", err
	}
	// mspID := id.GetMspid() // if you need the mspID
	signer := cert.Subject.CommonName
	return signer, nil
}
func (t *SimpleChaincode) verifyIdentity(
	ctx contractapi.TransactionContextInterface,
	identity string) error {

	creatorString, err := t.getSigner(ctx)
	if err != nil {
		return err
	}
	if creatorString != identity {
		errMsg := fmt.Sprintf(
			"Error: Identity of creator (%s) does not match %s",
			creatorString,
			identity)
		return errors.New(errMsg)
	}
	return nil
}

func (t *SimpleChaincode) getTxTimeStampAsTime(
	ctx contractapi.TransactionContextInterface) (time.Time, error) {

	timestampTime := time.Time{}
	timestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {
		return timestampTime, err
	}
	timestampTime, err = ptypes.Timestamp(timestamp)
	if err != nil {
		return timestampTime, err
	}

	return timestampTime, nil
}

func (t *SimpleChaincode) convertStringToArrayOfStrings(
	inputString string) ([]string, error) {

	var stringArray []string
	dec := json.NewDecoder(strings.NewReader(inputString))
	err := dec.Decode(&stringArray)

	return stringArray, err
}

func (t *SimpleChaincode) checkArgArrayLength(
	args []string,
	expectedArgLength int) error {

	argArrayLength := len(args)
	if argArrayLength != expectedArgLength {
		errMsg := fmt.Sprintf(
			"Incorrect number of arguments: Received %d, expecting %d",
			argArrayLength,
			expectedArgLength)
		return errors.New(errMsg)
	}
	return nil
}
