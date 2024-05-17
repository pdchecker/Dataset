package main

import (
	"errors"
	"encoding/json"

	"fmt"
	

	"strings"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
)

func insertData(stub *hypConnect, key string, privateCollection string, data []byte) error {
	myMSP, err2 := cid.GetMSPID(stub.Connection)
	if err2 != nil {
		return  err2
	}
	fmt.Println("MSP:" + myMSP)

	err := stub.Connection.PutPrivateData(privateCollection, key, data)
	if err != nil {
		return err
	}
	event := eventDataFormat{}
	event.Key = key
	event.Collection = privateCollection
	stub.EventList = stub.AddEvent(event)

	fmt.Println("Successfully Put State for Key: " + key + " and Private Collection " + privateCollection)
	return nil
}

func deleteData(stub *hypConnect, key string, privateCollection string) error {

	err := stub.Connection.DelPrivateData(privateCollection, key)
	if err != nil {
		return err
	}
	fmt.Println("Successfully Delete for Key: " + key + " and Private Collection " + privateCollection)

	event := eventDataFormat{}
	event.Key = key
	event.Collection = privateCollection
	stub.EventList = stub.AddEvent(event)

	fmt.Println("Successfully Put State for Key: " + key + " and Private Collection " + privateCollection)
	return nil
}

func fetchData(stub hypConnect, key string, privateCollection string) ([]byte, error) {
	bytes, err := stub.Connection.GetPrivateData(privateCollection, key)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func getArguments(stub shim.ChaincodeStubInterface) ([]string, error) {
	transMap, err := stub.GetTransient()
	if err != nil {
		return nil, err
	}
	if _, ok := transMap["PrivateArgs"]; !ok {
		return nil, errors.New("PrivateArgs must be a key in the transient map")
	}
	fmt.Println("Arguments: %v", transMap)
	generalInput := string(transMap["PrivateArgs"])
	retVal := strings.Split(generalInput, "|")
	return retVal, nil
}

func getOrgTypeByMSP(stub shim.ChaincodeStubInterface, MSP string) (string, error) {

	MSPMappingAsBytes, err := stub.GetState("MSPMapping")
	if err != nil {
		return "", err
	}

	if err != nil {
		fmt.Println("MSPMapping - Failed to get state MSP mapping information." + err.Error())
		return "", err
	} else if MSPMappingAsBytes != nil {
		fmt.Println("MSPMapping - This data Fetched from Transactions.")
		var MSPListUnmarshaled []MSPList

		err := json.Unmarshal(MSPMappingAsBytes, &MSPListUnmarshaled)
		if err != nil {
			fmt.Println("MSPMapping-Failed to UnMarshal state.")
			return "", err
		}
		fmt.Println("Unmarshaled: %v", MSPListUnmarshaled)
		for i := 0; i < len(MSPListUnmarshaled); i++ {
			if MSPListUnmarshaled[i].MSP == MSP {
				fmt.Println("OrgType for MSP " + MSP + " is " + MSPListUnmarshaled[i].OrgType)
				return MSPListUnmarshaled[i].OrgType, nil
			}
		}
	}
	return "", nil
}