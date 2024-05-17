package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
	//"reflect"
	"github.com/hyperledger/fabric-chaincode-go/shim"
	sc "github.com/hyperledger/fabric-protos-go/peer"
	"github.com/hyperledger/fabric/common/flogging"

	//"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}


// Assets Structure
type Asset struct {
	ID   string `json:"id"`
	Type  string `json:"type"`
	Attributes map[string]string `json:"attributes"`
	CID  string `json:"cid"`
}

// Policy Structure
type Policy struct {
	UserName   string `json:"username"`
	CID string `json:"cid"`
	SubjectAttributes map[string]string `json:"subjectattributes"`
	ObjectAttributes map[string]string `json:"obbjectattributes"`
	Rules []Rule`json:"rules"`
	//Rules  [] Rule `json:"rules"`
}

//Access rule structure
type Rule struct{
	Type   string `json:"type"`
	Field   string `json:"field"`
	Comparison string `json:"Comparison"`
	Value string `json:"value"`
	Effect string `json:"effect"`
}

//Access Respones structure
type AccessResponse struct{
	Effect   string `json:"effect"`
	Token   string `json:"token"`
}

//Access Record structure
type AccessRecord struct{
	Subject   string `json:"subject"`
	Object   string `json:"object"`
	Time   string `json:"time"`
	Result   string `json:"result"`
}

var logger = flogging.MustGetLogger("subject_cc")


// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	logger.Infof("Chiancode : accesscontrolcontract initiated")
	return shim.Success(nil)
}


// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()

	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
		case "accessControl":
			return s.accessControl(APIstub, args)
		case "accessHistory":
			return s.getHistoryForAsset(APIstub, args)
		case "getAccessRecords":
			return s.queryAccessRecords(APIstub, args)
	}
	
	return shim.Error("Invoke Function Not Success.")

}

func  generateResponse(s string) [] byte{
	accessresponse := AccessResponse{}
	responseAsBytes, _ := json.Marshal(accessresponse)
	if s=="Allow"{
		accessresponse.Token = "xascaassdwea"
		accessresponse.Effect = "Allow"
		responseAsBytes, _ = json.Marshal(accessresponse)
	}else{
		accessresponse.Token = ""
		accessresponse.Effect = "Deny"
		responseAsBytes, _ = json.Marshal(accessresponse)

	}

	return responseAsBytes
}



func (s *SmartContract) accessControl(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// if len(args) != 5 {
	// 	return shim.Error("Incorrect number of arguments. Expecting 5")
	// }
	
	// getting the object
	arguments := make([][]byte, 2)
	arguments[0] = []byte("queryObject")
	arguments[1] = []byte(args[1])

	logger.Infof("Getting the object CID from assetcontract")
	response := APIstub.InvokeChaincode("assetcontract", arguments, "mychannel")

	logger.Infof("Received a response from Assetcontract ")
	logger.Infof(fmt.Sprint(response.Status))
	logger.Infof(fmt.Sprint(shim.OK))
	if response.Status != shim.OK || len(response.Payload)==0{
		return shim.Success(generateResponse("Deny"))
	}

	logger.Infof(fmt.Sprint(response.Payload))
	object := Asset{}
	json.Unmarshal(response.Payload, &object)
	logger.Infof(object.ID)

	// getting the subject
	arguments[0] = []byte("querySubject")
	arguments[1] = []byte(args[0])

	logger.Infof("Getting the subject from assetcontract")
	response = APIstub.InvokeChaincode("assetcontract", arguments, "mychannel")

	logger.Infof("Received a response from Assetcontract ")
	if response.Status != shim.OK || len(response.Payload)==0{
		return shim.Success(generateResponse("Deny"))
	}

	subject := Asset{}
	json.Unmarshal(response.Payload, &subject)
	logger.Infof(subject.ID)


	//getting repsective policy
	logger.Infof("Getting the policy from policycontract")
	arguments[0] = []byte("queryPolicy")
	arguments[1] = []byte(object.CID)

	
	response = APIstub.InvokeChaincode("policycontract", arguments, "mychannel")

	logger.Infof("Received a response from policycontract")
	if response.Status != shim.OK || len(response.Payload)==0{
		return shim.Success(generateResponse("Deny"))
	}

	policy := Policy{}
	json.Unmarshal(response.Payload, &policy)

	logger.Infof(policy.CID)


	//Compare Subject Attributes
	for k, v := range policy.SubjectAttributes { 
		logger.Infof("Policy key[%s] value[%s]\n", k, v)
		if val, ok := subject.Attributes[k]; ok {
			logger.Infof("Subject key[%s] exist val[%s]", k, val)
			if v != val{
				return shim.Success(generateResponse("Deny"))
			}
		} else{
			return shim.Success(generateResponse("Deny"))
		}
	}

	//Compare Object Attributes
	for k, v := range policy.ObjectAttributes { 
		logger.Infof("Policy key[%s] value[%s]\n", k, v)
		if val, ok := object.Attributes[k]; ok {
			logger.Infof("Object key[%s] exist val[%s]", k, val)
			if v != val{
				return shim.Success(generateResponse("Deny"))
			}
		} else{
			return shim.Success(generateResponse("Deny"))
		}
	}


	//Evaluate the rule
	var flag = 0
	var val1 = ""
	var val2 = ""
	for i := 0; i < len(policy.Rules); i++ {

			if policy.Rules[i].Type == "subject"{
				val1 = subject.Attributes[policy.Rules[i].Field]
			} else{
				val1 = object.Attributes[policy.Rules[i].Field]
			}
			
			val2 = policy.Rules[i].Value
			logger.Infof("val1 is %s and val2 is %s", val1, val2)
			if policy.Rules[i].Comparison == "equals"{
				logger.Infof("In equals")
				if !isEqual(val1, val2){
					flag = 1
					break
				}
			}else if policy.Rules[i].Comparison == "greaterthan"{
				if !isGreater(val1, val2){
					flag = 1
					break
				}
			}else if policy.Rules[i].Comparison == "lessthan"{
				if !isGreater(val1, val2){
					flag = 1
					break
				}
			}else{
				flag = 1
			
			}

	}
		

	timestamp := time.Now()
	accessRecord := AccessRecord{Subject: subject.ID, Object: object.ID, Time: timestamp.Format("20060102150405"), Result:""}
	id := "access-"+object.ID
	

	logger.Infof("Getting the permisions to access the object %s", id)
	if flag==0{
		accessRecord.Result = "Allow"
		accessRecordAsBytes, _ := json.Marshal(accessRecord)
		APIstub.PutState(id, accessRecordAsBytes)
		return shim.Success(accessRecordAsBytes)
	} else{
		accessRecord.Result = "Deny"
		accessRecordAsBytes, _ := json.Marshal(accessRecord)
		APIstub.PutState(id, accessRecordAsBytes)
		return shim.Success(generateResponse("Deny"))
	} 
	
}

func isEqual(val1 string, val2 string) bool{
	return val1 == val2
}

func isGreater(val1 string, val2 string) bool{
	return val1 > val2
}

func isLesser(val1 string, val2 string) bool{
	return val1 < val2
}



func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	assetID := args[0]
	logger.Infof("searching for object id %s", args[0])
	resultsIterator, err := stub.GetHistoryForKey(assetID)
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

		logger.Infof(string(response.TxId))
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

	logger.Infof("- getHistoryForAsset returning:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}


// query Access
func (s *SmartContract) queryAccessRecords(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}
	logger.Infof("Quering the ID %s", args[0])
	asBytes, _ := APIstub.GetState(args[0])

	access := AccessRecord{}
	json.Unmarshal(asBytes, &access)
	logger.Infof("Sending the object with the this %s", access.Subject)
	return shim.Success(asBytes)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}


