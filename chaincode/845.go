package main


import (
    "encoding/base64"
    "encoding/json"
    "fmt"
    "strings"
    "strconv"
    "github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
    contractapi.Contract
}
type Asset struct {
    Name  string `json:"Name"`
    Value uint   `json:"Value"`
    OwnerID string `json:"OwnerID"`
    TransferStatus string `json:"TransferStatus"`
    RequestingUser string `json:RequestingUser`
}



// ADD REQUESTED OWNER to struct- WHEN IMPLEMENTING FUNC

const Prefix = "Key: "

// function that takes input as context of transaction and the name of the key, returns boolean value that implies whether the asset exists or not, otherwise- an error
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, Name string) (bool, error) {


    // all users can access this function irrespective of their approver and creator values

    assetJSON, err := ctx.GetStub().GetState(Prefix + Name)
    if err != nil {
    return false, fmt.Errorf("failed to read from world state: %v", err)
    }

    return assetJSON != nil, nil
}


func (s *SmartContract) GetAssetValue(ctx contractapi.TransactionContextInterface, Name string)(string, error){

	assets_list, err := s.GetAllAssets(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to read from world state: %v", err)
	}
	var asset *Asset
	for _, iteratorVar := range assets_list{
		if iteratorVar.Name == Name{
			asset = iteratorVar
			break
		}
    }
	return string(asset.Value), nil
	
}



// function to create an asset. Input= transaction context, name of the key to be created. Creates new asset if an asset with the name given does not exist
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, Name string) error {

    // those with creator set as true can access this function
    Is_creator := checkCreator(ctx)
    if Is_creator==false{
        return fmt.Errorf("Not enough permissions")
    }

    OwnerID, err := submittingClientIdentity(ctx)
	if err != nil {
		return err
	}
    // get id from func
    exists, err := s.AssetExists(ctx, Prefix + Name) // exists-> boolean value, err-> can be nil or the error, if present

    fmt.Printf("Asset exists returned : %t, %s\n", exists, err)

    if err != nil {
        return err
                }
        if exists {
            return fmt.Errorf("the asset %s already exists", Prefix + Name)
            }

            asset := Asset{ //creation of asset
                Name: Name,
                Value: 0,
                OwnerID: OwnerID,
                TransferStatus: "NA",   // default value; can be "NA", "approved" or "requested"
                RequestingUser: "NA",    // can be NA or the ID of the user that is requesting for transfer
            }
            assetJSON, err := json.Marshal(asset)
            if err != nil {
                return err
            }

        state_err := ctx.GetStub().PutState(Prefix + Name, assetJSON) // new state added

        fmt.Printf("Asset creation returned : %s\n", state_err)

        return state_err
}

// ReadAsset returns the asset stored in the world state with given Name.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, Name string) (*Asset, error) {

    // all users can access this function irrespective of their approver and creator values

    assetJSON, err := ctx.GetStub().GetState(Prefix + Name)
    if err != nil {
    return nil, fmt.Errorf("failed to read from world state: %v", err)
    }
    if assetJSON == nil {
    return nil, fmt.Errorf("the asset %s does not exist", Name)
    }

    var asset Asset
    err = json.Unmarshal(assetJSON, &asset)
    if err != nil {
    return nil, err
}

return &asset, nil
}



func (s *SmartContract )GetAssetsPagination(ctx contractapi.TransactionContextInterface, startname string, endname string, bookmark string) ([] *Asset, error) {

    // NOTE: BOOKMARK HAS TO BE SENT AS AN EMPTY STRING WHEN SENT AS A PARAMETER
    pageSizeInt := 5
                   iteratorVar, midvar, err:= ctx.GetStub().GetStateByRangeWithPagination(Prefix + startname, Prefix + endname, int32(pageSizeInt), bookmark)
    if err !=nil && midvar!=nil {
    return nil, err
}
defer iteratorVar.Close()


    var assets []*Asset

    for iteratorVar.HasNext() {
        queryResponse, err := iteratorVar.Next()
        if err != nil {
        return nil, err
    }

    var asset Asset
    err = json.Unmarshal(queryResponse.Value, &asset)
        if err != nil {
        return nil, err
    }
    assets = append(assets, &asset)
    }

    return assets, nil

}

func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([] *Asset, error) {

    // all users can access this function irrespective of their approver and creator values

    iteratorVar, err := ctx.GetStub().GetStateByRange("","")   // TRY RANGE PARAMETERS , other getstateby.... (rows etc.)
    if err !=nil {
        return nil, err
    }
    defer iteratorVar.Close()

    var assets []*Asset

    var assetCount = 0
    for iteratorVar.HasNext() {
        queryResponse, err := iteratorVar.Next()
        if err != nil {
        return nil, err
    }

    var asset Asset
    err = json.Unmarshal(queryResponse.Value, &asset)
        if err != nil {
        return nil, err
    }
    assets = append(assets, &asset)
                 assetCount++
    }

    if assetCount > 0 {
    return assets, nil
    } else {
    return nil, fmt.Errorf("No assets found")
    }

}



// IncreaseAsset increases the value of the asset by the specified value- with certain limits
func (s *SmartContract) IncreaseAsset(ctx contractapi.TransactionContextInterface, Name string, incrementValue string) (*Asset, error) {
    // NOTE: incrementValue is a string because SubmitTransaction accepts string parameters as input parameters


    // only the owner of the asset can do this
    reqOwnerID, err := submittingClientIdentity(ctx)
	if err != nil {
		return nil,err
	}


    asset_read, err := s.ReadAsset(ctx, Name) // asset is read
    if err != nil {
    return nil, err
    }
    owner_asset:= asset_read.OwnerID
    transferstatus:= asset_read.TransferStatus
    requser:= asset_read.RequestingUser

    if owner_asset != reqOwnerID{
        return nil, fmt.Errorf("User does not have the authority to do this.")
    }

    intermediateUpdateval, err := strconv.ParseUint(incrementValue, 10, 32)
    if err !=nil {
    fmt.Println(err)
    }
    incrementValueuInt := uint(intermediateUpdateval)
    newValue := uint(asset_read.Value) + incrementValueuInt

    if newValue > 20 {
    return nil, fmt.Errorf("You cannot have a value more than 20.")
    }

    // overwriting original asset with new value
    asset := Asset {
        Name:  Name,
        Value: newValue,
        OwnerID: owner_asset,
        TransferStatus: transferstatus,
        RequestingUser: requser,
    }
    assetJSON, err := json.Marshal(asset)
    if err != nil {
    return nil, err
}

updatestate_err := ctx.GetStub().PutState(Prefix + Name, assetJSON)
                       fmt.Printf("Increasing asset value returned the following: %s ", updatestate_err)

                       return &asset, nil
}

// DecreaseAsset decreases the value of the asset by the specified value
func (s *SmartContract) DecreaseAsset(ctx contractapi.TransactionContextInterface, Name string, decrementValue string) (*Asset, error) {

    reqOwnerID, err := submittingClientIdentity(ctx)
	if err != nil {
		return nil,err
	}


    asset_read, err := s.ReadAsset(ctx, Name)
    if err != nil {
    return nil, err
    }

    owner_asset:= asset_read.OwnerID
    transferstatus:= asset_read.TransferStatus
    requser:= asset_read.RequestingUser

    if owner_asset != reqOwnerID{
        return nil, fmt.Errorf("User does not have the authority to do this.")
    }

    intermediateval, err := strconv.ParseUint(decrementValue, 10, 32)
    if err !=nil {
    fmt.Println(err)
    }
    decrementValueuInt := uint(intermediateval)
    if decrementValueuInt > uint(asset_read.Value) {
        return nil, fmt.Errorf("You cannot decrement value to less than 0.")
    }
    newValue := uint(asset_read.Value) - decrementValueuInt


                // overwriting original asset with new value
    asset := Asset {
        Name:  Name,
        Value: newValue,
        OwnerID: owner_asset,
        TransferStatus: transferstatus,
        RequestingUser: requser,
    }
    assetJSON, err := json.Marshal(asset)
    if err != nil {
    return nil, err
    }

updatestate_Err := ctx.GetStub().PutState(Prefix + Name, assetJSON)
                       fmt.Printf("After decreasing asset value: %s", updatestate_Err)

                       return &asset , nil
}



func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, Name string) (*Asset, error) {

    // only creators can have assets transferred to them- this will be handled in RequestTransfer function

    newOwnerID, err := submittingClientIdentity(ctx)
	if err != nil {
		return nil,err
	}

    asset, err:= s.ReadAsset(ctx, Name)
    if err != nil {
        return nil, err
    }
    

    if asset.TransferStatus != "approved"{
        println("Not accepted")
        return nil, fmt.Errorf("Transfer has not been approved")
    }


    // if transfer has been approved
    // overwriting current asset with new owner id
    val_AssetInt:= asset.Value
    new_asset := Asset {
        Name:  Name,
        Value: val_AssetInt,
        OwnerID: newOwnerID,
        TransferStatus: "NA",
        RequestingUser: "NA",
    }

    assetJSON, err := json.Marshal(new_asset)
    if err != nil {
    return nil, err
    }

    ctx.GetStub().PutState(Prefix + Name, assetJSON)

    return &new_asset , nil
}



// DeleteAsset deletes the state from the ledger
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, name string) error {

    // can only be done by approvers and creators
    Is_creator := checkCreator(ctx)
    Is_approver := checkApprover(ctx)

    if Is_creator==true || Is_approver==true{
        exists, err := s.AssetExists(ctx, name)
        if err != nil {
            return err
        }
        if !exists {
            return fmt.Errorf("the asset %s does not exist", name)
        }

        delop:= ctx.GetStub().DelState(Prefix + name)
        fmt.Printf("Message received on deletion: %s", delop)
        return nil
    }
    return fmt.Errorf("User does not have the authority to delete asset.")
}


// trasnfer( asset id, destination owner)

// create functons to check if creator= true or approver= true


func checkApprover(ctx contractapi.TransactionContextInterface) bool{
    err := ctx.GetClientIdentity().AssertAttributeValue("approver", "true")
	if err != nil {
		return false
	}
    return true
}

func checkCreator(ctx contractapi.TransactionContextInterface) bool{
    err := ctx.GetClientIdentity().AssertAttributeValue("creator", "true")
	if err != nil {
		return false
	}
    return true
}


func submittingClientIdentity(ctx contractapi.TransactionContextInterface) (string, error) {

        b64ID, err := ctx.GetClientIdentity().GetID()
        if err != nil {
            return "", fmt.Errorf("Failed to read clientID: %v", err)
        }
        decodeID, err := base64.StdEncoding.DecodeString(b64ID)
        if err != nil {
            return "", fmt.Errorf("failed to base64 decode clientID: %v", err)
        }
        return string(decodeID), nil     // returns clientID as a string
}






// REQUEST TRANSFER- req transaction


func (s *SmartContract) RequestTransfer(ctx contractapi.TransactionContextInterface, Name string, reqUser string) (*Asset, error){

     // should only be done by appovers or creators
     Is_creator := checkCreator(ctx)
     Is_approver := checkApprover(ctx)
 
     if Is_creator==true || Is_approver==true{

            // extracting the id of the requesting owner

            requestingOwnerID, err := submittingClientIdentity(ctx)
            if err != nil {
                return nil,err
            }

            // acquiring the asset that is to be transferred

            asset, err:= s.ReadAsset(ctx, Name)
            if err != nil {
                return nil, err
            }

            asset_value:= asset.Value
            asset_ownerID:= asset.OwnerID

            if asset_ownerID == requestingOwnerID{
                println("This owner already owns the asset")
                return nil, fmt.Errorf("This owner already owns the asset")
            }
            
            // changing asset's transfer status from "NA" to "requested" and changing requesting user

            new_asset := Asset {
                Name:  Name,
                Value: asset_value,
                OwnerID:  asset_ownerID,
                TransferStatus: "requested",
                RequestingUser: reqUser,
            }

            assetJSON, err := json.Marshal(new_asset)
            if err != nil {
            return nil, err
            }

            ctx.GetStub().PutState(Prefix + Name, assetJSON)

            return &new_asset , nil

    }
    return nil, fmt.Errorf("This user does not have the authority to do this.")
}


// two modes- automatic or requires approval(based on condition- asset value<10 or >=10)



// APPROVE TRANSFER- for the user who owns the asset

func (s *SmartContract) ApproveTransfer(ctx contractapi.TransactionContextInterface, Name string) (*Asset, error){


    // should only be done by approver
    Is_approver := checkApprover(ctx)
 
    if Is_approver==true{
            // extracting the id of the owner that is supposed to perform the approval

            approvingOwnerID, err := submittingClientIdentity(ctx)
            if err != nil {
                
                return nil,err
            }

            Is_approver := checkApprover(ctx)
            if !Is_approver{
                return nil, fmt.Errorf("Not enough permissions")
            }

            // acquiring the asset that is to be transferred

            asset, err:= s.ReadAsset(ctx, Name)
            if err != nil {
                return nil, err
            }

            asset_value:= asset.Value
            asset_ownerID:= asset.OwnerID
            asset_status:= asset.TransferStatus
            requestingUser:= asset.RequestingUser
            
            // if the "approving" owner is not the same as current asset owner

            if asset_ownerID != approvingOwnerID{
                println("This owner cannot approve of the transfer")
                return nil, fmt.Errorf("This owner cannot approve of the transfer")
            }

            // if the asset is not being requested for transfer
            if asset_status != "requested"{
                println("This asset has not been requested")
                return nil, fmt.Errorf("This asset has not been requested")
            }


            // verifying requesting user --------- FOR NOW: STARTS WITH "user"
            if !strings.HasPrefix(requestingUser, "user"){
                println("This user cant make a transfer request")
                // restoring asset's original conditions
                new_asset := Asset {
                    Name:  Name,
                    Value: asset_value,
                    OwnerID:  asset_ownerID,
                    TransferStatus: "NA",
                    RequestingUser: requestingUser,
                }
                assetJSON, err := json.Marshal(new_asset)
                if err != nil {
                return nil, err
                }
                ctx.GetStub().PutState(Prefix + Name, assetJSON)

                return nil, fmt.Errorf("This user cant make a transfer request")
            }


            // changing asset's transfer status from "requested" to "approved"

            new_asset := Asset {
                Name:  Name,
                Value: asset_value,
                OwnerID:  asset_ownerID,
                TransferStatus: "approved",
                RequestingUser: requestingUser,
            }

            assetJSON, err := json.Marshal(new_asset)
            if err != nil {
            return nil, err
            }

            ctx.GetStub().PutState(Prefix + Name, assetJSON)

            return &new_asset , nil
        }

        return nil, fmt.Errorf("This user cannot approve of asset transfer")
}







// AFTER THESE TWO, READ THROUGH ABAC        (access control)