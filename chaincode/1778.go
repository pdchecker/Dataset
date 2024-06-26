package main

import (
	"errors"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	"github.com/hyperledger-labs/cckit/router"
	"github.com/hyperledger-labs/cckit/serialize"
)

var ErrInvokeMethodNotAllowed = errors.New(`invoke method not allowed`)

type (
	ChaincodeLocatorResolver func(ctx router.Context, serviceName string) (*ChaincodeLocator, error)
	// ChaincodeStubInvoker for cross chaincode calls - only query (read) methods
	// context argument is router.Context, not context.Context
	ChaincodeStubInvoker interface {
		Query(stub shim.ChaincodeStubInterface, fn string, args []interface{}, target interface{}) (interface{}, error)
	}

	LocatorChaincodeStubInvoker struct {
		Locator    *ChaincodeLocator
		Serializer serialize.Serializer
	}
)

func (c *LocatorChaincodeStubInvoker) Query(
	stub shim.ChaincodeStubInterface, fn string, args []interface{}, target interface{}) (interface{}, error) {

	// todo: remove hack
	if c.Serializer == nil {
		c.Serializer = serialize.DefaultSerializer
	}
	argsBytes, err := invokerArgs(fn, args, c.Serializer)
	if err != nil {
		return nil, fmt.Errorf(`query via stub: %w`, err)
	}

	// if target chaincode is encrypted we can only access to target chaincode via dummy `StateGet`,
	// without method name and arg decryption on target chaincode side
	// because transient data cannot be passed via stub.InvokeChaincode call

	response := stub.InvokeChaincode(c.Locator.Chaincode, argsBytes, c.Locator.Channel)
	if response.Status != shim.OK {
		return nil, fmt.Errorf(`cross chaincode=%s, channel=%s invoke: %w`,
			c.Locator.Chaincode, c.Locator.Channel, errors.New(response.Message))
	}

	return ccOutput(&response, target, c.Serializer)
}
