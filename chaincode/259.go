package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type Transaction struct {
	ObjectType  string  `json:"docType"`
	From        string  `json:"from"`
	To          string  `json:"to"`
	Qty         float64 `json:"qty"`
	Date        int     `json:"date"`
	TxType      string  `json:"txType"`
	ObjRef      string  `json:"objRef"`
	PhaseNumber int     `json:"phaseNumber"`
	Notes       string  `json:"notes"`
	PaymentMode string  `json:"paymentMode"`
	PaymentId   string  `json:"paymentId"`
}

type Notification struct {
	TxId        string   `json:"txId"`
	Description string   `json:"description"`
	Users       []string `json:"users"`
}

// Contains tells whether a contains x.
func contains(a []string, x string) bool {
	for _, n := range a {
		if x == n {
			return true
		}
	}
	return false
}

//generate query string
func gqs(a []string) string {
	res := "{\"selector\":{"

	for i := 0; i < len(a); i = i + 2 {
		res = res + "\"" + a[i] + "\":\"" + a[i+1] + "\""
		if i != len(a)-2 {
			res = res + ","
		}
	}
	return res + "}}"
}

//get all corporates
func getCorporates(ctx contractapi.TransactionContextInterface) []string {

	//Get corporate list
	corporatesBytes, _ := ctx.GetStub().GetState("corporates")
	corporates := []string{}

	if corporatesBytes != nil {
		json.Unmarshal(corporatesBytes, &corporates)
	}
	return corporates
}

//add a new Tx to the ledger
func createTransaction(ctx contractapi.TransactionContextInterface, fromAddress string, toAddress string, quantity float64, timestamp int, transactionType string, objRef string, txId string, phaseNumber int) error {

	txObjAsBytes, _ := ctx.GetStub().GetState(txId)
	if txObjAsBytes != nil {
		return fmt.Errorf("tx id already exists")
	}

	newTx := &Transaction{
		ObjectType:  "Transaction",
		From:        fromAddress,
		To:          toAddress,
		Qty:         quantity,
		Date:        timestamp,
		TxType:      transactionType,
		ObjRef:      objRef,
		PhaseNumber: phaseNumber,
	}
	txInBytes, _ := json.Marshal(newTx)

	//save the tx
	return ctx.GetStub().PutState(txId, txInBytes)
}

//get loggedin user info
func getTxCreatorInfo(ctx contractapi.TransactionContextInterface, creator []byte) (string, string, error) {
	var mspMap map[string]string
	mspMap = make(map[string]string)
	mspMap[CorporateMSP] = "." + corporate + "." + domain
	mspMap[NgoMSP] = "." + ngo + "." + domain
	mspMap[CreditsAuthorityMSP] = "." + creditsauthority + "." + domain

	identity := ctx.GetClientIdentity()

	clientId, er1 := identity.GetID() //asdrfdrgrxvxz
	if er1 == nil {
		fmt.Println("client id: " + clientId)
	}
	clientMSPId, er1 := identity.GetMSPID() //CorporateMSP
	if er1 == nil {
		fmt.Println("client msp id: " + clientMSPId)
	}

	data, er1 := base64.StdEncoding.DecodeString(clientId)
	if er1 != nil {
		fmt.Println("error:", er1)
	}
	fmt.Println("DATA: " + string(data))

	strArr := strings.Split(string(data), "::")
	fmt.Println(strArr)

	strArr2 := strings.Split(strArr[1], ",")
	fmt.Println(strArr2)

	strClientName := strings.Split(strArr2[0], "=")[1] //username

	return clientMSPId, strClientName + mspMap[clientMSPId], nil
}

// add a corporate with Email
func (s *SmartContract) AddCorporateEmail(ctx contractapi.TransactionContextInterface, arg string) (bool, error) {
	InfoLogger.Printf("*************** AddCorporateEmail Started ***************")
	InfoLogger.Printf("args received:", arg)

	//getusercontext to populate the required data
	creator, err := ctx.GetStub().GetCreator()
	if err != nil {
		return false, fmt.Errorf("Error getting transaction creator: " + err.Error())
	}

	mspId, commonName, _ := getTxCreatorInfo(ctx, creator)
	InfoLogger.Printf("current logged in user:", commonName, "with mspId:", mspId)

	if mspId != CreditsAuthorityMSP || !strings.HasPrefix(commonName, ca) {
		InfoLogger.Printf("only creditsauthority can initiate addCorporateEmail")
		return false, fmt.Errorf("only creditsauthority can initiate addCorporateEmail")
	}

	var args []string

	err = json.Unmarshal([]byte(arg), &args)
	if err != nil {
		return false, fmt.Errorf(err.Error())
	}

	if len(args) != 2 {
		return false, fmt.Errorf("Incorrect number of arguments. Expecting 2")
	} else if len(args[0]) <= 0 {
		return false, fmt.Errorf("Email must be non-empty")
	} else if len(args[1]) <= 0 {
		return false, fmt.Errorf("corporate name must be non-empty")
	}

	email := args[0]
	corpName := args[1]

	// set corporateEmail map
	corporateEmailBytes, _ := ctx.GetStub().GetState("corporateEmail")
	corporateEmail := make(map[string]string)

	if corporateEmailBytes != nil {
		json.Unmarshal(corporateEmailBytes, &corporateEmail)
	}
	if len(corporateEmail[email]) > 0 {
		return false, fmt.Errorf("This email already exists")
	}
	corporateEmail[email] = corpName

	corporateEmailBytes, err = json.Marshal(corporateEmail)
	if err != nil {
		return false, fmt.Errorf("error in marshalling: " + err.Error())
	}

	ctx.GetStub().PutState("corporateEmail", corporateEmailBytes)

	InfoLogger.Printf("*************** AddCorporateEmail Successfull ***************")
	return true, nil
}
