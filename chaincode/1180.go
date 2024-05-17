package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

	//"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

type Beacon struct {
	Sala string `json:"sala"` 
}

type DispositivoMedico struct {
	Nome   string `json:"nome"` 
	Medico string `json:"medico"` 
	Sala   string `json:"sala"` 
	Doente string `json:"doente"` 
}

type Medico struct{
	Nome string `json:"nome"` 
}

 type Doente struct{
	Sala string `json:"sala"`
 }


// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("fabcar_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "initLedger":
		return s.initLedger(APIstub)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "createBeacon":
		return s.createBeacon(APIstub, args)	
	case "createDispMedico":
		return s.createDispMedico(APIstub, args)	
	case "createMedico":
		return s.createMedico(APIstub, args)	
	case "createDoente":
		return s.createDoente(APIstub, args)
	case "changeBeaconSala":
		return s.changeBeaconSala(APIstub, args)
	case "changeDispMedMedico":
		return s.changeDispMedMedico(APIstub, args)
	case "changeDispMedSala":
		return s.changeDispMedSala(APIstub, args)
	case "changeDispMedDoente":
		return s.changeDispMedDoente(APIstub, args)
	case "changeDoenteSala":
		return s.changeDoenteSala(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

}



func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	beacons := []Beacon{
		Beacon{Sala: "48"},
		Beacon{Sala: "49"},
	}
	
	i := 0
	for i < len(beacons) {
		beaconAsBytes, _ := json.Marshal(beacons[i])
		APIstub.PutState("Beacon_"+strconv.Itoa(i), beaconAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createBeacon(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var beacon = Beacon{Sala: args[1]}

	beaconAsBytes, _ := json.Marshal(beacon)
	APIstub.PutState(args[0], beaconAsBytes)


	return shim.Success(beaconAsBytes)
}

func (s *SmartContract) createDispMedico(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var dispMed = DispositivoMedico{Nome: args[1], Medico: args[2], Sala: args[3], Doente: args[4]}

	dispMedAsBytes, _ := json.Marshal(dispMed)
	APIstub.PutState(args[0], dispMedAsBytes)


	return shim.Success(dispMedAsBytes)
}

func (s *SmartContract) createMedico(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var medico = Medico{Nome: args[1]}

	medicoAsBytes, _ := json.Marshal(medico)
	APIstub.PutState(args[0], medicoAsBytes)


	return shim.Success(medicoAsBytes)
}

func (s *SmartContract) createDoente(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	var doente = Doente{Sala: args[1]}

	doenteAsBytes, _ := json.Marshal(doente)
	APIstub.PutState(args[0], doenteAsBytes)


	return shim.Success(doenteAsBytes)
}

func (s *SmartContract) changeBeaconSala(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	beaconAsBytes, _ := APIstub.GetState(args[0])
	beacon := Beacon{}

	json.Unmarshal(beaconAsBytes, &beacon)
	beacon.Sala = args[1]

	beaconAsBytes, _ = json.Marshal(beacon)
	APIstub.PutState(args[0], beaconAsBytes)

	return shim.Success(beaconAsBytes)
}

func (s *SmartContract) changeDispMedMedico(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	dispMedAsBytes, _ := APIstub.GetState(args[0])
	dispMed := DispositivoMedico{}

	json.Unmarshal(dispMedAsBytes, &dispMed)
	dispMed.Medico = args[1]

	dispMedAsBytes, _ = json.Marshal(dispMed)
	APIstub.PutState(args[0], dispMedAsBytes)

	return shim.Success(dispMedAsBytes)
}

func (s *SmartContract) changeDispMedSala(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	dispMedAsBytes, _ := APIstub.GetState(args[0])
	dispMed := DispositivoMedico{}

	json.Unmarshal(dispMedAsBytes, &dispMed)
	dispMed.Sala = args[1]

	dispMedAsBytes, _ = json.Marshal(dispMed)
	APIstub.PutState(args[0], dispMedAsBytes)

	return shim.Success(dispMedAsBytes)
}

func (s *SmartContract) changeDispMedDoente(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	dispMedAsBytes, _ := APIstub.GetState(args[0])
	dispMed := DispositivoMedico{}

	json.Unmarshal(dispMedAsBytes, &dispMed)
	dispMed.Doente = args[1]

	dispMedAsBytes, _ = json.Marshal(dispMed)
	APIstub.PutState(args[0], dispMedAsBytes)

	return shim.Success(dispMedAsBytes)
}

func (s *SmartContract) changeDoenteSala(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	doenteAsBytes, _ := APIstub.GetState(args[0])
	doente := Doente{}

	json.Unmarshal(doenteAsBytes, &doente)
	doente.Sala = args[1]

	doenteAsBytes, _ = json.Marshal(doente)
	APIstub.PutState(args[0], doenteAsBytes)

	return shim.Success(doenteAsBytes)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	carName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(carName)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing historic values for the marble
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"TxId\":")
		buffer.WriteString("\"")
		buffer.WriteString(response.TxId)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Value\":")
		// if it was a delete operation on given key, then we need to set the
		//corresponding value null. Else, we will write the response.Value
		//as-is (as the Value itself a JSON marble)
		if response.IsDelete {
			buffer.WriteString("null")
		} else {
			buffer.WriteString(string(response.Value))
		}

		buffer.WriteString(", \"Timestamp\":")
		buffer.WriteString("\"")
		buffer.WriteString(time.Unix(response.Timestamp.Seconds, int64(response.Timestamp.Nanos)).String())
		buffer.WriteString("\"")

		buffer.WriteString(", \"IsDelete\":")
		buffer.WriteString("\"")
		buffer.WriteString(strconv.FormatBool(response.IsDelete))
		buffer.WriteString("\"")

		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- getHistoryForAsset returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}





// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
