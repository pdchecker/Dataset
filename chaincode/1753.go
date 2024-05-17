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
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	//	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// ============================================================================================================================
// write() - genric write variable into ledger
//
// Shows Off PutState() - writting a key/value into the ledger
//
// Inputs - Array of strings
//    0   ,    1
//   key  ,  value
//  "abc" , "test"
// ============================================================================================================================
func Write(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	var key, value string
	var err error
	fmt.Println("starting write")

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2. key of the variable and value to set")
	}

	// input sanitation
	err = sanitize_arguments(args)
	if err != nil {
		return shim.Error(err.Error())
	}

	key = args[0] //rename for funsies
	value = args[1]
	err = stub.PutState(key, []byte(value)) //write the variable into the ledger
	if err != nil {
		return shim.Error(err.Error())
	}

	fmt.Println("- end write")
	return shim.Success(nil)
}

/* -------------------------------------------------------------------------- */
/*                               Helper Methods                               */
/* -------------------------------------------------------------------------- */

func parseUserCategory(categoryStr string) (UserCategory, error) {
	if val, ok := userCategoryMap[categoryStr]; ok {
		return val, nil
	}
	return 0, errors.New("unknown user category")
}

func parseEnergySource(sourceStr string) (EnergySource, error) {
	if val, ok := energySourceMap[sourceStr]; ok {
		return val, nil
	}
	return 0, errors.New("unknown energy source")
}

/* -------------------------------------------------------------------------- */
/*                             User Write Methods                             */
/* -------------------------------------------------------------------------- */

func UpdateUserProfile(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting UpdateUserProfile")

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	err := sanitize_arguments(args)
	if err != nil {
		return shim.Error("Invalid argument: " + err.Error())
	}

	userID := args[0]
	existingUserAsBytes, err := stub.GetState(userID)

	var user User
	if err != nil || existingUserAsBytes == nil {
		// New user creation
		user.CreatedOn = time.Now().Unix()
		user.UpdatedOn = user.CreatedOn
	} else {
		// Existing user update
		err = json.Unmarshal(existingUserAsBytes, &user)
		if err != nil {
			return shim.Error("Failed to unmarshal user: " + err.Error())
		}
		user.UpdatedOn = time.Now().Unix()
	}

	user.ID, err = strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return shim.Error("Failed to convert user ID: " + err.Error())
	}
	user.Category, err = parseUserCategory(args[1])
	if err != nil {
		return shim.Error("Invalid user category: " + err.Error())
	}
	user.Location = args[2]
	user.MeterId = args[3]
	user.Source, err = parseEnergySource(args[4])
	if err != nil {
		return shim.Error("Invalid energy source: " + err.Error())
	}

	// Store the user in ledger
	userAsBytes, _ := json.Marshal(user)
	err = stub.PutState(strconv.Itoa(int(user.ID)), userAsBytes)
	if err != nil {
		return shim.Error("Could not store user: " + err.Error())
	}

	if existingUserAsBytes == nil {
		fmt.Println("- end CreateUser")
		return shim.Success(nil)
	} else {
		fmt.Println("- end UpdateUser")
		return shim.Success(nil)
	}
}

func SignPlatformContract(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting SignPlatformContract")

	// We are assuming that the only argument is the user ID.
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1 (UserID)")
	}

	// Check if user exists.
	userID := args[0]
	existingUserAsBytes, err := stub.GetState(userID)
	if err != nil || existingUserAsBytes == nil {
		return shim.Error("User with ID " + userID + " not found")
	}

	// Creating a new platform contract for the user.
	var contract PlatformContract
	contract.UserID, err = strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return shim.Error("Failed to convert user ID: " + err.Error())
	}
	contract.CreatedOn = time.Now().Unix()
	contract.UpdatedOn = contract.CreatedOn

	// Store the contract in the ledger using a composite key for uniqueness.
	// This will use "PlatformContract" as a prefix followed by the user ID.
	parsedUserID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return shim.Error("Failed to parse User ID: " + err.Error())
	}
	contractKey := "PlatformContract_" + strconv.FormatInt(parsedUserID, 10)

	contractAsBytes, _ := json.Marshal(contract)
	err = stub.PutState(contractKey, contractAsBytes)
	if err != nil {
		return shim.Error("Could not store platform contract: " + err.Error())
	}

	fmt.Println("- end SignPlatformContract")
	return shim.Success(nil)
}

/* -------------------------------------------------------------------------- */
/*                              Payment Methods                               */
/* -------------------------------------------------------------------------- */

func RecordPayment(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting RecordPayment")

	// Basic argument validation. We expect 12 arguments.
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 12.")
	}

	// Extracting required arguments.
	paymentID := args[0]
	paymentType, ok := paymentTypeMap[args[1]]
	if !ok {
		return shim.Error("Invalid payment type provided.")
	}
	totalAmount, err := strconv.ParseFloat(args[2], 64)
	if err != nil {
		return shim.Error("Failed to parse total amount: " + err.Error())
	}
	userID, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse user ID: " + err.Error())
	}
	debitedFrom := args[4]
	creditedTo := args[5]
	totalUnitCost, _ := strconv.ParseFloat(args[6], 64)
	platformFee, _ := strconv.ParseFloat(args[7], 64)
	tokenAmount, _ := strconv.ParseFloat(args[8], 64)
	bidRefundAmount, _ := strconv.ParseFloat(args[9], 64)
	platformFeeRefundAmount, _ := strconv.ParseFloat(args[10], 64)
	penaltyFromSeller, _ := strconv.ParseFloat(args[11], 64)

	// Create PaymentDetail entry.
	pd := PaymentDetail{
		ID:                      time.Now().UnixNano(), // Unique ID based on current timestamp.
		DebitedFrom:             debitedFrom,
		CreditedTo:              creditedTo,
		TotalUnitCost:           totalUnitCost,
		PlatformFee:             platformFee,
		TokenAmount:             tokenAmount,
		BidRefundAmount:         bidRefundAmount,
		PlatformFeeRefundAmount: platformFeeRefundAmount,
		TokenAmountRefund:       penaltyFromSeller,
		PenaltyFromSeller:       penaltyFromSeller,
	}

	// Store the PaymentDetail in the ledger.
	pdAsBytes, _ := json.Marshal(pd)
	err = stub.PutState("PaymentDetail_"+strconv.FormatInt(pd.ID, 10), pdAsBytes)
	if err != nil {
		return shim.Error("Could not store payment detail: " + err.Error())
	}

	// Create and store the Payment entry, using the PaymentDetail ID.
	p := Payment{
		CreatedOn:       time.Now().Unix(),
		ID:              paymentID,
		PaymentDetailId: pd.ID,
		PaymentType:     paymentType,
		TotalAmount:     totalAmount,
		UserID:          userID,
	}

	pAsBytes, _ := json.Marshal(p)
	err = stub.PutState("Payment_"+paymentID, pAsBytes)
	if err != nil {
		return shim.Error("Could not store payment: " + err.Error())
	}

	fmt.Println("- end RecordPayment")
	return shim.Success(nil)
}

/* -------------------------------------------------------------------------- */
/*                            Energy Bid  Methods                             */
/* -------------------------------------------------------------------------- */

func RegisterOrder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting RegisterOrder")

	// We expect 13 arguments.
	if len(args) != 12 {
		return shim.Error("Incorrect number of arguments. Expecting 12.")
	}

	// Parsing ID first to check existence.
	orderID, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse order ID: " + err.Error())
	}

	// Check if order with given ID already exists.
	existingOrderAsBytes, err := stub.GetState("Order_" + strconv.FormatInt(orderID, 10))
	if err != nil {
		return shim.Error("Error accessing state: " + err.Error())
	}

	var order Order
	if existingOrderAsBytes != nil {
		// Order exists, so we will update it.
		err = json.Unmarshal(existingOrderAsBytes, &order)
		if err != nil {
			return shim.Error("Failed to unmarshal existing order: " + err.Error())
		}
	} else {
		// Order doesn't exist, so we will create a new one.
		order.CreatedOn = time.Now().Unix()
		order.ID = orderID
	}

	bidMatchID, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BidMatchID: " + err.Error())
	}

	bidStatus, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BidStatus: " + err.Error())
	}

	// BidStatus check
	if existingOrderAsBytes == nil {
		if bidStatus != 0 && bidStatus != 1 {
			return shim.Error("Invalid BidStatus provided for new Order. It should be 0 or 1.")
		}
	}

	onMarketPrice := args[3]

	orderCost, err := strconv.ParseFloat(args[4], 64)
	if err != nil {
		return shim.Error("Failed to parse OrderCost: " + err.Error())
	}

	paymentID, err := strconv.ParseInt(args[5], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse PaymentID: " + err.Error())
	}

	slotID := args[6]
	totalQuantity, err := strconv.ParseInt(args[7], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse TotalQuantity: " + err.Error())
	}

	unitCost, err := strconv.ParseFloat(args[8], 64)
	if err != nil {
		return shim.Error("Failed to parse UnitCost: " + err.Error())
	}

	userID, err := strconv.ParseInt(args[9], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse UserID: " + err.Error())
	}

	slotExecDate, err := strconv.ParseInt(args[10], 10, 64) // Add this line to parse SlotExecDate
	if err != nil {
		return shim.Error("Failed to parse SlotExecDate: " + err.Error())
	}

	action, err := strconv.ParseInt(args[11], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse action: " + err.Error())
	}

	// Assign parsed values to the order struct
	order.BidMatchID = bidMatchID
	order.BidStatus = EnergyBidStatus(bidStatus)
	order.OnMarketPrice = onMarketPrice
	order.OrderCost = orderCost
	order.PaymentID = paymentID
	order.SlotID = slotID
	order.TotalQuantity = totalQuantity
	order.UnitCost = unitCost
	order.UpdatedOn = time.Now().Unix()
	order.UserID = userID
	order.SlotExecDate = slotExecDate // Set the SlotExecDate
	order.UserAction = Action(action)

	// Store the order back in the ledger.
	orderAsBytes, _ := json.Marshal(order)
	err = stub.PutState("Order_"+strconv.FormatInt(order.ID, 10), orderAsBytes)
	if err != nil {
		return shim.Error("Could not store order: " + err.Error())
	}

	fmt.Println("- end RegisterOrder")
	return shim.Success(nil)
}

func ProcessBidMatch(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	fmt.Println("starting ProcessBidMatch")

	// We expect 11 arguments.
	if len(args) != 11 {
		return shim.Error("Incorrect number of arguments. Expecting 11.")
	}

	// Parsing ID first to check existence.
	bidMatchID, err := strconv.ParseInt(args[6], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BidMatch ID: " + err.Error())
	}

	// Check if BidMatch with the given ID already exists.
	existingBidMatchAsBytes, err := stub.GetState("BidMatch_" + strconv.FormatInt(bidMatchID, 10))
	if err != nil {
		return shim.Error("Error accessing state: " + err.Error())
	}

	var bidMatch BidMatch
	if existingBidMatchAsBytes != nil {
		// BidMatch exists, so we will update it.
		err = json.Unmarshal(existingBidMatchAsBytes, &bidMatch)
		if err != nil {
			return shim.Error("Failed to unmarshal existing BidMatch: " + err.Error())
		}
	} else {
		// BidMatch doesn't exist, so we will create a new one.
		bidMatch.BidMatchTms = time.Now().Unix()
		bidMatch.ID = bidMatchID
	}

	bidMatchTms, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BidStatus: " + err.Error())
	}
	bidSlot := args[1]
	bidStatus, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BidStatus: " + err.Error())
	}

	bidUnitPrice, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BidUnitPrice: " + err.Error())
	}
	buyerUserId, err := strconv.ParseInt(args[4], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse BuyerUserId: " + err.Error())
	}
	deliveredBidUnits, err := strconv.ParseFloat(args[5], 64)
	if err != nil {
		return shim.Error("Failed to parse DeliveredBidUnits: " + err.Error())
	}
	originalBidUnits, err := strconv.ParseFloat(args[7], 64)
	if err != nil {
		return shim.Error("Failed to parse OriginalBidUnits: " + err.Error())
	}
	sellerUserId, err := strconv.ParseInt(args[8], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse SellerUserId: " + err.Error())
	}
	transactionBuyID, err := strconv.ParseInt(args[9], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse TransactionBuyID: " + err.Error())
	}
	transactionSellID, err := strconv.ParseInt(args[10], 10, 64)
	if err != nil {
		return shim.Error("Failed to parse TransactionSellID: " + err.Error())
	}

	// Assign parsed values to bidMatch
	bidMatch.BidMatchTms = bidMatchTms
	bidMatch.BidSlot = bidSlot
	bidMatch.BidStatus = EnergyBidStatus(bidStatus)
	bidMatch.BidUnitPrice = bidUnitPrice
	bidMatch.BuyerUserId = buyerUserId
	bidMatch.DeliveredBidUnits = deliveredBidUnits
	bidMatch.OriginalBidUnits = originalBidUnits
	bidMatch.SellerUserId = sellerUserId
	bidMatch.TransactionBuyID = transactionBuyID
	bidMatch.TransactionSellID = transactionSellID

	// Store the bidMatch back in the ledger.
	bidMatchAsBytes, _ := json.Marshal(bidMatch)
	err = stub.PutState("BidMatch_"+strconv.FormatInt(bidMatch.ID, 10), bidMatchAsBytes)
	if err != nil {
		return shim.Error("Could not store BidMatch: " + err.Error())
	}

	fmt.Println("- end ProcessBidMatch")
	return shim.Success(nil)
}
