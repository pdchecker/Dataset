package main

import(
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
    contractapi.Contract
}


// NOTE: Write the asset properties in CAMEL CASE- otherwise, chaincode will not get deployed 
type TallyScoreAsset struct {
    LicenseId string `json:"LicenseId"`
	Score uint `json:"Score"`
}

//function to check whether the company has already been registered
func (s *SmartContract) companyAssetExists(ctx contractapi.TransactionContextInterface, LicenseId string) (bool, error) {

    companyAssetJSON, err := ctx.GetStub().GetState(LicenseId)
    if err != nil {
    	return false, fmt.Errorf("failed to read from world state: %v", err)
    }

    return companyAssetJSON != nil, nil
}

// function to register company and initialize it's Score to 500 
func (s *SmartContract) RegisterCompany(ctx contractapi.TransactionContextInterface, LicenseId string) error{

	// ------------------------- use client.Context to retrieve the PAN value of the business ------------------------

	//checking if licenseID is valid
	companyAssetExists,err:= s.companyAssetExists(ctx, LicenseId)
	if err!=nil{
		return fmt.Errorf("error in checking whether asset exists: %v", err)
	}
	if companyAssetExists {
		return fmt.Errorf("This company already exists!")
	}

	// if the company is unregistered
	companyScoreAsset := TallyScoreAsset{
		LicenseId: LicenseId,
		Score: 500,
	}
	companyScoreAssetJSON, err := json.Marshal(companyScoreAsset)
    if err != nil {
        return err
    }

	putStateErr := ctx.GetStub().PutState(LicenseId, companyScoreAssetJSON) // new state added to the tallyscore ledger
    fmt.Printf("Asset creation returned : %s\n", putStateErr)
    return putStateErr

} 

// function to unregister a company (deleting it's Score asset)
func (s *SmartContract) UnregisterCompany(ctx contractapi.TransactionContextInterface, LicenseId string) error{
	//checking if licenseID is valid
	// var sumOfDigits int
	// for _, charDigit:= range LicenseId{
	// 	digit:= int(charDigit- '0')
	// 	sumOfDigits+= digit
	// }
	// if sumOfDigits%9 !=0{
	// 	return fmt.Errorf("Invalid license ID")
	// }

	exists, err := s.companyAssetExists(ctx, LicenseId)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", LicenseId)
	}
		
	delStateOp:= ctx.GetStub().DelState(LicenseId)
    fmt.Printf("Message received on deletion: %s", delStateOp)
    return nil
}

// function to read companyasset
func (s *SmartContract) ReadCompanyAsset(ctx contractapi.TransactionContextInterface, licenseID string) (*TallyScoreAsset, error){
	companyScoreAssetJSON, err := ctx.GetStub().GetState(licenseID)
    if err != nil {
    	return nil, fmt.Errorf("Failed to read from world state: %v", err)
    }
    if companyScoreAssetJSON == nil {
    	return nil, fmt.Errorf("The company with ID %s is not registered.", licenseID)
    }

    var companyScoreAsset TallyScoreAsset
    err = json.Unmarshal(companyScoreAssetJSON, &companyScoreAsset)
    if err != nil {
    	return nil, err
	}

	return &companyScoreAsset, nil
}

// function to increase tallyScore of a company
func (s *SmartContract) IncreaseScore(ctx contractapi.TransactionContextInterface, licenseID string, incrementValue string) (*TallyScoreAsset, error) {
    companyAssetRead, err := s.ReadCompanyAsset(ctx, licenseID) // asset is read
    if err != nil {
    	return nil, err
    }

    intermediateUpdateval, err := strconv.ParseUint(incrementValue, 10, 32)
    if err !=nil {
    	fmt.Println(err)
    }
	newScore:= uint(companyAssetRead.Score) + ((1000- companyAssetRead.Score) * uint(intermediateUpdateval))/100
    if newScore > 1000 {
    	return nil, fmt.Errorf("You cannot have a value more than 1000.")
    }

    // overwriting original asset with new value
    companyAsset := TallyScoreAsset {
        LicenseId: licenseID,
		Score: newScore,
    }
    companyAssetJSON, err := json.Marshal(companyAsset)
    if err != nil {
    	return nil, err
	}

	updatestate_err := ctx.GetStub().PutState(licenseID, companyAssetJSON)
	fmt.Printf("Increasing company asset Score returned the following: %s ", updatestate_err)
	return &companyAsset, nil
}

// function to decrease tallyScore of a company
func (s *SmartContract) DecreaseScore(ctx contractapi.TransactionContextInterface, licenseID string, decrementValue string) (*TallyScoreAsset, error) {
    companyAssetRead, err := s.ReadCompanyAsset(ctx, licenseID) // asset is read
    if err != nil {
    	return nil, err
    }

    intermediateUpdateval, err := strconv.ParseUint(decrementValue, 10, 32)
    if err !=nil {
    	fmt.Println(err)
    }

	updateVal:= ((1000- companyAssetRead.Score) * uint(intermediateUpdateval))/100
	if updateVal>uint(companyAssetRead.Score){
		return nil, fmt.Errorf("You cannot have a value lesser than 0.")
	}
	newScore:= uint(companyAssetRead.Score) - updateVal


    // overwriting original asset with new value
    companyAsset := TallyScoreAsset {
        LicenseId: licenseID,
		Score: newScore,
    }
    companyAssetJSON, err := json.Marshal(companyAsset)
    if err != nil {
    	return nil, err
	}

	updatestate_err := ctx.GetStub().PutState(licenseID, companyAssetJSON)
	fmt.Printf("Decreasing company asset Score returned the following: %s ", updatestate_err)
	return &companyAsset, nil
}