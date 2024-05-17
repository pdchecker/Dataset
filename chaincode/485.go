package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/segmentio/ksuid"
)

// Define the Smart Contract structure
type SmartContract struct {
	contractapi.Contract
}

func (s *SmartContract) Test(ctx contractapi.TransactionContextInterface, id string, carData string) (string, error) {

	if len(carData) == 0 {
		return "", fmt.Errorf("please pass the correct car data")
	}

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(id, []byte(carData))
}

func (s *SmartContract) pavan(ctx contractapi.TransactionContextInterface, id string, carData string) (string, error) {

	if len(carData) == 0 {
		return "", fmt.Errorf("please pass the correct car data")
	}

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(id, []byte(carData))
}

// create parcel allows the seller to create parcel immutable properties on implicit private data and parcel agreement with customer on parcel collection
func (s *SmartContract) CreateParcel(ctx contractapi.TransactionContextInterface) (string, error) {
	//verify that submitting client has the role of seller
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Seller")
	if err != nil {
		return "", fmt.Errorf("submitting client not authorized to create a parcel, does not have a seller role")
	}
	//get client MSP id
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// Parcel private properties includes parcel description, value, size and quantity
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}
	// Parcel immutable properties must be retrieved from the transient field as they are private
	// and are not kept as part of a valid transaction inside a block, which stay in the ledger forever
	immutablePropertiesJSON, ok := transientMap["parcel_properties"]
	if !ok {
		return "", fmt.Errorf("parcel_properties key not found in the transient map")
	}
	// ParcelID will be the hash of the parcel's immutable properties
	hash := sha256.New()
	hash.Write(immutablePropertiesJSON)
	parcelID := hex.EncodeToString(hash.Sum(nil))

	// create a composite key using parcel id to query the parcel
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}
	// err = ctx.GetStub().PutPrivateData(collection, parcelKey, immutablePropertiesJSON)
	err = ctx.GetStub().PutPrivateData("_implicit_org_Org1MSP", parcelKey, immutablePropertiesJSON)

	if err != nil {
		return "", fmt.Errorf("failed to put Parcel in implicit collection: %v", err)
	}
	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get client identity %v", err)
	}

	parcel := Parcel{
		Type:        "parcel",
		ParcelID:    parcelID,
		SellerOrg:   clientOrgID,
		Seller:      clientID,
		Customer:    "",
		ShipDate:    "",
		Destination: "",
		State:       "Waiting Customer Input",
	}

	parcelJSON, err := json.Marshal(parcel)
	if err != nil {
		return "", fmt.Errorf(" failed to marshal parcel json")
	}
	err = ctx.GetStub().PutPrivateData(parcelCollection, parcelKey, parcelJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put Parcel in parcelCollection: %v", err)
	}

	return parcelID, nil
}

// customer will add his agreement by adding himself as Customer, destination and ship date in parcelCollection
func (s *SmartContract) CustomerAgreement(ctx contractapi.TransactionContextInterface, parcelID string, shipDate string, destination string) error {
	//verify that submitting client has the role of customer
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Customer")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to make an agreement, does not have customer role")
	}
	// Get ID of submitting client identity and comapre it to order seller
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}

	// Get parcel immutable properties from transient map to verify that it match what has been agreed on and compare it with parcel ID and private data hash
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	// private details is private, therefore it get passed in transient field
	immutablePropertiesJSON, ok := transientMap["parcel_properties"]
	if !ok {
		return fmt.Errorf("parcel_properties is not found in the transient map input")
	}
	hash := sha256.New()
	hash.Write(immutablePropertiesJSON)
	calculatedPropertiesHash := hash.Sum(nil)

	parcel, err := s.QueryParcel(ctx, parcelID)
	if err != nil {
		return err
	}
	if parcel.State != "Waiting Customer Input" {
		return fmt.Errorf("failed to update the parcel, the state must be Waiting Customer Input")
	}
	// create a composite key using the transaction ID and order id to query the order private details
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	sellerCollection := buildCollectionName(parcel.SellerOrg)
	sellerImmutablePropertiesHash, err := ctx.GetStub().GetPrivateDataHash(sellerCollection, parcelKey)
	if err != nil {
		return fmt.Errorf("failed to read parcel immutable properties hash from seller's collection: %v", err)
	}
	if sellerImmutablePropertiesHash == nil {
		return fmt.Errorf("parcel immutable properties hash does not exist: %s", parcelKey)
	}
	// verify that the calculated hash of the passed immutable properties of the parcel matches the hash of parcel properties on seller private data collection
	if !bytes.Equal(calculatedPropertiesHash, sellerImmutablePropertiesHash) {
		return fmt.Errorf("hash %x for parcel JSON %s does not match hash : %x",
			calculatedPropertiesHash,
			immutablePropertiesJSON,
			sellerImmutablePropertiesHash,
		)
	}
	// verify that the calculated hash of the passed immutable properties of the parcel matches the parcel id in the parcel collection
	if !(hex.EncodeToString(sellerImmutablePropertiesHash) == parcel.ParcelID) {
		return fmt.Errorf("hash %x for passed immutable properties %s does match seller hash %x but do not match parcel %s: parcel was altered from its initial form",
			calculatedPropertiesHash,
			immutablePropertiesJSON,
			sellerImmutablePropertiesHash,
			parcel.ParcelID)
	}

	//Add remaining information to parcel
	parcel.Customer = clientID
	parcel.ShipDate = shipDate
	parcel.Destination = destination
	parcel.State = "Customer Agreed"

	newParcelJSON, err := json.Marshal(parcel)
	if err != nil {
		return fmt.Errorf("faild to marshal new parcel json")
	}
	err = ctx.GetStub().PutPrivateData(parcelCollection, parcelKey, newParcelJSON)
	if err != nil {
		return fmt.Errorf("failed to update parcel")
	}

	return nil
}

func (s *SmartContract) CreateOrder(ctx contractapi.TransactionContextInterface, parcelID string, minRep float64, picloc string) (string, error) {
	//verify that submitting client has the role of seller
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Seller")
	if err != nil {
		return "", fmt.Errorf("submitting client not authorized to create a bid, does not have courier role")
	}
	//get client MSP id
	clientOrgID, err := getClientOrgID(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get client MSP ID: %v", err)
	}

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get client identity %v", err)
	}
	//verify that the submitting client is same seller in parcel collection and state is customer agreed
	parcel, err := s.QueryParcel(ctx, parcelID)
	if err != nil {
		return "", err
	}
	if clientID != parcel.Seller {
		return "", fmt.Errorf("client is not authorized to create order for this parcel, he is not the seller")
	}
	if parcel.State != "Customer Agreed" {
		return "", fmt.Errorf("failed to create order, the customer must agreed on parcel details first")
	}
	//generate OrderID
	orderID := ksuid.New().String()
	//record transaction timestamp to set order date
	txTimestamp, err := ctx.GetStub().GetTxTimestamp()
	if err != nil {

		return "", fmt.Errorf("failed to create timestamp for order: %v", err)
	}
	timestamp, err := ptypes.Timestamp(txTimestamp)
	if err != nil {
		return "", err
	}
	privateOrder := ShippingOrderPrivateDetails{
		Type:         "orderPrivateDetails",
		OrderID:      orderID,
		OrderDate:    timestamp,
		ParcelID:     parcelID,
		Seller:       clientID,
		Courier:      "",
		ShippingCost: 0,
		OrderState:   "Waiting Courier Assignment",
	}

	orderPrivateDetailsJSON, err := json.Marshal(privateOrder)
	if err != nil {
		return "", fmt.Errorf("failed to marshal private order JSON: %v", err)
	}
	// the transaction ID is used as a unique index for the order private details
	orderTxID := ctx.GetStub().GetTxID()
	// create a composite key using the transaction ID and order ID
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}
	// put the order private details into the order data collection
	err = ctx.GetStub().PutPrivateData(orderCollection, orderKey, orderPrivateDetailsJSON)
	if err != nil {
		return "", fmt.Errorf("failed to input order private details into order collection: %v", err)
	}

	// Create bids attributes for the public order details
	bidders := make(map[string]BidHash)
	revealedBids := make(map[string]FullBid)

	publicOrder := ShippingOrder{
		Type:             "shippingOrder",
		OrderID:          orderID,
		MinRep:           minRep,
		PickupDate:       parcel.ShipDate,
		PickupLocation:   picloc,
		ShippingLocation: parcel.Destination,
		Orgs:             []string{clientOrgID},
		PrivateBids:      bidders,
		RevealedBids:     revealedBids,
		BidState:         "Open",
	}
	orderPublicDetailsJSON, err := json.Marshal(publicOrder)
	if err != nil {
		return "", fmt.Errorf("failed to marshal public order json: %v", err)
	}
	// put Puclic order details into state
	err = ctx.GetStub().PutState(orderID, orderPublicDetailsJSON)
	if err != nil {
		return "", fmt.Errorf("failed to put order in public data: %v", err)
	}

	// Set the endorsement policy such that an owner org peer is required to endorse future updates of the order
	err = setOrderStateBasedEndorsement(ctx, orderID, clientOrgID)
	if err != nil {
		return "", fmt.Errorf("failed setting order state based endorsement for new organization: %v", err)
	}

	//update Parcel State
	parcel.State = "Waiting Courier Assignment"
	newParcelJSON, err := json.Marshal(parcel)
	if err != nil {
		return "", fmt.Errorf("failed to marshal new parcel JSON: %v", err)
	}
	// put the new parcel state into the parcel data collection
	// create a composite key using the transaction ID and order id to query the order private details
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(parcelCollection, parcelKey, newParcelJSON)
	if err != nil {
		return "", fmt.Errorf("failed to update parcel state in parcel collection: %v", err)
	}
	// return orderID, nil

	data := map[string]string{"orderId": orderID, "txId": ctx.GetStub().GetTxID()}
	jsonStr, err := json.Marshal(data)
	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}

	return string(jsonStr), nil
}

// Bid is used to add a courier's bid to the order. The bid is stored in the private
// data collection on the peer of the courier's organization. The function returns
// the transaction ID so that couriers can identify and query their bid
func (s *SmartContract) Bid(ctx contractapi.TransactionContextInterface, orderID string) (string, error) {
	//verify that submitting client has the role of courier
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Courier")
	if err != nil {
		return "", fmt.Errorf("submitting client not authorized to create a bid, does not have courier role")
	}
	// get courier bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return "", fmt.Errorf("error getting transient: %v", err)
	}
	BidJSON, ok := transientMap["bid"]
	if !ok {
		return "", fmt.Errorf("bid key not found in the transient map")
	}
	// get the implicit collection name using the courier's organization ID and verify that courier is targeting their peer to store the bid
	collection, err := getClientImplicitCollectionNameAndVerifyClientOrg(ctx)
	if err != nil {
		return "", err
	}
	// the transaction ID is used as a unique index for the bid
	bidTxID := ctx.GetStub().GetTxID()

	// create a composite key using the transaction ID
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{orderID, bidTxID})
	if err != nil {
		return "", fmt.Errorf("failed to create composite key: %v", err)
	}
	// put the bid into the organization's implicit data collection
	err = ctx.GetStub().PutPrivateData(collection, bidKey, []byte(BidJSON))
	if err != nil {
		return "", fmt.Errorf("failed to input bid price into collection: %v", err)
	}
	// return the trannsaction ID so couriers can identify their bid
	return bidTxID, nil
}

// SubmitBid is used by the courier to add the hash of that bid stored in private data to the
// order. Note that this function alters the order in private state, and needs
// to meet the order endorsement policy. Transaction ID is used identify the bid
func (s *SmartContract) SubmitBid(ctx contractapi.TransactionContextInterface, orderID string, bidTxID string) error {
	//verify that submitting client has the role of courier
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Courier")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to submit a bid, does not have courier role")
	}
	// get the MSP ID of the courier's org
	clientOrgID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return fmt.Errorf("failed to get client MSP ID: %v", err)
	}
	// get the implicit collection name using the courier's organization ID and verify that courier is targeting their peer to read the bid
	collection, err := getClientImplicitCollectionNameAndVerifyClientOrg(ctx)
	if err != nil {
		return err
	}
	// get the order from public state
	order, err := s.QueryOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order from public state %v", err)
	}

	// the order needs to be at an open state for couriers to add their bid
	state := order.BidState
	if state != "Open" {
		return fmt.Errorf("cannot submit any bids toward thie order, it must be open")
	}
	// use the transaction ID and order ID passed as a parameter to create composite bid key to retrive private data hash
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{orderID, bidTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	// get the hash of the bid stored in private data collection
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid bash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}
	// store the hash along with the bidder's organization
	NewHash := BidHash{
		Org:  clientOrgID,
		Hash: fmt.Sprintf("%x", bidHash),
	}

	bidders := make(map[string]BidHash)
	bidders = order.PrivateBids
	bidders[bidKey] = NewHash
	order.PrivateBids = bidders

	// Add the bidding organization to the list of participating organizations if it is not already
	Orgs := order.Orgs
	if !(contains(Orgs, clientOrgID)) {
		newOrgs := append(Orgs, clientOrgID)
		order.Orgs = newOrgs
		err = addOrderStateBasedEndorsement(ctx, orderID, clientOrgID)
		if err != nil {
			return fmt.Errorf("failed setting state based endorsement for new organization: %v", err)
		}
	}
	newOrderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal new order json: %v", err)
	}
	err = ctx.GetStub().PutState(orderID, newOrderJSON)
	if err != nil {
		return fmt.Errorf("failed to update order: %v", err)
	}
	return nil
}

// CloseOrderBid can be used by the seller to close accepting bid toward the order. This prevents
// bids from being added to the order, and allows users to reveal their bids
func (s *SmartContract) CloseOrderBid(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string) error {
	//verify that submitting client has the role of seller
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Seller")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to create a parcel, does not have a seller role")
	}
	//verify that client is the order seller
	orderPrivateDetails, err := s.QueryOrderPrivateDetails(ctx, orderID, orderTxID)
	if err != nil {
		return err
	}
	// Get ID of submitting client identity and comapre it to order seller
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != orderPrivateDetails.Seller {
		return fmt.Errorf("submitting client not authorized to close this order, he is not the seller")
	}
	// get order from public state
	order, err := s.QueryOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order from public state %v", err)
	}
	state := order.BidState
	if state != "Open" {
		return fmt.Errorf("cannot close order's bid that is not open")
	}

	order.BidState = "Closed"
	closedOrderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order JSON: %v", err)
	}

	err = ctx.GetStub().PutState(orderID, closedOrderJSON)
	if err != nil {
		return fmt.Errorf("failed to close order bid: %v", err)
	}
	return nil
}

// RevealBid is used by a courier to reveal their bid after the order bidding is closed
func (s *SmartContract) RevealBid(ctx contractapi.TransactionContextInterface, orderID string, bidTxID string) error {
	//verify that submitting client has the role of courier
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Courier")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to reveal bid, does not have a courier role")
	}
	// get bid from transient map
	transientMap, err := ctx.GetStub().GetTransient()
	if err != nil {
		return fmt.Errorf("error getting transient: %v", err)
	}
	transientBidJSON, ok := transientMap["bid"]
	if !ok {
		return fmt.Errorf("bid key not found in the transient map")
	}
	// get the implicit collection name using the courier's organization ID and verify that courier is targeting their peer
	collection, err := getClientImplicitCollectionNameAndVerifyClientOrg(ctx)
	if err != nil {
		return err
	}
	// use transaction ID to create composit bid key
	bidKey, err := ctx.GetStub().CreateCompositeKey(bidKeyType, []string{orderID, bidTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	// get bid hash of bid if private bid on the public ledger
	bidHash, err := ctx.GetStub().GetPrivateDataHash(collection, bidKey)
	if err != nil {
		return fmt.Errorf("failed to read bid hash from collection: %v", err)
	}
	if bidHash == nil {
		return fmt.Errorf("bid hash does not exist: %s", bidKey)
	}
	// get order from public state
	order, err := s.QueryOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order from public state %v", err)
	}
	// Complete a series of three checks before we add the bid to the order

	// check 1: check that the order is closed.
	// We cannot reveal abid to an open order.
	state := order.BidState
	if state != "Closed" {
		return fmt.Errorf("cannot reveal bid. The order state must be closed")
	}
	// check 2: check that hash of revealed bid matches hash of private bid
	// on the public ledger. This checks that the courier is telling the truth
	// about the price they offere in their bid

	hash := sha256.New()
	hash.Write(transientBidJSON)
	calculatedBidJSONHash := hash.Sum(nil)

	// verify that the hash of the passed immutable properties of the bid matches the on-chain hash
	if !bytes.Equal(calculatedBidJSONHash, bidHash) {
		return fmt.Errorf("hash %x for bid JSON %s does not match hash in order: %x",
			calculatedBidJSONHash,
			transientBidJSON,
			bidHash,
		)
	}
	// check 3; check hash of relvealed bid matches hash of private bid that was
	// submitted earlier to the order. This ensures that the bid has not changed since it
	// was submitted to the order

	bidders := order.PrivateBids
	privateBidHashString := bidders[bidKey].Hash

	//hash of bid stored in courier's collection private data
	onChainBidHashString := fmt.Sprintf("%x", bidHash)
	if privateBidHashString != onChainBidHashString {
		return fmt.Errorf("hash %s for bid JSON %s does not match hash in order: %s, courier must have changed bid",
			privateBidHashString,
			transientBidJSON,
			onChainBidHashString,
		)
	}
	// if all three checks passed we can add the bid to the order
	type transientBidInput struct {
		Price   int    `json:"price"`
		Org     string `json:"org"`
		Courier string `json:"courier"`
	}

	// unmarshal bid input
	var bidInput transientBidInput
	err = json.Unmarshal(transientBidJSON, &bidInput)
	if err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %v", err)
	}
	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	// check 4: make sure that the transaction is being submitted is the courier
	if bidInput.Courier != clientID {
		return fmt.Errorf("Permission denied, client id %v is not the owner of the bid", clientID)
	}
	// marshal transient parameters and ID and MSPID into bid object
	NewBid := FullBid{
		Type:    "bid",
		Price:   bidInput.Price,
		Org:     bidInput.Org,
		Courier: bidInput.Courier,
	}
	revealedBids := make(map[string]FullBid)
	revealedBids = order.RevealedBids
	revealedBids[bidKey] = NewBid
	order.RevealedBids = revealedBids

	newOrderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order json")
	}

	// put order with bid added back into state
	err = ctx.GetStub().PutState(orderID, newOrderJSON)
	if err != nil {
		return fmt.Errorf("failed to update the order : %v", err)
	}

	return nil
}

// AssignCourier select the lowest bid that match the reputation requirement  and assign it in the order private details
func (s *SmartContract) AssignCourier(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string, parcelID string) error {
	//verify that submitting client has the role of seller
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Seller")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to assign a courier, does not have a seller role")
	}
	// get order from public state
	order, err := s.QueryOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order from public state %v", err)
	}
	state := order.BidState
	if state != "Closed" {
		return fmt.Errorf("Can only assign courier if the order in closed state")
	}
	orderPrivateDetails, err := s.QueryOrderPrivateDetails(ctx, orderID, orderTxID)
	if err != nil {
		return err
	}
	// Get ID of submitting client identity and comapre it to order seller
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != orderPrivateDetails.Seller {
		return fmt.Errorf("submitting client not authorized to assign a courier for this order, he is not the seller ")
	}
	// get the list of revealed bids
	revealedBidMap := order.RevealedBids
	if len(order.RevealedBids) == 0 {
		return fmt.Errorf("No bids have been revealed, cannot assign any courier: %v", err)
	}
	//Assign the first bid price as shipping cost to compare it later with the winning bid
	for _, bid := range revealedBidMap {
		if bid.Price > 0 {
			orderPrivateDetails.Courier = bid.Courier
			orderPrivateDetails.ShippingCost = bid.Price
			break
		}
	}
	// determine the lowest bid
	for _, bid := range revealedBidMap {
		if bid.Price < orderPrivateDetails.ShippingCost {
			orderPrivateDetails.Courier = bid.Courier
			orderPrivateDetails.ShippingCost = bid.Price
		}
	}
	// check if there is a lower bid that has yet to be revealed
	err = checkForLowestBid(ctx, orderPrivateDetails.ShippingCost, order.RevealedBids, order.PrivateBids)
	if err != nil {
		return fmt.Errorf("Cannot assign courier for the order: %v", err)
	}
	orderPrivateDetails.OrderState = "Courier Assigned"

	newOrderPrivateDetailsJSON, err := json.Marshal(orderPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal order private details json")
	}
	// create a composite key using the transaction ID and order id to verify that submitting client is the seller
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(orderCollection, orderKey, newOrderPrivateDetailsJSON)
	if err != nil {
		return fmt.Errorf("failed to assign courier for the order: %v", err)
	}
	//update parcel state
	parcel, err := s.QueryParcel(ctx, parcelID)
	if err != nil {
		return err
	}
	parcel.State = "Courier Assigned"

	newParcelJSON, err := json.Marshal(parcel)
	if err != nil {
		return fmt.Errorf("faild to marshal new parcel json")
	}
	// create a composite key using the transaction ID and order id to query the order private details
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(parcelCollection, parcelKey, newParcelJSON)
	if err != nil {
		return fmt.Errorf("failed to update parcel")
	}

	return nil
}

func (s *SmartContract) CourierArrived(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string) error {
	//verify that submitting client has the role of courier
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Courier")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to confirm location arrival, does not have a courier role")
	}
	//verify that client is the order assigned courier
	orderPrivateDetails, err := s.QueryOrderPrivateDetails(ctx, orderID, orderTxID)
	if err != nil {
		return err
	}
	// Get ID of submitting client identity and comapre it to order assigned courier
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != orderPrivateDetails.Courier {
		return fmt.Errorf("submitting client not authorized to confirm location arrival, he is not the assigned courier")
	}
	// change the order state to Arrived Location
	orderPrivateDetails.OrderState = "Arrived Location"

	newOrderPrivateDetailsJSON, err := json.Marshal(orderPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to m arshal order private details json")
	}
	// create a composite key using the transaction ID and order id to verify that submitting client is the seller
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(orderCollection, orderKey, newOrderPrivateDetailsJSON)
	if err != nil {
		return fmt.Errorf("failed to assign courier for the order: %v", err)
	}

	return nil
}

func (s *SmartContract) OutForDelivery(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string, parcelID string) error {
	//verify that submitting client has the role of Seller
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Seller")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to set order and parcel in out for delivery state, does not have a seller role")
	}
	//verify that client is the order seller
	orderPrivateDetails, err := s.QueryOrderPrivateDetails(ctx, orderID, orderTxID)
	if err != nil {
		return err
	}
	// Get ID of submitting client identity and comapre it to order seller
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != orderPrivateDetails.Seller {
		return fmt.Errorf("submitting client not authorized to set order and parcel in out for delivery state, he is not the seller")
	}
	//verify that order state is Arrived Location
	if orderPrivateDetails.OrderState != "Arrived Location" {
		return fmt.Errorf("failed to change order and parcel state to Out For Delivery, courier should confirm location arrival first")
	}

	// change the order state to Out for Delivery
	orderPrivateDetails.OrderState = "Out For Delivery"

	newOrderPrivateDetailsJSON, err := json.Marshal(orderPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to m arshal order private details json")
	}
	// create a composite key using the transaction ID and order id to verify that submitting client is the seller
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(orderCollection, orderKey, newOrderPrivateDetailsJSON)
	if err != nil {
		return fmt.Errorf("failed to assign courier for the order: %v", err)
	}
	//update parcel state
	parcel, err := s.QueryParcel(ctx, parcelID)
	if err != nil {
		return err
	}
	if parcel.State != "Courier Assigned" {
		return fmt.Errorf("Failed to update parcel state, current state should be Courier Assigned ")
	}
	parcel.State = "Out For Delivery"

	newParcelJSON, err := json.Marshal(parcel)
	if err != nil {
		return fmt.Errorf("faild to marshal new parcel json")
	}
	// create a composite key using the transaction ID and order id to query the order private details
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(parcelCollection, parcelKey, newParcelJSON)
	if err != nil {
		return fmt.Errorf("failed to update parcel")
	}

	return nil
}

func (s *SmartContract) ReceiveParcel(ctx contractapi.TransactionContextInterface, parcelID string) error {
	//verify that submitting client has the role of customer
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Customer")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to recieve the parcel, does not have a customer role")
	}
	//verify that client is the parcel customer
	parcel, err := s.QueryParcel(ctx, parcelID)
	if err != nil {
		return err
	}

	// Get ID of submitting client identity and comapre it to parcel customer
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != parcel.Customer {
		return fmt.Errorf("submitting client not authorized to receive this parcel, he is not the expected customer")
	}
	//verify that parcel state is Out for Delivery
	if parcel.State != "Out For Delivery" {
		return fmt.Errorf("Failed to update parcel state, current state must be Out For Delivery")
	}
	// change the parcel state to Received
	parcel.State = "Recieved by Customer"

	newParcelJSON, err := json.Marshal(parcel)
	if err != nil {
		return fmt.Errorf("faild to marshal new parcel json")
	}
	// create a composite key using the transaction ID and order id to query the order private details
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(parcelCollection, parcelKey, newParcelJSON)
	if err != nil {
		return fmt.Errorf("failed to update parcel")
	}
	return nil
}

func (s *SmartContract) Handover(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string) error {
	//verify that submitting client has the role of courier
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Courier")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to handover this order, does not have a courier role")
	}
	//verify that client is the order assigned courier
	orderPrivateDetails, err := s.QueryOrderPrivateDetails(ctx, orderID, orderTxID)
	if err != nil {
		return err
	}
	// Get ID of submitting client identity and comapre it to order assigned courier
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != orderPrivateDetails.Courier {
		return fmt.Errorf("submitting client not authorized to handover this order, he is not the assigned courier")
	}
	if orderPrivateDetails.OrderState != "Out For Delivery" {
		return fmt.Errorf("failed to update order state, the order current state must be Out For Delivery")
	}
	// change the order state to Shipping Request Completed
	orderPrivateDetails.OrderState = "Parcel Handovered to Customer"

	newOrderPrivateDetailsJSON, err := json.Marshal(orderPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal order private details json")
	}
	// create a composite key using the transaction ID and order id to verify that submitting client is the seller
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(orderCollection, orderKey, newOrderPrivateDetailsJSON)
	if err != nil {
		return fmt.Errorf("failed to update this order: %v", err)
	}

	return nil
}
func (s *SmartContract) CompleteOrder(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string, parcelID string) error {
	//verify that submitting client has the role of Seller
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Seller")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to set order and parcel in complete state, does not have a seller role")
	}
	//verify that client is the order seller
	orderPrivateDetails, err := s.QueryOrderPrivateDetails(ctx, orderID, orderTxID)
	if err != nil {
		return err
	}
	// Get ID of submitting client identity and comapre it to order seller
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != orderPrivateDetails.Seller {
		return fmt.Errorf("submitting client not authorized to set order and parcel in out for delivery state, he is not the seller")
	}
	//verify that order state is Arrived Location
	if orderPrivateDetails.OrderState != "Parcel Handovered to Customer" {
		return fmt.Errorf("failed to update order state, courier should handover the parcel to customer first")
	}

	// change the order state to Out for Delivery
	orderPrivateDetails.OrderState = "Shipping Order Completed"

	newOrderPrivateDetailsJSON, err := json.Marshal(orderPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to m arshal order private details json")
	}
	// create a composite key using the transaction ID and order id to verify that submitting client is the seller
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(orderCollection, orderKey, newOrderPrivateDetailsJSON)
	if err != nil {
		return fmt.Errorf("failed to update the order: %v", err)
	}
	//update parcel state
	parcel, err := s.QueryParcel(ctx, parcelID)
	if err != nil {
		return err
	}
	if parcel.State != "Recieved by Customer" {
		return fmt.Errorf("Failed to update parcel state, it should be received by Customer first")
	}

	parcel.State = "Parcel Delivered"

	newParcelJSON, err := json.Marshal(parcel)
	if err != nil {
		return fmt.Errorf("faild to marshal new parcel json")
	}
	// create a composite key using the transaction ID and order id to query the order private details
	parcelKey, err := ctx.GetStub().CreateCompositeKey(parcelKeyType, []string{parcelID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}

	err = ctx.GetStub().PutPrivateData(parcelCollection, parcelKey, newParcelJSON)
	if err != nil {
		return fmt.Errorf("failed to update parcel")
	}
	return nil
}

// CancelOrder can be used by the seller to cancel the order and prevents accepting bid toward the order.
func (s *SmartContract) CancelOrder(ctx contractapi.TransactionContextInterface, orderID string, orderTxID string) error {
	//verify that submitting client has the role of seller
	err := ctx.GetClientIdentity().AssertAttributeValue("role", "Seller")
	if err != nil {
		return fmt.Errorf("submitting client not authorized to cancel this order, does not have a seller role")
	}
	//verify that client is the order seller
	orderPrivateDetails, err := s.QueryOrderPrivateDetails(ctx, orderID, orderTxID)
	if err != nil {
		return err
	}
	// Get ID of submitting client identity and comapre it to order seller
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return fmt.Errorf("failed to get client identity %v", err)
	}
	if clientID != orderPrivateDetails.Seller {
		return fmt.Errorf("submitting client not authorized to cancel this order, he is not the seller")
	}
	// get order from public state
	order, err := s.QueryOrder(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to get order from public state %v", err)
	}
	state := order.BidState
	if state != "Open" {
		return fmt.Errorf("cannot cancel this order, it should be in an open state")
	}

	order.BidState = "Cancelled"
	cancelledOrderJSON, err := json.Marshal(order)
	if err != nil {
		return fmt.Errorf("failed to marshal order JSON: %v", err)
	}

	err = ctx.GetStub().PutState(orderID, cancelledOrderJSON)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %v", err)
	}
	orderPrivateDetails.OrderState = "Cancelled"

	newOrderPrivateDetailsJSON, err := json.Marshal(orderPrivateDetails)
	if err != nil {
		return fmt.Errorf("failed to marshal order private details json")
	}
	// create a composite key using the transaction ID and order id to verify that submitting client is the seller
	orderKey, err := ctx.GetStub().CreateCompositeKey(orderKeyType, []string{orderID, orderTxID})
	if err != nil {
		return fmt.Errorf("failed to create composite key: %v", err)
	}
	err = ctx.GetStub().PutPrivateData(orderCollection, orderKey, newOrderPrivateDetailsJSON)
	if err != nil {
		return fmt.Errorf("failed to cancel this order: %v", err)
	}
	return nil
}

func main() {
	orderSmartContract, err := contractapi.NewChaincode(new(SmartContract))
	if err != nil {
		log.Panicf("Error creating order chaincode: %v", err)
	}

	if err := orderSmartContract.Start(); err != nil {
		log.Panicf("Error starting order chaincode: %v", err)
	}
}
