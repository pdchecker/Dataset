/*
SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"reflect"
	"runtime"
	"strings"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type serverConfig struct {
	CCID    string
	Address string
}

// SmartContract provides functions for managing an asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID                 string  `json:"ID"`
	Owner              string  `json:"owner"`
	Buyer              string  `json:"buyer"`
	Hash               int     `json:"hash"`
	InvoiceNumber      string  `json:"invoiceNumber"`
	Tax                float32 `json:"tax"`
	Netto              float32 `json:"netto"`
	CountryOrigin      string  `json:"countryOrigin"`
	CountryBuyer       string  `json:"countryBuyer"`
	Status             string  `json:"status"`
	Received           bool    `json:"received"`
	ReceivedOrder      bool    `json:"receivedOrder"`
	Sold               bool    `json:"sold"`
	ClaimPaid          bool    `json:"claimPaid"`
	ClaimPaidBy        string  `json:"claimPaidBy"`
	TaxExemptionReason string  `json:"taxExemptionReason"`
	TaxReceived        bool    `json:"taxReceived"`
}

type Role int

const (
	Buyer Role = iota
	Seller
	Factor
	TaxInspector
)

var roleTransactions map[string][]string = map[string][]string{
	"Buyer": {getFunctionNameByInterface((*SmartContract).ListInvoice),
		getFunctionNameByInterface((*SmartContract).ListInvoices),
		getFunctionNameByInterface((*SmartContract).ReceivedInvoice),
		getFunctionNameByInterface((*SmartContract).ReceivedOrder)},

	"Seller": {getFunctionNameByInterface((*SmartContract).ListInvoice),
		getFunctionNameByInterface((*SmartContract).ListInvoices),
		getFunctionNameByInterface((*SmartContract).CreateInvoice),
		getFunctionNameByInterface((*SmartContract).TransferInvoice),
		getFunctionNameByInterface((*SmartContract).ReceivedPayment)},

	"Factor": {getFunctionNameByInterface((*SmartContract).ListInvoice),
		getFunctionNameByInterface((*SmartContract).ListInvoices),
		getFunctionNameByInterface((*SmartContract).TransferInvoice),
		getFunctionNameByInterface((*SmartContract).ReceivedPayment)},

	"TaxInspector": {getFunctionNameByInterface((*SmartContract).ListInvoice),
		getFunctionNameByInterface((*SmartContract).ListInvoices),
		getFunctionNameByInterface((*SmartContract).TaxReceived)},
}

//= make(map[string][]string)

var roles map[string][]Role = make(map[string][]Role)

type RoleTransactions struct {
	Role         string   `json:"role"`
	Transactions []string `json:"transactions"`
}

func Role2String(e Role) string {
	switch e {
	case Buyer:
		return "Buyer"
	case Seller:
		return "Seller"
	case Factor:
		return "Factor"
	case TaxInspector:
		return "TaxInspector"
	default:
		return ""
	}
}

func String2Role(e string) Role {
	switch e {
	case "Buyer":
		return Buyer
	case "Seller":
		return Seller
	case "Factor":
		return Factor
	case "TaxInspector":
		return TaxInspector
	default:
		return -1
	}
}

type RoleResult struct {
	Role string
}

type RoleResult2 struct {
	Name  string   `json:"name"`
	Roles []string `json:"roles"`
}

// QueryResult structure used for handling result of query
type QueryResult struct {
	Key    string `json:"Key"`
	Record *Asset
}

// InitLedger adds a base set of cars to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{ID: "asset1", Owner: "company", Hash: 0, InvoiceNumber: "0", Tax: 0.0, Netto: 0.0, CountryOrigin: "DE", CountryBuyer: "DE", Status: "", Received: false,
			ReceivedOrder: false, Sold: false, ClaimPaid: false, ClaimPaidBy: "", TaxExemptionReason: "", TaxReceived: false},
	}
	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.ID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state: %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateInvoice(ctx contractapi.TransactionContextInterface, id, owner string, buyer string, hash int,
	invoiceNumber string, tax float32, netto float32, countryOrigin string, CountryBuyer string, status string, received bool,
	receivedOrder bool, sold bool, claimPaid bool, claimPaidBy string, taxExemptionReason string, taxReceived bool,
) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	// test
	//clientOrgID, err := ctx.GetClientIdentity().GetMSPID()

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	asset := Asset{
		ID:                 id,
		Owner:              clientID,
		Buyer:              buyer,
		Hash:               hash,
		InvoiceNumber:      invoiceNumber,
		Tax:                tax,
		Netto:              netto,
		CountryOrigin:      countryOrigin,
		CountryBuyer:       CountryBuyer,
		Status:             status,
		Received:           false,
		ReceivedOrder:      false,
		Sold:               false,
		ClaimPaid:          claimPaid,
		ClaimPaidBy:        claimPaidBy,
		TaxExemptionReason: taxExemptionReason,
		TaxReceived:        false,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) GetEmptyInvoice(ctx contractapi.TransactionContextInterface) (*Asset, error) {

	var asset Asset

	asset = Asset{
		ID:                 "",
		Owner:              "",
		Buyer:              "",
		Hash:               0,
		InvoiceNumber:      "",
		Tax:                0,
		Netto:              0,
		CountryOrigin:      "",
		CountryBuyer:       "",
		Status:             "",
		Received:           false,
		ReceivedOrder:      false,
		Sold:               false,
		ClaimPaid:          false,
		ClaimPaidBy:        "",
		TaxExemptionReason: "",
		TaxReceived:        false,
	}

	return &asset, nil
}

func (s *SmartContract) GetMSPID(ctx contractapi.TransactionContextInterface) (string, error) {

	var clientOrgID, err = ctx.GetClientIdentity().GetMSPID()

	return clientOrgID, err
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) GetRoleTransactions(ctx contractapi.TransactionContextInterface) ([]RoleTransactions, error) {

	print(getFunctionName())

	var results []RoleTransactions
	for key, value := range roleTransactions {

		var stringList []string
		//roleTransactions[key]
		for _, element := range value {
			stringList = append(stringList, element)
		}

		//store role
		var result RoleTransactions
		result.Role = key
		result.Transactions = stringList
		results = append(results, result)
	}

	return results, nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ReadAssetTest(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// ReadAsset returns the asset stored in the world state with given id.
func (s *SmartContract) ListInvoice(ctx contractapi.TransactionContextInterface, id string) (*Asset, error) {

	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", id)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return nil, err
	}

	if clientID == asset.Owner || clientID == asset.Buyer {
		return &asset, nil
	} else {
		return nil, fmt.Errorf("Only InvoiceOwner or ProductBuyer are allow to read this invoice ")
	}
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) ListInvoices(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	// range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []QueryResult

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return nil, err
	}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		if strings.HasPrefix(queryResponse.Key, "roles_") {
			continue
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}

		print("ListInvoices: ", queryResponse.Key, " : ", clientID, "-->", asset.Owner, "-->", asset.Buyer, "\n")

		if clientID == asset.Owner || clientID == asset.Buyer {
			print("ListInvoices: ", queryResponse.Key, "found \n")
			queryResult := QueryResult{Key: queryResponse.Key, Record: &asset}
			results = append(results, queryResult)
		}

	}

	return results, nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllKeys(ctx contractapi.TransactionContextInterface) ([]string, error) {
	// range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []string

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}

		//queryResult := QueryResult{Key: queryResponse.Key, Record: &asset}
		results = append(results, queryResponse.Key)
	}

	return results, nil
}

// UpdateAsset updates an existing asset in the world state with provided parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, id, owner string, hash int,
	invoiceNumber string, Tax float32, netto float32, countryOrigin string, CountryBuyer string, status string, received bool,
	receivedOrder bool, sold bool, claimPaid bool, claimPaidBy string, taxExemptionReason string, taxReceived bool) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	// overwritting original asset with new asset
	asset := Asset{
		ID:                 id,
		Owner:              owner,
		Hash:               hash,
		InvoiceNumber:      invoiceNumber,
		Tax:                Tax,
		Netto:              netto,
		CountryOrigin:      countryOrigin,
		CountryBuyer:       CountryBuyer,
		Status:             status,
		Received:           received,
		ReceivedOrder:      receivedOrder,
		Sold:               sold,
		ClaimPaid:          claimPaid,
		ClaimPaidBy:        claimPaidBy,
		TaxExemptionReason: taxExemptionReason,
		TaxReceived:        taxReceived,
	}

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func LocalStoreAsset(ctx contractapi.TransactionContextInterface, id string, asset *Asset) error {
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) ReceivedInvoice(ctx contractapi.TransactionContextInterface, id string) error {

	var asset, err = s.ListInvoice(ctx, id)

	if err != nil {
		return err
	}

	asset.Received = true

	return LocalStoreAsset(ctx, id, asset)
}

func (s *SmartContract) ReceivedOrder(ctx contractapi.TransactionContextInterface, id string) error {

	var asset, err = s.ListInvoice(ctx, id)

	if err != nil {
		return err
	}

	asset.ReceivedOrder = true

	return LocalStoreAsset(ctx, id, asset)
}

func (s *SmartContract) ReceivedPayment(ctx contractapi.TransactionContextInterface, id string, payer string) error {

	var asset, err = s.ListInvoice(ctx, id)

	if err != nil {
		return err
	}

	asset.ClaimPaid = true
	asset.ClaimPaidBy = payer

	return LocalStoreAsset(ctx, id, asset)
}

func (s *SmartContract) TaxReceived(ctx contractapi.TransactionContextInterface, id string) error {

	var asset, err = s.ListInvoice(ctx, id)

	if err != nil {
		return err
	}

	asset.TaxReceived = true

	return LocalStoreAsset(ctx, id, asset)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, id string) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", id)
	}

	return ctx.GetStub().DelState(id)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the owner field of asset with given id in world state.
func (s *SmartContract) TransferInvoice(ctx contractapi.TransactionContextInterface, id string, newOwner string) error {
	asset, err := s.ListInvoice(ctx, id)
	if err != nil {
		return err
	}

	// Get ID of submitting client identity
	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return err
	}

	if clientID != asset.Owner {
		return fmt.Errorf("submitting client not authorized to send invoice")
	}

	asset.Owner = newOwner
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

func (s *SmartContract) GetSubmittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

	b64ID, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", fmt.Errorf("Failed to read clientID: %v", err)
	}
	decodeID, err := base64.StdEncoding.DecodeString(b64ID)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
	}
	return string(decodeID), nil
}

func LocalGetRoles(ctx contractapi.TransactionContextInterface, name string) (RoleResult2, error) {
	// trial
	var id = "roles_" + name
	rolesJSON, err := ctx.GetStub().GetState(id)
	var roles1 RoleResult2

	if err != nil {
		return roles1, fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if rolesJSON == nil {
		return roles1, fmt.Errorf("the asset %s does not exist", id)
	}

	err = json.Unmarshal(rolesJSON, &roles1)
	if err != nil {
		return roles1, err
	}

	return roles1, nil
}

func (s *SmartContract) GetRoles(ctx contractapi.TransactionContextInterface) (RoleResult2, error) {

	var roles1 RoleResult2
	var clientID, err = s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return roles1, err
	}

	return LocalGetRoles(ctx, clientID)

	/* var result RoleResult2
	var rolesStringList []string


	clientID, err := s.GetSubmittingClientIdentity(ctx)
	if err != nil {
		return nil, err
	}

	var name = clientID
	for _, element := range roles[name] {
		rolesStringList = append(rolesStringList, Role2String(element))
	}

	result.Name = name
	result.Roles = rolesStringList

	return &result, nil */
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllRoles(ctx contractapi.TransactionContextInterface) ([]RoleResult2, error) {
	// range query with empty string for startKey and endKey does an open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var results []RoleResult2

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		print("GetAllRoles", queryResponse.Key)

		if !strings.HasPrefix(queryResponse.Key, "roles_") {
			continue
		}

		var roles RoleResult2
		err = json.Unmarshal(queryResponse.Value, &roles)
		if err != nil {
			return nil, err
		}

		//queryResult := QueryResult{Key: queryResponse.Key, Record: &asset}
		results = append(results, roles)
	}

	return results, nil
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func getFunctionNameByInterface(i interface{}) string {
	return runtime.FuncForPC(reflect.ValueOf(i).Pointer()).Name()
}

func getFunctionName() string {
	fpcs := make([]uintptr, 1)

	// Skip 2 levels to get the caller
	n := runtime.Callers(2, fpcs)
	if n == 0 {
		return "MSG: NO CALLER"
	}

	caller := runtime.FuncForPC(fpcs[0] - 1)
	if caller == nil {
		return "MSG CALLER WAS NIL"
	}

	// Print the name of the function
	return caller.Name()
}

// store role in blockchain
func LocalStoreRoles(ctx contractapi.TransactionContextInterface, name string, result1 RoleResult2) error {

	a, err1 := json.Marshal(result1)
	if err1 != nil {
		return err1
	}

	var err = ctx.GetStub().PutState("roles_"+name, a)
	if err != nil {
		return fmt.Errorf("failed to put to world state: %v", err)
	}

	return nil
}

func (s *SmartContract) AppendRole(ctx contractapi.TransactionContextInterface, name string, role string) error {

	var result, error = LocalGetRoles(ctx, name)

	if error != nil {
		print("error")
		result.Name = name
	} else {
		print("hi", result.Name, result.Roles)
	}

	print("OUT", result.Name, result.Roles)

	if !contains(result.Roles, role) {
		result.Roles = append(result.Roles, role)
		LocalStoreRoles(ctx, name, result)
	}

	/* roles[name] = append(roles[name], String2Role(role))
	println(name, roles[name]) */

	/// trial
	//var results []RoleResult2
	// Order not specified
	/* for key, value := range roles {
		fmt.Println(key, value)

		var rolesStringList []string

		for _, element := range roles[key] {
			rolesStringList = append(rolesStringList, Role2String(element))
		}

		//store role
		var result1 RoleResult2
		result1.Name = key
		result1.Roles = rolesStringList
		LocalStoreRoles(ctx, key, result1)
	} */

	return nil
}

/* func (s *SmartContract) GetAllRoles(ctx contractapi.TransactionContextInterface) ([]string, error) {

	var results []string

	for key, value := range roles { // Order not specified
		fmt.Println(key, value)

		result := &RoleResult{Role: Conv(roles[key])}
		b, err := json.Marshal(result)
		if err != nil {
			return nil, err
		}
		results = append(results, string(b))
	}

	return results, nil
} */

func main() {

	// See chaincode.env.example
	config := serverConfig{
		CCID:    os.Getenv("CHAINCODE_ID"),
		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(&SmartContract{})

	if err != nil {
		log.Panicf("error create asset-transfer-basic chaincode: %s", err)
	}

	server := &shim.ChaincodeServer{
		CCID:    config.CCID,
		Address: config.Address,
		CC:      chaincode,
		TLSProps: shim.TLSProperties{
			Disabled: true,
		},
	}

	if err := server.Start(); err != nil {
		log.Panicf("error starting asset-transfer-basic chaincode: %s", err)
	}
}
