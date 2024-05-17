package main

import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"log"
	"strconv"
)

type GoodContract struct {
	contractapi.Contract
}

type Good struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Unit        string `json:"unit"`
	ImportLimit string `json:"importLimit"`
	ExportLimit string `json:"exportLimit"`
}

func (c *GoodContract) CreateGood(
	ctx contractapi.TransactionContextInterface,
	id string,
	name string,
	unit string,
	importLimit string,
	exportLimit string) error {

	log.Printf("CreateGood -> INFO: creating good id: %s, name %s, unit: %s, importLimit: %s, exportLimit: %s ...\n",
		id, name, unit, importLimit, exportLimit)

	log.Println("CreateGood -> INFO: authenticating client...")
	x509Cert, err := ctx.GetClientIdentity().GetX509Certificate()
	if err != nil {
		log.Println("CreateGood -> ERROR: failed to get client certificate")
		return err
	}
	if x509Cert == nil {
		log.Println("CreateGood -> ERROR: client unidentified")
		return fmt.Errorf("client unidentified")
	}

	// Ideally, Subject field would be used but
	// cryptogen does not generate Organizations to the certificate Subjects
	organizations := x509Cert.Issuer.Organization
	authorized := false
	for _, organization := range organizations {
		if organization == "singlewindow.example.com" {
			authorized = true
		}
	}

	if !authorized {
		log.Println("CreateGood -> ERROR: client unauthorized: unauthorized organization")
		return fmt.Errorf("client unauthorized: unauthorized organization")
	} else {
		log.Println("CreateGood -> INFO: client authorized")
	}

	log.Println("CreateGood -> INFO: checking for existing good")

	exists, err := c.GoodExists(ctx, id)
	if err != nil {
		return err
	}
	if exists {
		log.Printf("CreateGood -> ERROR: good %s already exists\n", id)
		return fmt.Errorf("the good %s already exists", id)
	}

	_, err = strconv.ParseFloat(importLimit, 64)
	if err != nil {
		log.Println("CreateGood -> ERROR: failed to parse importLimit")
		return err
	}

	_, err = strconv.ParseFloat(exportLimit, 64)
	if err != nil {
		log.Println("CreateGood -> ERROR: failed to parse exportLimit")
		return err
	}

	good := Good{
		ID:          id,
		Name:        name,
		Unit:        unit,
		ImportLimit: importLimit,
		ExportLimit: exportLimit,
	}

	log.Println("CreateGood -> INFO: converting into json...")

	goodJson, err := json.Marshal(good)
	if err != nil {
		return err
	}

	log.Println("CreateGood -> INFO: adding good to the ledger")
	err = ctx.GetStub().PutState(id, goodJson)
	if err != nil {
		log.Println("CreateGood -> ERROR: failed to put state to the ledger")
	} else {
		log.Println("CreateGood -> INFO: successfully added good to the ledger")
	}
	return err
}

func (c *GoodContract) UpdateGood(ctx contractapi.TransactionContextInterface, id, importLimit, exportLimit string) error {
	log.Printf("UpdateGood -> INFO: updating good id: %s with new importLimit: %s and exportLimit: %s ...\n",
		id, importLimit, exportLimit)

	log.Println("UpdateGood -> INFO: authenticating client...")
	x509Cert, err := ctx.GetClientIdentity().GetX509Certificate()
	if err != nil {
		log.Println("UpdateGood -> ERROR: failed to get client certificate")
		return err
	}
	if x509Cert == nil {
		log.Println("UpdateGood -> ERROR: client unidentified")
		return fmt.Errorf("client unidentified")
	}

	organizations := x509Cert.Issuer.Organization
	authorized := false
	for _, organization := range organizations {
		if organization == "singlewindow.example.com" {
			authorized = true
		}
	}

	if !authorized {
		log.Println("UpdateGood -> ERROR: client unauthorized: unauthorized organization")
		return fmt.Errorf("client unauthorized: unauthorized organization")
	} else {
		log.Println("UpdateGood -> INFO: client authorized")
	}

	log.Println("UpdateGood -> INFO: retrieving good from world state")
	good, err := c.GetGoodById(ctx, id)
	if err != nil {
		return err
	}
	if good == nil {
		log.Printf("UpdateGood -> ERROR: good id: %s not found\n", id)
		return fmt.Errorf("good id: %s not found", id)
	}

	log.Println("UpdateGood -> INFO: validating importLimit value")
	if importLimit != "" {
		_, err = strconv.ParseFloat(importLimit, 64)
		if err != nil {
			log.Println("UpdateGood -> ERROR: invalid value")
			return err
		}
		good.ImportLimit = importLimit
	}

	log.Println("UpdateGood -> INFO: validating exportLimit value")
	if exportLimit != "" {
		_, err = strconv.ParseFloat(exportLimit, 64)
		if err != nil {
			log.Println("UpdateGood -> ERROR: invalid value")
			return err
		}
		good.ExportLimit = exportLimit
	}

	log.Println("UpdateGood -> INFO: marshalling good into json")
	goodJson, err := json.Marshal(good)
	if err != nil {
		log.Println("UpdateGood -> ERROR: failed to marshall json")
		return err
	}
	log.Println("UpdateGood -> INFO: putting updated value to world state...")
	err = ctx.GetStub().PutState(id, goodJson)
	if err != nil {
		log.Println("UpdateGood -> ERROR: failed to put data to world state")
	} else {
		log.Println("UpdateGood -> INFO: data updated successfully")
	}
	return err
}

func (c *GoodContract) GetGoodById(ctx contractapi.TransactionContextInterface, id string) (*Good, error) {
	log.Printf("INFO: reading good id: %s from world state...\n", id)
	goodJson, err := ctx.GetStub().GetState(id)
	if err != nil {
		log.Printf("GetGoodById -> ERROR: failed to get data from world state\n")
		return nil, err
	}
	if goodJson == nil {
		log.Printf("GetGoodById -> INFO: good id: %s not found\n", id)
		return nil, nil
	}
	log.Println("GetGoodById -> INFO: unmarshalling data from json...")
	var good *Good
	err = json.Unmarshal(goodJson, &good)
	if err != nil {
		log.Println("GetGoodById -> ERROR: failed to unmarshall json data")
		return nil, err
	}
	log.Println("GetGoodById -> INFO: data successfully unmarshalled")
	return good, nil
}

func (c *GoodContract) GoodExists(ctx contractapi.TransactionContextInterface, id string) (bool, error) {
	log.Printf("GoodExists -> INFO: checking good id: %s existence...\n", id)
	goodJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		log.Println("GoodExists -> ERROR: failed to read from world state")
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	if goodJSON == nil {
		log.Printf("GoodExists -> INFO: good id: %s does not exist\n", id)
	} else {
		log.Printf("GoodExists -> INFO: found good with id: %s\n", id)
	}

	return goodJSON != nil, nil
}

// Not included in main feature, for simulation purpose
func (c *GoodContract) ClearData(ctx contractapi.TransactionContextInterface, startKey, endKey string) error {
	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return err
	}
	defer resultsIterator.Close()
	var keys []string
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return err
		}

		var good Good
		err = json.Unmarshal(queryResponse.Value, &good)
		if err != nil {
			return err
		}
		keys = append(keys, good.ID)
		if err != nil {
			return err
		}
	}
	for i := 0; i < len(keys); i++ {
		err = ctx.GetStub().DelState(keys[i])
		if err != nil {
			return err
		}
	}
	return nil
}
