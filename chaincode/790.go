package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

const id_index = "id~name"

type CertContract struct {
	contractapi.Contract
}

type Cert struct {
	CertID string `json:"certID"`
	Unit   string `json:"unit"`
	Name   string `json:"name"`
	NID    string `json:"nid"`
}

func (c *CertContract) CreateCert(ctx contractapi.TransactionContextInterface, certID string, unit string, name string, nid string) (string, error) {
	exists, err := c.HashExists(ctx, certID)

	if err != nil {
		return "", fmt.Errorf("failed to read from world state. %s", err.Error())
	}
	if exists {
		return "", fmt.Errorf("the certificate already exists: %s", certID)
	}

	cert := &Cert{
		CertID: certID,
		Unit:   unit,
		Name:   name,
		NID:    nid,
	}

	certAsBytes, err := json.Marshal(cert)

	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(certID, certAsBytes)

	if err != nil {
		return "", err
	}

	idNameIndexKey, err := ctx.GetStub().CreateCompositeKey(id_index, []string{cert.NID, cert.CertID})
	if err != nil {
		return "", err
	}

	value := []byte{0x00}

	return ctx.GetStub().GetTxID(), ctx.GetStub().PutState(idNameIndexKey, value)
}

func (c *CertContract) ReadCert(ctx contractapi.TransactionContextInterface, certID string) (*Cert, error) {
	certAsBytes, err := ctx.GetStub().GetState(certID)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if certAsBytes == nil {
		return nil, fmt.Errorf("certificate %s does not exist", certID)
	}

	cert := new(Cert)
	_ = json.Unmarshal(certAsBytes, cert)

	return cert, nil
}

func (c *CertContract) HashExists(ctx contractapi.TransactionContextInterface, certID string) (bool, error) {
	hashRecordAsBytes, err := ctx.GetStub().GetState(certID)

	if err != nil {
		return false, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	return hashRecordAsBytes != nil, nil
}

func (c *CertContract) GetCertByPartialKey(ctx contractapi.TransactionContextInterface, nid string) ([]*Cert, error) {
	idNameIndexIterator, err := ctx.GetStub().GetStateByPartialCompositeKey(id_index, []string{nid})

	if err != nil {
		return nil, err
	}
	defer idNameIndexIterator.Close()

	var certs []*Cert

	for idNameIndexIterator.HasNext() {
		responseRange, err := idNameIndexIterator.Next()

		if err != nil {
			return nil, fmt.Errorf("failed to read the next value from the iterator. %s", err.Error())
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)

		if err != nil {
			return nil, fmt.Errorf("failed to split the composite key. %s", err.Error())
		}

		certID := compositeKeyParts[1]

		cert, err := c.ReadCert(ctx, certID)

		if err != nil {
			return nil, err
		}

		certs = append(certs, cert)
	}

	return certs, nil
}

func main() {
	contract := new(CertContract)

	certChaincode, err := contractapi.NewChaincode(contract)
	if err != nil {
		log.Panicf("Error creating cert chaincode: %v", err)
	}

	if err := certChaincode.Start(); err != nil {
		log.Panicf("Error creating cert chaincode: %v", err)
	}
}
