package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides authentication for IoT Tracker
type DeviceStore struct {
	contractapi.Contract
}

// Device represents a device
type Device struct {
	DeviceID     string `json:"deviceID"`
	PasswordHash string `json:"password"`
	Role         string `json:"role"`
}

// Init initializes the device store
func (r *DeviceStore) Init(ctx contractapi.TransactionContextInterface) error {
	return nil
}

// RegisterDevice registers a device (avoid overwriting existing device)
func (r *DeviceStore) RegisterDevice(ctx contractapi.TransactionContextInterface, deviceID string, password string, role string) error {

	// hash the password
	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	if hash == "" || hash == " " || hash == "  " {
		return fmt.Errorf("password hash is empty")
	}

	device := Device{
		DeviceID:     deviceID,
		PasswordHash: string(hash),
		Role:         role,
	}

	// Check if device already exists
	deviceBytes, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return err
	}
	if deviceBytes != nil {
		return fmt.Errorf("device already exists")
	}

	deviceJSON, err := json.Marshal(device)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(deviceID, deviceJSON)
	if err != nil {
		return err
	}

	return nil
}

// AuthenticateDevice authenticates a device
func (r *DeviceStore) AuthenticateDevice(ctx contractapi.TransactionContextInterface, deviceID string, password string) (string, error) {

	device, err := ctx.GetStub().GetState(deviceID)
	if err != nil {
		return "", err
	}
	if device == nil {
		return "", fmt.Errorf("device does not exist")
	}

	hash := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	if hash == "" || hash == " " || hash == "  " {
		return "", fmt.Errorf("password hash is empty")
	}

	var deviceStruct Device
	err = json.Unmarshal(device, &deviceStruct)
	if err != nil {
		return "", err
	}

	if deviceStruct.PasswordHash != hash {
		return "", fmt.Errorf("incorrect password")
	}

	return deviceStruct.Role, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(DeviceStore))
	if err != nil {
		panic(err.Error())
	}

	if err := chaincode.Start(); err != nil {
		panic(err.Error())
	}
}
