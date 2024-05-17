package main


import (
	"encoding/json"
	"fmt"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}


type PrivateDetails struct {
	SecretKey string `json:"secretKey"`
	Nonce string `json:"nonce"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("private_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "createPrivateKeyStorage":
		return s.createPrivateKeyStorage(APIstub, args)
	case "readPrivateKey":
		return s.readPrivateKey(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}
}


func (s *SmartContract) readPrivateKey(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	keyAsBytes, err := APIstub.GetPrivateData(args[0], args[1])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if keyAsBytes == nil {
		jsonResp := "{\"Error\":\"" + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(keyAsBytes)
}



func (s *SmartContract) createPrivateKeyStorage(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type TransientInput struct {
		SecretKey string `json:"secretKey"`
		Nonce string `json:"nonce"`
		Key   string `json:"key"`
	}

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error(err.Error())
	}
	privateDataAsBytes, errr := transMap["keys"]
	
	if !errr {
		return shim.Error("")
	}
	
	var Input TransientInput
	err = json.Unmarshal(privateDataAsBytes, &Input)

	
	if len(Input.SecretKey) == 0 {
		return shim.Error("Nonce field must be a non-empty string")
	}
	if len(Input.Nonce) == 0 {
		return shim.Error("Key field must be a non-empty string")
	}

	PrivateDetails := &PrivateDetails{SecretKey: Input.SecretKey, Nonce: Input.Nonce}
	PrivateDetailsAsBytes, err := json.Marshal(PrivateDetails)
	err = APIstub.PutPrivateData("collectionPrivateDetails", Input.Key, PrivateDetailsAsBytes)
	return shim.Success(PrivateDetailsAsBytes)
}

func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}