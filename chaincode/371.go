package main

import "github.com/hyperledger/fabric-contract-api-go/contractapi"

func main() {

	// simple chaincode
	simpleContract := new(SimpleContract)
	simpleContract.TransactionContextHandler = new(CustomTransactionContext)
	simpleContract.BeforeTransaction = GetWorldState
	simpleContract.UnknownTransaction = UnknownTransactionHandler
	simpleContract.Name = "org.chyidl.com.SimpleContract"

	// complex chaincode
	complexContract := new(ComplexContract)
	complexContract.TransactionContextHandler = new(CustomTransactionContext)
	complexContract.BeforeTransaction = GetWorldState
	complexContract.Name = "org.chyidl.com.ComplexContract"

	cc, err := contractapi.NewChaincode(simpleContract, complexContract)

	if err != nil {
		panic(err.Error())
	}

	cc.DefaultContract = complexContract.GetName()

	if err := cc.Start(); err != nil {
		panic(err.Error())
	}
}
