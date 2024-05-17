package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type PersonAsset struct {
	ID                 string
	EmailAddress       string
	FirstName          string
	LastName           string
	AmountOfMoneyOwned float32
}

type Defect struct {
	Description string
	RepairPrice float32
}

type CarAsset struct {
	ID         string
	OwnerID    string
	Brand      string
	Model      string
	Color      string
	Price      float32
	Year       int
	DefectList []Defect
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {

	personAssets := []PersonAsset{
		{ID: "person1", FirstName: "Osoba1", LastName: "Osobic1", EmailAddress: "osoba1@osoba1.com", AmountOfMoneyOwned: 10000.10},
		{ID: "person2", FirstName: "Osoba2", LastName: "Osobic2", EmailAddress: "osoba2@osoba2.com", AmountOfMoneyOwned: 3000.10},
		{ID: "person3", FirstName: "Osoba3", LastName: "Osobic3", EmailAddress: "osoba3@osoba3.com", AmountOfMoneyOwned: 15000.10},
	}

	for _, personAsset := range personAssets {
		personAssetJSON, err := json.Marshal(personAsset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(personAsset.ID, personAssetJSON)
		if err != nil {
			return fmt.Errorf("failed to put persons to world state. %v", err)
		}
	}

	carAssets := []CarAsset{
		{ID: "car1", Brand: "Toyota", Model: "Yaris", Year: 2020, Color: "white", OwnerID: "person2", Price: 2222, DefectList: []Defect{
			{Description: "Probusena desna prednja guma", RepairPrice: 30},
		}},
		{ID: "car2", Brand: "Kia", Model: "Sportage", Year: 2021, Color: "blue", OwnerID: "person1", Price: 1111, DefectList: []Defect{
			{Description: "Probusena leva prednja guma", RepairPrice: 30},
			{Description: "Probusena desna prednja guma", RepairPrice: 50},
		}},
		{ID: "car3", Brand: "Hyundai", Model: "Tucson", Year: 2020, Color: "red", OwnerID: "person2", Price: 2222, DefectList: []Defect{
			{Description: "Probusena desna zadnja guma", RepairPrice: 100},
		}},
		{ID: "car4", Brand: "Ferrari", Model: "Ferrari", Year: 2021, Color: "green", OwnerID: "person1", Price: 1111, DefectList: []Defect{
			{Description: "Probusena leva zadnja guma", RepairPrice: 300},
			{Description: "Probusena desna zadnja guma", RepairPrice: 2000},
		}},
		{ID: "car5", Brand: "Fiat", Model: "Grand Punto", Year: 2023, Color: "blue", OwnerID: "person1", Price: 1111, DefectList: []Defect{
			{Description: "Probusena leva zadnja guma", RepairPrice: 1},
		}},
		{ID: "car6", Brand: "Citroen", Model: "C5", Year: 2018, Color: "grey", OwnerID: "person3", Price: 2222, DefectList: []Defect{}},
	}

	for _, carAsset := range carAssets {
		carAssetJSON, err := json.Marshal(carAsset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(carAsset.ID, carAssetJSON)
		if err != nil {
			return fmt.Errorf("failed to put cars to world state. %v", err)
		}

		indexName := "color~owner~ID"
		colorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{carAsset.Color, carAsset.OwnerID, carAsset.ID})
		if err != nil {
			return err
		}

		value := []byte{0x00}
		err = ctx.GetStub().PutState(colorOwnerIndexKey, value)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SmartContract) ReadPersonAsset(ctx contractapi.TransactionContextInterface, id string) (*PersonAsset, error) {
	personAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read person from world state: %v", err)
	}
	if personAssetJSON == nil {
		return nil, fmt.Errorf("the person asset %s does not exist", id)
	}

	var personAsset PersonAsset
	err = json.Unmarshal(personAssetJSON, &personAsset)
	if err != nil {
		return nil, err
	}

	return &personAsset, nil
}

func (s *SmartContract) ReadCarAsset(ctx contractapi.TransactionContextInterface, id string) (*CarAsset, error) {
	carAssetJSON, err := ctx.GetStub().GetState(id)
	if err != nil {
		return nil, fmt.Errorf("failed to read car from world state: %v", err)
	}
	if carAssetJSON == nil {
		return nil, fmt.Errorf("the car asset %s does not exist", id)
	}

	var carAsset CarAsset
	err = json.Unmarshal(carAssetJSON, &carAsset)
	if err != nil {
		return nil, err
	}

	return &carAsset, nil
}

func (s *SmartContract) GetCarsByColor(ctx contractapi.TransactionContextInterface, color string) ([]*CarAsset, error) {
	coloredCarIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color})
	if err != nil {
		return nil, err
	}

	defer coloredCarIter.Close()

	retVal := make([]*CarAsset, 0)

	for i := 0; coloredCarIter.HasNext(); i++ {
		responseRange, err := coloredCarIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		retCarID := compositeKeyParts[2]

		carAsset, err := s.ReadCarAsset(ctx, retCarID)
		if err != nil {
			return nil, err
		}

		retVal = append(retVal, carAsset)
	}

	return retVal, nil
}

func (s *SmartContract) GetCarsByColorAndOwner(ctx contractapi.TransactionContextInterface, color string, ownerID string) ([]*CarAsset, error) {

	coloredCarByOwnerIter, err := ctx.GetStub().GetStateByPartialCompositeKey("color~owner~ID", []string{color, ownerID})
	if err != nil {
		return nil, err
	}

	defer coloredCarByOwnerIter.Close()

	retVal := make([]*CarAsset, 0)

	for i := 0; coloredCarByOwnerIter.HasNext(); i++ {
		responseRange, err := coloredCarByOwnerIter.Next()
		if err != nil {
			return nil, err
		}

		_, compositeKeyParts, err := ctx.GetStub().SplitCompositeKey(responseRange.Key)
		if err != nil {
			return nil, err
		}

		retCarID := compositeKeyParts[2]

		carAsset, err := s.ReadCarAsset(ctx, retCarID)
		if err != nil {
			return nil, err
		}

		retVal = append(retVal, carAsset)
	}

	return retVal, nil
}

func (s *SmartContract) TransferCarAsset(ctx contractapi.TransactionContextInterface, id string, newOwnerID string, acceptDefect bool) (bool, error) {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return false, err
	}

	if carAsset.OwnerID == newOwnerID {
		return false, fmt.Errorf("%s already owner", newOwnerID)
	}

	seller, err := s.ReadPersonAsset(ctx, carAsset.OwnerID)
	if err != nil {
		return false, err
	}

	buyer, err := s.ReadPersonAsset(ctx, newOwnerID)
	if err != nil {
		return false, err
	}

	carPrice := float32(0)

	if carAsset.DefectList == nil || len(carAsset.DefectList) == 0 {
		carPrice = carAsset.Price
	} else if acceptDefect {
		defectPrice := float32(0)
		for _, carDefect := range carAsset.DefectList {
			defectPrice += carDefect.RepairPrice
		}
		carPrice = carAsset.Price - defectPrice
	} else {
		return false, fmt.Errorf("car is defected")
	}

	oldOwnerID := carAsset.OwnerID
	carAsset.OwnerID = newOwnerID

	if carPrice <= buyer.AmountOfMoneyOwned {
		seller.AmountOfMoneyOwned += carPrice
		buyer.AmountOfMoneyOwned -= carPrice
	} else {
		return false, fmt.Errorf("not enough money")
	}

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return false, err
	}

	sellerJSON, err := json.Marshal(seller)
	if err != nil {
		return false, err
	}

	buyerJSON, err := json.Marshal(buyer)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return false, err
	}
	err = ctx.GetStub().PutState(seller.ID, sellerJSON)
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().PutState(buyer.ID, buyerJSON)
	if err != nil {
		return false, err
	}

	indexName := "color~owner~ID"
	colorNewOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{carAsset.Color, newOwnerID, carAsset.ID})
	if err != nil {
		return false, err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(colorNewOwnerIndexKey, value)
	if err != nil {
		return false, err
	}

	colorOldOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{carAsset.Color, oldOwnerID, carAsset.ID})
	if err != nil {
		return false, err
	}

	err = ctx.GetStub().DelState(colorOldOwnerIndexKey)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SmartContract) AddCarDefect(ctx contractapi.TransactionContextInterface, id string, description string, repairPrice float32) error {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return err
	}

	newDefect := Defect{
		Description: description,
		RepairPrice: repairPrice,
	}

	carAsset.DefectList = append(carAsset.DefectList, newDefect)

	totalRepairPrice := float32(0)
	for _, carDefect := range carAsset.DefectList {
		totalRepairPrice += carDefect.RepairPrice
	}

	if totalRepairPrice > carAsset.Price {
		return ctx.GetStub().DelState(id)
	}

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return err
	}

	return nil
}

func (s *SmartContract) ChangeCarColor(ctx contractapi.TransactionContextInterface, id string, newColor string) (string, error) {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return "", err
	}

	oldColor := carAsset.Color
	carAsset.Color = newColor

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return "", err
	}

	indexName := "color~owner~ID"
	newColorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{newColor, carAsset.OwnerID, carAsset.ID})
	if err != nil {
		return "", err
	}

	value := []byte{0x00}
	err = ctx.GetStub().PutState(newColorOwnerIndexKey, value)
	if err != nil {
		return "", err
	}

	oldColorOwnerIndexKey, err := ctx.GetStub().CreateCompositeKey(indexName, []string{oldColor, carAsset.OwnerID, carAsset.ID})
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().DelState(oldColorOwnerIndexKey)
	if err != nil {
		return "", err
	}

	return oldColor, nil
}

func (s *SmartContract) RepairCar(ctx contractapi.TransactionContextInterface, id string) error {
	carAsset, err := s.ReadCarAsset(ctx, id)
	if err != nil {
		return err
	}

	personAsset, err := s.ReadPersonAsset(ctx, carAsset.OwnerID)
	if err != nil {
		return err
	}

	repairPriceSum := float32(0)
	for _, carDefect := range carAsset.DefectList {
		repairPriceSum += carDefect.RepairPrice
		if repairPriceSum > personAsset.AmountOfMoneyOwned {
			return fmt.Errorf("not enough owner resources")
		}
	}

	carAsset.DefectList = []Defect{}
	personAsset.AmountOfMoneyOwned -= repairPriceSum

	carAssetJSON, err := json.Marshal(carAsset)
	if err != nil {
		return err
	}

	personAssetJSON, err := json.Marshal(personAsset)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(id, carAssetJSON)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(personAsset.ID, personAssetJSON)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	assetChaincode, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		log.Panicf("Error creating chaincode: %v", err)
	}

	if err := assetChaincode.Start(); err != nil {
		log.Panicf("Error starting chaincode: %v", err)
	}
}
