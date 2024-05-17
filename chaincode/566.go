package main

import (
	"encoding/json"
	"fmt"

	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

// SmartContract provides functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

type BatchHistory struct {
	Batch_id             string              `json:"batchId"`
	CurrentStage         string              `json:"currentStage"`
	FarmerSeedBatch      FarmerSeedData      `json:"farmerSeedBatch"`
	InspectorSeedBatch   InspectorSeedData   `json:"inspectorSeedBatch"`
	TransporterSeedBatch TransporterSeedData `json:"transporterSeedBatch"`
	ExporterSeedBatch    ExporterSeedData    `json:"exporterSeedBatch"`
	ImporterSeedBatch    ImporterSeedData    `json:"importerSeedBatch"`
	RetailorSeedBatch    RetailorSeedData    `json:"retailorSeedBatch"`
}

type FarmerBasicData struct {
	Farmer_id   string           `json:"farmerID"`
	Name        string           `json:"name"`
	Address     string           `json:"address"`
	FarmerSeeds []FarmerSeedData `json:"farmerSeeds"`
}

type FarmerSeedData struct {
	Batch_id             string  `json:"batchId"`
	FarmerName           string  `json:"farmerName"`
	Seed_type            string  `json:"seedType"`
	PlaceGrow            string  `json:"placeGrow"`
	Cutting              string  `json:"cutting"`
	LandArea             string  `json:"landArea"`
	TotalQuantitySeed    float64 `json:"totalQuantitySeed"`
	SeedQualityType      string  `json:"seedQualityType"`
	FertilizerName       string  `json:"fertilizerName"`
	FertilizerQuantity   string  `json:"fertilizerQuantity"`
	PestisideName        string  `json:"pestisideName"`
	PestisideQuantity    string  `json:"pestisideQuantity"`
	ProduceQuantityTypeA float64 `json:"produceQuantityTypeA"`
	ProduceQuantityTypeB float64 `json:"produceQuantityTypeB"`
	ProduceQuantityTypeC float64 `json:"produceQuantityTypeC"`
	GrowDate             string  `json:"growDate"`
	CuttingDate          string  `json:"cuttingDate"`
}

type InspectorBasicData struct {
	Inspector_id   string              `json:"inspectorID"`
	Name           string              `json:"name"`
	Address        string              `json:"address"`
	InspectorSeeds []InspectorSeedData `json:"inspectorSeeds"`
}

type InspectorSeedData struct {
	Batch_id                     string `json:"batchId"`
	InspectorName                string `json:"inspectorName"`
	QualityVerified              bool   `json:"qualityVerified"`
	PestisideVerified            bool   `json:"pestisideVerified"`
	FertilizerVerified           bool   `json:"fertilizerVerified"`
	ProduceQuantityTypeAVerified bool   `json:"produceQuantityTypeAVerified"`
	ProduceQuantityTypeBVerified bool   `json:"produceQuantityTypeBVerified"`
	ProduceQuantityTypeCVerified bool   `json:"produceQuantityTypeCVerified"`
	InspectDate                  string `json:"inspectDate"`
}

type TransporterBasicData struct {
	Transporter_id   string                `json:"transporterID"`
	Name             string                `json:"name"`
	Address          string                `json:"address"`
	TransporterSeeds []TransporterSeedData `json:"transporterSeeds"`
}

type TransporterSeedData struct {
	Batch_id                 string  `json:"batchId"`
	TransporterName          string  `json:"transporterName"`
	Location                 string  `json:"location"`
	VehicleUsedTransporter   string  `json:"vehicleUsedTransporter"`
	VehicleNumberTransporter string  `json:"vehicleNumberTransporter"`
	Quantity                 float64 `json:"quantity"`
	TransportDate            string  `json:"transportDate"`
}

type ExporterBasicData struct {
	Exporter_id   string             `json:"exporterID"`
	Name          string             `json:"name"`
	Address       string             `json:"address"`
	ExporterSeeds []ExporterSeedData `json:"exporterSeeds"`
}

type ExporterSeedData struct {
	Batch_id              string `json:"batchId"`
	ExporterName          string `json:"exporterName"`
	PlaceKeptExporter     string `json:"placeKeptExporter"`
	ArrivalDateExporter   string `json:"arrivalDateExporter"`
	DestinationAddress    string `json:"destinationAddress"`
	VehicleUsedExporter   string `json:"vehicleUsedExporter"`
	VehicleNumberExporter string `json:"vehicleNumberExporter"`
	DepartureDate         string `json:"departureDate"`
}

type ImporterBasicData struct {
	Importer_id   string             `json:"importer_ID"`
	Name          string             `json:"name"`
	Address       string             `json:"address"`
	ImporterSeeds []ImporterSeedData `json:"importerSeeds"`
}

type ImporterSeedData struct {
	Batch_id            string `json:"batchId"`
	ImporterName        string `json:"importerName"`
	PlaceKeptImporter   string `json:"placeKeptImporter"`
	ArrivalDateImporter string `json:"arrivalDateImporter"`
}

type RetailorBasicData struct {
	Retailor_id   string             `json:"Retailor_ID"`
	Name          string             `json:"name"`
	Address       string             `json:"address"`
	RetailorSeeds []RetailorSeedData `json:"retailorSeeds"`
}

type RetailorSeedData struct {
	Batch_id            string `json:"batchId"`
	RetailorName        string `json:"retailorName"`
	PlaceKeptRetailor   string `json:"placeKeptImporter"`
	ArrivalDateRetailor string `json:"arrivalDateImporter"`
}

func (s *SmartContract) Init(ctx contractapi.TransactionContextInterface) error {
	fmt.Println("Initializing chaincode...")

    // Farmer init data:
	farmerSeed := FarmerSeedData{
		Batch_id:             "1",
		FarmerName:           "John Farmer",
		Seed_type:            "Wheat",
		PlaceGrow:            "Farm A",
		Cutting:              "Early",
		LandArea:             "50 acres",
		TotalQuantitySeed:    100.0,
		SeedQualityType:      "A",
		FertilizerName:       "SuperGrow",
		FertilizerQuantity:   "10 kg",
		PestisideName:        "BugOff",
		PestisideQuantity:    "5 liters",
		ProduceQuantityTypeA: 80.0,
		ProduceQuantityTypeB: 15.0,
		ProduceQuantityTypeC: 5.0,
		GrowDate:             "2023-01-01",
		CuttingDate:          "2023-03-01",
	}

	farmerSeedJSON, err := json.Marshal(farmerSeed)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("farmerSeed1", farmerSeedJSON); err != nil {
		return err
	}

    farmerBasicData := FarmerBasicData{
		Farmer_id: "F1",
		Name:      "John Farmer",
		Address:   "Farm A, Country X",
		FarmerSeeds: []FarmerSeedData{
			farmerSeed,
		},
	}

	farmerBasicDataJSON, err := json.Marshal(farmerBasicData)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("farmerBasicData1", farmerBasicDataJSON); err != nil {
		return err
	}

    // Inspector init data:
	inspectorSeed := InspectorSeedData{
		Batch_id:                     "1",
		InspectorName:                "Inspector Smith",
		QualityVerified:              true,
		PestisideVerified:            true,
		FertilizerVerified:           true,
		ProduceQuantityTypeAVerified: true,
		ProduceQuantityTypeBVerified: true,
		ProduceQuantityTypeCVerified: true,
		InspectDate:                  "2023-04-01",
	}

	inspectorSeedJSON, err := json.Marshal(inspectorSeed)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("inspectorSeed1", inspectorSeedJSON); err != nil {
		return err
	}

	inspectorBasicData := InspectorBasicData{
		Inspector_id: "I1",
		Name:         "Inspector Smith",
		Address:      "Inspectors Lane, City Y",
		InspectorSeeds: []InspectorSeedData{
			inspectorSeed,
		},
	}

	inspectorBasicDataJSON, err := json.Marshal(inspectorBasicData)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("inspectorBasicData1", inspectorBasicDataJSON); err != nil {
		return err
	}

    // Transporter init data:
    transporterSeed := TransporterSeedData{
		Batch_id:                 "1",
		TransporterName:          "Transporter Logistics",
		Location:                 "Logistics City",
		VehicleUsedTransporter:   "Truck",
		VehicleNumberTransporter: "XYZ123",
		Quantity:                 500.0,
		TransportDate:            "2023-05-01",
	}

	transporterSeedJSON, err := json.Marshal(transporterSeed)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("transporterSeed1", transporterSeedJSON); err != nil {
		return err
	}

	transporterBasicData := TransporterBasicData{
		Transporter_id: "T1",
		Name:           "Transporter Logistics",
		Address:        "Logistics Avenue, Transport City",
		TransporterSeeds: []TransporterSeedData{
			transporterSeed,
		},
	}

	transporterBasicDataJSON, err := json.Marshal(transporterBasicData)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("transporterBasicData1", transporterBasicDataJSON); err != nil {
		return err
	}

    // Exporter init data:
    exporterSeed := ExporterSeedData{
		Batch_id:              "1",
		ExporterName:          "Exporters Inc.",
		PlaceKeptExporter:     "Export Warehouse",
		ArrivalDateExporter:   "2023-06-01",
		DestinationAddress:    "International Market",
		VehicleUsedExporter:   "Ship",
		VehicleNumberExporter: "ABC456",
		DepartureDate:         "2023-06-10",
	}

	exporterSeedJSON, err := json.Marshal(exporterSeed)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("exporterSeed1", exporterSeedJSON); err != nil {
		return err
	}

	exporterBasicData := ExporterBasicData{
		Exporter_id: "E1",
		Name:        "Exporters Inc.",
		Address:     "Exporters Street, Export City",
		ExporterSeeds: []ExporterSeedData{
			exporterSeed,
		},
	}

	exporterBasicDataJSON, err := json.Marshal(exporterBasicData)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("exporterBasicData1", exporterBasicDataJSON); err != nil {
		return err
	}

    // Importer init data:
    importerSeed := ImporterSeedData{
		Batch_id:            "1",
		ImporterName:        "Importers LLC",
		PlaceKeptImporter:   "Import Warehouse",
		ArrivalDateImporter: "2023-07-01",
	}

	importerSeedJSON, err := json.Marshal(importerSeed)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("importerSeed1", importerSeedJSON); err != nil {
		return err
	}

	importerBasicData := ImporterBasicData{
		Importer_id: "I1",
		Name:        "Importers LLC",
		Address:     "Importers Street, Import City",
		ImporterSeeds: []ImporterSeedData{
			importerSeed,
		},
	}

	importerBasicDataJSON, err := json.Marshal(importerBasicData)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("importerBasicData1", importerBasicDataJSON); err != nil {
		return err
	}

    // Retailor init data:
    retailorSeed := RetailorSeedData{
		Batch_id:            "1",
		RetailorName:        "Retailors Inc.",
		PlaceKeptRetailor:   "Retail Store",
		ArrivalDateRetailor: "2023-08-01",
	}

	retailorSeedJSON, err := json.Marshal(retailorSeed)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("retailorSeed1", retailorSeedJSON); err != nil {
		return err
	}

	retailorBasicData := RetailorBasicData{
		Retailor_id: "R1",
		Name:        "Retailors Inc.",
		Address:     "Retail Street, Retail City",
		RetailorSeeds: []RetailorSeedData{
			retailorSeed,
		},
	}

	retailorBasicDataJSON, err := json.Marshal(retailorBasicData)
	if err != nil {
		return err
	}

	if err := ctx.GetStub().PutState("retailorBasicData1", retailorBasicDataJSON); err != nil {
		return err
	}

    batchHistory := BatchHistory{
		Batch_id:             "1",
		CurrentStage:         "Completed",
		FarmerSeedBatch:      farmerSeed,
		InspectorSeedBatch:   inspectorSeed,
		TransporterSeedBatch: transporterSeed,
		ExporterSeedBatch:    exporterSeed,
		ImporterSeedBatch:    importerSeed,
		RetailorSeedBatch:    retailorSeed,
	}

	// Convert the BatchHistory structure to JSON
	batchHistoryJSON, err := json.Marshal(batchHistory)
	if err != nil {
		return fmt.Errorf("error marshaling BatchHistory: %v", err)
	}

	// Save the BatchHistory JSON in the world state
	err = ctx.GetStub().PutState("batchHistoryKey", batchHistoryJSON)
	if err != nil {
		return fmt.Errorf("failed to put BatchHistory to world state: %v", err)
	}

	fmt.Println("Chaincode initialization complete.")
	return nil
}

// Funtion for a creating a Batch
func (s *SmartContract) CreateBatch(ctx contractapi.TransactionContextInterface, batchID string) error {

	exists, err := s.BatchExists(ctx, batchID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Batch with ID %s already exists", batchID)
	}

	batch := BatchHistory{
		Batch_id:     batchID,
		CurrentStage: "New Batch Created",
	}

	batchJSON, err := json.Marshal(batch)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(batchID, batchJSON)
}

func (s *SmartContract) BatchExists(ctx contractapi.TransactionContextInterface, batchID string) (bool, error) {
	batchJSON, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return false, err
	}
	return batchJSON != nil, nil
}

// Farmer code started.

// Function for creating the farmer
func (s *SmartContract) CreateFarmer(ctx contractapi.TransactionContextInterface, farmerID string, name string, address string) error {

	exists, err := s.FarmerExists(ctx, farmerID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Farmer with ID %s already exists", farmerID)
	}

	// Create a new farmer
	farmer := FarmerBasicData{
		Farmer_id:   farmerID,
		Name:        name,
		Address:     address,
		FarmerSeeds: []FarmerSeedData{},
	}

	farmerJSON, err := json.Marshal(farmer)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(farmerID, farmerJSON)
	if err != nil {
		return err
	}

	return nil
}

// Function for checking the farmer that wheather it exist or not
func (s *SmartContract) FarmerExists(ctx contractapi.TransactionContextInterface, farmerID string) (bool, error) {

	farmerJSON, err := ctx.GetStub().GetState(farmerID)
	if err != nil {
		return false, fmt.Errorf("Failed to read from the ledger: %v", err)
	}

	return farmerJSON != nil, nil
}

// As farmer involved two time so this function is for before data
func (s *SmartContract) CreateFarmerSeedBefore(ctx contractapi.TransactionContextInterface, batchID string, name string, seedType string, placeGrow string, landArea string, totalQuantitySeed float64, seedQualityType string, growDate string) error {
    
    batchHistoryJSON, err := ctx.GetStub().GetState(batchID)
    var batchHistory BatchHistory

    if err != nil {
        return fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    if batchHistoryJSON != nil {
        // Batch exists, so unmarshal it
        if err := json.Unmarshal(batchHistoryJSON, &batchHistory); err != nil {
            return fmt.Errorf("Errored occur while unmarshalling the data: %v", err)
        }
    } else {
        // Batch doesn't exist, so create it
        if err := s.CreateBatch(ctx, batchID); err != nil {
            return fmt.Errorf("Failed to create new batch: %v", err)
        }
        fmt.Printf("Batch created successfully!!")
    }

    // Create a new FarmerSeedData "before"
    farmerSeedData := FarmerSeedData{
        Batch_id:           batchID,
        FarmerName:         name,
        Seed_type:          seedType,
        PlaceGrow:          placeGrow,
        LandArea:           landArea,
        TotalQuantitySeed:  totalQuantitySeed,
        SeedQualityType:    seedQualityType,
        GrowDate:           growDate,
    }

    batchHistory.FarmerSeedBatch = farmerSeedData
	batchHistory.CurrentStage = "Farmer Stage"

    updatedBatchHistoryJSON, err := json.Marshal(batchHistory)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(batchID, updatedBatchHistoryJSON)
    if err != nil {
        return err
    }

    return nil
}

// This function is for farmer second time
func (s *SmartContract) CreateFarmerSeedDataAfter(ctx contractapi.TransactionContextInterface, farmerID string, batchID string, fertilizerName string, fertilizerQuantity string, pestisideName string, pestisideQuantity string, produceQuantityTypeA float64, produceQuantityTypeB float64, produceQuantityTypeC float64, cutting string, cuttingDate string) error {
    
    batchHistoryJSON, err := ctx.GetStub().GetState(batchID)
    if err != nil {
        return fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    if batchHistoryJSON == nil {
        return fmt.Errorf("Batch with ID %s does not exist. Cannot add Farmer later data.", batchID)
    }

    var batchHistory BatchHistory
    err = json.Unmarshal(batchHistoryJSON, &batchHistory)
    if err != nil {
        return err
    }

	var beforeData = &batchHistory.FarmerSeedBatch

	// Create a new FarmerSeedData "After" and merge with "Before" data
    farmerSeedData := FarmerSeedData{
        Batch_id:             batchID,
        FertilizerName:       fertilizerName,
        FertilizerQuantity:   fertilizerQuantity,
        PestisideName:        pestisideName,
        PestisideQuantity:    pestisideQuantity,
        ProduceQuantityTypeA: produceQuantityTypeA,
        ProduceQuantityTypeB: produceQuantityTypeB,
        ProduceQuantityTypeC: produceQuantityTypeC,
        Cutting: cutting,
        CuttingDate: cuttingDate,
        FarmerName: beforeData.FarmerName,
        Seed_type: beforeData.Seed_type,
        PlaceGrow: beforeData.PlaceGrow,
        LandArea: beforeData.LandArea,
        TotalQuantitySeed: beforeData.TotalQuantitySeed,
        SeedQualityType: beforeData.SeedQualityType,
        GrowDate: beforeData.GrowDate,
    }

    batchHistory.FarmerSeedBatch = farmerSeedData
	batchHistory.CurrentStage = "Farmer Stage completed"

	// After this I also want the farmerSeedData to be appended to the farmer list, so first retrieve the farmer data
	farmerDataJSON, err := ctx.GetStub().GetState(farmerID)
    if err != nil {
        return fmt.Errorf("Failed to read FarmerBasicData from the ledger: %v", err)
    }

    if farmerDataJSON == nil {
        return fmt.Errorf("Farmer with ID %s does not exist. Cannot add farmerSeedData to the list", farmerID)
    }

    var farmerData FarmerBasicData
    err = json.Unmarshal(farmerDataJSON, &farmerData)
    if err != nil {
        return err
    }

    // Append farmerSeedData to the FarmerSeeds list
    farmerData.FarmerSeeds = append(farmerData.FarmerSeeds, farmerSeedData)

    // Update the FarmerBasicData in the ledger
    updatedFarmerDataJSON, err := json.Marshal(farmerData)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(farmerID, updatedFarmerDataJSON)
    if err != nil {
        return err
    }

	// Now updating the batch history,
    updatedBatchHistoryJSON, err := json.Marshal(batchHistory)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(batchID, updatedBatchHistoryJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function for quering farmer all seeds:
func (s *SmartContract) GetFarmerSeedsList(ctx contractapi.TransactionContextInterface, farmerID string) ([]*FarmerSeedData, error) {
    // Retrieve the FarmerBasicData by farmerID
    farmerDataJSON, err := ctx.GetStub().GetState(farmerID)
    if err != nil {
        return nil, fmt.Errorf("Failed to read Farmer Basic Data from the ledger: %v", err)
    }

    if farmerDataJSON == nil {
        return nil, fmt.Errorf("Farmer with ID %s does not exist.", farmerID)
    }

    var farmerData FarmerBasicData
    err = json.Unmarshal(farmerDataJSON, &farmerData)
    if err != nil {
        return nil, err
    }

    // Create a new slice of pointers and copy data from FarmerSeeds
    seedsList := make([]*FarmerSeedData, len(farmerData.FarmerSeeds))
    for i, seed := range farmerData.FarmerSeeds {
        seedCopy := seed
        seedsList[i] = &seedCopy
    }

    return seedsList, nil
}
// Farmer code ended

// Inspector code started:

// Function for creating the inspector
func (s *SmartContract) CreateInspector(ctx contractapi.TransactionContextInterface, inspectorID string, name string, address string) error {
    
    exists, err := s.InspectorExists(ctx, inspectorID)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("Inspector with ID %s already exists", inspectorID)
    }

    inspector := InspectorBasicData{
        Inspector_id: inspectorID,
        Name:         name,
        Address:      address,
        InspectorSeeds: []InspectorSeedData{},
    }

    inspectorJSON, err := json.Marshal(inspector)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(inspectorID, inspectorJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function to check if an Inspector with the given ID exists
func (s *SmartContract) InspectorExists(ctx contractapi.TransactionContextInterface, inspectorID string) (bool, error) {

    inspectorJSON, err := ctx.GetStub().GetState(inspectorID)
    if err != nil {
        return false, fmt.Errorf("Failed to read from the ledger: %v", err)
    }
    return inspectorJSON != nil, nil
}

func (s *SmartContract) UpdateInspectorSeedData(ctx contractapi.TransactionContextInterface, name string, inspectorID string, batchID string, qualityVerified bool, pestisideVerified bool, fertilizerVerified bool, produceQuantityTypeAVerified bool, produceQuantityTypeBVerified bool, produceQuantityTypeCVerified bool, inspectDate string) error {
    
    batchHistoryJSON, err := ctx.GetStub().GetState(batchID)
    if err != nil {
        return fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    if batchHistoryJSON == nil {
        return fmt.Errorf("Batch with ID %s does not exist. Cannot add inspector data.", batchID)
    }

    var batchHistory BatchHistory
    err = json.Unmarshal(batchHistoryJSON, &batchHistory)
    if err != nil {
        return err
    }

    inspectorSeedData := InspectorSeedData{
        Batch_id:                     batchID,
        InspectorName:                name,
        QualityVerified:              qualityVerified,
        PestisideVerified:            pestisideVerified,
        FertilizerVerified:           fertilizerVerified,
        ProduceQuantityTypeAVerified: produceQuantityTypeAVerified,
        ProduceQuantityTypeBVerified: produceQuantityTypeBVerified,
        ProduceQuantityTypeCVerified: produceQuantityTypeCVerified,
        InspectDate:                  inspectDate,
    }

    batchHistory.InspectorSeedBatch = inspectorSeedData
    batchHistory.CurrentStage = "Inspector Stage Complete"

    // Append the updated InspectorSeedData to the Inspector's list, so to keep track of the inspector inspected seeds.
    inspectorDataJSON, err := ctx.GetStub().GetState(inspectorID)
    if err != nil {
        return fmt.Errorf("Failed to read InspectorBasicData from the ledger: %v", err)
    }

    if inspectorDataJSON == nil {
        return fmt.Errorf("Inspector with ID %s does not exist. Cannot add InspectorSeedData to the list", inspectorID)
    }

    var inspectorData InspectorBasicData
    err = json.Unmarshal(inspectorDataJSON, &inspectorData)
    if err != nil {
        return err
    }

    inspectorData.InspectorSeeds = append(inspectorData.InspectorSeeds, inspectorSeedData)

    updatedInspectorDataJSON, err := json.Marshal(inspectorData)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(inspectorID, updatedInspectorDataJSON)
    if err != nil {
        return err
    }

    // Update the batch history in the ledger
    updatedBatchHistoryJSON, err := json.Marshal(batchHistory)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(batchID, updatedBatchHistoryJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function for getting specific inspector seed data:
func (s *SmartContract) GetInspectorSeeds(ctx contractapi.TransactionContextInterface, inspectorID string) ([]*InspectorSeedData, error) {
    inspectorDataJSON, err := ctx.GetStub().GetState(inspectorID)
    if err != nil {
        return nil, fmt.Errorf("Failed to read InspectorBasicData from the ledger: %v", err)
    }

    if inspectorDataJSON == nil {
        return nil, fmt.Errorf("Inspector with ID %s does not exist.", inspectorID)
    }

    var inspectorData InspectorBasicData
    err = json.Unmarshal(inspectorDataJSON, &inspectorData)
    if err != nil {
        return nil, err
    }

    seedsList := make([]*InspectorSeedData, len(inspectorData.InspectorSeeds))
    for i, seed := range inspectorData.InspectorSeeds {
        seedCopy := seed // Create a copy of 'seed'
        seedsList[i] = &seedCopy
    }

    return seedsList, nil
}
// End of inspector code.

// Transporter code started.

// Function for creating a Transporter
func (s *SmartContract) CreateTransporter(ctx contractapi.TransactionContextInterface, transporterID string, name string, address string) error {

	exists, err := s.TransporterExists(ctx, transporterID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("Transporter with ID %s already exists", transporterID)
	}

	transporter := TransporterBasicData{
		Transporter_id:   transporterID,
		Name:             name,
		Address:          address,
		TransporterSeeds: []TransporterSeedData{},
	}

	transporterJSON, err := json.Marshal(transporter)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(transporterID, transporterJSON)
	if err != nil {
		return err
	}

	return nil
}

// Function for checking if a Transporter exists
func (s *SmartContract) TransporterExists(ctx contractapi.TransactionContextInterface, transporterID string) (bool, error) {

	transporterJSON, err := ctx.GetStub().GetState(transporterID)
	if err != nil {
		return false, fmt.Errorf("Failed to read from the ledger: %v", err)
	}

	return transporterJSON != nil, nil
}

// Function for updating TransporterSeedData
func (s *SmartContract) UpdateTransporterSeedData(ctx contractapi.TransactionContextInterface, transporterID string, batchID string, transporterName string, location string, vehicleUsedTransporter string, vehicleNumberTransporter string, quantity float64, transportDate string) error {

	batchHistoryJSON, err := ctx.GetStub().GetState(batchID)
	if err != nil {
		return fmt.Errorf("Failed to read from the ledger: %v", err)
	}

	if batchHistoryJSON == nil {
		return fmt.Errorf("Batch with ID %s does not exist. Cannot add Transporter data.", batchID)
	}

	var batchHistory BatchHistory
	err = json.Unmarshal(batchHistoryJSON, &batchHistory)
	if err != nil {
		return err
	}

	transporterSeedData := TransporterSeedData{
		Batch_id:                 batchID,
		TransporterName:          transporterName,
		Location:                 location,
		VehicleUsedTransporter:   vehicleUsedTransporter,
		VehicleNumberTransporter: vehicleNumberTransporter,
		Quantity:                 quantity,
		TransportDate:            transportDate,
	}

	batchHistory.TransporterSeedBatch = transporterSeedData
	batchHistory.CurrentStage = "Transporter Stage Complete"

	// Append the updated TransporterSeedData to the Transporter's list
	transporterDataJSON, err := ctx.GetStub().GetState(transporterID)
	if err != nil {
		return fmt.Errorf("Failed to read TransporterBasicData from the ledger: %v", err)
	}

	if transporterDataJSON == nil {
		return fmt.Errorf("Transporter with ID %s does not exist. Cannot add TransporterSeedData to the list", transporterID)
	}

	var transporterData TransporterBasicData
	err = json.Unmarshal(transporterDataJSON, &transporterData)
	if err != nil {
		return err
	}

	transporterData.TransporterSeeds = append(transporterData.TransporterSeeds, transporterSeedData)

	updatedTransporterDataJSON, err := json.Marshal(transporterData)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(transporterID, updatedTransporterDataJSON)
	if err != nil {
		return err
	}

	// Update the batch history in the ledger
	updatedBatchHistoryJSON, err := json.Marshal(batchHistory)
	if err != nil {
		return err
	}

	err = ctx.GetStub().PutState(batchID, updatedBatchHistoryJSON)
	if err != nil {
		return err
	}

	return nil
}

// Function for getting transporter data against transporter w.r.t transporter ID.
func (s *SmartContract) GetTransporterSeeds(ctx contractapi.TransactionContextInterface, transporterID string) ([]*TransporterSeedData, error) {

    transporterDataJSON, err := ctx.GetStub().GetState(transporterID)
    if err != nil {
        return nil, fmt.Errorf("Failed to read Transporter Basic Data from the ledger: %v", err)
    }

    if transporterDataJSON == nil {
        return nil, fmt.Errorf("Transporter with ID %s does not exist.", transporterID)
    }

    var transporterData TransporterBasicData
    err = json.Unmarshal(transporterDataJSON, &transporterData)
    if err != nil {
        return nil, err
    }

    seedsList := make([]*TransporterSeedData, len(transporterData.TransporterSeeds))
    for i, seed := range transporterData.TransporterSeeds {
        seedCopy := seed // Create a copy of 'seed'
        seedsList[i] = &seedCopy
    }

    return seedsList, nil
}
// End of transporter code.

// Exporter code started:

// Function for creating the Exporter
func (s *SmartContract) CreateExporter(ctx contractapi.TransactionContextInterface, exporterID string, name string, address string) error {
    
    exists, err := s.ExporterExists(ctx, exporterID)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("Exporter with ID %s already exists", exporterID)
    }

    // Create a new exporter
    exporter := ExporterBasicData{
        Exporter_id: exporterID,
        Name:        name,
        Address:     address,
        ExporterSeeds: []ExporterSeedData{},
    }

    exporterJSON, err := json.Marshal(exporter)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(exporterID, exporterJSON)
    if err != nil {
        return err
    }

    return nil
}

func (s *SmartContract) ExporterExists(ctx contractapi.TransactionContextInterface, exporterID string) (bool, error) {
    exporterJSON, err := ctx.GetStub().GetState(exporterID)
    if err != nil {
        return false, fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    return exporterJSON != nil, nil
}

// Function for updating ExporterSeedData
func (s *SmartContract) UpdateExporterSeedData(ctx contractapi.TransactionContextInterface, exporterID string, batchID string, exporterName string, placeKeptExporter string, arrivalDateExporter string, destinationAddress string, vehicleUsedExporter string, vehicleNumberExporter string, departureDate string) error {
    
    batchHistoryJSON, err := ctx.GetStub().GetState(batchID)
    if err != nil {
        return fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    if batchHistoryJSON == nil {
        return fmt.Errorf("Batch with ID %s does not exist. Cannot add Exporter data.", batchID)
    }

    var batchHistory BatchHistory
    err = json.Unmarshal(batchHistoryJSON, &batchHistory)
    if err != nil {
        return err
    }

    exporterSeedData := ExporterSeedData{
        Batch_id:              batchID,
        ExporterName:          exporterName,
        PlaceKeptExporter:     placeKeptExporter,
        ArrivalDateExporter:   arrivalDateExporter,
        DestinationAddress:    destinationAddress,
        VehicleUsedExporter:   vehicleUsedExporter,
        VehicleNumberExporter: vehicleNumberExporter,
        DepartureDate:         departureDate,
    }

    batchHistory.ExporterSeedBatch = exporterSeedData
    batchHistory.CurrentStage = "Exporter Stage Complete"

    exporterDataJSON, err := ctx.GetStub().GetState(exporterID)
    if err != nil {
        return fmt.Errorf("Failed to read ExporterBasicData from the ledger: %v", err)
    }

    if exporterDataJSON == nil {
        return fmt.Errorf("Exporter with ID %s does not exist. Cannot add Exporter Seed Data to the list", exporterID)
    }

    var exporterData ExporterBasicData
    err = json.Unmarshal(exporterDataJSON, &exporterData)
    if err != nil {
        return err
    }

    exporterData.ExporterSeeds = append(exporterData.ExporterSeeds, exporterSeedData)

    updatedExporterDataJSON, err := json.Marshal(exporterData)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(exporterID, updatedExporterDataJSON)
    if err != nil {
        return err
    }

    updatedBatchHistoryJSON, err := json.Marshal(batchHistory)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(batchID, updatedBatchHistoryJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function for getting exporter seeds list data:
func (s *SmartContract) GetExporterSeeds(ctx contractapi.TransactionContextInterface, exporterID string) ([]*ExporterSeedData, error) {
    
    exporterDataJSON, err := ctx.GetStub().GetState(exporterID)
    if err != nil {
        return nil, fmt.Errorf("Failed to read Exporter data from the ledger: %v", err)
    }

    if exporterDataJSON == nil {
        return nil, fmt.Errorf("Exporter with ID %s does not exist. Cannot retrieve Exporter Seed data.", exporterID)
    }

    var exporterData ExporterBasicData
    err = json.Unmarshal(exporterDataJSON, &exporterData)
    if err != nil {
        return nil, err
    }

    seedsList := make([]*ExporterSeedData, len(exporterData.ExporterSeeds))
    for i, seed := range exporterData.ExporterSeeds {
        seedCopy := seed // Create a copy of 'seed'
        seedsList[i] = &seedCopy
    }

    return seedsList, nil
}
// Exporter code ended.

// Importer code started:

// Function for creating an Importer
func (s *SmartContract) CreateImporter(ctx contractapi.TransactionContextInterface, importerID string, name string, address string) error {
    exists, err := s.ImporterExists(ctx, importerID)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("Importer with ID %s already exists", importerID)
    }

    // Create a new importer
    importer := ImporterBasicData{
        Importer_id: importerID,
        Name:        name,
        Address:     address,
        ImporterSeeds: []ImporterSeedData{},
    }

    importerJSON, err := json.Marshal(importer)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(importerID, importerJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function for checking if an Importer exists
func (s *SmartContract) ImporterExists(ctx contractapi.TransactionContextInterface, importerID string) (bool, error) {
    importerJSON, err := ctx.GetStub().GetState(importerID)
    if err != nil {
        return false, fmt.Errorf("Failed to read from the ledger: %v", err)
    }
    return importerJSON != nil, nil
}

// Function for updating ImporterSeedData
func (s *SmartContract) UpdateImporterSeedData(ctx contractapi.TransactionContextInterface, importerID string, batchID string, importerName string, placeKeptImporter string, arrivalDateImporter string) error {
    
    batchHistoryJSON, err := ctx.GetStub().GetState(batchID)
    if err != nil {
        return fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    if batchHistoryJSON == nil {
        return fmt.Errorf("Batch with ID %s does not exist. Cannot add Importer data.", batchID)
    }

    var batchHistory BatchHistory
    err = json.Unmarshal(batchHistoryJSON, &batchHistory)
    if err != nil {
        return err
    }

    importerSeedData := ImporterSeedData{
        Batch_id:            batchID,
        ImporterName:        importerName,
        PlaceKeptImporter:   placeKeptImporter,
        ArrivalDateImporter: arrivalDateImporter,
    }

    batchHistory.ImporterSeedBatch = importerSeedData
    batchHistory.CurrentStage = "Importer Stage Complete"

    importerDataJSON, err := ctx.GetStub().GetState(importerID)
    if err != nil {
        return fmt.Errorf("Failed to read ImporterBasicData from the ledger: %v", err)
    }

    if importerDataJSON == nil {
        return fmt.Errorf("Importer with ID %s does not exist. Cannot add Importer Seed Data to the list", importerID)
    }

    var importerData ImporterBasicData
    err = json.Unmarshal(importerDataJSON, &importerData)
    if err != nil {
        return err
    }

    importerData.ImporterSeeds = append(importerData.ImporterSeeds, importerSeedData)

    updatedImporterDataJSON, err := json.Marshal(importerData)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(importerID, updatedImporterDataJSON)
    if err != nil {
        return err
    }

    updatedBatchHistoryJSON, err := json.Marshal(batchHistory)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(batchID, updatedBatchHistoryJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function for getting the importer seed list data:
func (s *SmartContract) GetImporterSeeds(ctx contractapi.TransactionContextInterface, importerID string) ([]*ImporterSeedData, error) {
    
    importerDataJSON, err := ctx.GetStub().GetState(importerID)
    if err != nil {
        return nil, fmt.Errorf("Failed to read Importer data from the ledger: %v", err)
    }

    if importerDataJSON == nil {
        return nil, fmt.Errorf("Importer with ID %s does not exist. Cannot retrieve Importer Basic data.", importerID)
    }

    var importerData ImporterBasicData
    err = json.Unmarshal(importerDataJSON, &importerData)
    if err != nil {
        return nil, err
    }

    seedsList := make([]*ImporterSeedData, len(importerData.ImporterSeeds))
    for i, seed := range importerData.ImporterSeeds {
        seedCopy := seed // Create a copy of 'seed'
        seedsList[i] = &seedCopy
    }

    return seedsList, nil
}
// Importer code ended.

// Retailor code started.

// Function for creating a Retailor
func (s *SmartContract) CreateRetailor(ctx contractapi.TransactionContextInterface, retailorID string, name string, address string) error {

    exists, err := s.RetailorExists(ctx, retailorID)
    if err != nil {
        return err
    }
    if exists {
        return fmt.Errorf("Retailor with ID %s already exists", retailorID)
    }

    retailor := RetailorBasicData{
        Retailor_id: retailorID,
        Name:        name,
        Address:     address,
        RetailorSeeds: []RetailorSeedData{},
    }

    retailorJSON, err := json.Marshal(retailor)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(retailorID, retailorJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function for checking if a Retailor exists
func (s *SmartContract) RetailorExists(ctx contractapi.TransactionContextInterface, retailorID string) (bool, error) {
    retailorJSON, err := ctx.GetStub().GetState(retailorID)
    if err != nil {
        return false, fmt.Errorf("Failed to read from the ledger: %v", err)
    }
    return retailorJSON != nil, nil
}

// Function for updating RetailorSeedData
func (s *SmartContract) UpdateRetailorSeedData(ctx contractapi.TransactionContextInterface, retailorID string, batchID string, retailorName string, placeKeptRetailor string, arrivalDateRetailor string) error {

    batchHistoryJSON, err := ctx.GetStub().GetState(batchID)
    if err != nil {
        return fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    if batchHistoryJSON == nil {
        return fmt.Errorf("Batch with ID %s does not exist. Cannot add Retailor data.", batchID)
    }

    var batchHistory BatchHistory
    err = json.Unmarshal(batchHistoryJSON, &batchHistory)
    if err != nil {
        return err
    }

    retailorSeedData := RetailorSeedData{
        Batch_id:            batchID,
        RetailorName:        retailorName,
        PlaceKeptRetailor:   placeKeptRetailor,
        ArrivalDateRetailor: arrivalDateRetailor,
    }

    batchHistory.RetailorSeedBatch = retailorSeedData
    batchHistory.CurrentStage = "Retailor Stage Complete"

    retailorDataJSON, err := ctx.GetStub().GetState(retailorID)
    if err != nil {
        return fmt.Errorf("Failed to read RetailorBasicData from the ledger: %v", err)
    }

    if retailorDataJSON == nil {
        return fmt.Errorf("Retailor with ID %s does not exist. Cannot add RetailorSeedData to the list", retailorID)
    }

    var retailorData RetailorBasicData
    err = json.Unmarshal(retailorDataJSON, &retailorData)
    if err != nil {
        return err
    }

    retailorData.RetailorSeeds = append(retailorData.RetailorSeeds, retailorSeedData)

    updatedRetailorDataJSON, err := json.Marshal(retailorData)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(retailorID, updatedRetailorDataJSON)
    if err != nil {
        return err
    }

    updatedBatchHistoryJSON, err := json.Marshal(batchHistory)
    if err != nil {
        return err
    }

    err = ctx.GetStub().PutState(batchID, updatedBatchHistoryJSON)
    if err != nil {
        return err
    }

    return nil
}

// Function for getting the retailor seed data:
func (s *SmartContract) GetRetailorSeeds(ctx contractapi.TransactionContextInterface, retailorID string) ([]*RetailorSeedData, error) {
    
    retailorDataJSON, err := ctx.GetStub().GetState(retailorID)
    if err != nil {
        return nil, fmt.Errorf("Failed to read Retailor data from the ledger: %v", err)
    }

    if retailorDataJSON == nil {
        return nil, fmt.Errorf("Retailor with ID %s does not exist. Cannot retrieve Retailor Basic data.", retailorID)
    }

    var retailorData RetailorBasicData
    err = json.Unmarshal(retailorDataJSON, &retailorData)
    if err != nil {
        return nil, err
    }

    seedsList := make([]*RetailorSeedData, len(retailorData.RetailorSeeds))
    for i, seed := range retailorData.RetailorSeeds {
        seedCopy := seed // Create a copy of 'seed'
        seedsList[i] = &seedCopy
    }

    return seedsList, nil
}
// Code ended for retailor.

// This function will retrieve the data for a batch
func (s *SmartContract) GetBatchData(ctx contractapi.TransactionContextInterface, batchID string) (*BatchHistory, error) {

    batchDataJSON, err := ctx.GetStub().GetState(batchID)
    if err != nil {
        return nil, fmt.Errorf("Failed to read from the ledger: %v", err)
    }

    if batchDataJSON == nil {
        return nil, fmt.Errorf("Batch with ID %s does not exist", batchID)
    }

    var batchData BatchHistory
    if err := json.Unmarshal(batchDataJSON, &batchData); err != nil {
        return nil, fmt.Errorf("Failed to unmarshal BatchHistory: %v", err)
    }

    return &batchData, nil
}

func main() {
	// Create a new Smart Contract
	smartContract, err := contractapi.NewChaincode(&SmartContract{})
	if err != nil {
		fmt.Printf("Error creating SmartContract chaincode: %v\n", err)
		return
	}

	if err := smartContract.Start(); err != nil {
		fmt.Printf("Error starting SmartContract chaincode: %v\n", err)
		return
	}
}
