package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Agreement struct
type Agreement struct {
	Buyer            string  `json:"buyer"`            //Who is getting access to the data
	Buyer_addr       string  `json:"eth_buyer_addr"`   //Ethereum wallet address of buyer
	BuyerSign        string  `json:"buyer_sign"`       //Digital Signature if Buyer has agreed to the PN
	Seller           string  `json:"seller"`           //Who is giving access to the data
	Seller_addr      string  `json:"eth_seller_addr"`  //Ethereum wallet address of seller
	SellerSign       string  `json:"seller_sign"`      //Digital Signature if Seller has agreed to the PN
	ID               int     `json:"id"`               //Agreement's ID
	BatchID          string  `json:" batchID "`        //Batch ID of the data accorded
	IsActive         bool    `json:"isActive"`         //True when agreed by both parties
	Amount           int     `json:"amount"`           //Units that are accorded to be sold
	Price            int     `json:"price"`            //Price for each asset sold
	Payment          float64 `json:"payment"`          //How much will seller receive for each asset claimed
	Percentage       string  `json:"percentage"`       //Percentage of the 'Price' that will result in payment
	Percentage_Bonus string  `json:"percentage_bonus"` //Percentage of the 'Price' that will result in payment if Total_Devices > Amount
	AssetType        string  `json:"assetType"`        //What type of asset will be designed with this data
	TotalDevices     int     `json:"TotalDevices"`     //How many devices were designed with this data
}

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
type Asset struct {
	ID             string `json:"ID"`
	Color          string `json:"color"`
	Size           int    `json:"size"`
	Owner          string `json:"owner"`
	AppraisedValue int    `json:"appraisedValue"`
}

type AriesAgent struct {
	AgentID     string `json:"agentID"`
	AgentType   string `json:"agentType"`
	AgentStatus string `json:"status"`
}

//agentType = [CONSORTIUM, OEM_GW, OEM_SD, CONSUMER, AI_SERVICE_PROVIDER, DATA_PURCHASER]
//agentStatus = [IN_GOOD_STANDING]

type Device struct {
	DeviceID       string `json:" deviceID "`
	ControllerID   string `json:" controllerID "`
	DeviceModelID  string `json:" DeviceModelID "`
	DeviceType     string `json:" DeviceType "`
	DeviceStatus   string `json:" status "`
	DTID           string `json:" DTID "`
	SellInvitation string `json:" sell_invitation "`
}

//deviceStatus = [AVAILABLE, CLAIMED, TWINNED, IN_TRANSIT, DECOMMISSIONED]

type DataBatch struct {
	BatchID      string `json:" batchID "`
	BatchURL     string `json:" batchURL "`
	Hash         string `json:" hash "`
	ControllerID string `json:" ControllerID "`
	timestamp    string `json:" ControllerID "`
}

type DeviceModel struct {
	DeviceModelID string   `json:"deviceModelID"`
	Description   string   `json:"description"`
	Features      []string `json:"features"`
	Images        []string `json:"images"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	counter := 0
	counterAsBytes := []byte(strconv.Itoa(counter)) //counter to string and  then converting to byte slice format
	ctx.GetStub().PutState("agreementCounter", counterAsBytes)

	ariesAgents := []AriesAgent{
		{AgentID: "agent1",
			AgentType:   "CONSORTIUM",
			AgentStatus: "IN_GOOD_STANDING"},
	}

	deviceModels := []DeviceModel{
		{DeviceModelID: "devicemodel1",
			Description: "iWatch Device Model",
			Features:    []string{"feature1"},
			Images:      []string{"image1"}},
	}

	devices := []Device{
		{DeviceID: "device1",
			ControllerID:   "controller1",
			DeviceModelID:  "devicemodel1",
			DeviceType:     "deviceType",
			DeviceStatus:   "deviceStatus",
			DTID:           "digitalTwin1",
			SellInvitation: "url1"},
	}

	dataBatchs := []DataBatch{
		{BatchID: "batch1",
			BatchURL: "batch url 1",
			Hash:     "hash1"},
	}

	for _, ariesagent := range ariesAgents {
		assetJSON, err := json.Marshal(ariesagent)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(ariesagent.AgentID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	for _, devicemodel := range deviceModels {
		assetJSON, err := json.Marshal(devicemodel)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(devicemodel.DeviceModelID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	for _, device := range devices {
		assetJSON, err := json.Marshal(device)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(device.DeviceID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	for _, databatch := range dataBatchs {
		assetJSON, err := json.Marshal(databatch)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(databatch.BatchID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, id string, color string, size int, owner string, appraisedValue int) error {
	exists, err := s.AssetExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", id)
	}

	asset := Asset{
		ID:             id,
		Color:          color,
		Size:           size,
		Owner:          owner,
		AppraisedValue: appraisedValue,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(id, assetJSON)
}

// AssetExists returns true when asset with given ID exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// DeviceExists checks if a device with deviceid already exists in the worldstate
func (s *SmartContract) DeviceExists(ctx contractapi.TransactionContextInterface, deviceid string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(deviceid)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}
	return assetJSON != nil, nil
}

// DeviceGenesisRegistration registers a new device with given details
func (s *SmartContract) DeviceGenesisRegistration(ctx contractapi.TransactionContextInterface, deviceid string,
	controllerid string,
	devicemodelid string,
	devicetype string,
	devicestatus string,
	dtid string,
	sellinvitation string) error {
	exists, err := s.DeviceExists(ctx, deviceid)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the device %s already exists", deviceid)
	}

	device := Device{
		DeviceID:       deviceid,
		ControllerID:   controllerid,
		DeviceModelID:  devicemodelid,
		DeviceType:     devicetype,
		DeviceStatus:   devicestatus,
		DTID:           dtid,
		SellInvitation: sellinvitation,
	}
	assetJSON, err := json.Marshal(device)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(deviceid, assetJSON)
}

// GetAllDevices returns all the devices that exist in the worldstate
func (s *SmartContract) GetAllDevices(ctx contractapi.TransactionContextInterface) ([]*Device, error) {
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var devices []*Device
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var device Device
		err = json.Unmarshal(queryResponse.Value, &device)
		if err != nil {
			return nil, err
		}
		if device.DeviceID == "" {
			continue
		} else {
			devices = append(devices, &device)
		}
	}

	return devices, nil
}

// ReadDevice returns the pointer to a device with deviceid
func (s *SmartContract) ReadDevice(ctx contractapi.TransactionContextInterface, deviceid string) (*Device, error) {
	assetJSON, err := ctx.GetStub().GetState(deviceid)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the device %s does not exist", deviceid)
	}

	var device Device
	err = json.Unmarshal(assetJSON, &device)
	if err != nil {
		return nil, err
	}

	return &device, nil
}

// BuyDevice changes the status of a given deviceid with the status given
func (s *SmartContract) BuyDevice(ctx contractapi.TransactionContextInterface, deviceid string, status string) error {
	device, err := s.ReadDevice(ctx, deviceid)
	if err != nil {
		return fmt.Errorf("failed to buy device: %v", err)
	}
	device.DeviceStatus = status
	assetJSON, err := json.Marshal(device)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(deviceid, assetJSON)
}

// ClaimDevice issues the claim of a device that is registered in the network.
// All the promissory notes created and active that have the same device model as the one that was claimed
// are identified and the payments are updated. An event is set every time a device is claimed to inform
// Fabric Application, which is listening to this specific event.
func (s *SmartContract) ClaimDevice(ctx contractapi.TransactionContextInterface, deviceid string,
	status string,
	controllerid string) float64 {
	device, _ := s.ReadDevice(ctx, deviceid)

	device.DeviceStatus = status
	device.ControllerID = controllerid
	assetJSON, _ := json.Marshal(device)

	ctx.GetStub().PutState(deviceid, assetJSON)

	payload_device := Device{
		DeviceID:       device.DeviceID,
		ControllerID:   device.ControllerID,
		DeviceModelID:  device.DeviceModelID,
		DeviceType:     device.DeviceType,
		DeviceStatus:   device.DeviceStatus,
		DTID:           device.DTID,
		SellInvitation: device.SellInvitation,
	}

	deviceAsBytes, _ := json.Marshal(payload_device)

	// After the device has been successfully claimed, we can check the promissory notes:
	agreementCounterAsBytes, _ := ctx.GetStub().GetState("agreementCounter")
	agreementCounter, _ := strconv.Atoi(string(agreementCounterAsBytes))

	fmt.Printf("Counter: %v \n", agreementCounter)

	result := 0.0
	for i := 1; i <= agreementCounter; i++ {
		agreementAsBytes, _ := ctx.GetStub().GetState("Agreement" + strconv.Itoa(i))
		var agreement Agreement
		json.Unmarshal(agreementAsBytes, &agreement)

		//From all the promises that exist, we are only interested in those with device.DeviceModelID
		if agreement.AssetType == device.DeviceModelID {
			agreement.TotalDevices = agreement.TotalDevices + 1 // assuming totalDevices has been added to your Agreement struct

			percent := ParseStringToFloat(agreement.Percentage)
			percent_bonus := ParseStringToFloat(agreement.Percentage_Bonus)

			agreement.Payment = calculatePayment(percent,
				percent_bonus,
				agreement.Amount,
				agreement.TotalDevices,
				agreement.Price)

			fmt.Printf("Payment: %v", agreement.Payment)

			agreementAsBytes, _ = json.Marshal(agreement)
			ctx.GetStub().PutState("Agreement"+strconv.Itoa(i), agreementAsBytes)
		}
	}

	//Emit an event
	ctx.GetStub().SetEvent("ClaimedDevice", []byte(deviceAsBytes))

	return result
}

// SignPromissoryNote simulates the agreement from both parties. Once signed by both parties, the
// promissory note is considered active
func (s *SmartContract) SignPromissoryNote(ctx contractapi.TransactionContextInterface, agreementID string) error {

	invokerID, _ := ctx.GetClientIdentity().GetID()
	decoded, err := base64.StdEncoding.DecodeString(invokerID)
	if err != nil {
		return fmt.Errorf("failed to decode base64 invokerID: %v", err)
	}

	parts := strings.Split(string(decoded), "::")
	if len(parts) < 2 {
		return fmt.Errorf("invalid invoker ID format")
	}

	signer := ""
	for _, part := range parts {
		if strings.HasPrefix(part, "CN=") {
			cnValue := strings.TrimPrefix(part, "CN=")
			signer = strings.Split(cnValue, ",")[0]
			break
		}
	}

	agreementAsBytes, err := ctx.GetStub().GetState("Agreement" + agreementID)
	if err != nil {
		return fmt.Errorf("could not read Agreement with ID %s from world state: %v", agreementID, err)
	}

	var agreement Agreement
	err = json.Unmarshal(agreementAsBytes, &agreement)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Agreement: %v", err)
	}

	if signer == agreement.Buyer && agreement.BuyerSign == "" {
		agreement.BuyerSign = signer
		agreement.SellerSign = agreement.Seller
		agreement.IsActive = true
	}

	// Check if both the buyer and seller have signed
	if agreement.BuyerSign != "" && agreement.SellerSign != "" {
		agreement.IsActive = true
	}

	agreementAsBytes, err = json.Marshal(agreement)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState("Agreement"+agreementID, agreementAsBytes)
}

// CreatePromissoryNote creates a PN with given details
func (s *SmartContract) CreatePromissoryNote(ctx contractapi.TransactionContextInterface,
	buyer string,
	b_addr string,
	seller string,
	s_addr string,
	batchID string,
	amount string,
	price int,
	assetType string,
	percent string,
	percent_bonus string) (int, error) {

	userID := ""
	// Get the full invoker ID
	invokerID, _ := ctx.GetClientIdentity().GetID()

	decoded, err := base64.StdEncoding.DecodeString(invokerID)
	if err != nil {
		return -1, fmt.Errorf("failed to decode base64 invokerID: %v", err)
	}

	// Now split the decoded string and extract the CN
	parts := strings.Split(string(decoded), "::")
	if len(parts) < 2 {
		return -1, fmt.Errorf("invalid invoker ID format")
	}

	for _, part := range parts {
		// Check if the string starts with "CN="
		if strings.HasPrefix(part, "CN=") {
			cnValue := strings.TrimPrefix(part, "CN=")
			userID = strings.Split(cnValue, ",")[0]
			break
		}
	}

	// Check if appUserID matches the buyer parameter
	if buyer != userID {
		return -1, fmt.Errorf("authorization error: the invoker is not the specified buyer")
	}

	counterAsBytes, _ := ctx.GetStub().GetState("agreementCounter")
	counter, _ := strconv.Atoi(string(counterAsBytes))

	counter++

	_amount := ParseStringToInt(amount)

	_percent := ParseStringToFloat(percent)

	_percent_bonus := ParseStringToFloat(percent_bonus)

	total := 0

	_payment := calculatePayment(_percent, _percent_bonus, _amount, total, price)

	agreement := Agreement{
		Buyer:            buyer,
		Buyer_addr:       b_addr,
		BuyerSign:        "",
		Seller:           seller,
		Seller_addr:      s_addr,
		SellerSign:       "",
		ID:               counter,
		BatchID:          batchID,
		IsActive:         false,
		Amount:           _amount,
		Price:            price,
		Payment:          _payment,
		Percentage:       percent,
		Percentage_Bonus: percent_bonus,
		AssetType:        assetType,
		TotalDevices:     total,
	}

	agreementAsBytes, err := json.Marshal(agreement)

	if err != nil {
		return -1, err
	}

	ctx.GetStub().PutState("Agreement"+strconv.Itoa(counter), agreementAsBytes) //Store in ledger as key value; Key: Agreement'x', Value: agreementAsBytes

	counterAsBytes = []byte(strconv.Itoa(counter))
	ctx.GetStub().PutState("agreementCounter", counterAsBytes)

	return counter, nil
}

//calculatePayment returns a value which represents the payment in each PN
func calculatePayment(percentage float64, percentageBonus float64, amount int, total int, pricePerUnit int) float64 {
	var payment float64
	if total > amount {
		payment = float64(pricePerUnit) * (percentageBonus / 100.0)
	} else {
		payment = float64(pricePerUnit) * (percentage / 100.0)
	}
	return payment
}

func ParseStringToInt(s string) int {
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return value
}

func ParseStringToFloat(s string) float64 {
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return value
}

// GetAgreementByID() returns a PN with a given id
func (s *SmartContract) GetAgreementByID(ctx contractapi.TransactionContextInterface, id string) (string, error) {
	key := "Agreement" + id

	// Fetch the agreement from the ledger using the ID
	agreementBytes, err := ctx.GetStub().GetState(key)

	if err != nil {
		return "null", fmt.Errorf("Failed to read from world state: %v", err)
	}

	if agreementBytes == nil {
		return "null", fmt.Errorf("The Agreement with ID %s does not exist", id)
	}

	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, agreementBytes, "", "\t")
	if err != nil {
		return "null", fmt.Errorf("Failed to generate pretty JSON: %s", err)
	}

	return prettyJSON.String(), nil

}

// GetAllAgreements() returns all the PNs eligible in the worldstate
func (s *SmartContract) GetAllAgreements(ctx contractapi.TransactionContextInterface) ([]*Agreement, error) {
	startKey := "Agreement0" //Lexicographically smaller than any Agreement_ value (ASCII table)
	endKey := "AgreementZ"   //Lexicographically bigger than any Agreement_ value (ASCII table)

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var agreements []*Agreement
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var agreement Agreement
		err = json.Unmarshal(queryResponse.Value, &agreement)
		if err != nil {
			return nil, err
		}
		agreements = append(agreements, &agreement)
	}

	return agreements, nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating merged chaincode: %v", err)
	}
	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting merged chaincode: %v", err)
	}
}
