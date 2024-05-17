/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package main

import (
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

/* -------------------------------------------------------------------------- */
/*                             User Read Methods                              */
/* -------------------------------------------------------------------------- */

func ReadUserProfile(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting ReadUserProfile")

	// We expect 1 argument: the user ID.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	// Parsing the user ID.
	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse User ID: " + err.Error())
	}

	// Attempt to retrieve the user profile from the state using the user ID.
	userProfileAsBytes, err := stub.GetState("User_" + strconv.FormatInt(userID, 10))
	if err != nil {
		return shim.Error("Error accessing state: " + err.Error())
	}
	if userProfileAsBytes == nil {
		return shim.Error("User with ID " + strconv.FormatInt(userID, 10) + " does not exist.")
	}

	fmt.Println("- end ReadUserProfile")
	return shim.Success(userProfileAsBytes)
}

func ReadPlatformContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting ReadPlatformContract")

	// We expect 1 argument: the user ID.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	// Parsing the user ID.
	userID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse User ID: " + err.Error())
	}

	// Attempt to retrieve the platform contract from the state using the user ID.
	platformContractAsBytes, err := stub.GetState("PlatformContract_" + strconv.FormatInt(userID, 10))
	if err != nil {
		return shim.Error("Error accessing state: " + err.Error())
	}
	if platformContractAsBytes == nil {
		return shim.Error("Platform Contract for User with ID " + strconv.FormatInt(userID, 10) + " does not exist.")
	}

	fmt.Println("- end ReadPlatformContract")
	return shim.Success(platformContractAsBytes)
}

/* -------------------------------------------------------------------------- */
/*                           Payment Read Methods                             */
/* -------------------------------------------------------------------------- */

func ReadPayment(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting ReadPayment")

	// We expect 1 argument: the payment ID.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	// Retrieve the payment ID from the arguments.
	paymentID := args[0]

	// Attempt to retrieve the payment from the state using the payment ID.
	paymentAsBytes, err := stub.GetState("Payment_" + paymentID)
	if err != nil {
		return shim.Error("Error accessing state: " + err.Error())
	}
	if paymentAsBytes == nil {
		return shim.Error("Payment with ID " + paymentID + " does not exist.")
	}

	fmt.Println("- end ReadPayment")
	return shim.Success(paymentAsBytes)
}

func ReadPaymentDetail(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting ReadPaymentDetail")

	// We expect 1 argument: the ID of the PaymentDetail to retrieve.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	// Parsing ID.
	paymentDetailID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse PaymentDetail ID: " + err.Error())
	}

	// Retrieve the paymentDetail from state.
	paymentDetailAsBytes, err := stub.GetState("PaymentDetail_" + strconv.FormatInt(paymentDetailID, 10))
	if err != nil {
		return shim.Error("Failed to fetch PaymentDetail with ID " + args[0] + " from the ledger: " + err.Error())
	}

	if paymentDetailAsBytes == nil {
		return shim.Error("PaymentDetail with ID " + args[0] + " not found.")
	}

	fmt.Println("- end ReadPaymentDetail")
	return shim.Success(paymentDetailAsBytes)
}

/* -------------------------------------------------------------------------- */
/*                          Energy Bid Read Methods                           */
/* -------------------------------------------------------------------------- */

func ReadOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting ReadOrder")

	// We expect 1 argument: the ID of the Order to retrieve.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	// Parsing ID.
	orderID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse Order ID: " + err.Error())
	}

	// Retrieve the order from state.
	orderAsBytes, err := stub.GetState("Order_" + strconv.FormatInt(orderID, 10))
	if err != nil {
		return shim.Error("Failed to fetch Order with ID " + args[0] + " from the ledger: " + err.Error())
	}

	if orderAsBytes == nil {
		return shim.Error("Order with ID " + args[0] + " not found.")
	}

	fmt.Println("- end ReadOrder")
	return shim.Success(orderAsBytes)
}

func ReadBidMatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting ReadBidMatch")

	// We expect 1 argument: the ID of the BidMatch to retrieve.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	// Parsing ID.
	bidMatchID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BidMatch ID: " + err.Error())
	}

	// Retrieve the bidMatch from state.
	bidMatchAsBytes, err := stub.GetState("BidMatch_" + strconv.FormatInt(bidMatchID, 10))
	if err != nil {
		return shim.Error("Failed to fetch BidMatch with ID " + args[0] + " from the ledger: " + err.Error())
	}

	if bidMatchAsBytes == nil {
		return shim.Error("BidMatch with ID " + args[0] + " not found.")
	}

	fmt.Println("- end ReadBidMatch")
	return shim.Success(bidMatchAsBytes)
}
