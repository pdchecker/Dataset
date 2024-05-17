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

	"github.com/hyperledger/fabric-chaincode-go/pkg/cid"
)

// SmartContract Define the Smart Contract structure
type SmartContract struct {
}

// Node :  Define the Node structure, with 4 properties.  Structure tags are used by encoding/json library
type Node struct {
	ID   string `json:"id"`
	Platform  string `json:"platform"`
	Architecture string `json:"architecture"`
	Info  string `json:"info"`
}

type NodePrivateDetails struct {
	Info string `json:"info"`
	Price string `json:"price"`
}

// Init ;  Method for initializing smart contract
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

var logger = flogging.MustGetLogger("node_cc")

// Invoke :  Method for INVOKING smart contract
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	function, args := APIstub.GetFunctionAndParameters()
	logger.Infof("Function name is:  %d", function)
	logger.Infof("Args length is : %d", len(args))

	switch function {
	case "queryNode":
		return s.queryNode(APIstub, args)
	case "initLedger":
		return s.initLedger(APIstub)
	case "createNode":
		return s.createNode(APIstub, args)
	case "queryAllNodes":
		return s.queryAllNodes(APIstub)
	case "changeNodeInfo":
		return s.changeNodeInfo(APIstub, args)
	case "getHistoryForAsset":
		return s.getHistoryForAsset(APIstub, args)
	case "queryNodesByInfo":
		return s.queryNodesByInfo(APIstub, args)
	case "restictedMethod":
		return s.restictedMethod(APIstub, args)
	case "test":
		return s.test(APIstub, args)
	case "createPrivateNode":
		return s.createPrivateNode(APIstub, args)
	case "readPrivateNode":
		return s.readPrivateNode(APIstub, args)
	case "updatePrivateData":
		return s.updatePrivateData(APIstub, args)
	case "readNodePrivateDetails":
		return s.readNodePrivateDetails(APIstub, args)
	case "createPrivateNodeImplicitForOrg1":
		return s.createPrivateNodeImplicitForOrg1(APIstub, args)
	case "createPrivateNodeImplicitForOrg2":
		return s.createPrivateNodeImplicitForOrg2(APIstub, args)
	case "queryPrivateDataHash":
		return s.queryPrivateDataHash(APIstub, args)
	default:
		return shim.Error("Invalid Smart Contract function name.")
	}

	// return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) queryNode(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	NodeAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) readPrivateNode(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	// collectionNodes, collectionNodePrivateDetails, _implicit_org_Org1MSP, _implicit_org_Org2MSP
	NodeAsBytes, err := APIstub.GetPrivateData(args[0], args[1])
	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[1] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if NodeAsBytes == nil {
		jsonResp := "{\"Error\":\"Node private details does not exist: " + args[1] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) readPrivateNodeIMpleciteForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	NodeAsBytes, _ := APIstub.GetPrivateData("_implicit_org_Org1MSP", args[0])
	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) readNodePrivateDetails(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	NodeAsBytes, err := APIstub.GetPrivateData("collectionNodePrivateDetails", args[0])

	if err != nil {
		jsonResp := "{\"Error\":\"Failed to get private details for " + args[0] + ": " + err.Error() + "\"}"
		return shim.Error(jsonResp)
	} else if NodeAsBytes == nil {
		jsonResp := "{\"Error\":\"Marble private details does not exist: " + args[0] + "\"}"
		return shim.Error(jsonResp)
	}
	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) test(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	NodeAsBytes, _ := APIstub.GetState(args[0])
	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	Nodes := []Node{
		Node{ID: "Toyota", Platform: "Prius", Architecture: "blue", Info: "Tomoko"},
		Node{ID: "Ford", Platform: "Mustang", Architecture: "red", Info: "Brad"},
		Node{ID: "Hyundai", Platform: "Tucson", Architecture: "green", Info: "Jin Soo"},
		Node{ID: "Volkswagen", Platform: "Passat", Architecture: "yellow", Info: "Max"},
		Node{ID: "Tesla", Platform: "S", Architecture: "black", Info: "Adriana"},
		Node{ID: "Peugeot", Platform: "205", Architecture: "purple", Info: "Michel"},
		Node{ID: "Chery", Platform: "S22L", Architecture: "white", Info: "Aarav"},
		Node{ID: "Fiat", Platform: "Punto", Architecture: "violet", Info: "Pari"},
		Node{ID: "Tata", Platform: "Nano", Architecture: "indigo", Info: "Valeria"},
		Node{ID: "Holden", Platform: "Barina", Architecture: "brown", Info: "Shotaro"},
	}

	i := 0
	for i < len(Nodes) {
		NodeAsBytes, _ := json.Marshal(Nodes[i])
		APIstub.PutState("Node"+strconv.Itoa(i), NodeAsBytes)
		i = i + 1
	}

	return shim.Success(nil)
}

func (s *SmartContract) createPrivateNode(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	type NodeTransientInput struct {
		ID  string `json:"id"` //the fieldtags are needed to keep case from bouncing around
		Platform string `json:"platform"`
		Color string `json:"color"`
		Info string `json:"info"`
		Price string `json:"price"`
		Key   string `json:"key"`
	}
	if len(args) != 0 {
		return shim.Error("1111111----Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	NodeDataAsBytes, ok := transMap["Node"]
	if !ok {
		return shim.Error("Node must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(NodeDataAsBytes))

	if len(NodeDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var NodeInput NodeTransientInput
	err = json.Unmarshal(NodeDataAsBytes, &NodeInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(NodeDataAsBytes) + "Error is : " + err.Error())
	}

	logger.Infof("3333")

	if len(NodeInput.Key) == 0 {
		return shim.Error("name field must be a non-empty string")
	}
	if len(NodeInput.ID) == 0 {
		return shim.Error("color field must be a non-empty string")
	}
	if len(NodeInput.Platform) == 0 {
		return shim.Error("Platform field must be a non-empty string")
	}
	if len(NodeInput.Color) == 0 {
		return shim.Error("color field must be a non-empty string")
	}
	if len(NodeInput.Info) == 0 {
		return shim.Error("Info field must be a non-empty string")
	}
	if len(NodeInput.Price) == 0 {
		return shim.Error("price field must be a non-empty string")
	}

	logger.Infof("444444")

	// ==== Check if Node already exists ====
	NodeAsBytes, err := APIstub.GetPrivateData("collectionNodes", NodeInput.Key)
	if err != nil {
		return shim.Error("Failed to get marble: " + err.Error())
	} else if NodeAsBytes != nil {
		fmt.Println("This Node already exists: " + NodeInput.Key)
		return shim.Error("This Node already exists: " + NodeInput.Key)
	}

	logger.Infof("55555")

	var Node = Node{ID: NodeInput.ID, Platform: NodeInput.Platform, Architecture: NodeInput.Color, Info: NodeInput.Info}

	NodeAsBytes, err = json.Marshal(Node)
	if err != nil {
		return shim.Error(err.Error())
	}
	err = APIstub.PutPrivateData("collectionNodes", NodeInput.Key, NodeAsBytes)
	if err != nil {
		logger.Infof("6666666")
		return shim.Error(err.Error())
	}

	NodePrivateDetails := &NodePrivateDetails{Info: NodeInput.Info, Price: NodeInput.Price}

	NodePrivateDetailsAsBytes, err := json.Marshal(NodePrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionNodePrivateDetails", NodeInput.Key, NodePrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) updatePrivateData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	type NodeTransientInput struct {
		Info string `json:"info"`
		Price string `json:"price"`
		Key   string `json:"key"`
	}
	if len(args) != 0 {
		return shim.Error("1111111----Incorrect number of arguments. Private marble data must be passed in transient map.")
	}

	logger.Infof("11111111111111111111111111")

	transMap, err := APIstub.GetTransient()
	if err != nil {
		return shim.Error("222222 -Error getting transient: " + err.Error())
	}

	NodeDataAsBytes, ok := transMap["Node"]
	if !ok {
		return shim.Error("Node must be a key in the transient map")
	}
	logger.Infof("********************8   " + string(NodeDataAsBytes))

	if len(NodeDataAsBytes) == 0 {
		return shim.Error("333333 -marble value in the transient map must be a non-empty JSON string")
	}

	logger.Infof("2222222")

	var NodeInput NodeTransientInput
	err = json.Unmarshal(NodeDataAsBytes, &NodeInput)
	if err != nil {
		return shim.Error("44444 -Failed to decode JSON of: " + string(NodeDataAsBytes) + "Error is : " + err.Error())
	}

	NodePrivateDetails := &NodePrivateDetails{Info: NodeInput.Info, Price: NodeInput.Price}

	NodePrivateDetailsAsBytes, err := json.Marshal(NodePrivateDetails)
	if err != nil {
		logger.Infof("77777")
		return shim.Error(err.Error())
	}

	err = APIstub.PutPrivateData("collectionNodePrivateDetails", NodeInput.Key, NodePrivateDetailsAsBytes)
	if err != nil {
		logger.Infof("888888")
		return shim.Error(err.Error())
	}

	return shim.Success(NodePrivateDetailsAsBytes)

}

func (s *SmartContract) createNode(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect number of arguments. Expecting 5")
	}

	var Node = Node{ID: args[1], Platform: args[2], Architecture: args[3], Info: args[4]}

	NodeAsBytes, _ := json.Marshal(Node)
	APIstub.PutState(args[0], NodeAsBytes)

	indexName := "Info~key"
	colorNameIndexKey, err := APIstub.CreateCompositeKey(indexName, []string{Node.Info, args[0]})
	if err != nil {
		return shim.Error(err.Error())
	}
	value := []byte{0x00}
	APIstub.PutState(colorNameIndexKey, value)

	return shim.Success(NodeAsBytes)
}

func (S *SmartContract) queryNodesByInfo(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments")
	}
	Info := args[0]

	InfoAndIdResultIterator, err := APIstub.GetStateByPartialCompositeKey("Info~key", []string{Info})
	if err != nil {
		return shim.Error(err.Error())
	}

	defer InfoAndIdResultIterator.Close()

	var i int
	var id string

	var Nodes []byte
	bArrayMemberAlreadyWritten := false

	Nodes = append([]byte("["))

	for i = 0; InfoAndIdResultIterator.HasNext(); i++ {
		responseRange, err := InfoAndIdResultIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}

		objectType, compositeKeyParts, err := APIstub.SplitCompositeKey(responseRange.Key)
		if err != nil {
			return shim.Error(err.Error())
		}

		id = compositeKeyParts[1]
		assetAsBytes, err := APIstub.GetState(id)

		if bArrayMemberAlreadyWritten == true {
			newBytes := append([]byte(","), assetAsBytes...)
			Nodes = append(Nodes, newBytes...)

		} else {
			// newBytes := append([]byte(","), NodesAsBytes...)
			Nodes = append(Nodes, assetAsBytes...)
		}

		fmt.Printf("Found a asset for index : %s asset id : ", objectType, compositeKeyParts[0], compositeKeyParts[1])
		bArrayMemberAlreadyWritten = true

	}

	Nodes = append(Nodes, []byte("]")...)

	return shim.Success(Nodes)
}

func (s *SmartContract) queryAllNodes(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "Node0"
	endKey := "Node999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllNodes:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) restictedMethod(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	// get an ID for the client which is guaranteed to be unique within the MSP
	//id, err := cid.GetID(APIstub) -

	// get the MSP ID of the client's identity
	//mspid, err := cid.GetMSPID(APIstub) -

	// get the value of the attribute
	//val, ok, err := cid.GetAttributeValue(APIstub, "attr1") -

	// get the X509 certificate of the client, or nil if the client's identity was not based on an X509 certificate
	//cert, err := cid.GetX509Certificate(APIstub) -

	val, ok, err := cid.GetAttributeValue(APIstub, "role")
	if err != nil {
		// There was an error trying to retrieve the attribute
		shim.Error("Error while retriving attributes")
	}
	if !ok {
		// The client identity does not possess the attribute
		shim.Error("Client identity doesnot posses the attribute")
	}
	// Do something with the value of 'val'
	if val != "approver" {
		fmt.Println("Attribute role: " + val)
		return shim.Error("Only user with role as APPROVER have access this method!")
	} else {
		if len(args) != 1 {
			return shim.Error("Incorrect number of arguments. Expecting 1")
		}

		NodeAsBytes, _ := APIstub.GetState(args[0])
		return shim.Success(NodeAsBytes)
	}

}

func (s *SmartContract) changeNodeInfo(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}

	NodeAsBytes, _ := APIstub.GetState(args[0])
	Node := Node{}

	json.Unmarshal(NodeAsBytes, &Node)
	Node.Info = args[1]

	NodeAsBytes, _ = json.Marshal(Node)
	APIstub.PutState(args[0], NodeAsBytes)

	return shim.Success(NodeAsBytes)
}

func (t *SmartContract) getHistoryForAsset(stub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) < 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1")
	}

	NodeName := args[0]

	resultsIterator, err := stub.GetHistoryForKey(NodeName)
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

func (s *SmartContract) createPrivateNodeImplicitForOrg1(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var Node = Node{ID: args[1], Platform: args[2], Architecture: args[3], Info: args[4]}

	NodeAsBytes, _ := json.Marshal(Node)
	// APIstub.PutState(args[0], NodeAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org1MSP", args[0], NodeAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) createPrivateNodeImplicitForOrg2(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 5 {
		return shim.Error("Incorrect arguments. Expecting 5 arguments")
	}

	var Node = Node{ID: args[1], Platform: args[2], Architecture: args[3], Info: args[4]}

	NodeAsBytes, _ := json.Marshal(Node)
	APIstub.PutState(args[0], NodeAsBytes)

	err := APIstub.PutPrivateData("_implicit_org_Org2MSP", args[0], NodeAsBytes)
	if err != nil {
		return shim.Error("Failed to add asset: " + args[0])
	}
	return shim.Success(NodeAsBytes)
}

func (s *SmartContract) queryPrivateDataHash(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	NodeAsBytes, _ := APIstub.GetPrivateDataHash(args[0], args[1])
	return shim.Success(NodeAsBytes)
}

// func (s *SmartContract) CreateNodeAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	if len(args) != 1 {
// 		return shim.Error("Incorrect number of arguments. Expecting 1")
// 	}

// 	var Node Node
// 	err := json.Unmarshal([]byte(args[0]), &Node)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	NodeAsBytes, err := json.Marshal(Node)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	err = APIstub.PutState(Node.ID, NodeAsBytes)
// 	if err != nil {
// 		return shim.Error(err.Error())
// 	}

// 	return shim.Success(nil)
// }

// func (s *SmartContract) addBulkAsset(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
// 	logger.Infof("Function addBulkAsset called and length of arguments is:  %d", len(args))
// 	if len(args) >= 500 {
// 		logger.Errorf("Incorrect number of arguments in function CreateAsset, expecting less than 500, but got: %b", len(args))
// 		return shim.Error("Incorrect number of arguments, expecting 2")
// 	}

// 	var eventKeyValue []string

// 	for i, s := range args {

// 		key :=s[0];
// 		var Node = Node{ID: s[1], Platform: s[2], Architecture: s[3], Info: s[4]}

// 		eventKeyValue = strings.SplitN(s, "#", 3)
// 		if len(eventKeyValue) != 3 {
// 			logger.Errorf("Error occured, Please ID sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
// 			return shim.Error("Error occured, Please ID sure that you have provided the array of strings and each string should be  in \"EventType#Key#Value\" format")
// 		}

// 		assetAsBytes := []byte(eventKeyValue[2])
// 		err := APIstub.PutState(eventKeyValue[1], assetAsBytes)
// 		if err != nil {
// 			logger.Errorf("Error coocured while putting state for asset %s in APIStub, error: %s", eventKeyValue[1], err.Error())
// 			return shim.Error(err.Error())
// 		}
// 		// logger.infof("Adding value for ")
// 		fmt.Println(i, s)

// 		indexName := "Event~Id"
// 		eventAndIDIndexKey, err2 := APIstub.CreateCompositeKey(indexName, []string{eventKeyValue[0], eventKeyValue[1]})

// 		if err2 != nil {
// 			logger.Errorf("Error coocured while putting state in APIStub, error: %s", err.Error())
// 			return shim.Error(err2.Error())
// 		}

// 		value := []byte{0x00}
// 		err = APIstub.PutState(eventAndIDIndexKey, value)
// 		if err != nil {
// 			logger.Errorf("Error coocured while putting state in APIStub, error: %s", err.Error())
// 			return shim.Error(err.Error())
// 		}
// 		// logger.Infof("Created Composite key : %s", eventAndIDIndexKey)

// 	}

// 	return shim.Success(nil)
// }

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
