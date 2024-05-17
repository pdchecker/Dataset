package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid/crypto"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

// DeliveryOrderChaincode implements the chaincode interface for managing delivery orders
type DeliveryOrderChaincode struct {
}

// RequestDeliveryOrder is a function to request a delivery order
func (cc *DeliveryOrderChaincode) RequestDeliveryOrder(ctx contractapi.TransactionContextInterface, args []string) pb.Response {
	// Check if the caller is the cargo owner
	if !cc.checkCallerIsCargoOwner(ctx) {
		return shim.Error("Only cargo owners can request a delivery order")
	}

	// Parse the request details from the input args
	var requestDetails RequestDetails
	err := json.Unmarshal([]byte(args[0]), &requestDetails)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to parse request details: %s", err.Error()))
	}

	// Perform analysis and verification of the delivery order request by the shipping line
	// ...

	// Verify and validate the request by INSW as an administrator
	// ...

	// Receive the delivery order file by INAPortnet and the terminal operator
	// ...

	// Return success response
	return shim.Success(nil)
}

// ReleaseDeliveryOrder is a function to release a delivery order
func (cc *DeliveryOrderChaincode) ReleaseDeliveryOrder(ctx contractapi.TransactionContextInterface, args []string) pb.Response {
	// Check if the caller is the shipping line or INSW as an administrator
	if !cc.checkCallerIsShippingLine(ctx) && !cc.checkCallerIsINSWAdmin(ctx) {
		return shim.Error("Only the shipping line or INSW administrators can release a delivery order")
	}

	// Parse the request details from the input args
	var requestDetails RequestDetails
	err := json.Unmarshal([]byte(args[0]), &requestDetails)
	if err != nil {
		return shim.Error(fmt.Sprintf("Failed to parse request details: %s", err.Error()))
	}

	// Perform necessary checks and actions to release the delivery order
	// ...

	// Return success response
	return shim.Success(nil)
}

// Helper function to check if the caller is the cargo owner
func (cc *DeliveryOrderChaincode) checkCallerIsCargoOwner(ctx contractapi.TransactionContextInterface) bool {
	cert, err := cid.GetX509Certificate(ctx.GetStub())
	if err != nil {
		return false
	}

	// Perform necessary checks to validate the caller as the cargo owner
	// ...

	return true
}

// Helper function to check if the caller is the shipping line
func (cc *DeliveryOrderChaincode) checkCallerIsShippingLine(ctx contractapi.TransactionContextInterface) bool {
	// Use the Certificate Authority (CA) to verify the role of the caller
	cert, err := cid.GetX509Certificate(ctx.GetStub())
	if err != nil {
		return false
	}

	mspID, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return false
	}

	// Perform necessary checks to validate the caller as the shipping line
	// ...

	return true
}

// Helper function to check if the caller is INSW as an administrator
func (cc *DeliveryOrderChaincode) checkCallerIsINSWAdmin(ctx contractapi.TransactionContextInterface) bool {
	// Use the Certificate Authority (CA) to verify the role of the caller
	cert, err := cid.GetX509Certificate(ctx.GetStub())
	if err != nil {
		return false
	}

	mspID, err := cid.GetMSPID(ctx.GetStub())
	if err != nil {
		return false
	}

	// Perform necessary checks to validate the caller as INSW administrator
	// ...

	return true
}

// RequestDetails represents the structure of the delivery order request details
type RequestDetails struct {
	Requestor     Requestor     `json:"requestor"`
	ShippingLine  ShippingLine  `json:"shippingLine"`
	Payment       Payment       `json:"payment"`
	Document      Document      `json:"document"`
	Parties       Parties       `json:"parties"`
	CargoDetails  CargoDetails  `json:"cargoDetails"`
	Location      Location      `json:"location"`
	PaymentDetail PaymentDetail `json:"paymentDetail"`
	SupportingDoc SupportingDoc `json:"supportingDocument"`
}

type Requestor struct {
	RequestorType string `json:"requestorType"`
	URLFile       string `json:"urlFile"`
	NPWP          string `json:"npwp"`
	NIB           string `json:"nib"`
	RequestorName string `json:"requestorName"`
	RequestorAddr string `json:"requestorAddress"`
}

type ShippingLine struct {
	ShippingType string `json:"shippingType"`
	DOExpired    string `json:"doExpired"`
	VesselName   string `json:"vesselName"`
	VoyageNumber string `json:"voyageNumber"`
	Payment      string `json:"payment"`
}

type Payment struct {
	TermOfPayment string `json:"termOfPayment"`
}

type Document struct {
	LadingBillNumber string `json:"ladingBillNumber"`
	LadingBillDate   string `json:"ladingBillDate"`
	LadingBillType   string `json:"ladingBillType"`
	URLFile          string `json:"urlFile"`
}

type Parties struct {
	Shipper     Party `json:"shipper"`
	Consignee   Party `json:"consignee"`
	NotifyParty Party `json:"notifyParty"`
}

type Party struct {
	Name string `json:"name"`
	NPWP string `json:"npwp"`
}

type CargoDetails struct {
	Container Container `json:"container"`
}

type Container struct {
	ContainerSeq   string `json:"containerSeq"`
	ContainerNo    string `json:"containerNo"`
	SealNo         string `json:"sealNo"`
	SizeType       string `json:"sizeType"`
	GrossWeight    string `json:"grossWeight"`
	Ownership      string `json:"ownership"`
}

type Location struct {
	LocationType string `json:"locationType"`
	Location     string `json:"location"`
	CountryCode  string `json:"countryCode"`
	PortCode     string `json:"portCode"`
}

type PaymentDetail struct {
	Invoice Invoice `json:"invoice"`
}

type Invoice struct {
	InvoiceNo    string `json:"invoiceNo"`
	InvoiceDate  string `json:"invoiceDate"`
	TotalAmount  string `json:"totalAmount"`
	BankID       string `json:"bankId"`
	AccountNo    string `json:"accountNo"`
	URLFile      string `json:"urlFile"`
}

type SupportingDoc struct {
	DocumentType DocumentType `json:"documentType"`
}

type DocumentType struct {
	DocumentNo   string `json:"documentNo"`
	DocumentDate string `json:"documentDate"`
	URLFile      string `json:"urlFile"`
}

func main() {
	cc := new(DeliveryOrderChaincode)
	err := contractapi.CreateNewChaincode(cc)
	if err != nil {
		fmt.Printf("Error creating delivery order chaincode: %s", err.Error())
		return
	}

	if err := shim.Start(cc); err != nil {
		fmt.Printf("Error starting delivery order chaincode: %s", err.Error())
	}
}
