package main

import (
	// "github.com/hyperledger/fabric-contract-api-go/contractapi"
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

// type SmartContract struct {
// 	contractapi.Contract
// }

type User struct {
	UserUid         string `json:"user_uid,omitempty"  bson:"user_uid"  form:"user_uid"  binding:"user_uid"`
	FirstName       string `json:"first_name,omitempty"  bson:"first_name"  form:"first_name"  binding:"first_name"`
	LastName        string `json:"last_name,omitempty"  bson:"last_name"  form:"last_name"  binding:"last_name"`
	AddressLine1    string `json:"address_line_1,omitempty"  bson:"address_line_1"  form:"address_line_1"  binding:"address_line_1"`
	AddressLine2    string `json:"address_line_2,omitempty"  bson:"address_line_2"  form:"address_line_2"  binding:"address_line_2"`
	City            string `json:"city,omitempty"  bson:"city"  form:"city"  binding:"city"`
	Province        string `json:"province,omitempty"  bson:"province"  form:"province"  binding:"province"`
	PostalCode      int    `json:"postal_code,omitempty"  bson:"postal_code"  form:"postal_code"  binding:"postal_code"`
	Ttl             string `json:"ttl,omitempty"  bson:"ttl"  form:"ttl"  binding:"ttl"`
	Nik             string `json:"nik,omitempty"  bson:"nik"  form:"nik"  binding:"nik"`
	IdCard          string `json:"idcard,omitempty"  bson:"id_card"  form:"id_card"  binding:"id_card"`
	BusinessLicense string `json:"business_license,omitempty"  bson:"business_license"  form:"business_license"  binding:"business_license"`
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating asset-transfer-basic chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting asset-transfer-basic chaincode: %v", err)
	}
}

func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, userUid string, firstName string, lastName string,
	addressLine1 string, addressLine2 string, city string, province string, postalCode int, ttl string,
	nik string, idcard string, businessLicense string) error {
	exists, err := s.AssetExists(ctx, userUid)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the user %s is already exist", userUid)
	}

	user := User{
		UserUid:         userUid,
		FirstName:       firstName,
		LastName:        lastName,
		AddressLine1:    addressLine1,
		AddressLine2:    addressLine2,
		City:            city,
		Province:        province,
		PostalCode:      postalCode,
		Ttl:             ttl,
		Nik:             nik,
		IdCard:          idcard,
		BusinessLicense: businessLicense,
	}

	assetJson, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(userUid, assetJson)
}

func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, userUid string) (*User, error) {
	assetJson, err := ctx.GetStub().GetState(userUid)
	if err != nil {
		return nil, fmt.Errorf("faield to read from state database: %v", err)
	}
	if assetJson == nil {
		return nil, fmt.Errorf("the user %s does not exist", userUid)
	}

	var user User
	err = json.Unmarshal(assetJson, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, userUid string, firstName string, lastName string,
	addressLine1 string, addressLine2 string, city string, province string, postalCode int,
	ttl string, nik string, idcard string, businessLicense string) error {
	exists, err := s.AssetExists(ctx, userUid)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the user %s does not exist", userUid)
	}

	user := User{
		UserUid:         userUid,
		FirstName:       firstName,
		LastName:        lastName,
		AddressLine1:    addressLine1,
		AddressLine2:    addressLine2,
		City:            city,
		Province:        province,
		PostalCode:      postalCode,
		Ttl:             ttl,
		Nik:             nik,
		IdCard:          idcard,
		BusinessLicense: businessLicense,
	}

	assetJson, err := json.Marshal(user)
	if err != nil {
		return err
	}
	return ctx.GetStub().PutState(userUid, assetJson)
}

func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, userUid string) error {
	exists, err := s.AssetExists(ctx, userUid)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the user %s does not exist", userUid)
	}
	return ctx.GetStub().DelState(userUid)
}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*User, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*User
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset User
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}

func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, userUid string) (bool, error) {
	assetJson, err := ctx.GetStub().GetState(userUid)
	if err != nil {
		return false, fmt.Errorf("failed to read from state database: %v", err)
	}
	return assetJson != nil, nil
}
