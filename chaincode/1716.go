package main

import (
	"errors"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	"poc_kinder/contract/model"
)

type AuthService struct {
	basicRepo *BasicRepository
}

func NewAuthService(stub shim.ChaincodeStubInterface) *AuthService {
	return &AuthService{basicRepo: &BasicRepository{Stub: stub}}
}

func (service *AuthService) GetUser() (*model.User, error) {
	// Get the client ID object
	id, err := cid.New(service.basicRepo.Stub)
	if err != nil {
		return nil, err
	}
	mspid, err := id.GetMSPID()
	if err != nil {
		return nil, err
	}

	personId, found, err := id.GetAttributeValue("hf.EnrollmentID")
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.New("id attr was not found")
	}

	return model.NewUser(personId, mspid), nil
}
