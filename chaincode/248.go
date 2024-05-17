package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"securehealth.com/entities"
)

type PrimaryContract struct {
	contractapi.Contract
}

type QueryResult struct {
	Key    string             `json:"Key"`
	Record *entities.Patient
}

func (c *PrimaryContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("============= START : Initialize Ledger ===========")
	initPatients := []entities.Patient{
		{PatientID: "PAT0", FirstName: "Anouar1", LastName: "Oumaacha1", Age: 24, Password: "mypass1"},
		{PatientID: "PAT1", FirstName: "Anouar2", LastName: "Oumaacha2", Age: 23, Password: "mypass2"},
		{PatientID: "PAT2", FirstName: "Anouar3", LastName: "Oumaacha3", Age: 23, Password: "mypass3"},
		{PatientID: "PAT3", FirstName: "Anouar4", LastName: "Oumaacha4", Age: 18, Password: "mypass4"},
		{PatientID: "PAT4", FirstName: "Anouar5", LastName: "Oumaacha5", Age: 28, Password: "mypass5"},
	}

	for i, patient := range initPatients {
		patientBytes, err := json.Marshal(patient)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState("PAT"+strconv.Itoa(i), patientBytes)
		if err != nil {
			return err
		}

		fmt.Println("Added <--> ", patient)
	}

	fmt.Println("============= END : Initialize Ledger ===========")
	return nil
}

func (c *PrimaryContract) GrantAccessToDoctor(ctx contractapi.TransactionContextInterface, patientId string, doctorId string) error {
	exists, err := c.PatientExists(ctx, patientId)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("The patient " + patientId + " does not exist")
	}
	patientBytes, err := ctx.GetStub().GetState(patientId)
	if err != nil {
		return err
	}

	patient := new(entities.Patient)
	err = json.Unmarshal(patientBytes, patient)
	if err != nil {
		return err
	}

	patient.PermissionGranted = append(patient.PermissionGranted, doctorId)
	patient.ChangedBy = patientId

	patientBytes, err = json.Marshal(patient)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(patientId, patientBytes)
	if err != nil {
		return err
	}

	return nil
}

func (c *PrimaryContract) ReadPatient(ctx contractapi.TransactionContextInterface, patientID string) (*entities.Patient, error) {
	exists, err := c.PatientExists(ctx, patientID)
	if err != nil {
		return nil,err
	}
	if !exists {
		return nil,errors.New("The patient " + patientID + " does not exist")
	}

	// patientID is the key
	patientBytes, err := ctx.GetStub().GetState(patientID)
	if err != nil {
		return nil,err
	}

	patient := new(entities.Patient)
	err = json.Unmarshal(patientBytes, patient)
	if err != nil {
		return nil,err
	}

	return patient, nil
}

func (c *PrimaryContract) PatientExists(ctx contractapi.TransactionContextInterface, patientID string) (bool, error) {
	patientBytes, err := ctx.GetStub().GetState(patientID)
	if err != nil {
		return false, err
	}

	return patientBytes != nil && len(patientBytes) > 0, nil
}

func (c *PrimaryContract) QueryAllPatients(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		patient := new(entities.Patient)
		err = json.Unmarshal(queryResponse.Value, patient)
		if err != nil {
			return nil, err
		}

		queryResult := QueryResult{Key: queryResponse.Key, Record: patient}
		results = append(results, queryResult)
	}

	return results, nil
}

func (c *PrimaryContract) DeletePatient(ctx contractapi.TransactionContextInterface, patientId string) error {
	exists, err := c.PatientExists(ctx, patientId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("The patient %s does not exist", patientId)
	}

	err = ctx.GetStub().DelState(patientId)
	if err != nil {
		return err
	}
	err = c.SetPatientAsDeleted(ctx, patientId)
	if err != nil {
		return err
	}
	return nil
}

func (c *PrimaryContract) GetPatientPassword(ctx contractapi.TransactionContextInterface, patientID string) (string, error) {
	patient, err := c.ReadPatient(ctx, patientID)
	if err != nil {
		return "",err
	}
	password := patient.Password

	return password,nil
}

func (c *PrimaryContract) UpdatePatient(ctx contractapi.TransactionContextInterface, patientID string, firstName string, lastName string, age int) error {
        exists, err := c.PatientExists(ctx, patientID)
        if err != nil {
                return err
        }
        if !exists {
                return fmt.Errorf("The patient %s does not exist", patientID)
        }

        patient, err := c.ReadPatient(ctx, patientID)
        if err != nil {
                return err
        }

        // Update the patient details
        patient.FirstName = firstName
        patient.LastName = lastName
        patient.Age = age

        // Marshal the updated patient object into JSON bytes
        updatedPatientBytes, err := json.Marshal(patient)
        if err != nil {
                return err
        }

        // Update the state on the ledger with the updated patient object
        err = ctx.GetStub().PutState(patientID, updatedPatientBytes)
        if err != nil {
                return err
        }

        return nil
}

func (c *PrimaryContract) SetPatientAsDeleted(ctx contractapi.TransactionContextInterface, patientID string) error {
        exists, err := c.PatientExists(ctx, patientID)
        if err != nil {
                return err
        }
        if !exists {
                return fmt.Errorf("The patient %s does not exist", patientID)
        }

        patient, err := c.ReadPatient(ctx, patientID)
        if err != nil {
                return err
        }

        // Set the IsDeleted flag to true
        patient.IsDeleted = true

        // Marshal the updated patient object into JSON bytes
        updatedPatientBytes, err := json.Marshal(patient)
        if err != nil {
                return err
        }

        // Update the state on the ledger with the updated patient object
        err = ctx.GetStub().PutState(patientID, updatedPatientBytes)
        if err != nil {
                return err
        }

        return nil
}

func (c *PrimaryContract) CreatePatient(ctx contractapi.TransactionContextInterface, patientID string, firstName string, lastName string, age int, password string) error {
    exists, err := c.PatientExists(ctx, patientID)
    if err != nil {
      return err
    }
    if exists {
      return fmt.Errorf("the patient %s already exists", patientID)
    }

    patient := entities.Patient{
		PatientID:        patientID,
		FirstName:        firstName,
		LastName:         lastName,
		Age:              age,
		Password:	  password,
    }
    patientJSON, err := json.Marshal(patient)
    if err != nil {
      return err
    }

    return ctx.GetStub().PutState(patientID, patientJSON)
}

func (c *PrimaryContract) GetLatestPatientID(ctx contractapi.TransactionContextInterface) (string, error) {
    allResults, err := c.QueryAllPatients(ctx)
    if err != nil {
        return "", err
    }

    if len(allResults) == 0 {
        return "", errors.New("No patients found")
    }

    latestPatient := allResults[len(allResults)-1]
    return latestPatient.Record.PatientID, nil
}

func (c *PrimaryContract) QueryPatientsByLastName(ctx contractapi.TransactionContextInterface, lastName string) ([]*entities.Patient, error) {
    queryString := map[string]interface{}{
        "selector": map[string]interface{}{
            "docType": "patient",
            "lastName": lastName,
        },
    }
    queryStringStr, _ := json.Marshal(queryString)
    queryResult, err := ctx.GetStub().GetQueryResult(string(queryStringStr))
    if err != nil {
        return nil, err
    }

    patients := []*entities.Patient{}
    for queryResult.HasNext() {
        queryResponse, err := queryResult.Next()
        if err != nil {
            return nil, err
        }

        patient := new(entities.Patient)
        err = json.Unmarshal(queryResponse.Value, patient)
        if err != nil {
            return nil, err
        }

        patients = append(patients, patient)
    }

    return patients, nil
}

func main() {

        chaincode, err := contractapi.NewChaincode(new(PrimaryContract))

        if err != nil {
                fmt.Printf("Error create fabcar chaincode: %s", err.Error())
                return
        }

        if err := chaincode.Start(); err != nil {
                fmt.Printf("Error starting fabcar chaincode: %s", err.Error())
        }
}
