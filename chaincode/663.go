package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	log "github.com/sirupsen/logrus"
	"io"
	"os"
	"strconv"
)

type KYC struct {
	contractapi.Contract
	NextClientID int `default:"1"`
	NextBankID   int `default:"1"`
}

type CustomerData struct {
	docType      string
	Name         string         `json:"name"`
	DateOfBirth  string         `json:"dateOfBirth"`
	Address      string         `json:"address"`
	IdNumber     int            `json:"idNumber"`
	PhoneNumber  string         `json:"phoneNumber"`
	RegisteredBy OrgCredentials `json:"registeredBy"`
}

type BankData struct {
	docType        string
	Name           string         `json:"name"`
	IdNumber       int            `json:"idNumber"`
	OrgCredentials OrgCredentials `json:"orgCredentials"`
}

type OrgCredentials struct {
	OrgName string `json:"orgName"`
	OrgNum  int    `json:"orgNum"`
}

// InitLedger adds initial customers and financial institutions to the ledger
func (s *KYC) InitLedger(ctx contractapi.TransactionContextInterface) error {
	file, err := os.OpenFile("data/customers.json", os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			panic(err)
		}
	}(file)

	// Read the contents of the file
	content, err := io.ReadAll(file)
	if err != nil {
		return err
	}

	var customers []CustomerData
	err = json.Unmarshal(content, &customers)
	if err != nil {
		return err
	}

	for _, customer := range customers {
		customer.docType = "customer"
		customerID := s.NextClientID
		customerJSON, err := json.Marshal(customer)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(strconv.Itoa(customerID), customerJSON)
		if err != nil {
			return fmt.Errorf("failed to insert the customer into world state: #{err}")
		}
		s.NextClientID++

		customerBankIndexKey, err := ctx.GetStub().CreateCompositeKey("customer~bank", []string{strconv.Itoa(customerID), customer.RegisteredBy.OrgName})
		if err != nil {
			return err
		}
		bankCustomerIndexKey, err := ctx.GetStub().CreateCompositeKey("bank~customer", []string{customer.RegisteredBy.OrgName, strconv.Itoa(customerID)})
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(customerBankIndexKey, []byte{0x00})
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(bankCustomerIndexKey, []byte{0x00})
		if err != nil {
			return err
		}
	}

	bankFile, err := os.OpenFile("data/bankData.json", os.O_RDONLY, 0644)
	if err != nil {
		return err
	}
	defer func(bankFile *os.File) {
		err := bankFile.Close()
		if err != nil {
			panic(err)
		}
	}(bankFile)

	bankContents, err := io.ReadAll(bankFile)
	if err != nil {
		return err
	}
	var banks []BankData
	err = json.Unmarshal(bankContents, &banks)
	if err != nil {
		return err
	}
	for _, bank := range banks {
		bank.docType = "bank"
		bankID := s.NextBankID
		bankJSON, err := json.Marshal(bank)
		if err != nil {
			return err
		}
		err = ctx.GetStub().PutState(strconv.Itoa(bankID), bankJSON)
		if err != nil {
			return fmt.Errorf("failed to insert the bank into world state: #{err}")
		}
		s.NextBankID++
	}
	return nil
}

func (s *KYC) GetCallerId(ctx contractapi.TransactionContextInterface) (string, error) {
	callerId, err := ctx.GetClientIdentity().GetID()
	if err != nil {
		return "", err
	}
	return callerId, nil
}

// IsRegisteredBy returns {boolean} is who registered or not, return null if client does not exists or does not have data
func (s *KYC) IsRegisteredBy(ctx contractapi.TransactionContextInterface, clientId string) (bool, error) {
	client, err := ctx.GetStub().GetState(clientId)
	if err != nil || client == nil {
		return false, err
	}
	callerId, err := s.GetCallerId(ctx)
	if err != nil {
		return false, err
	}
	var clientData CustomerData
	err = json.Unmarshal(client, &clientData)
	if clientData.RegisteredBy.OrgName == callerId {
		return true, nil
	}
	return false, nil
}

// CreateClient creates a new client and returns its ID
func (s *KYC) CreateClient(ctx contractapi.TransactionContextInterface, clientDataJson string) (string, error) {
	var customerData CustomerData
	err := json.Unmarshal([]byte(clientDataJson), &customerData)
	if err != nil {
		return "", fmt.Errorf("error parsing customer Data: %s", err)
	}
	callerId, err := s.GetCallerId(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting callerId: %s", err)
	}

	// Check if the caller is authorized to register a customer
	if customerData.RegisteredBy.OrgName != callerId {
		return "", fmt.Errorf("you are not allowed to register this client")
	}

	client := CustomerData{
		docType: "customer",
	}
	newId := s.NextClientID
	s.NextClientID++

	clientJSON, err := json.Marshal(client)
	if err != nil {
		return "", fmt.Errorf("error serializing customer Data to JSON: %s", err)
	}
	err = ctx.GetStub().PutState(strconv.Itoa(newId), clientJSON)
	if err != nil {
		return "", fmt.Errorf("error inserting client into world state: %s", err)
	}

	// Create the composite key that will allow us to query for all clients registered by a specific bank
	clientBankIndexKey, err := ctx.GetStub().CreateCompositeKey("customer~bank", []string{strconv.Itoa(newId), callerId})
	if err != nil {
		return "", fmt.Errorf("error creating client composite key: %s", err)
	}
	bankClientIndexKey, err := ctx.GetStub().CreateCompositeKey("bank~customer", []string{callerId, strconv.Itoa(newId)})
	if err != nil {
		return "", fmt.Errorf("error creating bank composite key: %s", err)
	}

	err = ctx.GetStub().PutState(clientBankIndexKey, []byte{0x00})
	if err != nil {
		return "", fmt.Errorf("error inserting client index into world state: %s", err)
	}
	err = ctx.GetStub().PutState(bankClientIndexKey, []byte{0x00})
	if err != nil {
		return "", fmt.Errorf("error inserting bank index into world state: %s", err)
	}

	return strconv.Itoa(newId), nil
}

// GetClientData returns customer's data as requested by a bank
func (s *KYC) GetClientData(ctx contractapi.TransactionContextInterface, clientId string, fields []string) (string, error) {
	client, err := ctx.GetStub().GetState(clientId)
	if err != nil || client == nil {
		return "", fmt.Errorf("error getting client data: %s", err)
	}
	callerId, err := s.GetCallerId(ctx)
	if err != nil {
		return "", fmt.Errorf("error getting callerId: %s", err)
	}
	var clientData CustomerData
	err = json.Unmarshal(client, &clientData)
	if err != nil {
		return "", fmt.Errorf("error parsing client data: %s", err)
	}
	if clientData.RegisteredBy.OrgName != callerId {
		return "", fmt.Errorf("you are not allowed to get this client data")
	}
	var clientDataJson []byte
	if len(fields) == 0 {
		clientDataJson, err = json.Marshal(clientData)
		if err != nil {
			return "", fmt.Errorf("error serializing client data to JSON: %s", err)
		}
	} else {
		clientDataJson, err = json.Marshal(clientData)
		if err != nil {
			return "", fmt.Errorf("error serializing client data to JSON: %s", err)
		}
	}
	return string(clientDataJson), nil
}

// GetAllClients returns the data of all customers
func (s *KYC) GetAllClients(ctx contractapi.TransactionContextInterface) ([]CustomerData, error) {
	// Get a range of all the keys in the ledger
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer func(resultsIterator shim.StateQueryIteratorInterface) {
		err := resultsIterator.Close()
		if err != nil {
			panic(err)
		}
	}(resultsIterator)

	var clients []CustomerData
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var client CustomerData
		err = json.Unmarshal(queryResponse.Value, &client)
		if err != nil {
			return nil, err
		}
		clients = append(clients, client)
	}

	return clients, nil
}

func main() {
	KYCchaincode, err := contractapi.NewChaincode(&KYC{
		NextBankID:   1,
		NextClientID: 1,
	})
	if err != nil {
		log.Panicf("Error creating KYC chaincode: #{err}")
	}

	if err := KYCchaincode.Start(); err != nil {
		log.Panicf("Error starting KYC chaincode: #{err}")
	}

}
