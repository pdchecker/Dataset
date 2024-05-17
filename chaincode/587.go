package main

import (
    "sample/counter"
    "fmt"

    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

func main() {
    ccCounter := new(counter.CcCounter)

    cc, err := contractapi.NewChaincode(ccCounter)

    fmt.Printf("Start")

    if err != nil {
        fmt.Printf("Error1")
        panic(err.Error())
    }

    if err := cc.Start(); err != nil {
        fmt.Printf("Error2")
        panic(err.Error())
    }
}
