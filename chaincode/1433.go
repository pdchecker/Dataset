package main

import (
	"encoding/json"
	"strings"

	// "errors"
	"fmt"
	//  "time"
	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	pb "github.com/hyperledger/fabric-protos-go/peer"
)

//<<Standard Code Section - Start>>

/* ===================================================================================
	This is the system generated code section. Please don't make any change in this section
  ===================================================================================*/

//Struct For Chain Code
type PRChainCode struct {
}

//Standard Functions
func main() {
	//TestAll(new(PRChainCode))

	fmt.Println("SupplyChain ChainCode Started")
	err := shim.Start(new(PRChainCode))
	if err != nil {
		fmt.Println("Error starting SupplyChain chaincode: %s", err)
	}

}

//Init is called during chaincode instantiation to initialize any data.
func (t *PRChainCode) Init(stub shim.ChaincodeStubInterface) pb.Response {
	fmt.Println("SupplyChain ChainCode Initiated")

	_, args := stub.GetFunctionAndParameters()
	fmt.Println("Init: %v", args)
	if len(args[0]) <= 0 {
		return shim.Error("MSP Mapping information is required for initiating the chain code")
	}

	var MSPListUnmarshaled []MSPList
	err := json.Unmarshal([]byte(args[0]), &MSPListUnmarshaled)

	if err != nil {
		return shim.Error("An error occurred while Unmarshiling MSPMapping: " + err.Error())
	}
	MSPMappingJSONasBytes, err := json.Marshal(MSPListUnmarshaled)
	if err != nil {
		return shim.Error("An error occurred while Marshiling MSPMapping :" + err.Error())
	}

	_Key := "MSPMapping"
	err = stub.PutState(_Key, []byte(MSPMappingJSONasBytes))
	if err != nil {
		return shim.Error("An error occurred while inserting MSPMapping:" + err.Error())
	}
	return shim.Success(nil)
}

//Invoke is called per transaction on the chaincode
func (t *PRChainCode) Invoke(stub shim.ChaincodeStubInterface) pb.Response {

	//getting MSP
	certOrgType, err := cid.GetMSPID(stub)
	if err != nil {
		return shim.Error("Enrolment mspid Type invalid!!! " + err.Error())
	}
	fmt.Println("MSP:" + certOrgType)

	orgType, err := getOrgTypeByMSP(stub, string(certOrgType))
	if err != nil {
		return shim.Error(err.Error())
	}

	function, args := stub.GetFunctionAndParameters()
	fmt.Println("Invoke is running for function: " + function)
	//<<Function Validation Logic-Start>>

	// args, errTrans := getArguments(stub)
	// if errTrans != nil {
	// 	return shim.Error(errTrans.Error())
	// }
	fmt.Println("Arguments Loaded Successfully!!", args)

	connection := hypConnect{}
	connection.Connection = stub
	if orgType == "Airline" || orgType == "Airport" || orgType == "Interliner" {

		switch functionName := function; functionName {

		case "RegisterAirports":
			return t.RegisterAirports(connection, args)
		case "RegisterAirlines":
			return t.RegisterAirlines(connection, args)
		case "CreateBaggage":
			return t.CreateBaggage(connection, args)
		case "ChangeBaggageStatusByAirlines":
			return t.ChangeBaggageStatusByAirlines(connection, args)
		case "ChangeBaggageStatusByAirport":
			return t.ChangeBaggageStatusByAirport(connection, args)	
		case "GetBaggageDetails":
			return t.GetBaggageDetails(connection, args)

		default:
			return defaultConditionInvoke(function)
		}
	} else {
		return shim.Error("Invalid MSP: " + orgType)
	}
}
func defaultConditionInvoke(function string) pb.Response {
	//logger.Warning("Invoke did not find function: " + function)
	return shim.Error("Received unknown function invocation: " + function)
}

func (t *PRChainCode) RegisterAirports(stub hypConnect, args []string) pb.Response {

	fmt.Printf("RegisterAirports: %v", args)

	if len(args) != 2 {
		return shim.Error("Invalid Argument")
	}

	AirportId := sanitize(args[0], "string").(string)
	Location := sanitize(args[1], "string").(string)

	exists, err := t.AirportsExists(stub, []string{AirportId})

	var airports Airports

	if !exists {
		fmt.Printf("the Airports for %s does not exist; creating for first time", AirportId)

		airports = Airports{
			AirportId: AirportId,
			Location:  Location,
		}
	} else {
		return shim.Success([]byte(`Airports Already Exist`))
	}

	assetJSON, err := json.Marshal(airports)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = insertData(&stub, strings.ToLower(AirportId), "Airport", []byte(assetJSON))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// AirportsExists returns true when GovBit with given user exists in world state
func (t *PRChainCode) AirportsExists(stub hypConnect, args []string) (bool, error) {
	fmt.Printf("AirportsExists: %v", args)

	key := sanitize(args[0], "string").(string) //Fenergo

	trnxAsBytes, err := fetchData(stub, strings.ToLower(key), "Airport")
	if err != nil {
		return false, err
	}

	return trnxAsBytes != nil, nil
}

func (t *PRChainCode) RegisterAirlines(stub hypConnect, args []string) pb.Response {

	fmt.Printf("RegisterAirlines: %v", args)

	if len(args) != 1 {
		return shim.Error("Invalid Argument")
	}

	airlineId := sanitize(args[0], "string").(string)

	exists, err := t.AirlinesExists(stub, []string{airlineId})

	var airlines Airlines

	if !exists {
		fmt.Printf("the Airlines for %s does not exist; creating for first time", airlineId)

		airlines = Airlines{
			AirlineId: airlineId,
		}
	} else {
		return shim.Success([]byte(`Airlines Already Exist`))
	}

	assetJSON, err := json.Marshal(airlines)
	if err != nil {
		return shim.Error(err.Error())
	}

	err = insertData(&stub, strings.ToLower(airlineId), "Airport", []byte(assetJSON))
	if err != nil {
		return shim.Error(err.Error())
	}
	return shim.Success(nil)
}

// AirlinesExists returns true when GovBit with given user exists in world state
func (t *PRChainCode) AirlinesExists(stub hypConnect, args []string) (bool, error) {
	fmt.Printf("AirlinesExists: %v", args)

	key := sanitize(args[0], "string").(string) //Fenergo

	trnxAsBytes, err := fetchData(stub, strings.ToLower(key), "Airport")
	if err != nil {
		return false, err
	}

	return trnxAsBytes != nil, nil
}

func (t *PRChainCode) CreateBaggage(stub hypConnect, args []string) pb.Response {

	certOrgType, err := cid.GetMSPID(stub.Connection)
	if err != nil {
		return shim.Error("Enrolment mspid Type invalid!!! " + err.Error())
	}
	role, _, Roleerr := cid.GetAttributeValue(stub.Connection, "Role")
	if Roleerr != nil {
		return shim.Error("GetAttributeValue Role Type invalid!!! " + Roleerr.Error())
	}

	if certOrgType == `InterlinerMSP` && role == `Interlining Agent` {

		fmt.Printf("CreateBaggage: %v", args)

		if len(args) != 4 {
			return shim.Error("Invalid Argument")
		}

		// airlineId         := sanitize(args[0], "string").(string)
		// airlinesFee         := sanitize(args[2], "float").(float64)
		userId, _, useriderr := cid.GetAttributeValue(stub.Connection, "UserId")
		if Roleerr != nil {
			return shim.Error("User Id Not Found!!! " + useriderr.Error())
		}
		baggageId := sanitize(args[0], "string").(string)
		source := sanitize(args[1], "string").(string)
		destination := sanitize(args[2], "string").(string)
		path := sanitize(args[3], "string").(string)

		var route []Route
		var airportFees []float64
		var airlineFees []float64

		patherr := json.Unmarshal([]byte(path), &route)
		if patherr != nil {
			return shim.Error(patherr.Error())
		}

		var baggageRoute BaggageRoute
		exists, err := t.BaggageExists(stub, []string{baggageId})

		if !exists {
			fmt.Printf("the Airlines for %s does not exist; creating for first time", baggageId)

			baggageRoute = BaggageRoute{
				BaggageId:   baggageId,
				UserId:      userId,
				Source:      source,
				Destination: destination,
				TotalExpence:0,
				AirportFees: airportFees,
				AirlineFees: airlineFees,
				Path:        route,
			}
		} else {
			return shim.Success([]byte(`BaggageExists Already Exist`))
		}

		assetJSON, err := json.Marshal(baggageRoute)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = insertData(&stub, strings.ToLower(baggageId), "Baggage", []byte(assetJSON))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	} else {
		return shim.Error("Do not access to this feature")

	}
}

// BaggageExists returns true when GovBit with given user exists in world state
func (t *PRChainCode) BaggageExists(stub hypConnect, args []string) (bool, error) {
	fmt.Printf("BaggageExists: %v", args)

	key := sanitize(args[0], "string").(string) //Fenergo

	trnxAsBytes, err := fetchData(stub, strings.ToLower(key), "Baggage")
	if err != nil {
		return false, err
	}

	return trnxAsBytes != nil, nil
}


func (t *PRChainCode) ChangeBaggageStatusByAirlines(stub hypConnect, args []string) pb.Response {

	certOrgType, err := cid.GetMSPID(stub.Connection)
	if err != nil {
		return shim.Error("Enrolment mspid Type invalid!!! " + err.Error())
	}
	role, _, Roleerr := cid.GetAttributeValue(stub.Connection, "Role")
	if Roleerr != nil {
		return shim.Error("GetAttributeValue Role Type invalid!!! " + Roleerr.Error())
	}

	if certOrgType == `AirlineMSP` && role == `Airline` {

		fmt.Printf("ChangeBaggageStatusByAirlines: %v", args)

		if len(args) != 3 {
			return shim.Error("Invalid Argument")
		}



		baggageId := sanitize(args[0], "string").(string)
		status := sanitize(args[1], "bool").(bool)
		fees := sanitize(args[2], "float").(float64)

		airlineId, _, airlineIderr := cid.GetAttributeValue(stub.Connection, "UserId")
		if airlineIderr != nil {
			return shim.Error("User Id Not Found!!! " + airlineIderr.Error())
		}
		var baggageRoute BaggageRoute

		trnxAsBytes, BaggageErr := fetchData(stub, strings.ToLower(baggageId), "Baggage")
		if BaggageErr != nil {
			return shim.Error("Baggage not exit")
		}
		BaggageErr = json.Unmarshal(trnxAsBytes, &baggageRoute)
		if BaggageErr != nil {
			return shim.Error(BaggageErr.Error())
		}
		isExist :=false
		for i:=0 ; i< len(baggageRoute.Path);i++{
			fmt.Println(strings.ToLower(baggageRoute.Path[i].AirlineId), strings.ToLower(airlineId))
			if strings.Compare(strings.ToLower(baggageRoute.Path[i].AirlineId),strings.ToLower(airlineId))==0{
				if i==0{
					baggageRoute.Path[i].AirlineStatus=status
					baggageRoute.AirlineFees=append(baggageRoute.AirlineFees,fees)
					baggageRoute.TotalExpence+=fees
					isExist=true
					break;
				}else{
					if baggageRoute.Path[i-1].AirlineStatus{
						baggageRoute.Path[i].AirlineStatus=status
						baggageRoute.AirlineFees=append(baggageRoute.AirlineFees,fees)
						baggageRoute.TotalExpence+=fees
						break
						isExist=true
					}else{
						return shim.Error("Last Airlines not confirm delivery")
					}

				}
			}
		}
		if !isExist{
			return shim.Error("Invalid Airlines Id not exit in Path")
		}

		assetJSON, err := json.Marshal(baggageRoute)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = insertData(&stub, strings.ToLower(baggageId), "Baggage", []byte(assetJSON))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	} else {
		return shim.Error("Do not access to this feature")

	}
}



func (t *PRChainCode) ChangeBaggageStatusByAirport(stub hypConnect, args []string) pb.Response {

	certOrgType, err := cid.GetMSPID(stub.Connection)
	if err != nil {
		return shim.Error("Enrolment mspid Type invalid!!! " + err.Error())
	}
	role, _, Roleerr := cid.GetAttributeValue(stub.Connection, "Role")
	if Roleerr != nil {
		return shim.Error("GetAttributeValue Role Type invalid!!! " + Roleerr.Error())
	}

	if certOrgType == `AirportMSP` && role == `Airport` {

		fmt.Printf("ChangeBaggageStatusByAirport: %v", args)

		if len(args) != 3 {
			return shim.Error("Invalid Argument")
		}



		baggageId := sanitize(args[0], "string").(string)
		status := sanitize(args[1], "bool").(bool)
		fees := sanitize(args[2], "float").(float64)


		airportId, _, airportIderr := cid.GetAttributeValue(stub.Connection, "UserId")
		if airportIderr != nil {
			return shim.Error("User Id Not Found!!! " + airportIderr.Error())
		}
		var baggageRoute BaggageRoute

		trnxAsBytes, BaggageErr := fetchData(stub, strings.ToLower(baggageId), "Baggage")
		if BaggageErr != nil {
			return shim.Error("Baggage not exit")
		}
		BaggageErr = json.Unmarshal(trnxAsBytes, &baggageRoute)
		if BaggageErr != nil {
			return shim.Error(BaggageErr.Error())
		}
		isExist :=false
		for i:=0 ; i< len(baggageRoute.Path);i++{
			fmt.Println()
			fmt.Println(strings.ToLower(baggageRoute.Path[i].AirportId), strings.ToLower(airportId),strings.Compare(strings.ToLower(baggageRoute.Path[i].AirportId),strings.ToLower(airportId)))

			if strings.Compare(strings.ToLower(baggageRoute.Path[i].AirportId),strings.ToLower(airportId))==0{
				if i==0{
					baggageRoute.Path[i].AirportStatus=status
					baggageRoute.AirportFees=append(baggageRoute.AirportFees,fees)
					baggageRoute.TotalExpence+=fees
					isExist=true
					break;
				}else{
					if baggageRoute.Path[i-1].AirportStatus{
						baggageRoute.Path[i].AirportStatus=status
						baggageRoute.AirportFees=append(baggageRoute.AirportFees,fees)
						baggageRoute.TotalExpence+=fees
						break
						isExist=true
					}else{
						return shim.Error("Last Airport not confirm delivery")
					}

				}
			}
		}
		if !isExist{
			return shim.Error("Invalid Airport Id not exit in Path")
		}

		assetJSON, err := json.Marshal(baggageRoute)
		if err != nil {
			return shim.Error(err.Error())
		}

		err = insertData(&stub, strings.ToLower(baggageId), "Baggage", []byte(assetJSON))
		if err != nil {
			return shim.Error(err.Error())
		}
		return shim.Success(nil)
	} else {
		return shim.Error("Do not access to this feature")

	}
}



func (t *PRChainCode) GetBaggageDetails(stub hypConnect, args []string) (pb.Response) {
	fmt.Printf("GetBaggageDetails: %v", args)

	if len(args) !=1  {
		return shim.Error("Invalid Argument")
	}
	key := sanitize(args[0], "string").(string) //Fenergo

	trnxAsBytes, err := fetchData(stub, strings.ToLower(key), "Baggage")
	if err != nil {
		return shim.Success(nil)
	}

	return shim.Success(trnxAsBytes)
}

