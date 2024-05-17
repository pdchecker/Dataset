/**
 *
 * Copyright (c) 2022, Oracle and/or its affiliates. All rights reserved.
 *
 */
package main

import (
	"testing"
	"sample.com/FiatMoneyToken/lib/trxcontext"
	"sample.com/FiatMoneyToken/lib/chaincode/chaincodetest"
	"github.com/hyperledger/fabric-chaincode-go/shimtest"
)

func TestControllerMethods(t *testing.T) {
	mockchaincode := new(chaincodetest.MockChainCode)
	mockStub := shimtest.NewMockStub("Test Stub", mockchaincode)

	/**
	 * t - testing interface
	 * Mockstub - Mock of shim's ChaincodeStubInterface for unit testing chaincode, provided by package shimtest
	 * MockChainCode- Mock Implementation of ChainCode Interface. defined in package lib/chaincode/chaincodetest
	 * controller - Instance of the conroller reciever on which methods to be tested are defined.

	 * You can test your controller methods like below by passing the required arguments
	 * t.Run("Testing CreateSupplier Function", func(t *testing.T) {
	 *	mockStub.MockTransactionStart("Txid1")
	 *	controller := new(Controller)
	 *	controller.Ctx = trxcontext.GetNewCtx(mockStub)
	 *		byt := []byte(`{"SupplierId":"s02","RawMaterialAvailable":5,"License":"valid supplier","ExpiryDate":"2020-05-30","Active":true}`)
	 *		var obj Supplier
	 *		if err := json.Unmarshal(byt, &obj); err != nil {
	 *			panic(err)
	 *		}
	 *		res, err := controller.CreateSupplier(obj)
	 *		if err != nil {
	 *			t.Errorf("CreateSupplier failed. Error %s \n", err.Error())
	 *		}
	 *		t.Logf("CreateSupplier success. Result %v \n", res)
	 *	})
	 *
	 *
	 * methodName - provide the method name as it is in your controller.
	 * arguments - should be as required by the controller method.
	 */

	//Substitute methodName with the required method's name to test successfully
	t.Run("test method: methodName", func(t *testing.T) {
		mockStub.MockTransactionStart("Txid1")
		controller := new(Controller)
		controller.Ctx = trxcontext.GetNewCtx(mockStub)
				
		//Create proper arguments here and pass in the call below.

		res, err := controller.methodName("method arguments")
		if err != nil {
			t.Errorf("methodName fail. Error %s \n", err.Error())
			t.FailNow()
		}
		t.Logf("methodName success. Result: %v \n", res)
	})
}
