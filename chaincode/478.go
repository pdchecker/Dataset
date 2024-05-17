package main

import (
    "errors"
    "strconv"
    "fmt"
    "os"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type CcCounter struct {
    contractapi.Contract
}

func getCount(ctx contractapi.TransactionContextInterface, key string) (int, error) {

    var bCount []byte
    var iCount int
    var err error

    bCount, err = ctx.GetStub().GetState(key)

    if err != nil {
        return 0, errors.New("Unable to interact with world state")
    }

    if bCount == nil {
        return 0, errors.New("Unknown key")
    }

    iCount, err = strconv.Atoi(string(bCount))

    if err != nil {
        return 0, errors.New("Internal error")
    }

    return iCount, nil
}

func (cc *CcCounter) InitLedger(ctx contractapi.TransactionContextInterface) error {

    err := ctx.GetStub().PutState("A", []byte("0"))

    if err != nil {
        return errors.New("Internal error")
    }

    return nil

}

func (cc *CcCounter) Get(ctx contractapi.TransactionContextInterface, key string) (string, error) {

    fmt.Fprintf(os.Stderr, "Get (key:%s)\n", key)

    iCount, err := getCount(ctx, key)

    if err != nil {
        return "", err
    }

    return strconv.Itoa(iCount), nil
}

func (cc *CcCounter) Reset(ctx contractapi.TransactionContextInterface, key string) error {

    fmt.Fprintf(os.Stderr, "Reset (key:%s)\n", key)

    err := ctx.GetStub().PutState(key, []byte("0"))

    if err != nil {
        return errors.New("Unable to interact with world state")
    }

    return nil
}

func (cc *CcCounter) Countup(ctx contractapi.TransactionContextInterface, key string) (string, error) {

    fmt.Fprintf(os.Stderr, "Countup (key:%s)\n", key)

    iCount, err := getCount(ctx, key)

    if err != nil {
        return "", err
    }

    iCount += 1

    err = ctx.GetStub().PutState(key, []byte(strconv.Itoa(iCount)))

    if err != nil {
        return "", errors.New("Unable to interact with world state")
    }

    return strconv.Itoa(iCount), nil
}

