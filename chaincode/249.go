package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type ParkingSlot struct {
	SlotID       string `json:"slot_id"`
	SlotNumber   string `json:"slot_number"`
	ProviderID   string `json:"provider_id"`
	IntegratorID string `json:"integrator_id"`
	OccupierID   string `json:"occupier_id"`
	Status       uint32 `json:"status"`
}

const (
	PARKING_SLOT_STATUS_FREE uint32 = iota
	PARKING_SLOT_STATUS_OCCUPIED
)

func (s *SmartContract) ParkingSlotExits(ctx contractapi.TransactionContextInterface, parkingSlotID string) error {
	parkingSlotJSON, err := ctx.GetStub().GetState(parkingSlotID)
	if err != nil {
		return err
	}

	if parkingSlotJSON != nil {
		return fmt.Errorf("parking slot %s already exists", parkingSlotID)
	}

	return nil
}

func (s *SmartContract) AddParkingSlot(ctx contractapi.TransactionContextInterface, parkingSlotNumber string) error {
	err := s.IsProvider(ctx)
	if err != nil {
		return err
	}

	clientID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	parkingSlotID, err := ctx.GetStub().CreateCompositeKey("parkingslot", []string{clientID, parkingSlotNumber})
	if err != nil {
		return err
	}

	err = s.ParkingSlotExits(ctx, parkingSlotID)
	if err != nil {
		return err
	}

	parkingSlot := ParkingSlot{
		SlotID:     parkingSlotID,
		ProviderID: clientID,
		SlotNumber: parkingSlotNumber,
		Status:     PARKING_SLOT_STATUS_FREE,
	}
	parkingSlotJSON, err := json.Marshal(parkingSlot)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(parkingSlotID, parkingSlotJSON)
}

func (s *SmartContract) ReadParkingSlot(ctx contractapi.TransactionContextInterface, parkingSlotID string) (*ParkingSlot, error) {
	parkingSlotJSON, err := ctx.GetStub().GetState(parkingSlotID)
	if err != nil {
		return nil, fmt.Errorf("failed to read parking slot from world state: %v", err)
	}
	if parkingSlotJSON == nil {
		return nil, fmt.Errorf("the parking slot with parking slot id %s does not exist", parkingSlotID)
	}

	parkingSlot := new(ParkingSlot)
	err = json.Unmarshal(parkingSlotJSON, parkingSlot)
	if err != nil {
		return nil, err
	}

	return parkingSlot, nil
}

func (s *SmartContract) GetAllParkingSlots(ctx contractapi.TransactionContextInterface) ([]*ParkingSlot, error) {
	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("parkingslot", []string{})
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var parkingSlots []*ParkingSlot
	for resultsIterator.HasNext() {
		parkingSlotJSON, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		parkingSlot := new(ParkingSlot)
		err = json.Unmarshal(parkingSlotJSON.Value, parkingSlot)
		if err != nil {
			return nil, err
		}

		parkingSlots = append(parkingSlots, parkingSlot)
	}

	return parkingSlots, nil
}

func (s *SmartContract) AssignParkingSlot(ctx contractapi.TransactionContextInterface,
	providerID string, slotNumber string, occupierID string) error {
	err := s.IsIntegrator(ctx)
	if err != nil {
		return err
	}

	clientID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	parkingSlotID, err := ctx.GetStub().CreateCompositeKey("parkingslot", []string{providerID, slotNumber})
	if err != nil {
		return err
	}

	parkingSlot, err := s.ReadParkingSlot(ctx, parkingSlotID)
	if err != nil {
		return err
	}

	if parkingSlot.Status != PARKING_SLOT_STATUS_FREE {
		return fmt.Errorf("failed to assign parking slot, parking slot is not free")
	}

	parkingSlot.Status = PARKING_SLOT_STATUS_OCCUPIED
	parkingSlot.IntegratorID = clientID
	parkingSlot.OccupierID = occupierID

	parkingSlotJSON, err := json.Marshal(parkingSlot)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(parkingSlotID, parkingSlotJSON)
	if err != nil {
		return err
	}

	return nil
}

// ReleaseParkingSlot is called by the parking provider owning the parkingSlotID
func (s *SmartContract) ReleaseParkingSlot(ctx contractapi.TransactionContextInterface,
	slotNumber string) error {
	err := s.IsProvider(ctx)
	if err != nil {
		return err
	}

	clientID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return err
	}

	parkingSlotID, err := ctx.GetStub().CreateCompositeKey("parkingslot", []string{clientID, slotNumber})
	if err != nil {
		return err
	}

	parkingSlot, err := s.ReadParkingSlot(ctx, parkingSlotID)
	if err != nil {
		return err
	}

	if parkingSlot.ProviderID != clientID {
		return fmt.Errorf("only provider of the parking slot can release the parking slot, owner = %s, client = %s", parkingSlot.ProviderID, clientID)
	}

	if parkingSlot.Status != PARKING_SLOT_STATUS_OCCUPIED {
		return fmt.Errorf("parking slot is not occupied")
	}

	parkingSlot.Status = PARKING_SLOT_STATUS_FREE
	parkingSlot.IntegratorID = ""
	parkingSlot.OccupierID = ""

	parkingSlotJSON, err := json.Marshal(parkingSlot)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(parkingSlotID, parkingSlotJSON)
	if err != nil {
		return err
	}

	return nil

}

func (s *SmartContract) GetMyParkingSlot(ctx contractapi.TransactionContextInterface) ([]*ParkingSlot, error) {
	err := s.IsProvider(ctx)
	if err != nil {
		return nil, err
	}

	clientID, err := ctx.GetClientIdentity().GetMSPID()
	if err != nil {
		return nil, err
	}

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("parkingslot", []string{clientID})
	if err != nil {
		return nil, err
	}

	parkingSlots := make([]*ParkingSlot, 0)
	for resultsIterator.HasNext() {
		parkingSlotResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		parkingSlot := new(ParkingSlot)
		err = json.Unmarshal(parkingSlotResult.Value, parkingSlot)
		if err != nil {
			return nil, err
		}

		parkingSlots = append(parkingSlots, parkingSlot)
	}

	return parkingSlots, nil
}

func (s *SmartContract) GetAvailableParkingSlotByProviderID(ctx contractapi.TransactionContextInterface, providerID string) ([]*ParkingSlot, error) {
	err := s.IsIntegrator(ctx)
	if err != nil {
		return nil, err
	}

	resultsIterator, err := ctx.GetStub().GetStateByPartialCompositeKey("parkingslot", []string{providerID})
	if err != nil {
		return nil, err
	}

	parkingSlots := make([]*ParkingSlot, 0)
	for resultsIterator.HasNext() {
		parkingSlotResult, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		parkingSlot := new(ParkingSlot)
		err = json.Unmarshal(parkingSlotResult.Value, parkingSlot)
		if err != nil {
			return nil, err
		}

		parkingSlots = append(parkingSlots, parkingSlot)
	}

	return parkingSlots, nil
}

func (s *SmartContract) GetParkingSlotByProviderIDAndSlotNumber(ctx contractapi.TransactionContextInterface, providerID, slotNumber string) (*ParkingSlot, error) {
	err := s.IsIntegrator(ctx)
	if err != nil {
		return nil, err
	}

	parkingSlotID, err := ctx.GetStub().CreateCompositeKey("parkingslot", []string{providerID, slotNumber})
	if err != nil {
		return nil, err
	}

	parkingSlot, err := s.ReadParkingSlot(ctx, parkingSlotID)
	if err != nil {
		return nil, err
	}

	return parkingSlot, nil
}
