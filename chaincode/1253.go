package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//var m1 map[string]string = make(map[string]string)

type PortSupplyChain struct {
}

type CounterNO struct {
	Counter int `json:"counter"`
}

type User struct {
	Name     string `json:"Name"`
	UserId   string `json:"UserID"`
	Email    string `json:"Email"`
	UserType string `json:"UserType"`
	Address  string `json:"Address"`
	Password string `json:"Password"`
}

type ProductDates struct {
	CreateProductDate                     string `json:"CreateProductDate"`
	SendToFreightForwarderDate            string `json:"SendToFreightForwarderDate"`
	SendToExportPortDate                  string `json:"SendToExportPortDate"`
	SendToShippingCompanyDate             string `json:"SendToShippingCompanyDate"`
	SendToImportPortDate                  string `json:"SendToImportPortDate"`
	SendToDestinationFreightForwarderDate string `json:"SendToDestinationFreightForwarderDate"`
	SendToConsigneeDate                   string `json:"SendToConsigneeDate"`
	OrderedDate                           string `json:"OrderedDate"`
	DeliveredDate                         string `json:"DeliveredDate"`
}

type Product struct {
	ProductId                     string       `json:"ProductID"`
	OrderId                       string       `json:"OrderID"`
	Name                          string       `json:"Name"`
	ConsignorId                   string       `json:"ConsignorID"`
	FreightForwarderId            string       `json:"FreightForwarderID"`
	ExportPortId                  string       `json:"ExportPortID"`
	ShippingCompanyId             string       `json:"ShippingCompanyID"`
	ImportPortID                  string       `json:"ImportPortID"`
	DestinationFreightForwarderId string       `json:"DestinationFreightForwarderId"`
	ConsigneeId                   string       `json:"ConsigneeID"`
	Status                        string       `json:"Status"`
	Date                          ProductDates `json:"Date"`
	Price                         float64      `json:"Price"`
}

func main() {
	err := shim.Start(new(PortSupplyChain))
	if err != nil {
		fmt.Printf("Error starting Simple chaincode: %s", err)
	}
}

func (t *PortSupplyChain) Init(stub shim.ChaincodeStubInterface) pb.Response {
	ProductCounterBytes, _ := stub.GetState("ProductCounterNO")
	if ProductCounterBytes == nil {
		var ProductCounter = CounterNO{Counter: 0}
		ProductCounterBytes, _ := json.Marshal(ProductCounter)
		err := stub.PutState("ProductCounterNO", ProductCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate Product Counter"))
		}
	}
	// Initializing Order Counter
	OrderCounterBytes, _ := stub.GetState("OrderCounterNO")
	if OrderCounterBytes == nil {
		var OrderCounter = CounterNO{Counter: 0}
		OrderCounterBytes, _ := json.Marshal(OrderCounter)
		err := stub.PutState("OrderCounterNO", OrderCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate Order Counter"))
		}
	}

	// Initializing User Counter
	UserCounterBytes, _ := stub.GetState("UserCounterNO")
	if UserCounterBytes == nil {
		UserCounter := CounterNO{Counter: 0}
		UserCounterBytes, _ := json.Marshal(UserCounter)
		err := stub.PutState("UserCounterNO", UserCounterBytes)
		if err != nil {
			return shim.Error(fmt.Sprintf("Failed to Intitate User Counter"))
		}
	}

	return shim.Success(nil)
}

// Invoke is called per transaction on the chaincode.
func (t *PortSupplyChain) Invoke(stub shim.ChaincodeStubInterface) pb.Response {
	function, args := stub.GetFunctionAndParameters()
	fmt.Println("invoke is running " + function)

	switch function {
	case "initLedger":
		return t.initLedger(stub, args)
	case "signIn":
		return t.signIn(stub, args)
	case "createUser":
		//create a new user
		return t.createUser(stub, args)
	case "createProduct":
		//create a new product
		return t.createProduct(stub, args)
	case "updateProduct":
		// update a product
		return t.updateProduct(stub, args)
	case "orderProduct":
		// order a product
		return t.orderProduct(stub, args)
	case "deliveredProduct":
		// Product is delivered
		return t.deliveredProduct(stub, args)
	case "sendToFreightForwarder":
		// send to FreightForwarder
		return t.sendToFreightForwarder(stub, args)
	case "sendToExportPort":
		// send to ExportPort
		return t.sendToExportPort(stub, args)
	case "sendToShippingCompany":
		// send to ShippingCompany
		return t.sendToShippingCompany(stub, args)
	case "sendToImportPort":
		// send to ImportPort
		return t.sendToImportPort(stub, args)
	case "sendToDestinationFreightForwarder":
		// send to DestinationFreightForwarder
		return t.sendToDestinationFreightForwarder(stub, args)
	case "sendToConsignee":
		// send to Consignee
		return t.sendToConsignee(stub, args)
	case "queryAsset":
		// query any using asset-id
		return t.queryAsset(stub, args)
	case "queryAll":
		// query all assests of a type
		return t.queryAll(stub, args)
	default:
		fmt.Println("invoke did not find func: " + function)
		return shim.Error("Received unknown function invocation")
	}
}

// create this func to count
func getCounter(stub shim.ChaincodeStubInterface, AssetType string) int {
	counterAsBytes, _ := stub.GetState(AssetType)
	counterAsset := CounterNO{}

	err := json.Unmarshal(counterAsBytes, &counterAsset)
	if err != nil {
		return 0
	}
	fmt.Printf("Counter Current Value %d of Asset Type %s", counterAsset.Counter, AssetType)

	return counterAsset.Counter
}

func incrementCounter(stub shim.ChaincodeStubInterface, assetType string) int {
	counterAsBytes, _ := stub.GetState(assetType)
	counterAsset := CounterNO{}

	err := json.Unmarshal(counterAsBytes, &counterAsset)

	if err != nil {
		shim.Error("")
	}
	counterAsset.Counter++
	counterAsBytes, _ = json.Marshal(counterAsset)
	err = stub.PutState(assetType, counterAsBytes)
	if err != nil {
		fmt.Println("Failed to Increment Counter")
	}

	fmt.Printf("Success in incrementing counter  %v\n", counterAsset)

	return counterAsset.Counter
}

// create this func to record the transaction time for PortSupplyChain struct

func (t *PortSupplyChain) GetTxTimestampChannel(stub shim.ChaincodeStubInterface) (string, error) {
	txTimeAsPtr, err := stub.GetTxTimestamp()
	if err != nil {
		fmt.Printf("Returning error in TimeStamp \n")
		return "Error", err
	}
	fmt.Printf("\t returned value from stub: %v\n", txTimeAsPtr)
	timeStr := time.Unix(txTimeAsPtr.Seconds, int64(txTimeAsPtr.Nanos)).String()

	return timeStr, nil
}

func (t *PortSupplyChain) initLedger(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// seed admin
	entityUser := User{Name: "admin", UserId: "admin", Email: "admin@pg.com", UserType: "admin", Address: "bangalore", Password: "adminpw"}

	entityUserAsBytes, errMarshal := json.Marshal(entityUser)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in user: %s", errMarshal))
	}

	errPut := stub.PutState(entityUser.UserId, entityUserAsBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to create Entity Asset: %s", entityUser.UserId))
	}

	fmt.Println("Added", entityUser)

	return shim.Success(nil)
}

// use different identities to sign in
func (t *PortSupplyChain) signIn(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expected 2 argument")
	}

	if len(args[0]) == 0 {
		return shim.Error("User ID must be provided")
	}

	if len(args[1]) == 0 {
		return shim.Error("Password must be provided")
	}

	entityUserBytes, _ := stub.GetState(args[0])
	if entityUserBytes == nil {
		return shim.Error("Cannot Find Entity")
	}
	entityUser := User{}
	err := json.Unmarshal(entityUserBytes, &entityUser)
	if err != nil {
		return shim.Error("entity user must not empty")
	}

	if entityUser.Password != args[1] {
		return shim.Error("Either id or password is wrong")
	}

	return shim.Success(entityUserBytes)
}

// create new users
func (t *PortSupplyChain) createUser(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments, Required 5 arguments")
	}

	if len(args[0]) == 0 {
		return shim.Error("Name must be provided to register user")
	}

	if len(args[1]) == 0 {
		return shim.Error("Email is mandatory")
	}

	if len(args[2]) == 0 {
		return shim.Error("User type must be specified")
	}

	if len(args[3]) == 0 {
		return shim.Error("Address must be non-empty ")
	}

	if len(args[4]) == 0 {
		return shim.Error("Password must be non-empty ")
	}

	userCounter := getCounter(stub, "UserCounterNO")
	userCounter++

	var comAsset = User{Name: args[0], UserId: "User" + strconv.Itoa(userCounter), Email: args[1], UserType: args[2], Address: args[3], Password: args[4]}

	comAssetAsBytes, errMarshal := json.Marshal(comAsset)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in Product: %s", errMarshal))
	}

	errPut := stub.PutState(comAsset.UserId, comAssetAsBytes)

	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to register user: %s", comAsset.UserId))
	}

	incrementCounter(stub, "UserCounterNO")

	fmt.Printf("User register successfully %v\n", comAsset)

	return shim.Success(comAssetAsBytes)

}

// create new products
func (t *PortSupplyChain) createProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments, Required 3 arguments")
	}
	if len(args[0]) == 0 {
		return shim.Error("productId must be provided to create a product")
	}
	if len(args[1]) == 0 {
		return shim.Error("Name must be provided to create a product")
	}

	if len(args[2]) == 0 {
		return shim.Error("Consignor_ID must be provided")
	}

	if len(args[3]) == 0 {
		return shim.Error("Price must be non-empty ")
	}
	productId := args[0]
	name := args[1]
	userBytes, _ := stub.GetState(args[2])

	if userBytes == nil {
		return shim.Error("Cannot Find User")
	}

	user := User{}

	err := json.Unmarshal(userBytes, &user)

	if err != nil {
		return shim.Error("Failed to parse userBytes")
	}

	if user.UserType != "Consignor" {
		return shim.Error("User type must be Consignor")
	}

	//Price conversion - Error handeling
	i1, errPrice := strconv.ParseFloat(args[3], 64)

	if errPrice != nil {
		return shim.Error(fmt.Sprintf("Failed to Convert Price: %s", errPrice))
	}

	//productCounter := getCounter(stub, "ProductCounterNO")
	//productCounter++
	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)

	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	dates := ProductDates{}

	dates.CreateProductDate = txTimeAsPtr

	var comAsset = Product{ProductId: productId,
		OrderId: "", Name: name, ConsigneeId: "", ConsignorId: args[1],
		FreightForwarderId: "", ExportPortId: "", ShippingCompanyId: "",
		ImportPortID: "", DestinationFreightForwarderId: "",
		Status: "Available", Date: dates, Price: i1}

	comAssetAsBytes, errMarshal := json.Marshal(comAsset)

	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error in Product: %s", errMarshal))
	}
	//productBytes, _ := stub.GetState(productId)
	errPut := stub.PutState(comAsset.ProductId, comAssetAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to create Product Asset: %s", comAsset.ProductId))
	}

	incrementCounter(stub, "ProductCounterNO")

	fmt.Printf("Success in creating Product Asset %v\n", comAsset)

	return shim.Success(comAssetAsBytes)
}

// Update the product information
func (t *PortSupplyChain) updateProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments, Required 4")
	}

	// parameter null check
	if len(args[0]) == 0 {
		return shim.Error("Product Id must be provided")
	}

	if len(args[1]) == 0 {
		return shim.Error("User Id must be provided")
	}

	if len(args[2]) == 0 {
		return shim.Error("Product Name must be provided")
	}

	if len(args[3]) == 0 {
		return shim.Error("Product Price must be provided")
	}

	userBytes, _ := stub.GetState(args[1])

	if userBytes == nil {
		return shim.Error("Cannot Find User")
	}

	user := User{}

	err := json.Unmarshal(userBytes, &user)
	if err != nil {
		return shim.Error("Failed to parse userBytes")
	}

	if user.UserType == "Consignee" {
		return shim.Error("User type cannot be Consignee")
	}

	productBytes, _ := stub.GetState(args[0])
	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}
	product := Product{}

	err = json.Unmarshal(productBytes, &product)

	if err != nil {
		return shim.Error("")
	}

	i1, errPrice := strconv.ParseFloat(args[3], 64)
	if errPrice != nil {
		return shim.Error(fmt.Sprintf("Failed to Convert Price: %s", errPrice))
	}

	product.Name = args[2]
	product.Price = i1

	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to Sell To Cosumer : %s", product.ProductId))
	}

	fmt.Printf("Success in updating Product %v \n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

// Consignee order the product
func (t *PortSupplyChain) orderProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments, Required 4")
	}

	if len(args[0]) == 0 {
		return shim.Error("Consignee Id must be provided")
	}

	if len(args[1]) == 0 {
		return shim.Error("Product Id must be provided")
	}

	userBytes, _ := stub.GetState(args[0])

	if userBytes == nil {
		return shim.Error("Cannot Find Consignee")
	}

	user := User{}

	err := json.Unmarshal(userBytes, &user)
	if err != nil {
		return shim.Error("")
	}

	if user.UserType != "Consignee" {
		return shim.Error("User type must be Consignee")
	}

	productBytes, _ := stub.GetState(args[1])
	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}
	product := Product{}

	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return shim.Error("Failed to parse productBytes")
	}

	orderCounter := getCounter(stub, "OrderCounterNO")
	orderCounter++

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	product.OrderId = "Order" + strconv.Itoa(orderCounter)
	product.ConsigneeId = user.UserId
	product.Status = "Ordered"
	product.Date.OrderedDate = txTimeAsPtr

	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	incrementCounter(stub, "OrderCounterNO")

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to place the order : %s", product.ProductId))
	}

	fmt.Printf("Order placed successfuly %v \n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

func (t *PortSupplyChain) deliveredProduct(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	// parameter length check
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, Required 4")
	}

	if len(args[0]) == 0 {
		return shim.Error("Product Id must be provided")
	}

	productBytes, _ := stub.GetState(args[0])
	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}
	product := Product{}

	err := json.Unmarshal(productBytes, &product)
	if err != nil {
		return shim.Error("")
	}

	if product.Status != "Delivered" {
		return shim.Error("Product is not delivered yet")
	}

	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	product.Date.DeliveredDate = txTimeAsPtr
	product.Status = "Delivered"
	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to update that product is delivered: %s", product.ProductId))
	}

	fmt.Printf("Success in delivering Product %v \n", product.ProductId)
	return shim.Success(updatedProductAsBytes)

}

func (t *PortSupplyChain) sendToFreightForwarder(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Less no of arguments provided")
	}

	if len(args[0]) == 0 {
		return shim.Error("Product Id must be provided")
	}

	if len(args[1]) == 0 {
		return shim.Error("Freightforwarder Id must be provided")
	}

	userBytes, _ := stub.GetState(args[1])

	if userBytes == nil {
		return shim.Error("Cannot Find Freightforwarder user")
	}

	user := User{}

	err := json.Unmarshal(userBytes, &user)
	if err != nil {
		return shim.Error("Failed to parse user bytes")
	}

	if user.UserType != "FreightForwarder" {
		return shim.Error("User type must be FreightForwarder")
	}

	productBytes, _ := stub.GetState(args[0])

	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}

	product := Product{}

	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return shim.Error("Failed to parse product bytes")
	}

	if product.FreightForwarderId != "" {
		return shim.Error("Product is send to FreightForwarder already")
	}

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	product.FreightForwarderId = user.UserId
	product.Date.SendToFreightForwarderDate = txTimeAsPtr
	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to Send to FreightForwarder: %s", product.ProductId))
	}

	fmt.Printf("Success in sending Product %v \n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

func (t *PortSupplyChain) sendToExportPort(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 2 {
		return shim.Error("Less no of arguments provided")
	}

	if len(args[0]) == 0 {
		return shim.Error("Product Id must be provided")
	}

	if len(args[1]) == 0 {
		return shim.Error("ExportPort Id must be provided")
	}

	userBytes, _ := stub.GetState(args[1])

	if userBytes == nil {
		return shim.Error("Cannot Find ExportPort user")
	}

	user := User{}

	err := json.Unmarshal(userBytes, &user)
	if err != nil {
		return shim.Error("Failed to parse ")
	}

	if user.UserType != "ExportPort" {
		return shim.Error("User type must be ExportPort")
	}

	productBytes, _ := stub.GetState(args[0])

	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}

	product := Product{}

	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return shim.Error("Failed to parse productBytes")
	}

	if product.ExportPortId != "" {
		return shim.Error("Product is send to exportport already")
	}

	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	product.ExportPortId = user.UserId
	product.Date.SendToExportPortDate = txTimeAsPtr
	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to Send to ExportPort: %s", product.ProductId))
	}

	fmt.Printf("Success in sending Product %v \n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

func (t *PortSupplyChain) sendToShippingCompany(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Less no of arguments provided")
	}

	if len(args[0]) == 0 {
		return shim.Error("Product Id must be specified")
	}

	if len(args[1]) == 0 {
		return shim.Error("ShippingCompany Id must be specified")
	}

	userBytes, _ := stub.GetState(args[1])
	if userBytes == nil {
		return shim.Error("Could not find the ShippingCompany")
	}

	user := User{}
	err := json.Unmarshal(userBytes, &user)
	if err != nil {
		return pb.Response{}
	}
	if user.UserType != "ShippingCompany" {
		return shim.Error("User must be a ShippingCompany")
	}

	productBytes, _ := stub.GetState(args[0])
	if productBytes == nil {
		return shim.Error("Could not find the product")
	}

	product := Product{}
	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return shim.Error("")
	}
	if product.ShippingCompanyId != "" {
		return shim.Error("Product has already been sent to shippingcompany")
	}

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	product.ShippingCompanyId = user.UserId
	product.Date.SendToShippingCompanyDate = txTimeAsPtr
	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to send to shippingcompany: %s", product.ProductId))
	}

	fmt.Printf("Sent product %v to shippingcompany successfully\n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

func (t *PortSupplyChain) sendToImportPort(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Less no of arguments provided")
	}

	if len(args[0]) == 0 {
		return shim.Error("Product Id must be specified")
	}

	if len(args[1]) == 0 {
		return shim.Error("ImportPort Id must be specified")
	}

	userBytes, _ := stub.GetState(args[1])
	if userBytes == nil {
		return shim.Error("Could not find the ImportPort")
	}

	user := User{}
	err := json.Unmarshal(userBytes, &user)
	if err != nil {
		return pb.Response{}
	}
	if user.UserType != "ImportPort" {
		return shim.Error("User must be a importport")
	}

	productBytes, _ := stub.GetState(args[0])
	if productBytes == nil {
		return shim.Error("Could not find the product")
	}

	product := Product{}
	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return pb.Response{}
	}
	if product.ImportPortID != "" {
		return shim.Error("Product has already been sent to ImportPort")
	}

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	product.ImportPortID = user.UserId
	product.Date.SendToImportPortDate = txTimeAsPtr
	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to send to ImportPort: %s", product.ProductId))
	}

	fmt.Printf("Sent product %v to ImportPort successfully\n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

func (t *PortSupplyChain) sendToDestinationFreightForwarder(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 2 {
		return shim.Error("Less no of arguments provided")
	}

	if len(args[0]) == 0 {
		return shim.Error("Product Id must be specified")
	}

	if len(args[1]) == 0 {
		return shim.Error("DestinationFreightForwarder Id must be specified")
	}

	userBytes, _ := stub.GetState(args[1])
	if userBytes == nil {
		return shim.Error("Could not find the Destination Freight Forwarder")
	}

	user := User{}
	err := json.Unmarshal(userBytes, &user)
	if err != nil {
		return shim.Error("Failed to  parse user bytes")
	}

	if user.UserType != "DestinationFreightForwarder" {
		return shim.Error("User must be a destinationfreightforwarder")
	}

	productBytes, _ := stub.GetState(args[0])
	if productBytes == nil {
		return shim.Error("Could not find the product")
	}

	product := Product{}
	err = json.Unmarshal(productBytes, &product)
	if err != nil {
		return shim.Error("product is wrong")
	}
	if product.DestinationFreightForwarderId != "" {
		return shim.Error("Product has already been sent to Destination Freight Forwarder")
	}

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	product.DestinationFreightForwarderId = user.UserId
	product.Date.SendToDestinationFreightForwarderDate = txTimeAsPtr
	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to send to DestinationFreightForwarder: %s", product.ProductId))
	}

	fmt.Printf("Sent product %v to Destination Freight Forwarder successfully\n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

// function to send the product to Consignee
// Input params , product id  Consignee id
func (t *PortSupplyChain) sendToConsignee(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, Required 2")
	}

	if len(args[0]) == 0 {
		return shim.Error("Product Id must be provided")
	}

	productBytes, _ := stub.GetState(args[0])

	if productBytes == nil {
		return shim.Error("Cannot Find Product")
	}

	product := Product{}

	err := json.Unmarshal(productBytes, &product)
	if err != nil {
		return shim.Error("Failed to parse productBytes")
	}

	if product.OrderId == "" {
		return shim.Error("Product has not been ordered yet")
	}

	if product.ConsigneeId == "" {
		return shim.Error("Consignee Id should be set to send to Consignee")
	}

	//To Get the transaction TimeStamp from the Channel Header
	txTimeAsPtr, errTx := t.GetTxTimestampChannel(stub)
	if errTx != nil {
		return shim.Error("Returning error in Transaction TimeStamp")
	}

	// Updating the product values to be updated after the function
	product.Date.SendToConsigneeDate = txTimeAsPtr
	product.Status = "Delivered"
	updatedProductAsBytes, errMarshal := json.Marshal(product)
	if errMarshal != nil {
		return shim.Error(fmt.Sprintf("Marshal Error: %s", errMarshal))
	}

	errPut := stub.PutState(product.ProductId, updatedProductAsBytes)
	if errPut != nil {
		return shim.Error(fmt.Sprintf("Failed to Send To Consignee : %s", product.ProductId))
	}

	fmt.Printf("Success in sending Product %v \n", product.ProductId)
	return shim.Success(updatedProductAsBytes)
}

// queryAsset
func (t *PortSupplyChain) queryAsset(stub shim.ChaincodeStubInterface, args []string) pb.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expected 1 argument")
	}
	productAsBytes, _ := stub.GetState(args[0])
	return shim.Success(productAsBytes)
}

// query all asset of a type
func (t *PortSupplyChain) queryAll(stub shim.ChaincodeStubInterface, args []string) pb.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments, Required 1")
	}

	// parameter null check
	if len(args[0]) == 0 {
		return shim.Error("Asset Type must be provided")
	}

	assetType := args[0]

	assetCounter := getCounter(stub, assetType+"CounterNO")

	startKey := assetType + "1"
	endKey := assetType + strconv.Itoa(assetCounter+1)

	resultsIterator, err := stub.GetStateByRange(startKey, endKey)

	if err != nil {

		return shim.Error(err.Error())

	}

	defer func(resultsIterator shim.StateQueryIteratorInterface) {
		err := resultsIterator.Close()
		if err != nil {
			fmt.Println("close error")
		}
	}(resultsIterator)

	// buffer is a JSON array containing QueryResults

	var buffer bytes.Buffer

	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false

	for resultsIterator.HasNext() {

		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())

		}

		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}

		buffer.WriteString("{\"Key\":")

		buffer.WriteString("\"")

		buffer.WriteString(queryResponse.Key)

		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")

		buffer.WriteString(string(queryResponse.Value))

		buffer.WriteString("}")

		bArrayMemberAlreadyWritten = true

	}

	buffer.WriteString("]")
	fmt.Printf("- queryAllAssets:\n%s\n", buffer.String())
	return shim.Success(buffer.Bytes())
}
