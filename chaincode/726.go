package main

import (
	"encoding/json"

	"errors"

	"fmt"

	"log"

	"os"

	"strconv"

	"github.com/hyperledger/fabric-chaincode-go/shim"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// Package ID: , Label: abstore_1

type serverConfig struct {
	CCID string

	Address string
}

// define transction, query entities of the chaincode

type DecapolisChaincode struct {
	contractapi.Contract
}

var staticJsonTransaction = `[



    {



        "id": "1530",



        "name": "lettuce",



        "name_alt": "خس ايسبرغ",



        "product": {



            "id": 379,



            "name": "Lettuce",



            "name_alt": "خس ايسبرغ",



            "img": ""



        },



        "company": {



            "id": 121,



            "name": "Delmonte",



            "owner": {



                "id": 128,



                "username": "Delmonte",



                "email": "Delmonte@Delmonte.com",



                "is_active": true,



                "is_superuser": false



            },



            "address": {



                "label": "Delmonte Farms",



                "country": "Jordan",



                "city": "Amman",



                "area": "Amman",



                "lat": "",



                "long": ""



            },



            "phone_number": "0786477291",



        ],



        "description": null,



        "expected_date": "2022-04-18",



        "expected_quantity": null,



        "expected_quantity_unit": null,



        "is_active": true,



        "status": "RU",



        "created_at": "2022-04-18T13:47:23.790533+03:00",



        "created_by": null,



        "batches": []



    }



]`

type Company struct {
	ID string `json:"id"`

	Name string `json:"name"`

	Owner struct {
		ID string `json:"id"`

		Username string `json:"username"`

		Email string `json:"email"`

		IsActive bool `json:"is_active"`

		IsSuperuser bool `json:"is_superuser"`
	} `json:"owner"`

	Address struct {
		Label string `json:"label,omitempty"`

		Country string `json:"country,omitempty"`

		City string `json:"city,omitempty"`

		Area string `json:"area,omitempty"`

		Lat string `json:"lat,omitempty"`

		Long string `json:"long,omitempty"`
	} `json:"address,omitempty"`

	PhoneNumber string `json:"phone_number"`
}

type CreatedBy struct {
	ID string `json:"id"`

	Username string `json:"username"`

	Email string `json:"email"`

	IsActive bool `json:"is_active"`

	IsSuperuser bool `json:"is_superuser"`
}

type Field struct {
	Name string `json:"name"`

	NameAlt interface{} `json:"name_alt"`

	Slug string `json:"slug"`

	FieldType string `json:"field_type"`

	ListFieldValues string `json:"list_field_values"`

	ValidationRule interface{} `json:"validation_rule"`

	ValidationRule1 interface{} `json:"validation_rule_1"`

	ValidationRule2 interface{} `json:"validation_rule_2"`

	Value string `json:"value"`
}

type Step struct {
	ID string `json:"id"`

	Name string `json:"name"`

	NameAlt string `json:"name_alt"`

	RequiresApproval bool `json:"requires_approval"`

	Steps []interface{} `json:"steps"`

	Fields []Field `json:"fields"`

	Order int `json:"order"`

	IsGroup bool `json:"is_group"`

	IsActive bool `json:"is_active"`

	IsPendingApproval bool `json:"is_pending_approval"`
}

type Process struct {
	ID string `json:"id"`

	Status string `json:"status"`

	Process struct {
		ID string `json:"id"`

		Name string `json:"name"`

		NameAr string `json:"name_ar"`
	} `json:"process"`

	Steps []Step `json:"steps"`

	Quantity interface{} `json:"quantity"`

	QuantityUnit interface{} `json:"quantity_unit"`

	CreatedBy CreatedBy `json:"created_by"`

	CreatedAt string `json:"created_at"`
}

type Product struct {
	ID string `json:"id"`

	Name string `json:"name"`

	NameAlt interface{} `json:"name_alt"`

	Product struct {
		ID string `json:"id"`

		Name string `json:"name"`

		NameAlt string `json:"name_alt"`

		Img string `json:"img"`
	} `json:"product"`

	Company Company `json:"company"`

	Processes []Process `json:"processes"`

	Description interface{} `json:"description"`

	ExpectedDate string `json:"expected_date"`

	ExpectedQuantity interface{} `json:"expected_quantity"`

	ExpectedQuantityUnit interface{} `json:"expected_quantity_unit"`

	IsActive bool `json:"is_active"`

	Status string `json:"status"`

	CreatedAt string `json:"created_at"`

	CreatedBy CreatedBy `json:"created_by"`

	Batches []interface{} `json:"batches,omitempty"`

	TestResults struct {
		LaboratoryName struct {
			Name string `json:"name"`

			TechnicalName string `json:"technical_name"`
		} `json:"laboratory_name"`

		SampleNumber interface{} `json:"sample_number"`

		SamplingDate interface{} `json:"sampling_date"`

		AnalysisDate interface{} `json:"analysis_date"`

		SoilTexture string `json:"soil_texture"`

		SoilSanitaryDsm interface{} `json:"soil_sanitary_dsm"`

		SoilSanitaryPpm interface{} `json:"soil_sanitary_ppm"`

		SoilPhDegree interface{} `json:"soil_ph_degree"`

		Nitrogen interface{} `json:"nitrogen"`

		Potassium interface{} `json:"potassium"`

		Phosphorus interface{} `json:"phosphorus"`

		Cadmium interface{} `json:"cadmium"`

		Copper interface{} `json:"copper"`

		Nickel interface{} `json:"nickel"`

		Lead interface{} `json:"lead"`

		Zinc interface{} `json:"zinc"`

		Aphelenchoides interface{} `json:"aphelenchoides"`

		Ditylenchus interface{} `json:"ditylenchus"`

		Helicotylenchu interface{} `json:"helicotylenchu"`

		Heterodera interface{} `json:"heterodera"`

		LongidorusSppAd interface{} `json:"longidorus_spp_ad"`

		LongidorusSppJv interface{} `json:"longidorus_spp_jv"`

		LongidorusSpp interface{} `json:"longidorus_spp"`

		Pratylenchus interface{} `json:"pratylenchus"`

		Tylenchus interface{} `json:"tylenchus"`

		Scutellonema interface{} `json:"scutellonema"`

		Tylenchorhynchns interface{} `json:"tylenchorhynchns"`

		ColonyFormingUnit interface{} `json:"colony_forming_unit"`

		WaterSalinityDegree interface{} `json:"water_salinity_degree"`

		WaterPhDegree interface{} `json:"water_ph_degree"`

		Company interface{} `json:"company"`
	} `json:"test_results,omitempty"`
}

type MyError struct{}

func (m *MyError) Error() string {

	return "Custom Error"

}

func SliceIndex(limit int, predicate func(i int) bool) int {

	for i := 0; i < limit; i++ {

		if predicate(i) {

			return i

		}

	}

	return -1

}

func getJsonTransactionId(jsonTransaction string) (string, error) {

	var transactions []Product

	var err = json.Unmarshal([]byte(jsonTransaction), &transactions)

	if err != nil {

		fmt.Println(err)

		return "err on getting the transaction-id", err

	}

	return transactions[0].ID, nil

}

func (t *DecapolisChaincode) Init(context contractapi.TransactionContextInterface) error {

	fmt.Println("Decapolis ChainCode initial ledger state")

	return t.Put(context, staticJsonTransaction)

}

func (t *DecapolisChaincode) Put(context contractapi.TransactionContextInterface, jsonTransaction string) error {

	fmt.Println("Decapolis ChainCode Put ledger state")

	//fmt.Println(jsonTransaction)

	var err error

	transactionID, err := getJsonTransactionId(jsonTransaction)

	if err != nil {

		fmt.Println(err)

		return err

	}

	fmt.Println("Init: transactionID : " + transactionID)

	// write the state to the ledger

	err = context.GetStub().PutState(transactionID, []byte(jsonTransaction))

	if err != nil {

		return err

	}

	return nil

}

func GetProduct(context contractapi.TransactionContextInterface, transactionID string) (string, error) {

	fmt.Println("Decapolis ChainCode Query ledger state" + transactionID)

	var err error

	var jsonTransaction []byte

	// write the state to the ledger

	jsonTransaction, err = context.GetStub().GetState(transactionID)

	if err != nil {

		fmt.Println(err)

		return "", errors.New(`{"Error":"` + transactionID + `"}"`)

	}

	fmt.Println("Decapolis Query" + transactionID + string(jsonTransaction))

	return string(jsonTransaction), nil

}

func (t *DecapolisChaincode) Query(context contractapi.TransactionContextInterface, transactionID string) (string, error) {

	fmt.Println("Decapolis ChainCode Query ledger state" + transactionID)

	var err error

	var jsonTransaction []byte

	// write the state to the ledger

	jsonTransaction, err = context.GetStub().GetState(transactionID)

	if err != nil {

		fmt.Println(err)

		return "", errors.New(`{"Error":"` + transactionID + `"}"`)

	}

	fmt.Println("Decapolis Query" + transactionID + string(jsonTransaction))

	return string(jsonTransaction), nil

}

func (t *DecapolisChaincode) AddProduct(context contractapi.TransactionContextInterface, product string) (string, error) {

	var data Product

	if err := json.Unmarshal([]byte(product), &data); err != nil {

		return "error", err

	}

	if data.ID == "" {

		return "Missing Product Identifier", &MyError{}

	}

	var err = context.GetStub().PutState(data.ID, []byte(product))

	// update to BC

	fmt.Printf(data.Name)

	return data.ID, err

}

func (t *DecapolisChaincode) AddProcess(context contractapi.TransactionContextInterface, productId string, process string) (string, error) {

	fmt.Println("getting Product Data" + productId)

	var product Product

	var productJson, error = GetProduct(context, productId)

	if error != nil {

		return "error", error

	}

	if err := json.Unmarshal([]byte(productJson), &product); err != nil {

		return "error", err

	}

	var data Process

	if err := json.Unmarshal([]byte(process), &data); err != nil {

		return "error", err

	}

	if data.ID == "" {

		return "Missing Process Identifier", &MyError{}

	}

	product.Processes = append(product.Processes, data)

	productJsonUpdated, marshalError := json.Marshal(product)

	if error != nil {

		return "error", marshalError

	}

	fmt.Println(product.Processes)

	var err = context.GetStub().PutState(product.ID, []byte(productJsonUpdated))

	// update to BC

	fmt.Printf(data.ID)

	return product.ID, err

}

func (t *DecapolisChaincode) AddStep(context contractapi.TransactionContextInterface, productId string, processId string, step string) (string, error) {

	var product Product

	var productJson, error = GetProduct(context, productId)

	if error != nil {

		return "error getting product info", error

	}

	if err := json.Unmarshal([]byte(productJson), &product); err != nil {

		return "error parsing JSON", err

	}

	var data Step

	if err := json.Unmarshal([]byte(step), &data); err != nil {

		return "error", err

	}

	if data.ID == "" {

		return "Missing Process Identifier", &MyError{}

	}

	var processIndex = SliceIndex(len(product.Processes), func(i int) bool { return product.Processes[i].ID == processId })

	if processIndex == -1 {

		return "Error, Process not exit", &MyError{}

	}

	product.Processes[processIndex].Steps = append(product.Processes[processIndex].Steps, data)

	fmt.Println(product.Processes[processIndex].Steps)

	productJsonUpdated, marshalError := json.Marshal(product)

	if error != nil {

		return "error Marchalling to json", marshalError

	}

	var err = context.GetStub().PutState(product.ID, []byte(productJsonUpdated))

	// update to BC

	fmt.Printf(data.ID)

	return product.ID, err

}

// GetAllAssets returns all assets found in world state

func (s *DecapolisChaincode) QueryAll(ctx contractapi.TransactionContextInterface) ([]*Product, error) {

	// range query with empty string for startKey and endKey does an

	// open-ended query of all assets in the chaincode namespace.

	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")

	if err != nil {

		return nil, err

	}

	defer resultsIterator.Close()

	var transactions []*Product

	for resultsIterator.HasNext() {

		queryResponse, err := resultsIterator.Next()

		if err != nil {

			return nil, err

		}

		var transaction Product

		err = json.Unmarshal(queryResponse.Value, &transaction)

		if err != nil {

			return nil, err

		}

		transactions = append(transactions, &transaction)

	}

	return transactions, nil

}

func (t *DecapolisChaincode) DeleteTransaction(context contractapi.TransactionContextInterface, transactionID string) error {

	var transaction, err = t.Query(context, transactionID)

	// check if transaction not found error etc..

	if err != nil {

		fmt.Println(transaction)

		return err

	}

	err = context.GetStub().DelState(transactionID)

	if err != nil {

		return fmt.Errorf("failed to delete transaction %s: %v", transactionID, err)

	}

	// success not errors

	return nil

}

func (t *DecapolisChaincode) DeleteAllTransaction(context contractapi.TransactionContextInterface) error {

	var resultsIterator, err = context.GetStub().GetStateByRange("", "")

	if err != nil {

		return err

	}

	defer resultsIterator.Close()

	for resultsIterator.HasNext() {

		queryResponse, err := resultsIterator.Next()

		fmt.Println("DeleteAllTransaction")

		if err != nil {

			return err

		}

		var transaction Product

		err = json.Unmarshal(queryResponse.Value, &transaction)

		if err != nil {

			// to be cleanup after the executed for the first time successfully

			t.DeleteTransaction(context, string(queryResponse.Value))

		} else {

			t.DeleteTransaction(context, transaction.ID)

		}

	}

	return nil

}

func main() {

	// See chaincode.env.example

	config := serverConfig{

		CCID: os.Getenv("CHAINCODE_ID"),

		Address: os.Getenv("CHAINCODE_SERVER_ADDRESS"),
	}

	chaincode, err := contractapi.NewChaincode(&DecapolisChaincode{})

	if err != nil {

		log.Panicf("error create asset-transfer-basic chaincode: %s", err)

	}

	server := &shim.ChaincodeServer{

		CCID: config.CCID,

		Address: config.Address,

		CC: chaincode,

		TLSProps: getTLSProperties(),
	}

	if err := server.Start(); err != nil {

		log.Panicf("error starting decaoplis chaincode: %s", err)

	}

}

func getTLSProperties() shim.TLSProperties {

	// Check if chaincode is TLS enabled

	tlsDisabledStr := getEnvOrDefault("CHAINCODE_TLS_DISABLED", "true")

	key := getEnvOrDefault("CHAINCODE_TLS_KEY", "")

	cert := getEnvOrDefault("CHAINCODE_TLS_CERT", "")

	clientCACert := getEnvOrDefault("CHAINCODE_CLIENT_CA_CERT", "")

	// convert tlsDisabledStr to boolean

	tlsDisabled := getBoolOrDefault(tlsDisabledStr, false)

	var keyBytes, certBytes, clientCACertBytes []byte

	var err error

	if !tlsDisabled {

		keyBytes, err = os.ReadFile(key)

		if err != nil {

			log.Panicf("error while reading the crypto file: %s", err)

		}

		certBytes, err = os.ReadFile(cert)

		if err != nil {

			log.Panicf("error while reading the crypto file: %s", err)

		}

	}

	// Did not request for the peer cert verification

	if clientCACert != "" {

		clientCACertBytes, err = os.ReadFile(clientCACert)

		if err != nil {

			log.Panicf("error while reading the crypto file: %s", err)

		}

	}

	return shim.TLSProperties{

		Disabled: tlsDisabled,

		Key: keyBytes,

		Cert: certBytes,

		ClientCACerts: clientCACertBytes,
	}

}

func getEnvOrDefault(env, defaultVal string) string {

	value, ok := os.LookupEnv(env)

	if !ok {

		value = defaultVal

	}

	return value

}

// Note that the method returns default value if the string

// cannot be parsed!

func getBoolOrDefault(value string, defaultVal bool) bool {

	parsed, err := strconv.ParseBool(value)

	if err != nil {

		return defaultVal

	}

	return parsed

}
