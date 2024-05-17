/*******************************************************************************
 * IBM Confidential
 *
 * OCO Source Materials
 *
 * (c) Copyright IBM Corporation 2022 All Rights Reserved.
 *
 * The source code for this program is not published or otherwise
 * divested of its trade secrets, irrespective of what has been
 * deposited with the U.S. Copyright Office.
 *******************************************************************************/

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const getStateError = "world state get error"
const CONSENT_ID_1 = "consent-1"
const CONSENT_ID_2 = "consent-2"

type MockStub struct {
	shim.ChaincodeStubInterface
	mock.Mock
}

func (ms *MockStub) GetState(pKey string) ([]byte, error) {
	args := ms.Called(pKey)

	return args.Get(0).([]byte), args.Error(1)
}

func (ms *MockStub) PutState(pKey string, value []byte) error {
	args := ms.Called(pKey, value)

	return args.Error(0)
}

func (ms *MockStub) DelState(pKey string) error {
	args := ms.Called(pKey)

	return args.Error(0)
}

type MockContext struct {
	contractapi.TransactionContextInterface
	mock.Mock
}

func (mc *MockContext) GetStub() shim.ChaincodeStubInterface {
	args := mc.Called()

	return args.Get(0).(*MockStub)
}

func configureStub() (*MockContext, *MockStub) {
	var nilBytes []byte

	consent1 := generateMockConsent(CONSENT_ID_1)
	consent1Bytes, _ := json.Marshal(consent1)

	consent2 := generateMockFHIRConsent(CONSENT_ID_2)
	consent2Bytes, _ := json.Marshal(consent2)

	ms := new(MockStub)
	ms.On("GetState", "statebad").Return(nilBytes, errors.New(getStateError))
	ms.On("GetState", "missingkey").Return(nilBytes, nil)
	ms.On("GetState", "existingkey").Return([]byte("some value"), nil)
	ms.On("GetState", CONSENT_ID_1).Return(consent1Bytes, nil)
	ms.On("GetState", CONSENT_ID_2).Return(consent2Bytes, nil)
	ms.On("PutState", mock.AnythingOfType("string"), mock.AnythingOfType("[]uint8")).Return(nil)
	ms.On("DelState", mock.AnythingOfType("string")).Return(nil)

	mc := new(MockContext)
	mc.On("GetStub").Return(ms)
	return mc, ms
}

func TestPing(t *testing.T) {
	ctx, _ := configureStub()
	contract := new(ConsentContract)

	_, err := contract.Ping(ctx)
	if err != nil {
		t.Error("Ping method failed")
	}

}

func TestConsentExists(t *testing.T) {
	var exists bool
	var err error

	ctx, _ := configureStub()
	contract := new(ConsentContract)

	exists, err = contract.ConsentExists(ctx, "statebad")
	assert.EqualError(t, err, getStateError)
	assert.False(t, exists, "should return false")

	exists, err = contract.ConsentExists(ctx, "missingkey")
	assert.Nil(t, err, "should not return error when can read from world state but no value for key")
	assert.False(t, exists, "should return false when no value for key in world state")

	exists, err = contract.ConsentExists(ctx, "existingkey")
	assert.Nil(t, err, "should not return error when can read from world state and value exists for key")
	assert.True(t, exists, "should return true when value for key in world state")
}

func TestCreateConsent(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	contract := new(ConsentContract)

	mockConsent := generateMockConsent(CONSENT_ID_1)
	mockFHIRConsent := generateMockFHIRConsent(CONSENT_ID_2)

	err = contract.CreateConsent(ctx, "statebad", mockConsent.PatientID, mockConsent.ServiceID, mockConsent.TenantID, mockConsent.DatatypeIDs, mockConsent.ConsentOption, mockConsent.Creation, mockConsent.Expiration, mockConsent.FHIRResourceID, mockConsent.FHIRResourceVersion, mockConsent.FHIRPolicy, mockConsent.FHIRStatus, mockConsent.FHIRProvisionType, mockConsent.FHIRProvisionAction, mockConsent.FHIRPerformerIDSystem, mockConsent.FHIRPerformerIDValue, mockConsent.FHIRPerformerDisplay, mockConsent.FHIRRecipientIDSystem, mockConsent.FHIRRecipientIDValue, mockConsent.FHIRRecipientDisplay)
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should return error")

	err = contract.CreateConsent(ctx, "existingkey", mockConsent.PatientID, mockConsent.ServiceID, mockConsent.TenantID, mockConsent.DatatypeIDs, mockConsent.ConsentOption, mockConsent.Creation, mockConsent.Expiration, mockConsent.FHIRResourceID, mockConsent.FHIRResourceVersion, mockConsent.FHIRPolicy, mockConsent.FHIRStatus, mockConsent.FHIRProvisionType, mockConsent.FHIRProvisionAction, mockConsent.FHIRPerformerIDSystem, mockConsent.FHIRPerformerIDValue, mockConsent.FHIRPerformerDisplay, mockConsent.FHIRRecipientIDSystem, mockConsent.FHIRRecipientIDValue, mockConsent.FHIRRecipientDisplay)
	assert.EqualError(t, err, "the Consent existingkey already exists", "should return error when consent already exists")

	missingConsent := generateMockConsent("missingkey")
	missingBytes, _ := json.Marshal(missingConsent)
	err = contract.CreateConsent(ctx, "missingkey", missingConsent.PatientID, missingConsent.ServiceID, missingConsent.TenantID, missingConsent.DatatypeIDs, missingConsent.ConsentOption, missingConsent.Creation, missingConsent.Expiration, missingConsent.FHIRResourceID, missingConsent.FHIRResourceVersion, missingConsent.FHIRPolicy, missingConsent.FHIRStatus, missingConsent.FHIRProvisionType, missingConsent.FHIRProvisionAction, missingConsent.FHIRPerformerIDSystem, missingConsent.FHIRPerformerIDValue, missingConsent.FHIRPerformerDisplay, missingConsent.FHIRRecipientIDSystem, missingConsent.FHIRRecipientIDValue, missingConsent.FHIRRecipientDisplay)
	assert.Nil(t, err, "should not return error")
	stub.AssertCalled(t, "PutState", "missingkey", missingBytes)

	err = contract.CreateConsent(ctx, "", mockConsent.PatientID, mockConsent.ServiceID, mockConsent.TenantID, mockConsent.DatatypeIDs, mockConsent.ConsentOption, mockConsent.Creation, mockConsent.Expiration, mockConsent.FHIRResourceID, mockConsent.FHIRResourceVersion, mockConsent.FHIRPolicy, mockConsent.FHIRStatus, mockConsent.FHIRProvisionType, mockConsent.FHIRProvisionAction, mockConsent.FHIRPerformerIDSystem, mockConsent.FHIRPerformerIDValue, mockConsent.FHIRPerformerDisplay, mockConsent.FHIRRecipientIDSystem, mockConsent.FHIRRecipientIDValue, mockConsent.FHIRRecipientDisplay)
	assert.EqualError(t, err, "CreateConsent: ERROR key minimum length is = 4", "should return error when consent is missing key")
	stub.AssertNotCalled(t, "PutState")

	err = contract.CreateConsent(ctx, mockConsent.ConsentID, mockConsent.PatientID, mockConsent.ServiceID, mockConsent.TenantID, mockConsent.DatatypeIDs, mockConsent.ConsentOption, mockConsent.Creation, mockConsent.Expiration, mockConsent.FHIRResourceID, mockConsent.FHIRResourceVersion, mockConsent.FHIRPolicy, mockConsent.FHIRStatus, mockConsent.FHIRProvisionType, mockConsent.FHIRProvisionAction, mockConsent.FHIRPerformerIDSystem, mockConsent.FHIRPerformerIDValue, mockConsent.FHIRPerformerDisplay, mockConsent.FHIRRecipientIDSystem, mockConsent.FHIRRecipientIDValue, mockConsent.FHIRRecipientDisplay)
	assert.EqualError(t, err, "the Consent consent-1 already exists", "should error when consent already exists")

	err = contract.CreateConsent(ctx, mockFHIRConsent.ConsentID, mockFHIRConsent.PatientID, mockFHIRConsent.ServiceID, mockFHIRConsent.TenantID, mockFHIRConsent.DatatypeIDs, mockFHIRConsent.ConsentOption, mockFHIRConsent.Creation, mockFHIRConsent.Expiration, mockFHIRConsent.FHIRResourceID, mockFHIRConsent.FHIRResourceVersion, mockFHIRConsent.FHIRPolicy, mockFHIRConsent.FHIRStatus, mockFHIRConsent.FHIRProvisionType, mockFHIRConsent.FHIRProvisionAction, mockFHIRConsent.FHIRPerformerIDSystem, mockFHIRConsent.FHIRPerformerIDValue, mockFHIRConsent.FHIRPerformerDisplay, mockFHIRConsent.FHIRRecipientIDSystem, mockFHIRConsent.FHIRRecipientIDValue, mockFHIRConsent.FHIRRecipientDisplay)
	assert.EqualError(t, err, "the Consent consent-2 already exists", "should error when consent already exists")
}

func TestReadConsent(t *testing.T) {
	log.Print("TestReadConsent: enter")
	defer log.Print("TestReadConsent: exit")

	var consent *Consent
	var err error

	ctx, _ := configureStub()
	contract := new(ConsentContract)

	consent, err = contract.ReadConsent(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should error when exists errors when reading")
	assert.Nil(t, consent, "should not return Consent when exists errors when reading")

	consent, err = contract.ReadConsent(ctx, "missingkey")
	assert.EqualError(t, err, "the Consent missingkey does not exist", "should error when exists returns true when reading")
	assert.Nil(t, consent, "should not return Consent when key does not exist in world state when reading")

	consent, err = contract.ReadConsent(ctx, "existingkey")
	assert.EqualError(t, err, "could not unmarshal world state data to type Consent", "should error when data in key is not Consent")
	assert.Nil(t, consent, "should not return Consent when data in key is not of type Consent")

	consent, err = contract.ReadConsent(ctx, CONSENT_ID_1)
	expectedConsent := generateMockConsent(CONSENT_ID_1)

	assert.Nil(t, err, "should not return error when Consent exists in world state when reading")
	assert.Equal(t, expectedConsent, consent, "should return deserialized Consent from world state")

}

func TestUpdateConsent(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	contract := new(ConsentContract)
	mockConsent := generateMockConsent(CONSENT_ID_1)
	mockFHIRConsent := generateMockFHIRConsent(CONSENT_ID_2)

	err = contract.UpdateConsent(ctx, "statebad", mockConsent.PatientID, mockConsent.ServiceID, mockConsent.TenantID, mockConsent.DatatypeIDs, mockConsent.ConsentOption, mockConsent.Creation, mockConsent.Expiration, mockConsent.FHIRResourceID, mockConsent.FHIRResourceVersion, mockConsent.FHIRPolicy, mockConsent.FHIRStatus, mockConsent.FHIRProvisionType, mockConsent.FHIRProvisionAction, mockConsent.FHIRPerformerIDSystem, mockConsent.FHIRPerformerIDValue, mockConsent.FHIRPerformerDisplay, mockConsent.FHIRRecipientIDSystem, mockConsent.FHIRRecipientIDValue, mockConsent.FHIRRecipientDisplay)
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should error when exists errors when updating")

	err = contract.UpdateConsent(ctx, "missingkey", mockConsent.PatientID, mockConsent.ServiceID, mockConsent.TenantID, mockConsent.DatatypeIDs, mockConsent.ConsentOption, mockConsent.Creation, mockConsent.Expiration, mockConsent.FHIRResourceID, mockConsent.FHIRResourceVersion, mockConsent.FHIRPolicy, mockConsent.FHIRStatus, mockConsent.FHIRProvisionType, mockConsent.FHIRProvisionAction, mockConsent.FHIRPerformerIDSystem, mockConsent.FHIRPerformerIDValue, mockConsent.FHIRPerformerDisplay, mockConsent.FHIRRecipientIDSystem, mockConsent.FHIRRecipientIDValue, mockConsent.FHIRRecipientDisplay)
	assert.EqualError(t, err, "the Consent missingkey does not exist", "should error when exists returns true when updating")

	err = contract.UpdateConsent(ctx, mockConsent.ConsentID, mockConsent.PatientID, mockConsent.ServiceID, mockConsent.TenantID, mockConsent.DatatypeIDs, mockConsent.ConsentOption, mockConsent.Creation, mockConsent.Expiration, mockConsent.FHIRResourceID, mockConsent.FHIRResourceVersion, mockConsent.FHIRPolicy, mockConsent.FHIRStatus, mockConsent.FHIRProvisionType, mockConsent.FHIRProvisionAction, mockConsent.FHIRPerformerIDSystem, mockConsent.FHIRPerformerIDValue, mockConsent.FHIRPerformerDisplay, mockConsent.FHIRRecipientIDSystem, mockConsent.FHIRRecipientIDValue, mockConsent.FHIRRecipientDisplay)

	expectedConsent := generateMockConsent(CONSENT_ID_1)
	expectedConsentBytes, _ := json.Marshal(expectedConsent)
	assert.Nil(t, err, "should not return error when Consent exists in world state when updating")
	stub.AssertCalled(t, "PutState", CONSENT_ID_1, expectedConsentBytes)

	err = contract.UpdateConsent(ctx, mockFHIRConsent.ConsentID, mockFHIRConsent.PatientID, mockFHIRConsent.ServiceID, mockFHIRConsent.TenantID, mockFHIRConsent.DatatypeIDs, mockFHIRConsent.ConsentOption, mockFHIRConsent.Creation, mockFHIRConsent.Expiration, mockFHIRConsent.FHIRResourceID, mockFHIRConsent.FHIRResourceVersion, mockFHIRConsent.FHIRPolicy, mockFHIRConsent.FHIRStatus, mockFHIRConsent.FHIRProvisionType, mockFHIRConsent.FHIRProvisionAction, mockFHIRConsent.FHIRPerformerIDSystem, mockFHIRConsent.FHIRPerformerIDValue, mockFHIRConsent.FHIRPerformerDisplay, mockFHIRConsent.FHIRRecipientIDSystem, mockFHIRConsent.FHIRRecipientIDValue, mockFHIRConsent.FHIRRecipientDisplay)

	expectedFHIRConsent := generateMockFHIRConsent(CONSENT_ID_2)
	expectedFHIRConsentBytes, _ := json.Marshal(expectedFHIRConsent)
	assert.Nil(t, err, "should not return error when Consent exists in world state when updating")
	stub.AssertCalled(t, "PutState", CONSENT_ID_2, expectedFHIRConsentBytes)
}

func TestDeleteConsent(t *testing.T) {
	var err error

	ctx, stub := configureStub()
	contract := new(ConsentContract)

	err = contract.DeleteConsent(ctx, "statebad")
	assert.EqualError(t, err, fmt.Sprintf("could not read from world state. %s", getStateError), "should error when exists errors")

	err = contract.DeleteConsent(ctx, "missingkey")
	assert.EqualError(t, err, "the Consent missingkey does not exist", "should error when exists returns true when deleting")

	err = contract.DeleteConsent(ctx, CONSENT_ID_1)
	assert.Nil(t, err, "should not return error when Consent exists in world state when deleting")
	stub.AssertCalled(t, "DelState", CONSENT_ID_1)
}

func generateMockFHIRConsent(pConsentID string) *Consent {
	mockConsent := new(Consent)

	startTime := time.Now()
	expirationTime := startTime.AddDate(1, 0, 0) // year, month, day
	datatypeIDs := []string{"datatypeID1", "datatypeID2"}

	mockConsent.ConsentID = pConsentID
	mockConsent.PatientID = "Patient1"
	mockConsent.TenantID = "TenantID1"
	mockConsent.DatatypeIDs = datatypeIDs
	mockConsent.Creation = startTime.Unix()
	mockConsent.Expiration = expirationTime.Unix()
	mockConsent.FHIRResourceID = "consent-001"
	mockConsent.FHIRResourceVersion = "1"
	mockConsent.FHIRPolicy = "regular"
	mockConsent.FHIRStatus = "active"
	mockConsent.FHIRProvisionType = "permit"
	mockConsent.FHIRProvisionAction = "disclose"
	mockConsent.FHIRPerformerIDSystem = "http://terminology.hl7.org/CodeSystem/v3-ParticipationType"
	mockConsent.FHIRPerformerIDValue = "0ba43008-1be2-4034-b50d-b76ff0110eae"
	mockConsent.FHIRPerformerDisplay = "Old Payer"
	mockConsent.FHIRRecipientIDSystem = "http://terminology.hl7.org/CodeSystem/v3-ParticipationType"
	mockConsent.FHIRRecipientIDValue = "93a4bb61-4cc7-469b-bf1b-c9cc24f8ace0"
	mockConsent.FHIRRecipientDisplay = "New Payer"

	return mockConsent
}

func generateMockConsent(pConsentID string) *Consent {
	mockConsent := new(Consent)

	startTime := time.Now()
	expirationTime := startTime.AddDate(1, 0, 0) // year, month, day
	datatypeIDs := []string{"datatypeID1", "datatypeID2"}

	mockConsent.ConsentID = pConsentID
	mockConsent.PatientID = "Patient1"
	mockConsent.ServiceID = "ServiceID1"
	mockConsent.TenantID = "TenantID1"
	mockConsent.DatatypeIDs = datatypeIDs
	mockConsent.ConsentOption = []string{"read, write"}
	mockConsent.Creation = startTime.Unix()
	mockConsent.Expiration = expirationTime.Unix()

	return mockConsent
}
